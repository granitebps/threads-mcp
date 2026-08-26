package threadscli

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"testing"

	"github.com/granitebps/threads-mcp/internal/config"
	"github.com/granitebps/threads-mcp/internal/domain"
	"github.com/tamnd/threads-cli/threads"
)

func TestClientSearchPostsUsesPageReaderAndMapsResults(t *testing.T) {
	var gotQuery string
	var gotLimit int
	client := testClient(&fakeSource{}, fakeSearchSource{searchFn: func(_ context.Context, query string, limit int) iter.Seq2[threads.SearchResult, error] {
		gotQuery, gotLimit = query, limit
		return sequence([]threads.SearchResult{{ID: "1", Text: "hello"}}, nil)
	}})

	page, err := client.SearchPosts(context.Background(), "  launch  ", 7)
	if err != nil || gotQuery != "launch" || gotLimit != 7 || page.ReturnedCount != 1 || page.Items[0].Text == nil || *page.Items[0].Text != "hello" {
		t.Fatalf("page=%+v err=%v query=%q limit=%d", page, err, gotQuery, gotLimit)
	}
	if page.Completeness != domain.CompletenessUnknown {
		t.Fatalf("completeness=%q", page.Completeness)
	}
}

func TestClientSearchPostsReportsEmptyWindow(t *testing.T) {
	client := testClient(&fakeSource{}, fakeSearchSource{searchFn: func(context.Context, string, int) iter.Seq2[threads.SearchResult, error] {
		return sequence[threads.SearchResult](nil, nil)
	}})
	page, err := client.SearchPosts(context.Background(), "none", 10)
	if err != nil || page.ReturnedCount != 0 || len(page.Warnings) != 1 || page.Warnings[0] != "Threads returned no public search results; completeness is unknown" {
		t.Fatalf("page=%+v err=%v", page, err)
	}
}

func TestClientSearchPostsPreservesPartialResults(t *testing.T) {
	client := testClient(&fakeSource{}, fakeSearchSource{searchFn: func(context.Context, string, int) iter.Seq2[threads.SearchResult, error] {
		return sequence([]threads.SearchResult{{ID: "1"}}, errors.New("page interrupted"))
	}})
	page, err := client.SearchPosts(context.Background(), "query", 10)
	if err != nil || page.ReturnedCount != 1 || page.Completeness != domain.CompletenessPartial || len(page.Warnings) != 1 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
}

func TestClientSearchPostsReturnsSafeErrorBeforeResults(t *testing.T) {
	client := testClient(&fakeSource{}, fakeSearchSource{searchFn: func(context.Context, string, int) iter.Seq2[threads.SearchResult, error] {
		return sequence[threads.SearchResult](nil, &threads.CodeError{Code: threads.ExitNotFound, Msg: "raw detail", Err: errSearchPageChanged})
	}})
	_, err := client.SearchPosts(context.Background(), "query", 10)
	var providerErr *domain.ProviderError
	if !errors.As(err, &providerErr) || providerErr.Code != domain.CodeUpstreamChanged || providerErr.Message == "raw detail" {
		t.Fatalf("error=%+v", err)
	}
}

func TestClientGetPostNormalizesURL(t *testing.T) {
	var got string
	source := &fakeSource{postFn: func(_ context.Context, post string) (*threads.Post, error) {
		got = post
		return &threads.Post{ID: "1", Text: "post"}, nil
	}}
	client := testClient(source, fakeSearchSource{})
	post, err := client.GetPost(context.Background(), "https://www.threads.net/@zuck/post/ABC123")
	if err != nil || got != "https://www.threads.com/@zuck/post/ABC123" || post.Text == nil || *post.Text != "post" {
		t.Fatalf("post=%+v err=%v input=%q", post, err, got)
	}
}

func TestClientGetProfileNormalizesUsername(t *testing.T) {
	var got string
	source := &fakeSource{profileFn: func(_ context.Context, username string) (*threads.Profile, error) {
		got = username
		return &threads.Profile{ID: "1", Username: username}, nil
	}}
	client := testClient(source, fakeSearchSource{})
	profile, err := client.GetProfile(context.Background(), "@zuck")
	if err != nil || got != "zuck" || profile.Username != "zuck" {
		t.Fatalf("profile=%+v err=%v input=%q", profile, err, got)
	}
}

func TestClientListMethodsPreservePartialResults(t *testing.T) {
	source := &fakeSource{
		profilePostsFn: func(context.Context, string, int) iter.Seq2[threads.Post, error] {
			return sequence([]threads.Post{{ID: "1"}}, errors.New("interrupted"))
		},
		postRepliesFn: func(context.Context, string, int) iter.Seq2[threads.Reply, error] {
			return sequence([]threads.Reply{{ID: "2"}}, errors.New("interrupted"))
		},
	}
	client := testClient(source, fakeSearchSource{})
	posts, postErr := client.GetProfilePosts(context.Background(), "zuck", 10)
	replies, replyErr := client.GetPostReplies(context.Background(), "https://www.threads.com/@zuck/post/ABC123", 10)
	if postErr != nil || replyErr != nil || posts.Completeness != domain.CompletenessPartial || replies.Completeness != domain.CompletenessPartial || posts.ReturnedCount != 1 || replies.ReturnedCount != 1 {
		t.Fatalf("posts=%+v postErr=%v replies=%+v replyErr=%v", posts, postErr, replies, replyErr)
	}
}

func TestClientInfoDisclosesHybridAnonymousMode(t *testing.T) {
	client := testClient(&fakeSource{mode: "anonymous", userAgent: threads.CrawlerUA}, fakeSearchSource{})
	info := client.Info()
	if info.Mode != "anonymous-crawler" || info.UserAgent != threads.CrawlerUA || len(info.KnownLimitations) == 0 {
		t.Fatalf("info=%+v", info)
	}
}

func TestNewIgnoresCredentialEnvironment(t *testing.T) {
	t.Setenv("THREADS_TOKEN", "must-not-be-read")
	t.Setenv("THREADS_SESSION", "must-not-be-read")
	t.Setenv("THREADS_CSRF", "must-not-be-read")
	cfg, err := config.Load("test")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Delay = 0
	client, err := New(cfg, "v0.1.1")
	if err != nil {
		t.Fatal(err)
	}
	if client.Info().Mode != "anonymous-crawler" {
		t.Fatalf("mode=%q", client.Info().Mode)
	}
}

func TestNewDisablesEnvironmentProxyForSearch(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")
	cfg, err := config.Load("test")
	if err != nil {
		t.Fatal(err)
	}
	client, err := New(cfg, "v0.1.1")
	if err != nil {
		t.Fatal(err)
	}
	searcher := client.searcher.(*searchPageSource)
	transport, ok := searcher.httpClient.Transport.(*http.Transport)
	if !ok || transport.Proxy != nil {
		t.Fatalf("search transport=%T", searcher.httpClient.Transport)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code string
	}{
		{"cancelled", context.Canceled, domain.CodeTimeout},
		{"deadline", context.DeadlineExceeded, domain.CodeTimeout},
		{"usage", &threads.CodeError{Code: threads.ExitUsage, Msg: "raw"}, domain.CodeInvalidInput},
		{"not found", &threads.CodeError{Code: threads.ExitNotFound, Msg: "raw"}, domain.CodeNotFound},
		{"changed page", &threads.CodeError{Code: threads.ExitNotFound, Msg: "raw", Err: errSearchPageChanged}, domain.CodeUpstreamChanged},
		{"wall", &threads.CodeError{Code: threads.ExitLoginWall, Msg: "raw"}, domain.CodeAccessRestricted},
		{"rate", &threads.CodeError{Code: threads.ExitRateLimit, Msg: "raw"}, domain.CodeRateLimited},
		{"network", &threads.CodeError{Code: threads.ExitNetwork, Msg: "raw"}, domain.CodeNetworkFailure},
		{"other", errors.New("raw"), domain.CodeUpstreamChanged},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classify(tc.err)
			if got.Code != tc.code || got.Message == "raw" || !errors.Is(got, tc.err) {
				t.Fatalf("classified=%+v", got)
			}
		})
	}
}

type fakeSource struct {
	profileFn      func(context.Context, string) (*threads.Profile, error)
	profilePostsFn func(context.Context, string, int) iter.Seq2[threads.Post, error]
	postFn         func(context.Context, string) (*threads.Post, error)
	postRepliesFn  func(context.Context, string, int) iter.Seq2[threads.Reply, error]
	mode           string
	userAgent      string
}

func (f *fakeSource) Profile(ctx context.Context, username string) (*threads.Profile, error) {
	return f.profileFn(ctx, username)
}
func (f *fakeSource) ProfilePosts(ctx context.Context, username string, limit int) iter.Seq2[threads.Post, error] {
	return f.profilePostsFn(ctx, username, limit)
}
func (f *fakeSource) Post(ctx context.Context, post string) (*threads.Post, error) {
	return f.postFn(ctx, post)
}
func (f *fakeSource) PostReplies(ctx context.Context, post string, limit int) iter.Seq2[threads.Reply, error] {
	return f.postRepliesFn(ctx, post, limit)
}
func (f *fakeSource) Mode() string      { return f.mode }
func (f *fakeSource) UserAgent() string { return f.userAgent }

type fakeSearchSource struct {
	searchFn func(context.Context, string, int) iter.Seq2[threads.SearchResult, error]
}

func (f fakeSearchSource) Search(ctx context.Context, query string, limit int) iter.Seq2[threads.SearchResult, error] {
	return f.searchFn(ctx, query, limit)
}

func testClient(source source, searcher searchSource) *Client {
	return &Client{source: source, searcher: searcher, version: "v0.1.1"}
}

func sequence[T any](items []T, finalErr error) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for _, item := range items {
			if !yield(item, nil) {
				return
			}
		}
		if finalErr != nil {
			var zero T
			yield(zero, finalErr)
		}
	}
}
