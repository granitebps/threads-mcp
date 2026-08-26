package domain

type Completeness string

const (
	CompletenessComplete Completeness = "complete"
	CompletenessPartial  Completeness = "partial"
	CompletenessUnknown  Completeness = "unknown"
)

type Post struct {
	ID           string   `json:"id"`
	Shortcode    *string  `json:"shortcode,omitempty"`
	Text         *string  `json:"text,omitempty"`
	Username     *string  `json:"username,omitempty"`
	UserID       *string  `json:"user_id,omitempty"`
	Permalink    *string  `json:"permalink,omitempty"`
	Timestamp    *string  `json:"timestamp,omitempty"`
	MediaType    *string  `json:"media_type,omitempty"`
	MediaURLs    []string `json:"media_urls,omitempty"`
	LikeCount    *int64   `json:"like_count,omitempty"`
	ReplyCount   *int64   `json:"reply_count,omitempty"`
	RepostCount  *int64   `json:"repost_count,omitempty"`
	QuoteCount   *int64   `json:"quote_count,omitempty"`
	IsReply      *bool    `json:"is_reply,omitempty"`
	IsQuotePost  *bool    `json:"is_quote_post,omitempty"`
	ReplyToID    *string  `json:"reply_to_id,omitempty"`
	QuotedPostID *string  `json:"quoted_post_id,omitempty"`
}

type Reply struct {
	Post
	ParentID *string `json:"parent_id,omitempty"`
	RootID   *string `json:"root_id,omitempty"`
}

type Profile struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	Name           *string `json:"name,omitempty"`
	Biography      *string `json:"biography,omitempty"`
	ProfilePicURL  *string `json:"profile_pic_url,omitempty"`
	IsVerified     *bool   `json:"is_verified,omitempty"`
	FollowerCount  *int64  `json:"follower_count,omitempty"`
	FollowingCount *int64  `json:"following_count,omitempty"`
	URL            string  `json:"url"`
}

type Page[T any] struct {
	Items          []T          `json:"items"`
	ReturnedCount  int          `json:"returned_count"`
	RequestedLimit int          `json:"requested_limit"`
	Completeness   Completeness `json:"completeness"`
	Warnings       []string     `json:"warnings"`
}
