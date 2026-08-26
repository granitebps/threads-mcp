package server

import (
	"errors"

	"github.com/granitebps/threads-mcp/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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

func errorResult(err error) (*mcp.CallToolResult, *ErrorOutput) {
	safe := &ErrorOutput{Code: domain.CodeInternalError, Message: "internal server error"}
	var providerErr *domain.ProviderError
	if errors.As(err, &providerErr) {
		safe.Code = providerErr.Code
		safe.Message = providerErr.Message
		safe.Retryable = providerErr.Retryable
		safe.RetryAfterSeconds = providerErr.RetryAfterSeconds
	}
	return &mcp.CallToolResult{IsError: true}, safe
}
