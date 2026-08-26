package server

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/granitebps/threads-mcp/internal/config"
	"github.com/granitebps/threads-mcp/internal/domain"
	"github.com/granitebps/threads-mcp/internal/provider"
	"github.com/granitebps/threads-mcp/internal/testutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var expectedTools = []string{"get_post", "get_post_replies", "get_profile", "get_profile_posts", "get_server_info", "search_posts"}

func TestToolList(t *testing.T) {
	p := fakeProvider(t)
	cs := connect(t, p)
	listed, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range listed.Tools {
		names = append(names, tool.Name)
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint || !tool.Annotations.IdempotentHint {
			t.Fatalf("tool %s is not marked read-only and idempotent", tool.Name)
		}
		if tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint {
			t.Fatalf("tool %s destructiveHint", tool.Name)
		}
		if tool.Annotations.OpenWorldHint == nil || !*tool.Annotations.OpenWorldHint {
			t.Fatalf("tool %s openWorldHint", tool.Name)
		}
		if tool.OutputSchema == nil {
			t.Fatalf("tool %s has no output schema", tool.Name)
		}
	}
	slices.Sort(names)
	if !slices.Equal(names, expectedTools) {
		t.Fatalf("tools=%v want=%v", names, expectedTools)
	}
}

func TestServerInfo(t *testing.T) {
	cs := connect(t, fakeProvider(t))
	result := callTool(t, cs, "get_server_info", map[string]any{})
	assertTextMatchesStructured(t, result)
	value := result.StructuredContent.(map[string]any)
	if value["sdk_version"] != "v1.7.0" || value["protocol_version"] != "2026-07-28" || value["provider_mode"] != "anonymous-crawler" {
		t.Fatalf("server info=%+v", value)
	}
	excluded := value["excluded_tools"].([]any)
	if len(excluded) != 1 || excluded[0] != "get_profile_replies" {
		t.Fatalf("excluded=%v", excluded)
	}
}

func TestSearchPostsDefaultsAndNormalizes(t *testing.T) {
	p := fakeProvider(t)
	p.SearchPostsFunc = func(_ context.Context, query string, limit int) (domain.Page[domain.Post], error) {
		if query != "launch" || limit != 10 {
			t.Fatalf("query=%q limit=%d", query, limit)
		}
		return domain.Page[domain.Post]{Items: []domain.Post{{ID: "1"}}, ReturnedCount: 1, RequestedLimit: limit, Completeness: domain.CompletenessUnknown, Warnings: []string{}}, nil
	}
	result := callTool(t, connect(t, p), "search_posts", map[string]any{"query": "  launch  "})
	if result.IsError || result.StructuredContent == nil {
		t.Fatalf("result=%+v", result)
	}
	assertTextMatchesStructured(t, result)
}

func TestSearchPostsReturnsSafeTimeout(t *testing.T) {
	p := fakeProvider(t)
	p.SearchPostsFunc = func(context.Context, string, int) (domain.Page[domain.Post], error) {
		return domain.Page[domain.Post]{}, &domain.ProviderError{Code: domain.CodeTimeout, Message: "request timed out", Retryable: true, Cause: errors.New("secret cause")}
	}
	result := callTool(t, connect(t, p), "search_posts", map[string]any{"query": "launch"})
	if !result.IsError {
		t.Fatal("expected tool error")
	}
	encoded, _ := json.Marshal(result.StructuredContent)
	if strings.Contains(string(encoded), "secret cause") || !strings.Contains(string(encoded), domain.CodeTimeout) {
		t.Fatalf("structured error=%s", encoded)
	}
}

func TestGetPostAndReplies(t *testing.T) {
	p := fakeProvider(t)
	p.GetPostFunc = func(_ context.Context, post string) (domain.Post, error) {
		if post != "https://www.threads.com/@zuck/post/ABC123" {
			t.Fatalf("post=%q", post)
		}
		return domain.Post{ID: "1"}, nil
	}
	p.GetPostRepliesFunc = func(_ context.Context, post string, limit int) (domain.Page[domain.Reply], error) {
		if limit != 20 {
			t.Fatalf("limit=%d", limit)
		}
		return domain.Page[domain.Reply]{Items: []domain.Reply{{Post: domain.Post{ID: "2"}}}, ReturnedCount: 1, RequestedLimit: limit, Completeness: domain.CompletenessPartial, Warnings: []string{"interrupted"}}, nil
	}
	cs := connect(t, p)
	post := callTool(t, cs, "get_post", map[string]any{"post": " https://www.threads.com/@zuck/post/ABC123 "})
	replies := callTool(t, cs, "get_post_replies", map[string]any{"post": "https://www.threads.com/@zuck/post/ABC123"})
	if post.IsError || replies.IsError {
		t.Fatalf("post=%+v replies=%+v", post, replies)
	}
	structured := replies.StructuredContent.(map[string]any)
	page := structured["result"].(map[string]any)
	if page["completeness"] != "partial" {
		t.Fatalf("replies=%+v", structured)
	}
}

func TestGetProfileTools(t *testing.T) {
	p := fakeProvider(t)
	p.GetProfileFunc = func(_ context.Context, username string) (domain.Profile, error) {
		if username != "zuck" {
			t.Fatalf("username=%q", username)
		}
		return domain.Profile{ID: "1", Username: username}, nil
	}
	p.GetProfilePostsFunc = func(_ context.Context, username string, limit int) (domain.Page[domain.Post], error) {
		if username != "zuck" || limit != 20 {
			t.Fatalf("username=%q limit=%d", username, limit)
		}
		return domain.Page[domain.Post]{Items: []domain.Post{}, RequestedLimit: limit, Completeness: domain.CompletenessUnknown, Warnings: []string{"recent public window"}}, nil
	}
	cs := connect(t, p)
	profile := callTool(t, cs, "get_profile", map[string]any{"username": " @zuck "})
	posts := callTool(t, cs, "get_profile_posts", map[string]any{"username": "@zuck"})
	if profile.IsError || posts.IsError {
		t.Fatalf("profile=%+v posts=%+v", profile, posts)
	}
}

func TestNoProfileRepliesTool(t *testing.T) {
	cs := connect(t, fakeProvider(t))
	listed, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range listed.Tools {
		if tool.Name == "get_profile_replies" {
			t.Fatal("get_profile_replies must not be registered")
		}
	}
}

func connect(t *testing.T, p provider.Provider) *mcp.ClientSession {
	t.Helper()
	cfg, err := config.Load("test")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ToolTimeout = 200 * time.Millisecond
	s := New(p, cfg, BuildInfo{Version: "test", Commit: "abc", Date: "today", SDKVersion: "v1.7.0", Protocol: "2026-07-28"})
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := s.Connect(context.Background(), serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "threads-mcp-test", Version: "test"}, nil)
	clientSession, err := client.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = clientSession.Close()
		_ = serverSession.Close()
	})
	return clientSession
}

func callTool(t *testing.T, cs *mcp.ClientSession, name string, arguments map[string]any) *mcp.CallToolResult {
	t.Helper()
	result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: arguments})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertTextMatchesStructured(t *testing.T, result *mcp.CallToolResult) {
	t.Helper()
	if len(result.Content) != 1 {
		t.Fatalf("content=%+v", result.Content)
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content type=%T", result.Content[0])
	}
	var decoded any
	if err := json.Unmarshal([]byte(text.Text), &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, result.StructuredContent) {
		t.Fatalf("text=%+v structured=%+v", decoded, result.StructuredContent)
	}
}

func fakeProvider(t *testing.T) *testutil.Provider {
	t.Helper()
	return &testutil.Provider{
		Fail: func(message string) { t.Error(message) },
		InfoValue: provider.Info{
			Name: "hybrid", Version: "v0.1.1", Mode: "anonymous-crawler", UserAgent: "crawler",
			SupportedInputs: []string{"query", "username", "post URL"}, KnownLimitations: []string{"public window"},
		},
	}
}
