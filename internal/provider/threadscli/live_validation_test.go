package threadscli_test

import (
	"encoding/json"
	"fmt"

	"github.com/granitebps/threads-mcp/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// validateLiveResult is shared by the live probe and its offline regression tests.
func validateLiveResult(name string, result *mcp.CallToolResult, callErr error, limit int) error {
	if callErr != nil || result == nil || result.IsError || result.StructuredContent == nil {
		return fmt.Errorf("tool %s returned no usable response", name)
	}
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		return fmt.Errorf("tool %s returned invalid structured content", name)
	}
	var envelope map[string]json.RawMessage
	if json.Unmarshal(data, &envelope) != nil || envelope == nil {
		return fmt.Errorf("tool %s returned an invalid envelope", name)
	}
	if payload, ok := envelope["error"]; ok && string(payload) != "null" {
		return fmt.Errorf("tool %s returned an error payload", name)
	}
	switch name {
	case "get_post":
		var post domain.Post
		if json.Unmarshal(envelope["post"], &post) != nil || post.ID == "" {
			return fmt.Errorf("tool %s returned no post ID", name)
		}
	case "get_profile":
		var profile domain.Profile
		if json.Unmarshal(envelope["profile"], &profile) != nil || profile.ID == "" || profile.Username == "" {
			return fmt.Errorf("tool %s returned no profile ID or username", name)
		}
	case "search_posts", "get_post_replies", "get_profile_posts":
		var page struct {
			Items          []domain.Post       `json:"items"`
			ReturnedCount  *int                `json:"returned_count"`
			RequestedLimit *int                `json:"requested_limit"`
			Completeness   domain.Completeness `json:"completeness"`
			Warnings       []string            `json:"warnings"`
		}
		if json.Unmarshal(envelope["result"], &page) != nil || page.Items == nil || page.ReturnedCount == nil || page.RequestedLimit == nil || page.Warnings == nil {
			return fmt.Errorf("tool %s returned an invalid page", name)
		}
		if limit <= 0 || *page.RequestedLimit != limit || *page.ReturnedCount != len(page.Items) || len(page.Items) > limit {
			return fmt.Errorf("tool %s returned inconsistent counts or limits", name)
		}
		if page.Completeness != domain.CompletenessUnknown && page.Completeness != domain.CompletenessPartial {
			return fmt.Errorf("tool %s returned unsupported completeness", name)
		}
		for _, post := range page.Items {
			if post.ID == "" {
				return fmt.Errorf("tool %s returned an item without a post ID", name)
			}
		}
	default:
		return fmt.Errorf("no live validation defined for tool %s", name)
	}
	return nil
}
