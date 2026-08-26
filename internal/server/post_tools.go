package server

import (
	"context"
	"strings"

	"github.com/granitebps/threads-mcp/internal/config"
	"github.com/granitebps/threads-mcp/internal/identifier"
	"github.com/granitebps/threads-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetPostInput struct {
	Post string `json:"post" jsonschema:"canonical public Threads post URL"`
}

type GetPostRepliesInput struct {
	Post  string `json:"post" jsonschema:"canonical public Threads post URL"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of replies"`
}

func addPostTools(server *mcp.Server, p provider.Provider, cfg config.Config) {
	postProperty := map[string]any{"type": "string", "minLength": 1}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_post",
		Description: "Read one public Threads post from a canonical Threads post URL.",
		Annotations: readAnnotations(),
		InputSchema: objectSchema(map[string]any{"post": postProperty}, []string{"post"}),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input GetPostInput) (*mcp.CallToolResult, PostOutput, error) {
		postURL, err := identifier.Post(strings.TrimSpace(input.Post))
		if err != nil {
			result, safe := errorResult(err)
			return result, PostOutput{Error: safe}, nil
		}
		requestCtx, cancel := context.WithTimeout(ctx, cfg.ToolTimeout)
		defer cancel()
		post, err := p.GetPost(requestCtx, postURL)
		if err != nil {
			result, safe := errorResult(err)
			return result, PostOutput{Error: safe}, nil
		}
		return nil, PostOutput{Post: &post}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_post_replies",
		Description: "Read the public replies beneath one Threads post. This does not return replies written by a user across Threads.",
		Annotations: readAnnotations(),
		InputSchema: objectSchema(map[string]any{
			"post":  postProperty,
			"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "default": 20},
		}, []string{"post"}),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input GetPostRepliesInput) (*mcp.CallToolResult, RepliesOutput, error) {
		postURL, err := identifier.Post(strings.TrimSpace(input.Post))
		if err != nil {
			result, safe := errorResult(err)
			return result, RepliesOutput{Error: safe}, nil
		}
		limit := input.Limit
		if limit == 0 {
			limit = 20
		}
		requestCtx, cancel := context.WithTimeout(ctx, cfg.ToolTimeout)
		defer cancel()
		page, err := p.GetPostReplies(requestCtx, postURL, limit)
		if err != nil {
			result, safe := errorResult(err)
			return result, RepliesOutput{Error: safe}, nil
		}
		return nil, RepliesOutput{Page: &page}, nil
	})
}
