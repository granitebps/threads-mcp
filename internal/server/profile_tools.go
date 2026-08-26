package server

import (
	"context"
	"strings"

	"github.com/granitebps/threads-mcp/internal/config"
	"github.com/granitebps/threads-mcp/internal/identifier"
	"github.com/granitebps/threads-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetProfileInput struct {
	Username string `json:"username" jsonschema:"public Threads username"`
}

type GetProfilePostsInput struct {
	Username string `json:"username" jsonschema:"public Threads username"`
	Limit    int    `json:"limit,omitempty" jsonschema:"maximum number of recent posts"`
}

func addProfileTools(server *mcp.Server, p provider.Provider, cfg config.Config) {
	usernameProperty := map[string]any{"type": "string", "minLength": 1}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_profile",
		Description: "Read one public Threads profile.",
		Annotations: readAnnotations(),
		InputSchema: objectSchema(map[string]any{"username": usernameProperty}, []string{"username"}),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input GetProfileInput) (*mcp.CallToolResult, ProfileOutput, error) {
		username, err := identifier.Username(strings.TrimSpace(input.Username))
		if err != nil {
			result, safe := errorResult(err)
			return result, ProfileOutput{Error: safe}, nil
		}
		requestCtx, cancel := context.WithTimeout(ctx, cfg.ToolTimeout)
		defer cancel()
		profileValue, err := p.GetProfile(requestCtx, username)
		if err != nil {
			result, safe := errorResult(err)
			return result, ProfileOutput{Error: safe}, nil
		}
		return nil, ProfileOutput{Profile: &profileValue}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_profile_posts",
		Description: "Read the recent public posts exposed for one Threads profile.",
		Annotations: readAnnotations(),
		InputSchema: objectSchema(map[string]any{
			"username": usernameProperty,
			"limit":    map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "default": 20},
		}, []string{"username"}),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input GetProfilePostsInput) (*mcp.CallToolResult, ProfilePostsOutput, error) {
		username, err := identifier.Username(strings.TrimSpace(input.Username))
		if err != nil {
			result, safe := errorResult(err)
			return result, ProfilePostsOutput{Error: safe}, nil
		}
		limit := input.Limit
		if limit == 0 {
			limit = 20
		}
		requestCtx, cancel := context.WithTimeout(ctx, cfg.ToolTimeout)
		defer cancel()
		page, err := p.GetProfilePosts(requestCtx, username, limit)
		if err != nil {
			result, safe := errorResult(err)
			return result, ProfilePostsOutput{Error: safe}, nil
		}
		return nil, ProfilePostsOutput{Page: &page}, nil
	})
}
