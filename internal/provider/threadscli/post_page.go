package threadscli

import (
	"context"
	"encoding/json"
	"errors"
	htmlpkg "html"
	"iter"
	"time"

	"github.com/tamnd/threads-cli/threads"
)

// errPostPageLayout reports that a post page did not carry the data.media
// layout. The caller then falls back to the threads-cli thread_items parser.
var errPostPageLayout = errors.New("public Threads post page did not contain a recognized data.media payload")

// postPageSource reads post pages that Threads now renders as data.media with
// text_post_app_info.self_thread and direct_replies. threads-cli v0.1.1 only
// recognizes thread_items, so it reports these pages as not found.
type postPageSource struct {
	*threads.Client
}

func (s *postPageSource) Post(ctx context.Context, postURL string) (*threads.Post, error) {
	page, err := s.GetRaw(ctx, postURL)
	if err != nil {
		return nil, err
	}
	root, _, err := parsePostPage(page)
	if errors.Is(err, errPostPageLayout) {
		return s.Client.Post(ctx, postURL)
	}
	return &root, err
}

func (s *postPageSource) PostReplies(ctx context.Context, postURL string, limit int) iter.Seq2[threads.Reply, error] {
	return func(yield func(threads.Reply, error) bool) {
		page, err := s.GetRaw(ctx, postURL)
		if err != nil {
			yield(threads.Reply{}, err)
			return
		}
		_, replies, err := parsePostPage(page)
		if errors.Is(err, errPostPageLayout) {
			for reply, err := range s.Client.PostReplies(ctx, postURL, limit) {
				if !yield(reply, err) {
					return
				}
			}
			return
		}
		for i, reply := range replies {
			if limit > 0 && i >= limit {
				return
			}
			if !yield(reply, nil) {
				return
			}
		}
	}
}

// parsePostPage returns the root post and the replies in page order: the
// author's self-thread continuation first, then each direct reply chain.
// Threads splits the root media across several Relay fragments that share an
// id: one carries the post fields, another carries self_thread and
// direct_replies.
func parsePostPage(page []byte) (threads.Post, []threads.Reply, error) {
	var fragments []map[string]any
	for _, match := range searchScriptPattern.FindAllSubmatch(page, -1) {
		var data any
		raw := htmlpkg.UnescapeString(string(match[1]))
		if json.Unmarshal([]byte(raw), &data) != nil {
			continue
		}
		collectPostPageMedia(data, 0, &fragments)
	}

	var media map[string]any
	for _, fragment := range fragments {
		if valueString(fragment["pk"]) != "" && valueString(fragment["code"]) != "" {
			media = fragment
			break
		}
	}
	if media == nil {
		return threads.Post{}, nil, errPostPageLayout
	}
	var info map[string]any
	for _, fragment := range fragments {
		if valueString(fragment["id"]) != valueString(media["id"]) {
			continue
		}
		candidate, _ := fragment["text_post_app_info"].(map[string]any)
		if candidate["direct_replies"] != nil || candidate["self_thread"] != nil {
			info = candidate
			break
		}
	}

	root := postFromPagePost(media)
	seen := map[string]bool{root.ID: true}
	var replies []threads.Reply
	appendChain := func(connection any) {
		parentID := root.ID
		for _, node := range connectionNodes(connection) {
			post := postFromPagePost(node)
			if post.ID == "" || seen[post.ID] {
				continue
			}
			seen[post.ID] = true
			replies = append(replies, replyFromPagePost(post, parentID, root.ID))
			parentID = post.ID
		}
	}
	appendChain(info["self_thread"])
	directReplies, _ := info["direct_replies"].(map[string]any)
	edges, _ := directReplies["edges"].([]any)
	for _, edgeValue := range edges {
		edge, _ := edgeValue.(map[string]any)
		appendChain(edge["node"])
	}
	return root, replies, nil
}

// collectPostPageMedia appends every {"data":{"media":{...}}} fragment in
// document order.
func collectPostPageMedia(data any, depth int, fragments *[]map[string]any) {
	if depth > 30 {
		return
	}
	switch value := data.(type) {
	case map[string]any:
		if payload, ok := value["data"].(map[string]any); ok {
			if media, ok := payload["media"].(map[string]any); ok {
				*fragments = append(*fragments, media)
			}
		}
		for _, child := range value {
			collectPostPageMedia(child, depth+1, fragments)
		}
	case []any:
		for _, child := range value {
			collectPostPageMedia(child, depth+1, fragments)
		}
	}
}

// connectionNodes reads {"posts":{"edges":[{"node":post}]}}.
func connectionNodes(connection any) []map[string]any {
	value, _ := connection.(map[string]any)
	posts, _ := value["posts"].(map[string]any)
	edges, _ := posts["edges"].([]any)
	nodes := make([]map[string]any, 0, len(edges))
	for _, edgeValue := range edges {
		edge, _ := edgeValue.(map[string]any)
		if node, ok := edge["node"].(map[string]any); ok {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func postFromPagePost(post map[string]any) threads.Post {
	result := threads.Post{ID: valueString(post["pk"]), Shortcode: valueString(post["code"]), FetchedAt: time.Now().UTC()}
	if caption, ok := post["caption"].(map[string]any); ok {
		result.Text = valueString(caption["text"])
	}
	if user, ok := post["user"].(map[string]any); ok {
		result.Username = valueString(user["username"])
		result.UserID = valueString(user["pk"])
		if result.UserID == "" {
			result.UserID = valueString(user["id"])
		}
	}
	if timestamp := valueFloat(post["taken_at"]); timestamp > 0 {
		result.Timestamp = time.Unix(int64(timestamp), 0).UTC()
	}
	switch int(valueFloat(post["media_type"])) {
	case 1:
		result.MediaType = "IMAGE"
		result.MediaURLs = nonEmpty(firstImageURL(post))
	case 2:
		result.MediaType = "VIDEO"
		result.MediaURLs = nonEmpty(firstVideoURL(post))
	case 8:
		result.MediaType = "CAROUSEL_ALBUM"
		carousel, _ := post["carousel_media"].([]any)
		for _, itemValue := range carousel {
			item, _ := itemValue.(map[string]any)
			if url := firstVideoURL(item); url != "" {
				result.MediaURLs = append(result.MediaURLs, url)
			} else if url := firstImageURL(item); url != "" {
				result.MediaURLs = append(result.MediaURLs, url)
			}
		}
	default:
		result.MediaType = "TEXT_POST"
	}
	result.HasMedia = result.MediaType != "TEXT_POST"
	if info, ok := post["text_post_app_info"].(map[string]any); ok {
		result.IsQuotePost, _ = info["is_quote_post"].(bool)
		if author, ok := info["reply_to_author"].(map[string]any); ok {
			result.IsReply = true
			result.ReplyToID = valueString(author["pk"])
			if result.ReplyToID == "" {
				result.ReplyToID = valueString(author["id"])
			}
		}
	}
	if result.Shortcode != "" && result.Username != "" {
		result.Permalink = threads.WebBase + "/@" + result.Username + "/post/" + result.Shortcode
	}
	return result
}

func replyFromPagePost(post threads.Post, parentID, rootID string) threads.Reply {
	return threads.Reply{
		ID:        post.ID,
		ParentID:  parentID,
		RootID:    rootID,
		Shortcode: post.Shortcode,
		Text:      post.Text,
		Username:  post.Username,
		UserID:    post.UserID,
		Permalink: post.Permalink,
		Timestamp: post.Timestamp,
		MediaType: post.MediaType,
		MediaURLs: post.MediaURLs,
		FetchedAt: post.FetchedAt,
	}
}

func firstImageURL(post map[string]any) string {
	versions, _ := post["image_versions2"].(map[string]any)
	candidates, _ := versions["candidates"].([]any)
	if len(candidates) == 0 {
		return ""
	}
	candidate, _ := candidates[0].(map[string]any)
	return valueString(candidate["url"])
}

func firstVideoURL(post map[string]any) string {
	versions, _ := post["video_versions"].([]any)
	if len(versions) == 0 {
		return ""
	}
	version, _ := versions[0].(map[string]any)
	return valueString(version["url"])
}

func nonEmpty(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}
