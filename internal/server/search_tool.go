package server

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/granitebps/threads-mcp/internal/config"
	"github.com/granitebps/threads-mcp/internal/domain"
	"github.com/granitebps/threads-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchPostsInput struct {
	Query string `json:"query" jsonschema:"keyword query containing 1 to 500 Unicode code points after trimming"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of results"`
}

func addSearchTool(server *mcp.Server, p provider.Provider, cfg config.Config) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_posts",
		Description: "Search public Threads posts. Results may be incomplete because the public search page is a limited ranked window.",
		Annotations: readAnnotations(),
		InputSchema: objectSchema(map[string]any{
			"query": map[string]any{"type": "string", "minLength": 1, "maxLength": 500},
			"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 50, "default": 10},
		}, []string{"query"}),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SearchPostsInput) (*mcp.CallToolResult, SearchPostsOutput, error) {
		query := strings.TrimSpace(input.Query)
		if query == "" || utf8.RuneCountInString(query) > 500 {
			result, safe := errorResult(&domain.ProviderError{Code: domain.CodeInvalidInput, Message: "query must contain 1 to 500 Unicode code points after trimming"})
			return result, SearchPostsOutput{Error: safe}, nil
		}
		limit := input.Limit
		if limit == 0 {
			limit = 10
		}
		requestCtx, cancel := context.WithTimeout(ctx, cfg.ToolTimeout)
		defer cancel()
		page, err := p.SearchPosts(requestCtx, query, limit)
		if err != nil {
			result, safe := errorResult(err)
			return result, SearchPostsOutput{Error: safe}, nil
		}
		return nil, SearchPostsOutput{Page: &page}, nil
	})
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}
