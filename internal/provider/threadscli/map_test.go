package threadscli

import (
	"testing"
	"time"

	"github.com/tamnd/threads-cli/threads"
)

func TestMapPostOmitsAmbiguousZeroValues(t *testing.T) {
	got := mapPost(threads.Post{ID: "1"})
	if got.LikeCount != nil || got.ReplyCount != nil || got.RepostCount != nil || got.QuoteCount != nil || got.IsReply != nil || got.IsQuotePost != nil {
		t.Fatalf("ambiguous zero values were exposed: %+v", got)
	}
}

func TestMapPostPreservesKnownValues(t *testing.T) {
	ts := time.Date(2026, 8, 26, 12, 30, 0, 0, time.FixedZone("test", 7*60*60))
	upstream := threads.Post{
		ID: "1", Shortcode: "ABC123", Text: "hello", Username: "zuck", UserID: "2",
		Permalink: "https://www.threads.com/@zuck/post/ABC123", Timestamp: ts,
		MediaType: "IMAGE", MediaURLs: []string{"https://cdn.example/image.jpg"},
		IsReply: true, IsQuotePost: true, ReplyToID: "3", QuotedPostID: "4",
	}
	got := mapPost(upstream)
	if got.Timestamp == nil || *got.Timestamp != "2026-08-26T05:30:00Z" {
		t.Fatalf("timestamp = %v", got.Timestamp)
	}
	if got.Text == nil || *got.Text != "hello" || got.IsReply == nil || !*got.IsReply || got.IsQuotePost == nil || !*got.IsQuotePost {
		t.Fatalf("known values were lost: %+v", got)
	}
	upstream.MediaURLs[0] = "changed"
	if got.MediaURLs[0] != "https://cdn.example/image.jpg" {
		t.Fatal("media URL slice aliases upstream storage")
	}
}

func TestMapReplyPreservesThreadLinks(t *testing.T) {
	got := mapReply(threads.Reply{ID: "1", ParentID: "2", RootID: "3", Text: "reply"})
	if got.ParentID == nil || *got.ParentID != "2" || got.RootID == nil || *got.RootID != "3" || got.Text == nil || *got.Text != "reply" {
		t.Fatalf("reply = %+v", got)
	}
}

func TestMapProfileOmitsAmbiguousCountsAndFalse(t *testing.T) {
	got := mapProfile(threads.Profile{ID: "1", Username: "zuck", URL: "https://www.threads.com/@zuck"})
	if got.FollowerCount != nil || got.FollowingCount != nil || got.IsVerified != nil {
		t.Fatalf("ambiguous profile values were exposed: %+v", got)
	}
}

func TestMapSearchResultHasNoEngagementCounters(t *testing.T) {
	got := mapSearchResult(threads.SearchResult{ID: "1", Text: "result", Username: "zuck"})
	if got.Text == nil || *got.Text != "result" || got.LikeCount != nil || got.ReplyCount != nil || got.RepostCount != nil || got.QuoteCount != nil {
		t.Fatalf("search result = %+v", got)
	}
}
