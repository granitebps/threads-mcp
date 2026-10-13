package threadscli

import (
	"errors"
	"os"
	"testing"
)

func TestParsePostPageReadsMediaLayout(t *testing.T) {
	page, err := os.ReadFile("testdata/post_page.html")
	if err != nil {
		t.Fatal(err)
	}
	root, replies, err := parsePostPage(page)
	if err != nil {
		t.Fatal(err)
	}
	if root.ID != "100" || root.Text != "root post" || root.Username != "alice" || root.UserID != "1" || root.Permalink != "https://www.threads.com/@alice/post/Root1" || root.Timestamp.IsZero() || root.IsReply {
		t.Fatalf("root = %+v", root)
	}
	want := []struct{ id, parent, user, text string }{
		{"101", "100", "alice", "self thread"},
		{"201", "100", "bob", "first reply"},
		{"202", "201", "carol", "nested reply"},
		{"203", "100", "dave", "second reply"},
	}
	if len(replies) != len(want) {
		t.Fatalf("got %d replies: %+v", len(replies), replies)
	}
	for i, w := range want {
		got := replies[i]
		if got.ID != w.id || got.ParentID != w.parent || got.RootID != "100" || got.Username != w.user || got.Text != w.text {
			t.Fatalf("reply %d = %+v, want %+v", i, got, w)
		}
	}
	if replies[1].MediaType != "IMAGE" || len(replies[1].MediaURLs) != 1 || replies[1].MediaURLs[0] != "https://cdn.example/a.jpg" {
		t.Fatalf("reply media = %+v", replies[1])
	}
}

func TestParsePostPageRejectsUnknownLayout(t *testing.T) {
	_, _, err := parsePostPage([]byte(`<script type="application/json">{"unrelated":{"thread_items":[{"post":{"pk":"1","like_count":0}}]}}</script>`))
	if !errors.Is(err, errPostPageLayout) {
		t.Fatalf("err = %v", err)
	}
}
