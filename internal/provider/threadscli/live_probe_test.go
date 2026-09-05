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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	appConfig, err := config.Load("live")
	if err != nil {
		t.Fatal(err)
	}
	appConfig.Delay = time.Second
	appConfig.CacheDir = t.TempDir()
	providerClient, err := provideradapter.New(appConfig, "v0.1.1")
	if err != nil {
		t.Fatal(err)
	}
	info := providerClient.Info()
	if info.Mode != "anonymous-crawler" {
		t.Fatalf("mode = %q, want anonymous-crawler", info.Mode)
	}
	if info.UserAgent != threads.CrawlerUA {
		t.Fatalf("user agent = %q", info.UserAgent)
	}
	profilePosts, err := providerClient.GetProfilePosts(ctx, "zuck", 1)
	if err != nil || profilePosts.ReturnedCount == 0 || profilePosts.Items[0].Permalink == nil || *profilePosts.Items[0].Permalink == "" {
		t.Fatalf("profile posts returned no fetchable public post: page=%+v err=%v", profilePosts, err)
	}
	postURL := *profilePosts.Items[0].Permalink
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
		{Name: "get_post", Arguments: map[string]any{"post": postURL}},
		{Name: "get_post_replies", Arguments: map[string]any{"post": postURL, "limit": 1}},
		{Name: "get_profile", Arguments: map[string]any{"username": "zuck"}},
		{Name: "get_profile_posts", Arguments: map[string]any{"username": "zuck", "limit": 1}},
	}
	for _, call := range calls {
		result, err := clientSession.CallTool(ctx, &call)
		if validationErr := validateLiveResult(call.Name, result, err, 1); validationErr != nil {
			t.Fatal(validationErr)
		}
		t.Logf("tool %s: successful response and valid payload", call.Name)
	}
}
