package provider

import (
	"context"

	"github.com/granitebps/threads-mcp/internal/domain"
)

type Info struct {
	Name             string   `json:"name"`
	Version          string   `json:"version"`
	Mode             string   `json:"mode"`
	UserAgent        string   `json:"user_agent"`
	SupportedInputs  []string `json:"supported_inputs"`
	KnownLimitations []string `json:"known_limitations"`
}

type Provider interface {
	SearchPosts(context.Context, string, int) (domain.Page[domain.Post], error)
	GetPost(context.Context, string) (domain.Post, error)
	GetPostReplies(context.Context, string, int) (domain.Page[domain.Reply], error)
	GetProfile(context.Context, string) (domain.Profile, error)
	GetProfilePosts(context.Context, string, int) (domain.Page[domain.Post], error)
	Info() Info
}
