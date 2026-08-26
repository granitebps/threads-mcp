package identifier

import (
	"net/url"
	"strings"

	"github.com/granitebps/threads-mcp/internal/domain"
	"github.com/tamnd/threads-cli/pkg/thid"
)

func Username(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if strings.Contains(trimmed, "://") && !isThreadsURL(trimmed) {
		return "", invalid("username must be a Threads handle or profile URL")
	}
	identity := thid.Classify(trimmed)
	if identity.Kind != thid.KindProfile || identity.Handle == "" {
		return "", invalid("username must be a Threads handle or profile URL")
	}
	return identity.Handle, nil
}

func Post(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if !isThreadsURL(trimmed) {
		return "", invalid("post must be a public Threads post URL")
	}
	identity := thid.Classify(trimmed)
	if identity.Kind != thid.KindPost || identity.URL == "" {
		return "", invalid("post must be a public Threads post URL")
	}
	return identity.URL, nil
}

func isThreadsURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil || u.Scheme != "https" {
		return false
	}
	switch strings.ToLower(u.Hostname()) {
	case "threads.com", "www.threads.com", "threads.net", "www.threads.net":
		return true
	default:
		return false
	}
}

func invalid(message string) *domain.ProviderError {
	return &domain.ProviderError{Code: domain.CodeInvalidInput, Message: message}
}
