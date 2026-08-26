package server

import (
	"github.com/granitebps/threads-mcp/internal/config"
	"github.com/granitebps/threads-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type BuildInfo struct {
	Version    string
	Commit     string
	Date       string
	SDKVersion string
	Protocol   string
}

var expectedToolNames = []string{"search_posts", "get_post", "get_post_replies", "get_profile", "get_profile_posts", "get_server_info"}

func New(p provider.Provider, cfg config.Config, build BuildInfo) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "threads-mcp", Version: build.Version}, nil)
	addSearchTool(server, p, cfg)
	addPostTools(server, p, cfg)
	addProfileTools(server, p, cfg)
	addServerInfoTool(server, p, build)
	return server
}

func readAnnotations() *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		ReadOnlyHint:    true,
		IdempotentHint:  true,
		DestructiveHint: boolPtr(false),
		OpenWorldHint:   boolPtr(true),
	}
}

func boolPtr(value bool) *bool { return &value }
