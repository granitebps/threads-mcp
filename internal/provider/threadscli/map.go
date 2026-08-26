package threadscli

import (
	"time"

	"github.com/granitebps/threads-mcp/internal/domain"
	"github.com/tamnd/threads-cli/threads"
)

// threads-cli v0.1.1 represents absent scalar values as zero values. See
// docs/design.md. Until presence is preserved upstream, engagement and profile
// counters are omitted even when non-zero so this server never overstates what
// it can prove about field availability.
func mapPost(post threads.Post) domain.Post {
	return domain.Post{
		ID:           post.ID,
		Shortcode:    stringPtr(post.Shortcode),
		Text:         stringPtr(post.Text),
		Username:     stringPtr(post.Username),
		UserID:       stringPtr(post.UserID),
		Permalink:    stringPtr(post.Permalink),
		Timestamp:    timePtr(post.Timestamp),
		MediaType:    stringPtr(post.MediaType),
		MediaURLs:    cloneStrings(post.MediaURLs),
		IsReply:      truePtr(post.IsReply),
		IsQuotePost:  truePtr(post.IsQuotePost),
		ReplyToID:    stringPtr(post.ReplyToID),
		QuotedPostID: stringPtr(post.QuotedPostID),
	}
}

func mapReply(reply threads.Reply) domain.Reply {
	post := mapPost(threads.Post{
		ID:        reply.ID,
		Shortcode: reply.Shortcode,
		Text:      reply.Text,
		Username:  reply.Username,
		UserID:    reply.UserID,
		Permalink: reply.Permalink,
		Timestamp: reply.Timestamp,
		MediaType: reply.MediaType,
		MediaURLs: reply.MediaURLs,
	})
	return domain.Reply{
		Post:     post,
		ParentID: stringPtr(reply.ParentID),
		RootID:   stringPtr(reply.RootID),
	}
}

func mapProfile(profile threads.Profile) domain.Profile {
	return domain.Profile{
		ID:            profile.ID,
		Username:      profile.Username,
		Name:          stringPtr(profile.Name),
		Biography:     stringPtr(profile.Biography),
		ProfilePicURL: stringPtr(profile.ProfilePicURL),
		IsVerified:    truePtr(profile.IsVerified),
		URL:           profile.URL,
	}
}

func mapSearchResult(result threads.SearchResult) domain.Post {
	return domain.Post{
		ID:          result.ID,
		Text:        stringPtr(result.Text),
		Username:    stringPtr(result.Username),
		Permalink:   stringPtr(result.Permalink),
		Timestamp:   timePtr(result.Timestamp),
		MediaType:   stringPtr(result.MediaType),
		IsReply:     truePtr(result.IsReply),
		IsQuotePost: truePtr(result.IsQuotePost),
	}
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func truePtr(value bool) *bool {
	if !value {
		return nil
	}
	return &value
}

func timePtr(value time.Time) *string {
	if value.IsZero() {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}
