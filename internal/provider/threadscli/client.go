package threadscli

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"strings"

	"github.com/granitebps/threads-mcp/internal/config"
	"github.com/granitebps/threads-mcp/internal/domain"
	"github.com/granitebps/threads-mcp/internal/identifier"
	"github.com/granitebps/threads-mcp/internal/provider"
	"github.com/tamnd/threads-cli/threads"
)

const (
	partialWarning       = "upstream stopped before the requested limit; returned partial public results"
	windowWarning        = "Threads exposes a limited public window; completeness is unknown"
	emptySearch          = "Threads returned no public search results; completeness is unknown"
	semanticReadAttempts = 5
)

type source interface {
	Profile(context.Context, string) (*threads.Profile, error)
	ProfilePosts(context.Context, string, int) iter.Seq2[threads.Post, error]
	Post(context.Context, string) (*threads.Post, error)
	PostReplies(context.Context, string, int) iter.Seq2[threads.Reply, error]
	Mode() string
	UserAgent() string
}

type searchSource interface {
	Search(context.Context, string, int) iter.Seq2[threads.SearchResult, error]
}

type Client struct {
	source   source
	searcher searchSource
	version  string
}

func New(cfg config.Config, version string) (*Client, error) {
	upstream, err := threads.NewClient(threads.Config{
		Delay:     cfg.Delay,
		Retries:   cfg.Retries,
		Timeout:   cfg.HTTPTimeout,
		UserAgent: threads.CrawlerUA,
		Proxy:     "",
		Lang:      "en-US",
		CacheDir:  cfg.CacheDir,
		CacheTTL:  cfg.CacheTTL,
		NoCache:   true, // An incomplete HTTP 200 response must not poison semantic retries.
		Token:     "",
		Session:   "",
		CSRF:      "",
	})
	if err != nil {
		return nil, err
	}
	searchTransport := http.DefaultTransport.(*http.Transport).Clone()
	searchTransport.Proxy = nil
	searchHTTP := &http.Client{Timeout: cfg.HTTPTimeout, Transport: searchTransport}
	return &Client{
		source:   upstream,
		searcher: newSearchPageSource(searchHTTP, threads.WebBase, threads.CrawlerUA, cfg.Delay),
		version:  version,
	}, nil
}

func (c *Client) SearchPosts(ctx context.Context, query string, limit int) (domain.Page[domain.Post], error) {
	query = strings.TrimSpace(query)
	if query == "" || limit < 1 {
		return domain.Page[domain.Post]{}, invalidInput("query must be non-empty and limit must be positive")
	}
	return retryPage(ctx, false, func() (domain.Page[domain.Post], error) {
		return collectPage(c.searcher.Search(ctx, query, limit), limit, mapSearchResult, true)
	})
}

func (c *Client) GetPost(ctx context.Context, input string) (domain.Post, error) {
	postURL, err := identifier.Post(input)
	if err != nil {
		return domain.Post{}, err
	}
	post, err := retryRead(ctx, func() (*threads.Post, error) {
		return c.source.Post(ctx, postURL)
	}, func(post *threads.Post, err error) bool {
		return isIncompleteCrawlerError(err) || err == nil && (post == nil || post.ID == "")
	})
	if err != nil {
		return domain.Post{}, classify(err)
	}
	if post == nil || post.ID == "" {
		return domain.Post{}, &domain.ProviderError{Code: domain.CodeNotFound, Message: "public Threads post was not found"}
	}
	return mapPost(*post), nil
}

func (c *Client) GetPostReplies(ctx context.Context, input string, limit int) (domain.Page[domain.Reply], error) {
	postURL, err := identifier.Post(input)
	if err != nil {
		return domain.Page[domain.Reply]{}, err
	}
	if limit < 1 {
		return domain.Page[domain.Reply]{}, invalidInput("limit must be positive")
	}
	return retryPage(ctx, false, func() (domain.Page[domain.Reply], error) {
		return collectPage(c.source.PostReplies(ctx, postURL, limit), limit, mapReply, false)
	})
}

func (c *Client) GetProfile(ctx context.Context, input string) (domain.Profile, error) {
	username, err := identifier.Username(input)
	if err != nil {
		return domain.Profile{}, err
	}
	profileValue, err := retryRead(ctx, func() (*threads.Profile, error) {
		return c.source.Profile(ctx, username)
	}, func(profileValue *threads.Profile, err error) bool {
		return isIncompleteCrawlerError(err) || err == nil && (profileValue == nil || profileValue.ID == "")
	})
	if err != nil {
		return domain.Profile{}, classify(err)
	}
	if profileValue == nil || profileValue.ID == "" {
		return domain.Profile{}, &domain.ProviderError{Code: domain.CodeNotFound, Message: "public Threads profile was not found"}
	}
	return mapProfile(*profileValue), nil
}

func (c *Client) GetProfilePosts(ctx context.Context, input string, limit int) (domain.Page[domain.Post], error) {
	username, err := identifier.Username(input)
	if err != nil {
		return domain.Page[domain.Post]{}, err
	}
	if limit < 1 {
		return domain.Page[domain.Post]{}, invalidInput("limit must be positive")
	}
	return retryPage(ctx, true, func() (domain.Page[domain.Post], error) {
		return collectPage(c.source.ProfilePosts(ctx, username, limit), limit, mapPost, false)
	})
}

func (c *Client) Info() provider.Info {
	mode := c.source.Mode()
	if mode == "anonymous" {
		mode = "anonymous-crawler"
	}
	return provider.Info{
		Name:            "threads-cli + public Threads search page",
		Version:         c.version,
		Mode:            mode,
		UserAgent:       c.source.UserAgent(),
		SupportedInputs: []string{"keyword query", "Threads username or profile URL", "canonical Threads post URL"},
		KnownLimitations: []string{
			"search and list results are limited public windows",
			"Threads may change or restrict its unofficial crawler-facing pages without notice",
			"missing scalar values are omitted instead of inferred",
		},
	}
}

func collectPage[Source, Target any](sequence iter.Seq2[Source, error], limit int, convert func(Source) Target, search bool) (domain.Page[Target], error) {
	page := domain.Page[Target]{
		Items:          make([]Target, 0),
		RequestedLimit: limit,
		Completeness:   domain.CompletenessUnknown,
		Warnings:       make([]string, 0),
	}
	for item, err := range sequence {
		if err != nil {
			if len(page.Items) == 0 {
				return domain.Page[Target]{}, classify(err)
			}
			page.Completeness = domain.CompletenessPartial
			page.Warnings = append(page.Warnings, partialWarning)
			page.ReturnedCount = len(page.Items)
			return page, nil
		}
		page.Items = append(page.Items, convert(item))
		if len(page.Items) >= limit {
			break
		}
	}
	page.ReturnedCount = len(page.Items)
	if search && len(page.Items) == 0 {
		page.Warnings = append(page.Warnings, emptySearch)
	} else {
		page.Warnings = append(page.Warnings, windowWarning)
	}
	return page, nil
}

func retryRead[T any](ctx context.Context, read func() (T, error), incomplete func(T, error) bool) (T, error) {
	var zero T
	for attempt := 1; attempt <= semanticReadAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return zero, err
		}
		value, err := read()
		if !incomplete(value, err) {
			return value, err
		}
		if attempt == semanticReadAttempts {
			return zero, incompleteCrawlerResponse(err)
		}
	}
	return zero, incompleteCrawlerResponse(nil)
}

func retryPage[T any](ctx context.Context, retryEmpty bool, read func() (domain.Page[T], error)) (domain.Page[T], error) {
	for attempt := 1; attempt <= semanticReadAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return domain.Page[T]{}, classify(err)
		}
		page, err := read()
		incomplete := isIncompleteCrawlerError(err)
		empty := retryEmpty && err == nil && page.ReturnedCount == 0
		if !incomplete && !empty {
			return page, err
		}
		if attempt == semanticReadAttempts {
			if empty {
				return page, nil
			}
			return domain.Page[T]{}, incompleteCrawlerResponse(err)
		}
	}
	return domain.Page[T]{}, incompleteCrawlerResponse(nil)
}

func isIncompleteCrawlerError(err error) bool {
	if errors.Is(err, errSearchPageChanged) {
		return true
	}
	var codeErr *threads.CodeError
	if !errors.As(err, &codeErr) || codeErr.Code != threads.ExitNotFound {
		return false
	}
	return strings.HasPrefix(codeErr.Msg, "not found: profile @") ||
		strings.HasPrefix(codeErr.Msg, "not found: post ")
}

func incompleteCrawlerResponse(cause error) *domain.ProviderError {
	return &domain.ProviderError{
		Code:      domain.CodeUpstreamChanged,
		Message:   "Threads returned an incomplete public page",
		Retryable: true,
		Cause:     cause,
	}
}

func classify(err error) *domain.ProviderError {
	if providerErr := (*domain.ProviderError)(nil); errors.As(err, &providerErr) {
		return providerErr
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &domain.ProviderError{Code: domain.CodeTimeout, Message: "Threads request timed out", Retryable: true, Cause: err}
	}
	if errors.Is(err, context.Canceled) {
		return &domain.ProviderError{Code: domain.CodeTimeout, Message: "request cancelled", Retryable: true, Cause: err}
	}
	if errors.Is(err, errSearchPageChanged) {
		return &domain.ProviderError{Code: domain.CodeUpstreamChanged, Message: "Threads public search page changed", Cause: err}
	}

	providerErr := &domain.ProviderError{Cause: err}
	switch threads.Code(err) {
	case threads.ExitUsage:
		providerErr.Code, providerErr.Message = domain.CodeInvalidInput, "Threads rejected the request input"
	case threads.ExitNotFound:
		providerErr.Code, providerErr.Message = domain.CodeNotFound, "public Threads content was not found"
	case threads.ExitLoginWall:
		providerErr.Code, providerErr.Message = domain.CodeAccessRestricted, "Threads restricted anonymous access to this content"
	case threads.ExitRateLimit:
		providerErr.Code, providerErr.Message, providerErr.Retryable = domain.CodeRateLimited, "Threads rate limited the request", true
	case threads.ExitNetwork:
		providerErr.Code, providerErr.Message, providerErr.Retryable = domain.CodeNetworkFailure, "Threads request failed", true
	default:
		providerErr.Code, providerErr.Message = domain.CodeUpstreamChanged, "Threads returned an unexpected response"
	}
	return providerErr
}

func invalidInput(message string) *domain.ProviderError {
	return &domain.ProviderError{Code: domain.CodeInvalidInput, Message: message}
}
