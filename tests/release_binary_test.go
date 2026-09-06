package tests

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"slices"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPackagedBinaryHandshakeAndTools(t *testing.T) {
	binary := os.Getenv("THREADS_MCP_TEST_BINARY")
	if binary == "" {
		t.Skip("THREADS_MCP_TEST_BINARY is not set")
	}
	wantVersion := os.Getenv("THREADS_MCP_TEST_VERSION")
	if wantVersion == "" {
		t.Fatal("THREADS_MCP_TEST_VERSION must be set with THREADS_MCP_TEST_BINARY")
	}

	var stderr bytes.Buffer
	command := exec.Command(binary)
	command.Stderr = &stderr
	client := mcp.NewClient(&mcp.Implementation{Name: "package-test", Version: "test"}, nil)
	session, err := client.Connect(context.Background(), &mcp.CommandTransport{
		Command:           command,
		TerminateDuration: time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("connect: %v stderr=%s", err, stderr.String())
	}
	t.Cleanup(func() { _ = session.Close() })

	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range listed.Tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	wantTools := []string{
		"get_post",
		"get_post_replies",
		"get_profile",
		"get_profile_posts",
		"get_server_info",
		"search_posts",
	}
	if !slices.Equal(names, wantTools) {
		t.Fatalf("tools=%v want=%v", names, wantTools)
	}

	info, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_server_info",
		Arguments: map[string]any{},
	})
	if err != nil || info == nil || info.IsError || info.StructuredContent == nil {
		t.Fatalf("info=%+v err=%v", info, err)
	}
	metadata := info.StructuredContent.(map[string]any)
	if metadata["server_version"] != wantVersion || metadata["provider_version"] != "v0.1.1" {
		t.Fatalf("incorrect packaged versions: %+v", metadata)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("close: %v stderr=%s", err, stderr.String())
	}
}
