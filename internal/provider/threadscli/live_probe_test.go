//go:build live

package threadscli_test

import (
	"context"
	"testing"
	"time"

	"github.com/granitebps/threads-mcp/internal/config"
	provideradapter "github.com/granitebps/threads-mcp/internal/provider/threadscli"
	serverpackage "github.com/granitebps/threads-mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

	appConfig, err := config.Load("live")
	if err != nil {
		t.Fatal(err)
	}
	appConfig.Delay = time.Second
	providerClient, err := provideradapter.New(appConfig, "v0.1.1")
	if err != nil {
		t.Fatal(err)
	}
	searchPage, err := providerClient.SearchPosts(ctx, "threads", 1)
	if err != nil || searchPage.ReturnedCount == 0 {
		t.Fatalf("public search page read failed: page=%+v err=%v", searchPage, err)
	}

	mcpServer := serverpackage.New(providerClient, appConfig, serverpackage.BuildInfo{
		Version: "live", SDKVersion: "v1.7.0", Protocol: "2026-07-28",
	})
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	clientSession, err := mcp.NewClient(&mcp.Implementation{Name: "live-test", Version: "live"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()

	calls := []mcp.CallToolParams{
		{Name: "search_posts", Arguments: map[string]any{"query": "threads", "limit": 1}},
		{Name: "get_post", Arguments: map[string]any{"post": post.Permalink}},
		{Name: "get_post_replies", Arguments: map[string]any{"post": post.Permalink, "limit": 1}},
		{Name: "get_profile", Arguments: map[string]any{"username": "zuck"}},
		{Name: "get_profile_posts", Arguments: map[string]any{"username": "zuck", "limit": 1}},
	}
	for _, call := range calls {
		result, err := clientSession.CallTool(ctx, &call)
		if err != nil || result.StructuredContent == nil {
			t.Fatalf("tool %s: result=%+v err=%v", call.Name, result, err)
		}
		if containsComplete(result.StructuredContent) {
			t.Fatalf("tool %s incorrectly claimed complete results", call.Name)
		}
	}

}

func containsComplete(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if key == "completeness" && child == "complete" {
				return true
			}
			if containsComplete(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsComplete(child) {
				return true
			}
		}
	}
	return false
}
