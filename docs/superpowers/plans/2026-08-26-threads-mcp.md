# threads-mcp implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` to implement this plan task by task. Steps use checkbox syntax for tracking.

**Goal:** Build and prepare an open-source stdio MCP server that reads public Threads posts, post replies, profiles, and recent profile posts without credentials.

**Architecture:** A Go command hosts the official MCP SDK over stdio. Server handlers depend on a server-owned provider interface. A local hybrid adapter imports the pinned `threads-cli` Go package for profile and conversation reads and parses server-rendered JSON from the public Threads search page for keyword search.

**Tech stack:** Go 1.26, `github.com/modelcontextprotocol/go-sdk v1.7.0`, `github.com/tamnd/threads-cli v0.1.1`, GoReleaser v2, GitHub Actions, Cosign, and Syft.

**Spec:** `docs/design.md`

## Global constraints

- Read public Threads data only.
- Never load Threads tokens, session cookies, or CSRF values from flags, files, or environment variables.
- Use `threads.CrawlerUA` only through `threads-cli`'s anonymous mode and disclose that access mechanism in server info and documentation.
- Stop implementation if profile or conversation reads fail in anonymous crawler mode, or if the public search page no longer contains parseable post records.
- Expose exactly six V1 tools: `search_posts`, `get_post`, `get_post_replies`, `get_profile`, `get_profile_posts`, and `get_server_info`.
- Keep `get_profile_replies` out of the server and documentation.
- Run MCP over stdio. Write protocol messages only to stdout and diagnostics only to stderr.
- Omit unavailable values or encode them as JSON `null`. Never convert missing provider values into zero or false claims.
- Mark uncertain list results with `completeness: "unknown"` and explain why in `warnings`.
- Default tool timeout is 45 seconds. `THREADS_MCP_TIMEOUT` accepts 5 seconds through 120 seconds.
- Make no more than two attempts per upstream request and wait at least one second between requests.
- Do not add Homebrew, npm, authentication, browser automation, proxy rotation, Docker images, or HTTP transport.
- Do not commit, push, publish, or create a pull request without separate user authorization. Commit commands below are checkpoints only.
- If implementation starts before this directory is a Git repository, include `git init -b main` in the explicit implementation authorization. Repository initialization does not authorize any commit or remote operation.

## Planned file map

```text
.
├── .github/workflows/ci.yml
├── .github/workflows/live-smoke.yml
├── .github/workflows/release.yml
├── .gitignore
├── .golangci.yml
├── .goreleaser.yaml
├── LICENSE
├── README.md
├── THIRD_PARTY_NOTICES.md
├── cmd/threads-mcp/main.go
├── docs/design.md
├── go.mod
├── go.sum
├── internal/config/config.go
├── internal/config/config_test.go
├── internal/domain/error.go
├── internal/domain/model.go
├── internal/identifier/identifier.go
├── internal/identifier/identifier_test.go
├── internal/provider/provider.go
├── internal/provider/threadscli/client.go
├── internal/provider/threadscli/client_test.go
├── internal/provider/threadscli/live_probe_test.go
├── internal/provider/threadscli/map.go
├── internal/provider/threadscli/map_test.go
├── internal/provider/threadscli/search_page.go
├── internal/provider/threadscli/search_page_test.go
├── internal/provider/threadscli/testdata/search_page.html
├── internal/server/info.go
├── internal/server/post_tools.go
├── internal/server/profile_tools.go
├── internal/server/result.go
├── internal/server/search_tool.go
├── internal/server/server.go
├── internal/server/server_test.go
├── internal/testutil/provider.go
├── internal/version/version.go
└── tests/stdio_test.go
```

## Task 1: Prove anonymous crawler access

**Gate result (2026-08-26): APPROVED HYBRID REVISION.** Profile, profile-post, post, and post-reply calls completed through anonymous crawler mode, but `threads-cli v0.1.1` keyword search failed with `graphql returned an unexpected shape (doc_id may be stale)`. Upstream `main` uses the same search document ID. A separate read-only probe of the public Threads search page returned HTTP 200 with embedded `thread_items` records. The user approved a local search-page reader while retaining `threads-cli` for the other reads.

**Files:**

- Create: `go.mod`
- Create: `.gitignore`
- Create: `internal/provider/threadscli/live_probe_test.go`

**Interfaces:**

- Consumes: `threads.Config`, `threads.NewClient`, `Client.Profile`, `Client.ProfilePosts`, `Client.Post`, and `Client.PostReplies` from `threads-cli v0.1.1`.
- Produces: live feasibility evidence for the upstream reads and the public search-page surface.

- [ ] **Step 1: Create the pinned Go module**

Create `go.mod` with:

```go
module github.com/granitebps/threads-mcp

go 1.26

require (
	github.com/modelcontextprotocol/go-sdk v1.7.0
	github.com/tamnd/threads-cli v0.1.1
)
```

Create `.gitignore` with:

```gitignore
/bin/
/dist/
/coverage.out
*.test
```

Run: `go mod download`

Expected: exit 0 and a new `go.sum` containing both direct modules.

- [ ] **Step 2: Add the live anonymous-crawler probe**

Create `internal/provider/threadscli/live_probe_test.go`:

```go
//go:build live

package threadscli_test

import (
	"context"
	"testing"
	"time"

	"github.com/tamnd/threads-cli/threads"
)

func TestAnonymousCrawlerPublicReads(t *testing.T) {
	cfg := threads.Config{
		Delay:     time.Second,
		Retries:   2,
		Timeout:   30 * time.Second,
		UserAgent: threads.CrawlerUA,
		Lang:      "en-US",
		NoCache:   true,
	}
	client, err := threads.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if client.Mode() != "anonymous" {
		t.Fatalf("mode = %q, want anonymous", client.Mode())
	}
	if client.UserAgent() != threads.CrawlerUA {
		t.Fatalf("user agent = %q", client.UserAgent())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	profile, err := client.Profile(ctx, "zuck")
	if err != nil || profile.ID == "" {
		t.Fatalf("profile read with anonymous crawler mode failed: profile=%+v err=%v", profile, err)
	}

	var post threads.Post
	for item, itemErr := range client.ProfilePosts(ctx, "zuck", 1) {
		if itemErr != nil {
			t.Fatalf("profile posts failed: %v", itemErr)
		}
		post = item
	}
	if post.Permalink == "" {
		t.Fatal("profile posts returned no fetchable public post")
	}

	gotPost, err := client.Post(ctx, post.Permalink)
	if err != nil || gotPost.ID == "" {
		t.Fatalf("post read failed: post=%+v err=%v", gotPost, err)
	}

	for _, itemErr := range client.PostReplies(ctx, post.Permalink, 1) {
		if itemErr != nil {
			t.Fatalf("post replies failed: %v", itemErr)
		}
	}
}
```

- [ ] **Step 3: Run the release gate**

Run: `go test -tags=live ./internal/provider/threadscli -run TestAnonymousCrawlerPublicReads -count=1 -v`

Expected: PASS with profile, profile-post, post, and post-reply calls using anonymous crawler mode. Search is verified through the local page reader in Task 4 because the upstream search call is known to fail.

If one of these upstream reads fails because Threads returns a login wall or generic page, stop. Record the response class without saving raw bodies, and ask the user whether to evaluate a hosted provider or end the project. Do not add credentials, proxies, CAPTCHA solving, or another access workaround.

- [ ] **Step 4: Run dependency checks**

Run: `go mod download && go list -m all && go version`

Expected: exit 0, `threads-cli v0.1.1`, Go SDK `v1.7.0`, and Go 1.26 or newer. Do not run `go mod tidy` until Task 5 imports the MCP SDK, because tidy correctly removes unused direct modules.

**Commit checkpoint, only with separate authorization:** `git add go.mod go.sum .gitignore internal/provider/threadscli/live_probe_test.go && git commit -m "test: prove anonymous Threads access"`

## Task 2: Define configuration, domain models, and provider contracts

**Files:**

- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `internal/domain/error.go`
- Create: `internal/domain/model.go`
- Create: `internal/provider/provider.go`

**Interfaces:**

- Consumes: process environment and `context.Context`.
- Produces: `config.Load`, domain records, `domain.ProviderError`, and `provider.Provider` used by all later tasks.

- [ ] **Step 1: Write configuration tests**

Test these exact cases in `internal/config/config_test.go`:

```go
func TestLoadTimeout(t *testing.T) {
	t.Setenv("THREADS_MCP_TIMEOUT", "30s")
	cfg, err := Load("1.2.3")
	if err != nil || cfg.ToolTimeout != 30*time.Second {
		t.Fatalf("cfg=%+v err=%v", cfg, err)
	}
}

func TestLoadRejectsTimeoutOutsideRange(t *testing.T) {
	for _, value := range []string{"4s", "121s", "not-a-duration"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("THREADS_MCP_TIMEOUT", value)
			if _, err := Load("dev"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
```

Run: `go test ./internal/config -count=1`

Expected: FAIL because `Load` does not exist.

- [ ] **Step 2: Implement configuration**

Define:

```go
type Config struct {
	ToolTimeout time.Duration
	HTTPTimeout time.Duration
	Delay       time.Duration
	Retries     int
	CacheDir    string
	CacheTTL    time.Duration
}

func Load(version string) (Config, error)
```

Use 45 seconds as `ToolTimeout`, 30 seconds as `HTTPTimeout`, one second as `Delay`, two retries, and one hour as `CacheTTL`. Resolve the cache directory with `os.UserCacheDir()` plus `threads-mcp`. Parse only `THREADS_MCP_TIMEOUT`. Do not read credential or proxy variables.

Run: `go test ./internal/config -count=1`

Expected: PASS.

- [ ] **Step 3: Define server-owned models**

Create `internal/domain/model.go` with these public shapes:

```go
type Completeness string

const (
	CompletenessComplete Completeness = "complete"
	CompletenessPartial  Completeness = "partial"
	CompletenessUnknown  Completeness = "unknown"
)

type Post struct {
	ID           string  `json:"id"`
	Shortcode    *string `json:"shortcode,omitempty"`
	Text         *string `json:"text,omitempty"`
	Username     *string `json:"username,omitempty"`
	UserID       *string `json:"user_id,omitempty"`
	Permalink    *string `json:"permalink,omitempty"`
	Timestamp    *string `json:"timestamp,omitempty"`
	MediaType    *string `json:"media_type,omitempty"`
	MediaURLs    []string `json:"media_urls,omitempty"`
	LikeCount    *int64  `json:"like_count,omitempty"`
	ReplyCount   *int64  `json:"reply_count,omitempty"`
	RepostCount  *int64  `json:"repost_count,omitempty"`
	QuoteCount   *int64  `json:"quote_count,omitempty"`
	IsReply      *bool   `json:"is_reply,omitempty"`
	IsQuotePost  *bool   `json:"is_quote_post,omitempty"`
	ReplyToID    *string `json:"reply_to_id,omitempty"`
	QuotedPostID *string `json:"quoted_post_id,omitempty"`
}

type Reply struct {
	Post
	ParentID *string `json:"parent_id,omitempty"`
	RootID   *string `json:"root_id,omitempty"`
}

type Profile struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	Name           *string `json:"name,omitempty"`
	Biography      *string `json:"biography,omitempty"`
	ProfilePicURL  *string `json:"profile_pic_url,omitempty"`
	IsVerified     *bool   `json:"is_verified,omitempty"`
	FollowerCount  *int64  `json:"follower_count,omitempty"`
	FollowingCount *int64  `json:"following_count,omitempty"`
	URL            string  `json:"url"`
}

type Page[T any] struct {
	Items          []T          `json:"items"`
	ReturnedCount  int          `json:"returned_count"`
	RequestedLimit int          `json:"requested_limit"`
	Completeness   Completeness `json:"completeness"`
	Warnings       []string     `json:"warnings"`
}
```

Create `internal/domain/error.go` with:

```go
const (
	CodeInvalidInput        = "INVALID_INPUT"
	CodeNotFound            = "NOT_FOUND"
	CodeAccessRestricted    = "ACCESS_RESTRICTED"
	CodeRateLimited         = "RATE_LIMITED"
	CodeNetworkFailure      = "NETWORK_FAILURE"
	CodeTimeout             = "TIMEOUT"
	CodeUnsupportedOperation = "UNSUPPORTED_OPERATION"
	CodeUpstreamChanged     = "UPSTREAM_CHANGED"
	CodeInternalError       = "INTERNAL_ERROR"
)

type ProviderError struct {
	Code              string
	Message           string
	Retryable         bool
	RetryAfterSeconds *int
	Cause             error
}

func (e *ProviderError) Error() string
func (e *ProviderError) Unwrap() error
```

- [ ] **Step 4: Define the provider interface**

Create `internal/provider/provider.go`:

```go
type Info struct {
	Name             string   `json:"name"`
	Version          string   `json:"version"`
	Mode             string   `json:"mode"`
	UserAgent        string   `json:"user_agent"`
	SupportedInputs  []string `json:"supported_inputs"`
	KnownLimitations []string `json:"known_limitations"`
}

type Provider interface {
	SearchPosts(context.Context, string, int) (domain.Page[domain.Post], error)
	GetPost(context.Context, string) (domain.Post, error)
	GetPostReplies(context.Context, string, int) (domain.Page[domain.Reply], error)
	GetProfile(context.Context, string) (domain.Profile, error)
	GetProfilePosts(context.Context, string, int) (domain.Page[domain.Post], error)
	Info() Info
}
```

Run: `gofmt -w internal/config internal/domain internal/provider && go test ./internal/config ./internal/domain ./internal/provider -count=1`

Expected: PASS.

**Commit checkpoint, only with separate authorization:** `git add internal/config internal/domain internal/provider/provider.go && git commit -m "feat: define server contracts"`

## Task 3: Validate identifiers and map provider data safely

**Files:**

- Create: `internal/identifier/identifier.go`
- Create: `internal/identifier/identifier_test.go`
- Create: `internal/provider/threadscli/map.go`
- Create: `internal/provider/threadscli/map_test.go`

**Interfaces:**

- Consumes: `thid.Classify`, `threads.Post`, `threads.Reply`, `threads.Profile`, and `threads.SearchResult`.
- Produces: `identifier.Username`, `identifier.Post`, and mapping functions that never claim missing zero or false values.

- [ ] **Step 1: Write identifier tests**

Cover these table entries:

```go
var usernameCases = []struct {
	in, want string
}{
	{"zuck", "zuck"},
	{"@zuck", "zuck"},
	{"https://www.threads.com/@zuck", "zuck"},
}

var postCases = []struct {
	in      string
	wantErr bool
}{
	{"https://www.threads.com/@zuck/post/ABC123", false},
	{"ABC123", true},
	{"123456789", true},
}
```

Run: `go test ./internal/identifier -count=1`

Expected: FAIL because the functions do not exist.

- [ ] **Step 2: Implement identifier validation**

Expose:

```go
func Username(input string) (string, error)
func Post(input string) (string, error)
```

Trim whitespace, use `thid.Classify`, return a normalized handle for profiles, and require `KindPost` with a non-empty URL for posts. Return `domain.ProviderError{Code: domain.CodeInvalidInput}` for invalid values.

Run: `go test ./internal/identifier -count=1`

Expected: PASS.

- [ ] **Step 3: Write missing-value mapping tests**

In `map_test.go`, pass zero-valued upstream counters and false booleans. Assert all corresponding domain pointers are nil. Pass non-empty strings, a non-zero timestamp, media URLs, and true relationship flags. Assert those values are preserved and timestamps equal `time.RFC3339` UTC strings.

Also assert `mapSearchResult` never adds engagement counters because `threads.SearchResult` does not expose them.

Run: `go test ./internal/provider/threadscli -run 'TestMap' -count=1`

Expected: FAIL because mapping functions do not exist.

- [ ] **Step 4: Implement conservative mapping**

Add unexported helpers `stringPtr`, `truePtr`, and `timePtr`. Map false booleans to nil. For `threads-cli v0.1.1`, map all engagement counters and profile counts to nil because the provider structs cannot distinguish missing values from real zero. Add a code comment that names the pinned provider limitation and points to `docs/design.md`.

Map a boolean pointer only when the upstream value is true. Map a timestamp only when `time.Time.IsZero()` is false. Copy media URL slices before returning them.

Run: `gofmt -w internal/identifier internal/provider/threadscli && go test ./internal/identifier ./internal/provider/threadscli -count=1`

Expected: PASS.

**Commit checkpoint, only with separate authorization:** `git add internal/identifier internal/provider/threadscli/map.go internal/provider/threadscli/map_test.go && git commit -m "feat: normalize and map public Threads data"`

## Task 4: Implement the hybrid local adapter

**Files:**

- Create: `internal/provider/threadscli/client.go`
- Create: `internal/provider/threadscli/client_test.go`
- Create: `internal/provider/threadscli/search_page.go`
- Create: `internal/provider/threadscli/search_page_test.go`
- Create: `internal/provider/threadscli/testdata/search_page.html`

**Interfaces:**

- Consumes: `config.Config`, domain mapping functions, and the provider interface.
- Produces: `func New(config.Config, string) (*Client, error)`, a bounded public search-page reader, and a complete `provider.Provider` implementation.

- [ ] **Step 1: Define a fakeable upstream interface**

Use this interface in `client.go`:

```go
type source interface {
	Profile(context.Context, string) (*threads.Profile, error)
	ProfilePosts(context.Context, string, int) iter.Seq2[threads.Post, error]
	Post(context.Context, string) (*threads.Post, error)
	PostReplies(context.Context, string, int) iter.Seq2[threads.Reply, error]
	Mode() string
	UserAgent() string
}

type searchSource interface {
	Search(context.Context, string, int) iter.Seq2[threads.SearchResult, error]
}

type Client struct {
	source   source
	searcher searchSource
	version  string
}
```

`New` must construct `threads.Config` directly. Set `Token`, `Session`, `CSRF`, and `Proxy` to empty strings. Set `UserAgent` to `threads.CrawlerUA`; set delay, retries, timeout, cache directory, cache TTL, and language from the approved configuration. Never call `threads.DefaultConfig()`.

- [ ] **Step 2: Write failing search-page parser tests**

Create a small, hand-minimized `testdata/search_page.html` containing:

- One irrelevant JSON script.
- One `application/json` script whose nested object contains `searchResults.edges`.
- Two edges containing `node.thread.thread_items[0].post` records.
- One duplicate post ID to prove deduplication.
- Post fields for `pk`, `code`, `caption.text`, `taken_at`, `media_type`, `user.pk`, `user.username`, and `text_post_app_info`.

Test `parseSearchPage` for two records in document order, stable IDs, canonical permalinks, text, usernames, timestamps, and deduplication. Add separate tests proving that a recognized `searchResults` object with empty `edges` is a valid empty result, while HTML with no recognized search payload returns `errSearchPageChanged`.

Run: `go test ./internal/provider/threadscli -run 'TestParseSearchPage' -count=1`

Expected: FAIL because `parseSearchPage` does not exist.

- [ ] **Step 3: Implement the minimal search-page parser**

In `search_page.go`, extract only `<script type="application/json">` bodies, HTML-unescape them, decode JSON with `encoding/json`, and recursively locate objects containing `searchResults`. Walk only `searchResults.edges[].node.thread.thread_items[].post`; do not collect unrelated `thread_items` elsewhere on the page. Convert each unique post into `threads.SearchResult` and stop at `limit`. Use a recursion depth cap of 30. Define the sentinel `errSearchPageChanged` for pages with no recognized search payload.

Run: `gofmt -w internal/provider/threadscli/search_page.go internal/provider/threadscli/search_page_test.go && go test ./internal/provider/threadscli -run 'TestParseSearchPage' -count=1`

Expected: PASS.

- [ ] **Step 4: Write failing bounded-fetch tests**

Use `httptest.Server` to verify `searchPageSource.Search`:

- Sends exactly one GET to `/search` with encoded `q` and `serp_type=default`.
- Sends `threads.CrawlerUA` and no cookies or authorization headers.
- Stops after the requested limit.
- Rejects bodies larger than 8 MiB.
- Maps HTTP 403 to `threads.ExitLoginWall`, 429 to `threads.ExitRateLimit`, other non-2xx responses to `threads.ExitNetwork`, and a changed page to `threads.ExitNotFound` without exposing the response body.
- Stops promptly when its context is cancelled while waiting for the one-second request pace.

Run: `go test ./internal/provider/threadscli -run 'TestSearchPageSource' -count=1`

Expected: FAIL because `searchPageSource` does not exist.

- [ ] **Step 5: Implement the bounded search-page fetcher**

Implement a `searchPageSource` with an injected `http.Client`, base URL, user agent, one-second request delay, and a mutex-protected last-request time. It performs no pagination, browser execution, cookies, credentials, proxies, or CAPTCHA handling. Limit the response with `io.LimitReader` to 8 MiB plus one byte and return safe coded errors that the existing `classify` function can map.

Run: `gofmt -w internal/provider/threadscli && go test ./internal/provider/threadscli -run 'TestSearchPageSource|TestParseSearchPage' -count=1`

Expected: PASS.

- [ ] **Step 6: Write adapter tests with fake upstream and search sources**

Test each method for:

- Normal mapping.
- Limit propagation.
- Iterator error before any item.
- Iterator error after one item, which returns a partial page and a warning rather than discarding the item.
- Empty search, which returns `CompletenessUnknown` and the warning `Threads returned no public search results; completeness is unknown`.
- Short successful search pages, which return `CompletenessUnknown` because the public page is a ranked window with no completeness proof.
- Short successful `PostReplies` and `ProfilePosts` calls, which return `CompletenessUnknown` because `threads-cli v0.1.1` hides pagination completion.
- `Info` reporting anonymous crawler mode, the active crawler user agent, URL-only post input, and provider limitations.

Run: `go test ./internal/provider/threadscli -run 'TestClient' -count=1`

Expected: FAIL because adapter methods do not exist.

- [ ] **Step 7: Implement error classification**

Add:

```go
func classify(err error) *domain.ProviderError
```

Map errors in this order:

- `context.DeadlineExceeded` to `TIMEOUT`, retryable true.
- `context.Canceled` to `TIMEOUT`, retryable true, with message `request cancelled`.
- `threads.ExitUsage` to `INVALID_INPUT`, retryable false.
- `threads.ExitNotFound` to `NOT_FOUND`, retryable false.
- `threads.ExitLoginWall` to `ACCESS_RESTRICTED`, retryable false.
- `threads.ExitRateLimit` to `RATE_LIMITED`, retryable true.
- `threads.ExitNetwork` to `NETWORK_FAILURE`, retryable true.
- Any other provider error to `UPSTREAM_CHANGED`, retryable false.

Use fixed safe messages. Keep the original error only in `Cause`. Do not copy the upstream message into the public message.

- [ ] **Step 8: Implement provider methods**

Validate inputs through `internal/identifier`. Collect iterator results until the adapter receives the requested limit or the iterator ends. Return:

- `CompletenessPartial` when an iterator yields items and then an explicit error.
- An error when the iterator fails before yielding an item.
- `CompletenessUnknown` for cleanly ended `Search`, `PostReplies`, and `ProfilePosts` calls because neither source proves that its public window is complete.
- The exact warnings tested in Step 2.

For a partial page, return the page and no Go error. Put a safe warning in the page so the MCP handler cannot turn useful partial data into a tool error.

Run: `gofmt -w internal/provider/threadscli && go test ./internal/provider/threadscli -count=1`

Expected: PASS, excluding the `live` build-tagged probe.

- [ ] **Step 9: Prove anonymous construction**

Add a constructor test that sets `THREADS_TOKEN`, `THREADS_SESSION`, and `THREADS_CSRF` to non-empty values before calling `New`. Assert `Client.Info().Mode == "anonymous-crawler"` and that the search request sends neither credentials nor cookies.

Run: `go test ./internal/provider/threadscli -run TestNewIgnoresCredentialEnvironment -count=1`

Expected: PASS.

**Commit checkpoint, only with separate authorization:** `git add internal/provider/threadscli && git commit -m "feat: add anonymous Threads provider"`

## Task 5: Build the MCP server core and server-info tool

**Files:**

- Create: `internal/version/version.go`
- Create: `internal/server/result.go`
- Create: `internal/server/info.go`
- Create: `internal/server/server.go`
- Create: `internal/server/server_test.go`
- Create: `internal/testutil/provider.go`

**Interfaces:**

- Consumes: `provider.Provider`, `config.Config`, and Go SDK `mcp.AddTool`.
- Produces: `server.New`, common typed tool outputs, safe tool errors, and `get_server_info`.

- [ ] **Step 1: Add build metadata**

Create:

```go
package version

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
```

- [ ] **Step 2: Define typed outputs and error conversion**

In `result.go`, define:

```go
type ErrorOutput struct {
	Code              string `json:"code"`
	Message           string `json:"message"`
	Retryable         bool   `json:"retryable"`
	RetryAfterSeconds *int   `json:"retry_after_seconds,omitempty"`
}

type SearchPostsOutput struct {
	Page  *domain.Page[domain.Post] `json:"result,omitempty"`
	Error *ErrorOutput              `json:"error,omitempty"`
}

type PostOutput struct {
	Post  *domain.Post `json:"post,omitempty"`
	Error *ErrorOutput `json:"error,omitempty"`
}

type RepliesOutput struct {
	Page  *domain.Page[domain.Reply] `json:"result,omitempty"`
	Error *ErrorOutput               `json:"error,omitempty"`
}

type ProfileOutput struct {
	Profile *domain.Profile `json:"profile,omitempty"`
	Error   *ErrorOutput    `json:"error,omitempty"`
}

type ProfilePostsOutput struct {
	Page  *domain.Page[domain.Post] `json:"result,omitempty"`
	Error *ErrorOutput              `json:"error,omitempty"`
}
```

Add:

```go
func errorResult(err error) (*mcp.CallToolResult, *ErrorOutput)
```

Use `errors.As` to read a `domain.ProviderError`. Unknown errors become `INTERNAL_ERROR` with message `internal server error` and no cause text. The helper returns `&mcp.CallToolResult{IsError: true}` plus the safe `ErrorOutput`. Each handler places that error in its own typed output and returns a nil Go error. This lets the SDK create matching `structuredContent` and JSON text without exposing raw errors.

- [ ] **Step 3: Create the fake provider**

`internal/testutil/provider.go` defines function fields for all five data methods plus `InfoValue`. Every method calls its function field and fails the test through a supplied callback if the field is nil. This prevents a handler test from silently making an unexpected provider call.

- [ ] **Step 4: Write server metadata tests**

Construct the server with in-memory transports. List tools through an MCP client and assert the exact six names. For every tool assert:

```go
if !tool.Annotations.ReadOnlyHint || !tool.Annotations.IdempotentHint {
	t.Fatalf("tool %s is not marked read-only and idempotent", tool.Name)
}
if tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint {
	t.Fatalf("tool %s destructiveHint", tool.Name)
}
if tool.Annotations.OpenWorldHint == nil || !*tool.Annotations.OpenWorldHint {
	t.Fatalf("tool %s openWorldHint", tool.Name)
}
```

Call `get_server_info` and assert SDK `v1.7.0`, protocol `2026-07-28`, anonymous provider mode, supported tools, excluded `get_profile_replies`, and known provider warnings.

Run: `go test ./internal/server -run 'TestToolList|TestServerInfo' -count=1`

Expected: FAIL because the server does not exist.

- [ ] **Step 5: Register the core server and info tool**

Expose:

```go
type BuildInfo struct {
	Version    string
	Commit     string
	Date       string
	SDKVersion string
	Protocol   string
}

func New(p provider.Provider, cfg config.Config, build BuildInfo) *mcp.Server
```

Construct the server with `mcp.NewServer`. Use implementation name `threads-mcp`. Register tools through generic `mcp.AddTool`. Supply the explicit JSON Schema 2020-12 input schemas defined in Tasks 6 and 7 so defaults and limits are validated by the SDK. Let the SDK infer output schemas from the typed output structs.

Use annotations with `ReadOnlyHint: true`, `IdempotentHint: true`, `DestructiveHint: boolPtr(false)`, and `OpenWorldHint: boolPtr(true)`.

Run: `gofmt -w internal/version internal/server internal/testutil && go test ./internal/server -count=1`

Expected: PASS for server-list and server-info tests. Data tool calls may still be absent until Tasks 6 and 7.

**Commit checkpoint, only with separate authorization:** `git add internal/version internal/server internal/testutil && git commit -m "feat: add MCP server core"`

## Task 6: Add search and post-reading tools

**Files:**

- Create: `internal/server/search_tool.go`
- Create: `internal/server/post_tools.go`
- Modify: `internal/server/server_test.go`

**Interfaces:**

- Consumes: `Provider.SearchPosts`, `Provider.GetPost`, and `Provider.GetPostReplies`.
- Produces: typed handlers for `search_posts`, `get_post`, and `get_post_replies`.

- [ ] **Step 1: Define exact inputs**

```go
type SearchPostsInput struct {
	Query string `json:"query" jsonschema:"keyword query containing 1 to 500 Unicode code points after trimming"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of results"`
}

type GetPostInput struct {
	Post string `json:"post" jsonschema:"canonical public Threads post URL"`
}

type GetPostRepliesInput struct {
	Post  string `json:"post" jsonschema:"canonical public Threads post URL"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of replies"`
}
```

Define explicit input schemas as `map[string]any` values. `search_posts` requires `query`, sets `minLength: 1`, `maxLength: 500`, and defines `limit` with `minimum: 1`, `maximum: 50`, and `default: 10`. `get_post` requires `post` with `minLength: 1`. `get_post_replies` requires `post` and defines `limit` with `minimum: 1`, `maximum: 100`, and `default: 20`. Every schema sets `type: "object"` and `additionalProperties: false`.

After schema defaults, handlers still normalize zero values to the stated defaults so direct unit calls and MCP calls behave the same. Semantic validation counts Unicode code points after trimming because JSON Schema `maxLength` and the Go SDK validator must be confirmed against that requirement in tests.

- [ ] **Step 2: Write handler tests**

For each tool, call through an in-memory MCP client and assert:

- The provider receives normalized input and the correct default limit.
- The success result has `IsError == false` and non-nil structured content.
- JSON text unmarshals to the same value as structured content.
- An invalid URL returns `INVALID_INPUT` with `IsError == true`.
- A provider timeout returns only the safe `TIMEOUT` message and does not contain the wrapped cause.
- A partial replies page remains a success with `completeness: "partial"` and warnings.

Run: `go test ./internal/server -run 'TestSearchPosts|TestGetPost' -count=1`

Expected: FAIL because the handlers are not registered.

- [ ] **Step 3: Implement handlers with request deadlines**

Each handler trims string input, applies its default limit, creates `context.WithTimeout(ctx, cfg.ToolTimeout)`, calls the provider, and converts all provider errors into typed tool-error output. It never returns a provider error as the handler's Go error.

Use these descriptions:

- `search_posts`: `Search public Threads posts. Results may be incomplete when anonymous search pagination is unavailable.`
- `get_post`: `Read one public Threads post from a canonical Threads post URL.`
- `get_post_replies`: `Read the public replies beneath one Threads post. This does not return replies written by a user across Threads.`

Run: `gofmt -w internal/server && go test ./internal/server -run 'TestSearchPosts|TestGetPost' -count=1`

Expected: PASS.

**Commit checkpoint, only with separate authorization:** `git add internal/server/search_tool.go internal/server/post_tools.go internal/server/server_test.go && git commit -m "feat: add post reading tools"`

## Task 7: Add profile tools

**Files:**

- Create: `internal/server/profile_tools.go`
- Modify: `internal/server/server_test.go`

**Interfaces:**

- Consumes: `Provider.GetProfile` and `Provider.GetProfilePosts`.
- Produces: typed handlers for `get_profile` and `get_profile_posts`.

- [ ] **Step 1: Define inputs**

```go
type GetProfileInput struct {
	Username string `json:"username" jsonschema:"public Threads username"`
}

type GetProfilePostsInput struct {
	Username string `json:"username" jsonschema:"public Threads username"`
	Limit    int    `json:"limit,omitempty" jsonschema:"maximum number of recent posts"`
}
```

Define explicit input schemas as `map[string]any`. Both schemas require `username` with `minLength: 1`. `get_profile_posts` defines `limit` with `minimum: 1`, `maximum: 100`, and `default: 20`. Both schemas set `type: "object"` and `additionalProperties: false`.

- [ ] **Step 2: Write profile handler tests**

Assert leading `@` normalization, default limit 20, output-schema presence, structured and text equality, recent-window completeness warnings, `ACCESS_RESTRICTED`, `NOT_FOUND`, and timeout mapping.

List the tools and assert `get_profile_replies` is absent.

Run: `go test ./internal/server -run 'TestGetProfile|TestNoProfileRepliesTool' -count=1`

Expected: FAIL because profile handlers are not registered.

- [ ] **Step 3: Implement profile handlers**

Use descriptions that say `public profile` and `recent public posts`. Do not use `history`, `all posts`, or any wording that implies completeness.

Apply the same request deadline and typed error behavior as Task 6.

Run: `gofmt -w internal/server && go test ./internal/server -count=1`

Expected: PASS with exactly six listed tools.

**Commit checkpoint, only with separate authorization:** `git add internal/server/profile_tools.go internal/server/server_test.go && git commit -m "feat: add profile reading tools"`

## Task 8: Add the stdio command and protocol integration tests

**Files:**

- Create: `cmd/threads-mcp/main.go`
- Create: `tests/stdio_test.go`

**Interfaces:**

- Consumes: `config.Load`, `threadscli.New`, `server.New`, `mcp.StdioTransport`, and build metadata.
- Produces: the `threads-mcp` executable and a child-process MCP test.

- [ ] **Step 1: Write the child-process test**

The test must:

1. Build `./cmd/threads-mcp` into `t.TempDir()`.
2. Start it with `mcp.CommandTransport{Command: exec.Command(binary)}`.
3. Connect an official `mcp.Client`.
4. List tools and assert the exact six names.
5. Call `get_server_info` and validate its structured output.
6. Close the client and assert the child exits.
7. Capture stderr separately.

Add a second process test that starts the binary with a pipe for stdout, sends no MCP input, cancels the process, and verifies every non-empty stdout line parses as a JSON object. This catches accidental startup logs on stdout.

Run: `go test ./tests -count=1 -v`

Expected: FAIL because the command does not exist.

- [ ] **Step 2: Implement the command**

`main.go` must:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

cfg, err := config.Load(version.Version)
if err != nil {
	fmt.Fprintln(os.Stderr, "threads-mcp:", err)
	os.Exit(1)
}

p, err := threadscli.New(cfg, version.Version)
if err != nil {
	fmt.Fprintln(os.Stderr, "threads-mcp:", err)
	os.Exit(1)
}

s := server.New(p, cfg, server.BuildInfo{
	Version: version.Version,
	Commit: version.Commit,
	Date: version.Date,
	SDKVersion: "v1.7.0",
	Protocol: "2026-07-28",
})
if err := s.Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, context.Canceled) {
	fmt.Fprintln(os.Stderr, "threads-mcp:", err)
	os.Exit(1)
}
```

Do not configure SDK logging to stdout.

- [ ] **Step 3: Run protocol verification**

Run: `gofmt -w cmd tests && go test ./tests -count=1 -v`

Expected: PASS.

Run: `go test -race -count=1 ./...`

Expected: PASS. The live probe remains excluded because the `live` build tag is absent.

**Commit checkpoint, only with separate authorization:** `git add cmd tests && git commit -m "feat: run threads-mcp over stdio"`

## Task 9: Add user documentation and licensing

**Files:**

- Create: `README.md`
- Create: `LICENSE`
- Create: `THIRD_PARTY_NOTICES.md`

**Interfaces:**

- Consumes: the final command path, environment variable, six tools, and release limitations.
- Produces: local build, `go install`, GitHub Release, and MCP client setup instructions.

- [ ] **Step 1: Write README verification commands first**

Add a shell verification section that runs:

```bash
go test ./...
go build -o ./bin/threads-mcp ./cmd/threads-mcp
```

Add local installation:

```bash
git clone https://github.com/granitebps/threads-mcp.git
cd threads-mcp
go build -o ./bin/threads-mcp ./cmd/threads-mcp
```

Add Go installation:

```bash
go install github.com/granitebps/threads-mcp/cmd/threads-mcp@latest
```

Add Scoop installation for the planned `granitebps/scoop-bucket` distribution:

```powershell
scoop bucket add granitebps https://github.com/granitebps/scoop-bucket
scoop install threads-mcp
```

- [ ] **Step 2: Document behavior and limits**

README must list exactly six tools and explain:

- `get_post_replies` reads the conversation beneath one post.
- `get_profile_replies` does not exist.
- Profile posts are a recent public window.
- Search and reply completeness may be unknown.
- Media URLs may expire.
- No authentication or private content is supported.
- The server uses `threads-cli`'s unofficial anonymous crawler mode, which Meta may change or block without notice.
- The server never uses account credentials, proxies, browser automation, or CAPTCHA solving.

- [ ] **Step 3: Add client configurations**

The client configuration section in `README.md` includes executable-path examples for the Codex desktop app and CLI, Claude Code, Claude Desktop, Cursor, and OpenCode. Include both the built local binary and `go run` development form where the client supports a working directory.

Use `/absolute/path/to/threads-mcp/bin/threads-mcp` in GUI examples. Do not assume GUI clients inherit `PATH`.

- [ ] **Step 4: Add Apache-2.0 licensing**

Copy the unmodified Apache License 2.0 text into `LICENSE`.

Create `THIRD_PARTY_NOTICES.md` with:

```markdown
# Third-party notices

## threads-cli

`threads-mcp` depends on `github.com/tamnd/threads-cli v0.1.1`.

- Repository: https://github.com/tamnd/threads-cli
- License: Apache License 2.0
- Copyright: retained by the upstream contributors

The upstream license is available at:
https://github.com/tamnd/threads-cli/blob/v0.1.1/LICENSE
```

- [ ] **Step 5: Verify documentation claims**

Run:

```bash
rg -n 'get_profile_replies|Googlebot|Homebrew|authentication|private|completeness' README.md
go build -o ./bin/threads-mcp ./cmd/threads-mcp
test -x ./bin/threads-mcp
```

Expected: `get_profile_replies` appears only as an explicit exclusion, the crawler access disclosure is present, Homebrew is absent, and the binary exists.

**Commit checkpoint, only with separate authorization:** `git add README.md LICENSE THIRD_PARTY_NOTICES.md && git commit -m "docs: add installation and usage guide"`

## Task 10: Add deterministic CI and release packaging

**Files:**

- Create: `.golangci.yml`
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`
- Create: `.goreleaser.yaml`

**Interfaces:**

- Consumes: Go module, command, tests, license, and documentation.
- Produces: pull-request checks and signed GitHub Release archives.

- [ ] **Step 1: Configure deterministic CI**

Create jobs with least-privilege `contents: read` permissions:

- Linux and macOS run `gofmt`, `go vet`, `go build`, and `go test -race -count=1 ./...`.
- Windows runs `go build ./...` and `go test -count=1 ./...`.
- A tidy job runs `go mod tidy` followed by `git diff --exit-code -- go.mod go.sum`.
- A vulnerability job runs `go install golang.org/x/vuln/cmd/govulncheck@v1.7.0` followed by `govulncheck ./...`.
- A lint job runs `golangci/golangci-lint-action@v8` with `version: v2.13.1`.

Use `actions/checkout@v5` and `actions/setup-go@v6` with `go-version-file: go.mod`.

- [ ] **Step 2: Configure GoReleaser**

Use GoReleaser schema version 2. Build `./cmd/threads-mcp` with `CGO_ENABLED=0`, `-trimpath`, and these link values:

```yaml
ldflags:
  - -s -w
  - -X github.com/granitebps/threads-mcp/internal/version.Version={{ .Version }}
  - -X github.com/granitebps/threads-mcp/internal/version.Commit={{ .ShortCommit }}
  - -X github.com/granitebps/threads-mcp/internal/version.Date={{ .CommitDate }}
```

Build only:

```yaml
targets:
  - linux_amd64
  - linux_arm64
  - darwin_amd64
  - darwin_arm64
  - windows_amd64
```

Archive `LICENSE`, `README.md`, and `THIRD_PARTY_NOTICES.md`. Produce `checksums.txt` with SHA-256, archive SBOMs, and a Cosign v3 bundle for the checksum file. Do not configure Homebrew, npm, OCI images, or Linux system packages.

Configure GoReleaser's Scoop publisher exactly as follows:

```yaml
scoops:
  - name: threads-mcp
    repository:
      owner: granitebps
      name: scoop-bucket
      token: '{{ envOrDefault "SCOOP_BUCKET_GITHUB_TOKEN" "" }}'
    homepage: https://github.com/granitebps/threads-mcp
    description: Read public Threads content through MCP.
    license: Apache-2.0
    skip_upload: '{{ if envOrDefault "SCOOP_BUCKET_GITHUB_TOKEN" "" }}false{{ else }}true{{ end }}'
```

This lets snapshot and unconfigured releases build archives without publishing to an external bucket.

- [ ] **Step 3: Configure release workflow**

Pull requests and pushes to `main` run `goreleaser check` with GoReleaser `v2.18.0`. Tags matching `v*` run `goreleaser release --clean` with `contents: write` and `id-token: write`. Install Cosign `v3.1.3` and Syft `v1.51.0` in the tagged release job. Pass optional `SCOOP_BUCKET_GITHUB_TOKEN` only to the tagged release job. Creating that secret or pushing a release tag requires separate publication approval.

- [ ] **Step 4: Verify CI and release files locally**

Run:

```bash
go install github.com/goreleaser/goreleaser/v2@v2.18.0
go install github.com/sigstore/cosign/v3/cmd/cosign@v3.1.3
go install github.com/anchore/syft/cmd/syft@v1.51.0
go test -race -count=1 ./...
go vet ./...
test -z "$(gofmt -l .)"
goreleaser check
goreleaser release --snapshot --clean --skip=sign
```

Expected: every command exits 0. `dist/` contains five platform archives, `checksums.txt`, SBOM files, and no Homebrew, npm, or container artifacts. Snapshot mode does not sign or upload a Scoop manifest; the tagged OIDC release job creates the Cosign bundle.

**Commit checkpoint, only with separate authorization:** `git add .github .golangci.yml .goreleaser.yaml && git commit -m "ci: verify and package releases"`

## Task 11: Add scheduled live checks and the release gate

**Files:**

- Modify: `internal/provider/threadscli/live_probe_test.go`
- Create: `.github/workflows/live-smoke.yml`
- Modify: `README.md`

**Interfaces:**

- Consumes: the Task 1 live probe and completed MCP binary.
- Produces: a scheduled signal for Threads changes and a documented release checklist.

- [ ] **Step 1: Extend the live test through MCP handlers**

After the provider-level assertions pass, construct `server.New` with the live provider and in-memory transports. Call all five data tools. Assert:

- Every call returns a structured result or a typed, safe tool error.
- No result claims `CompletenessComplete` unless the provider proves it.
- Empty search returns an `unknown` result with its warning instead of a clean complete result.
- The post-replies call accepts a zero-item public conversation without claiming completeness.

Run: `go test -tags=live ./internal/provider/threadscli -run TestAnonymousCrawlerPublicReads -count=1 -v`

Expected: PASS before the workflow is enabled.

- [ ] **Step 2: Add the scheduled workflow**

Run weekly and on manual dispatch. Grant `contents: read` only. Use `actions/checkout@v5`, `actions/setup-go@v6`, and:

```yaml
- name: Anonymous crawler access smoke test
  run: go test -tags=live ./internal/provider/threadscli -run TestAnonymousCrawlerPublicReads -count=1 -v
```

Do not store response bodies as artifacts. Do not add credentials or proxy secrets.

- [ ] **Step 3: Add the release checklist**

README's maintainer section requires:

```text
[ ] Deterministic CI passes.
[ ] Anonymous crawler-mode live smoke passes.
[ ] Official Go SDK pin is a stable release supporting MCP 2026-07-28.
[ ] threads-cli pin and Apache-2.0 notice are current.
[ ] Snapshot archives, checksums, Cosign bundles, and SBOMs verify.
[ ] Client configuration smoke checks are recorded for Codex, Claude, Cursor, and OpenCode.
```

- [ ] **Step 4: Run the complete local verification**

Run:

```bash
go mod tidy
git diff --exit-code -- go.mod go.sum
test -z "$(gofmt -l .)"
go vet ./...
go test -race -count=1 ./...
go build ./...
goreleaser check
goreleaser release --snapshot --clean --skip=sign
```

Then run the live test separately:

```bash
go test -tags=live ./internal/provider/threadscli -run TestAnonymousCrawlerPublicReads -count=1 -v
```

Expected: all commands exit 0. If the live test fails, the code may remain useful for diagnosis, but the project is not ready to release and must not add credentials, proxies, CAPTCHA solving, or another access workaround.

**Commit checkpoint, only with separate authorization:** `git add internal/provider/threadscli/live_probe_test.go .github/workflows/live-smoke.yml README.md && git commit -m "test: monitor public Threads access"`

## Coverage map

| Design requirement | Plan task |
|---|---|
| Anonymous crawler mode and hard stop | Tasks 1 and 11 |
| No credentials | Tasks 1, 2, and 4 |
| Six tools and no profile replies tool | Tasks 5, 6, and 7 |
| Typed optional data | Tasks 2 and 3 |
| Partial and unknown completeness | Tasks 2, 4, 6, 7, and 11 |
| Stable error contract | Tasks 2, 4, 5, 6, and 7 |
| MCP 2026-07-28 with stable Go SDK | Tasks 1, 5, 8, and 11 |
| Stdio hygiene | Tasks 5 and 8 |
| Deterministic and live tests | Tasks 1, 3 through 8, 10, and 11 |
| GitHub Releases and local repository use | Tasks 9 and 10 |
| No Homebrew or npm launcher | Tasks 9 and 10 |
| Apache-2.0 and third-party notice | Tasks 9 and 10 |

## Execution stop conditions

Stop and report to the user before continuing when:

- Task 1 fails in anonymous crawler mode.
- Meeting an output contract requires inventing a missing provider value.
- `threads-cli` must be forked or patched beyond a thin adapter.
- A dependency change beyond the two pinned direct modules becomes necessary.
- The task would add authentication, browser automation, proxy support, CAPTCHA solving, HTTP transport, Homebrew, npm, or registry publication.
- A commit, push, pull request, package publication, or release is requested by the workflow but has not been explicitly authorized.
