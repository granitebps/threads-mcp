package threadscli

import (
	"context"
	"encoding/json"
	"errors"
	htmlpkg "html"
	"io"
	"iter"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/tamnd/threads-cli/threads"
)

const maxSearchPageBytes = 8 << 20

var (
	errSearchPageChanged = errors.New("public Threads search page did not contain a recognized search payload")
	searchScriptPattern  = regexp.MustCompile(`(?is)<script[^>]*type=["']application/json["'][^>]*>(.*?)</script>`)
)

type searchPageSource struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	delay      time.Duration

	mu          sync.Mutex
	lastRequest time.Time
}

func newSearchPageSource(httpClient *http.Client, baseURL, userAgent string, delay time.Duration) *searchPageSource {
	return &searchPageSource{httpClient: httpClient, baseURL: baseURL, userAgent: userAgent, delay: delay}
}

func (s *searchPageSource) Search(ctx context.Context, query string, limit int) iter.Seq2[threads.SearchResult, error] {
	return func(yield func(threads.SearchResult, error) bool) {
		if err := s.waitForPace(ctx); err != nil {
			yield(threads.SearchResult{}, err)
			return
		}

		endpoint, err := url.Parse(s.baseURL)
		if err != nil {
			yield(threads.SearchResult{}, &threads.CodeError{Code: threads.ExitNetwork, Msg: "invalid Threads search endpoint"})
			return
		}
		endpoint.Path = "/search"
		params := endpoint.Query()
		params.Set("q", query)
		params.Set("serp_type", "default")
		endpoint.RawQuery = params.Encode()

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
		if err != nil {
			yield(threads.SearchResult{}, &threads.CodeError{Code: threads.ExitNetwork, Msg: "create Threads search request", Err: err})
			return
		}
		request.Header.Set("User-Agent", s.userAgent)

		response, err := s.httpClient.Do(request)
		if err != nil {
			if ctx.Err() != nil {
				yield(threads.SearchResult{}, ctx.Err())
				return
			}
			yield(threads.SearchResult{}, &threads.CodeError{Code: threads.ExitNetwork, Msg: "Threads search request failed", Err: err})
			return
		}
		defer func() { _ = response.Body.Close() }()

		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			code := threads.ExitNetwork
			message := "Threads search returned an unexpected status"
			switch response.StatusCode {
			case http.StatusForbidden, http.StatusUnauthorized:
				code, message = threads.ExitLoginWall, "Threads restricted anonymous search access"
			case http.StatusTooManyRequests:
				code, message = threads.ExitRateLimit, "Threads rate limited anonymous search"
			}
			yield(threads.SearchResult{}, &threads.CodeError{Code: code, Msg: message})
			return
		}

		page, err := io.ReadAll(io.LimitReader(response.Body, maxSearchPageBytes+1))
		if err != nil {
			yield(threads.SearchResult{}, &threads.CodeError{Code: threads.ExitNetwork, Msg: "read Threads search response", Err: err})
			return
		}
		if len(page) > maxSearchPageBytes {
			yield(threads.SearchResult{}, &threads.CodeError{Code: threads.ExitNetwork, Msg: "Threads search response exceeded the size limit"})
			return
		}
		results, err := parseSearchPage(page, query, limit)
		if err != nil {
			yield(threads.SearchResult{}, &threads.CodeError{Code: threads.ExitNotFound, Msg: "Threads public search page changed", Err: errSearchPageChanged})
			return
		}
		for _, result := range results {
			if !yield(result, nil) {
				return
			}
		}
	}
}

func (s *searchPageSource) waitForPace(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	remaining := s.delay - time.Since(s.lastRequest)
	if remaining > 0 {
		timer := time.NewTimer(remaining)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	s.lastRequest = time.Now()
	return nil
}

func parseSearchPage(page []byte, query string, limit int) ([]threads.SearchResult, error) {
	var results []threads.SearchResult
	seen := make(map[string]bool)
	found := false

	for _, match := range searchScriptPattern.FindAllSubmatch(page, -1) {
		var data any
		raw := htmlpkg.UnescapeString(string(match[1]))
		if json.Unmarshal([]byte(raw), &data) != nil {
			continue
		}
		walkSearchPayloads(data, 0, func(payload map[string]any) {
			found = true
			for _, post := range postsFromSearchPayload(payload) {
				id := valueString(post["pk"])
				if id == "" {
					id = valueString(post["id"])
				}
				if id == "" || seen[id] || (limit > 0 && len(results) >= limit) {
					continue
				}
				seen[id] = true
				results = append(results, searchResultFromPost(post, query))
			}
		})
	}
	if !found {
		return nil, errSearchPageChanged
	}
	return results, nil
}

func walkSearchPayloads(data any, depth int, visit func(map[string]any)) {
	if depth > 30 {
		return
	}
	switch value := data.(type) {
	case map[string]any:
		if payload, ok := value["searchResults"].(map[string]any); ok {
			visit(payload)
		}
		for _, child := range value {
			walkSearchPayloads(child, depth+1, visit)
		}
	case []any:
		for _, child := range value {
			walkSearchPayloads(child, depth+1, visit)
		}
	}
}

func postsFromSearchPayload(payload map[string]any) []map[string]any {
	edges, _ := payload["edges"].([]any)
	posts := make([]map[string]any, 0, len(edges))
	for _, edgeValue := range edges {
		edge, _ := edgeValue.(map[string]any)
		node, _ := edge["node"].(map[string]any)
		thread, _ := node["thread"].(map[string]any)
		items, _ := thread["thread_items"].([]any)
		for _, itemValue := range items {
			item, _ := itemValue.(map[string]any)
			if post, ok := item["post"].(map[string]any); ok {
				posts = append(posts, post)
			}
		}
	}
	return posts
}

func searchResultFromPost(post map[string]any, query string) threads.SearchResult {
	result := threads.SearchResult{
		ID:         valueString(post["pk"]),
		Query:      query,
		SearchedAt: time.Now().UTC(),
	}
	if result.ID == "" {
		result.ID = valueString(post["id"])
	}
	if caption, ok := post["caption"].(map[string]any); ok {
		result.Text = valueString(caption["text"])
	}
	if user, ok := post["user"].(map[string]any); ok {
		result.Username = valueString(user["username"])
	}
	if timestamp := valueFloat(post["taken_at"]); timestamp > 0 {
		result.Timestamp = time.Unix(int64(timestamp), 0).UTC()
	}
	switch int(valueFloat(post["media_type"])) {
	case 1:
		result.MediaType = "IMAGE"
	case 2:
		result.MediaType = "VIDEO"
	case 8:
		result.MediaType = "CAROUSEL_ALBUM"
	default:
		result.MediaType = "TEXT_POST"
	}
	if info, ok := post["text_post_app_info"].(map[string]any); ok {
		result.IsQuotePost, _ = info["is_quote_post"].(bool)
		result.IsReply = info["reply_to_author"] != nil
	}
	code := valueString(post["code"])
	if code != "" && result.Username != "" {
		result.Permalink = "https://www.threads.com/@" + result.Username + "/post/" + code
	}
	return result
}

func valueString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	default:
		return ""
	}
}

func valueFloat(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case string:
		parsed, _ := strconv.ParseFloat(typed, 64)
		return parsed
	default:
		return 0
	}
}
