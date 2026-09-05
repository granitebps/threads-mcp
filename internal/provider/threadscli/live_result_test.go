package threadscli_test

import (
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestValidateLiveResult(t *testing.T) {
	post := map[string]any{"post": map[string]any{"id": "123"}}
	page := func(items []any, count any, completeness string) map[string]any {
		return map[string]any{"result": map[string]any{
			"items": items, "returned_count": count, "requested_limit": 1,
			"completeness": completeness, "warnings": []string{"public window"},
		}}
	}
	for _, test := range []struct {
		name, tool string
		result     *mcp.CallToolResult
		err        error
		wantError  bool
	}{
		{"post", "get_post", &mcp.CallToolResult{StructuredContent: post}, nil, false},
		{"MCP error with valid data", "get_post", &mcp.CallToolResult{IsError: true, StructuredContent: post}, nil, true},
		{"transport error", "get_post", nil, errors.New("connection lost"), true},
		{"nil result", "get_post", nil, nil, true},
		{"missing structured content", "get_post", &mcp.CallToolResult{}, nil, true},
		{"error payload", "get_post", &mcp.CallToolResult{StructuredContent: map[string]any{"error": map[string]any{"code": "TIMEOUT"}}}, nil, true},
		{"missing post ID", "get_post", &mcp.CallToolResult{StructuredContent: map[string]any{"post": map[string]any{}}}, nil, true},
		{"wrong envelope", "get_profile", &mcp.CallToolResult{StructuredContent: post}, nil, true},
		{"profile", "get_profile", &mcp.CallToolResult{StructuredContent: map[string]any{"profile": map[string]any{"id": "1", "username": "zuck"}}}, nil, false},
		{"empty profile username", "get_profile", &mcp.CallToolResult{StructuredContent: map[string]any{"profile": map[string]any{"id": "1"}}}, nil, true},
		{"empty replies", "get_post_replies", &mcp.CallToolResult{StructuredContent: page([]any{}, 0, "unknown")}, nil, false},
		{"empty search", "search_posts", &mcp.CallToolResult{StructuredContent: page([]any{}, 0, "unknown")}, nil, false},
		{"profile posts", "get_profile_posts", &mcp.CallToolResult{StructuredContent: page([]any{map[string]any{"id": "1"}}, 1, "unknown")}, nil, false},
		{"partial replies", "get_post_replies", &mcp.CallToolResult{StructuredContent: page([]any{map[string]any{"id": "1"}}, 1, "partial")}, nil, false},
		{"missing item ID", "search_posts", &mcp.CallToolResult{StructuredContent: page([]any{map[string]any{}}, 1, "unknown")}, nil, true},
		{"incorrect count", "search_posts", &mcp.CallToolResult{StructuredContent: page([]any{}, 1, "unknown")}, nil, true},
		{"missing count", "search_posts", &mcp.CallToolResult{StructuredContent: page([]any{}, nil, "unknown")}, nil, true},
		{"missing items", "search_posts", &mcp.CallToolResult{StructuredContent: page(nil, 0, "unknown")}, nil, true},
		{"unsupported completeness", "search_posts", &mcp.CallToolResult{StructuredContent: page([]any{}, 0, "complete")}, nil, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validateLiveResult(test.tool, test.result, test.err, 1)
			if (err != nil) != test.wantError {
				t.Fatalf("validation error = %v, wantError = %v", err, test.wantError)
			}
		})
	}
}
