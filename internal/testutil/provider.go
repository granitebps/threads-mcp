package testutil

import (
	"context"

	"github.com/granitebps/threads-mcp/internal/domain"
	"github.com/granitebps/threads-mcp/internal/provider"
)

type Provider struct {
	Fail                func(string)
	SearchPostsFunc     func(context.Context, string, int) (domain.Page[domain.Post], error)
	GetPostFunc         func(context.Context, string) (domain.Post, error)
	GetPostRepliesFunc  func(context.Context, string, int) (domain.Page[domain.Reply], error)
	GetProfileFunc      func(context.Context, string) (domain.Profile, error)
	GetProfilePostsFunc func(context.Context, string, int) (domain.Page[domain.Post], error)
	InfoValue           provider.Info
}

func (p *Provider) SearchPosts(ctx context.Context, query string, limit int) (domain.Page[domain.Post], error) {
	if p.SearchPostsFunc == nil {
		p.unexpected("SearchPosts")
		return domain.Page[domain.Post]{}, nil
	}
	return p.SearchPostsFunc(ctx, query, limit)
}

func (p *Provider) GetPost(ctx context.Context, post string) (domain.Post, error) {
	if p.GetPostFunc == nil {
		p.unexpected("GetPost")
		return domain.Post{}, nil
	}
	return p.GetPostFunc(ctx, post)
}

func (p *Provider) GetPostReplies(ctx context.Context, post string, limit int) (domain.Page[domain.Reply], error) {
	if p.GetPostRepliesFunc == nil {
		p.unexpected("GetPostReplies")
		return domain.Page[domain.Reply]{}, nil
	}
	return p.GetPostRepliesFunc(ctx, post, limit)
}

func (p *Provider) GetProfile(ctx context.Context, username string) (domain.Profile, error) {
	if p.GetProfileFunc == nil {
		p.unexpected("GetProfile")
		return domain.Profile{}, nil
	}
	return p.GetProfileFunc(ctx, username)
}

func (p *Provider) GetProfilePosts(ctx context.Context, username string, limit int) (domain.Page[domain.Post], error) {
	if p.GetProfilePostsFunc == nil {
		p.unexpected("GetProfilePosts")
		return domain.Page[domain.Post]{}, nil
	}
	return p.GetProfilePostsFunc(ctx, username, limit)
}

func (p *Provider) Info() provider.Info { return p.InfoValue }

func (p *Provider) unexpected(method string) {
	if p.Fail != nil {
		p.Fail("unexpected provider call: " + method)
	}
}
