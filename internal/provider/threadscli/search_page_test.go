package threadscli

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tamnd/threads-cli/threads"
)

func TestParseSearchPageReturnsOnlySearchResults(t *testing.T) {
	page, err := os.ReadFile("testdata/search_page.html")
	if err != nil {
		t.Fatal(err)
	}
	got, err := parseSearchPage(page, "launch", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d results: %+v", len(got), got)
	}
	if got[0].ID != "101" || got[0].Text != "first result" || got[0].Username != "alice" || got[0].Permalink != "https://www.threads.com/@alice/post/Alpha1" {
		t.Fatalf("first result = %+v", got[0])
	}
	if got[0].Query != "launch" || got[0].Timestamp.IsZero() || got[0].MediaType != "IMAGE" {
		t.Fatalf("first result metadata = %+v", got[0])
	}
	if !got[1].IsReply || !got[1].IsQuotePost || got[1].MediaType != "VIDEO" {
		t.Fatalf("second result relationships = %+v", got[1])
	}
}

func TestParseSearchPageHonorsLimit(t *testing.T) {
	page, err := os.ReadFile("testdata/search_page.html")
	if err != nil {
		t.Fatal(err)
	}
	got, err := parseSearchPage(page, "launch", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "101" {
		t.Fatalf("results = %+v", got)
	}
}

func TestParseSearchPageAcceptsRecognizedEmptyResults(t *testing.T) {
	got, err := parseSearchPage([]byte(`<script type="application/json">{"searchResults":{"edges":[]}}</script>`), "none", 10)
	if err != nil || len(got) != 0 {
		t.Fatalf("results=%+v err=%v", got, err)
	}
}

func TestParseSearchPageRejectsChangedPage(t *testing.T) {
	_, err := parseSearchPage([]byte(`<html><script type="application/json">{"different":true}</script></html>`), "none", 10)
	if !errors.Is(err, errSearchPageChanged) {
		t.Fatalf("error = %v", err)
	}
}

func TestSearchPageSourceSendsBoundedAnonymousRequest(t *testing.T) {
	page, err := os.ReadFile("testdata/search_page.html")
	if err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodGet || r.URL.Path != "/search" || r.URL.Query().Get("q") != "a & b" || r.URL.Query().Get("serp_type") != "default" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		if r.UserAgent() != threads.CrawlerUA || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Errorf("unexpected headers: user-agent=%q authorization=%q cookie=%q", r.UserAgent(), r.Header.Get("Authorization"), r.Header.Get("Cookie"))
		}
		_, _ = w.Write(page)
	}))
	defer server.Close()

	source := newSearchPageSource(server.Client(), server.URL, threads.CrawlerUA, 0)
	got, gotErr := collectSearch(source.Search(context.Background(), "a & b", 1))
	if gotErr != nil || len(got) != 1 || got[0].ID != "101" || requests != 1 {
		t.Fatalf("results=%+v err=%v requests=%d", got, gotErr, requests)
	}
}

func TestSearchPageSourceRejectsOversizedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", maxSearchPageBytes+1)))
	}))
	defer server.Close()

	source := newSearchPageSource(server.Client(), server.URL, threads.CrawlerUA, 0)
	_, err := collectSearch(source.Search(context.Background(), "query", 10))
	if threads.Code(err) != threads.ExitNetwork {
		t.Fatalf("error = %v, code = %d", err, threads.Code(err))
	}
}

func TestSearchPageSourceClassifiesResponses(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		code   int
	}{
		{"forbidden", http.StatusForbidden, "secret wall body", threads.ExitLoginWall},
		{"rate limited", http.StatusTooManyRequests, "secret rate body", threads.ExitRateLimit},
		{"server failure", http.StatusBadGateway, "secret upstream body", threads.ExitNetwork},
		{"changed page", http.StatusOK, `<html>changed</html>`, threads.ExitNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			source := newSearchPageSource(server.Client(), server.URL, threads.CrawlerUA, 0)
			_, err := collectSearch(source.Search(context.Background(), "query", 10))
			if threads.Code(err) != tc.code || strings.Contains(err.Error(), "secret") {
				t.Fatalf("error = %v, code = %d", err, threads.Code(err))
			}
		})
	}
}

func TestSearchPageSourceHonorsCancellationWhilePacing(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(`<script type="application/json">{"searchResults":{"edges":[]}}</script>`))
	}))
	defer server.Close()

	source := newSearchPageSource(server.Client(), server.URL, threads.CrawlerUA, time.Hour)
	if _, err := collectSearch(source.Search(context.Background(), "first", 10)); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := collectSearch(source.Search(ctx, "second", 10))
	if !errors.Is(err, context.Canceled) || requests != 1 {
		t.Fatalf("error=%v requests=%d", err, requests)
	}
}

func collectSearch(sequence func(func(threads.SearchResult, error) bool)) ([]threads.SearchResult, error) {
	var results []threads.SearchResult
	var resultErr error
	sequence(func(result threads.SearchResult, err error) bool {
		if err != nil {
			resultErr = err
			return false
		}
		results = append(results, result)
		return true
	})
	return results, resultErr
}
