package tests

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestStdioHandshakeAndTools(t *testing.T) {
	binary := buildBinary(t)
	var stderr bytes.Buffer
	command := exec.Command(binary)
	command.Stderr = &stderr
	client := mcp.NewClient(&mcp.Implementation{Name: "stdio-test", Version: "test"}, nil)
	session, err := client.Connect(context.Background(), &mcp.CommandTransport{Command: command, TerminateDuration: time.Second}, nil)
	if err != nil {
		t.Fatalf("connect: %v stderr=%s", err, stderr.String())
	}
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range listed.Tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	want := []string{"get_post", "get_post_replies", "get_profile", "get_profile_posts", "get_server_info", "search_posts"}
	if !slices.Equal(names, want) {
		t.Fatalf("tools=%v want=%v", names, want)
	}
	info, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "get_server_info", Arguments: map[string]any{}})
	if err != nil || info.IsError || info.StructuredContent == nil {
		t.Fatalf("info=%+v err=%v", info, err)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("close: %v stderr=%s", err, stderr.String())
	}
}

func TestStartupWritesNoNonProtocolOutputToStdout(t *testing.T) {
	binary := buildBinary(t)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	_, _ = output.ReadFrom(stdout)
	_ = command.Wait()

	scanner := bufio.NewScanner(bytes.NewReader(output.Bytes()))
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var object map[string]any
		if json.Unmarshal(line, &object) != nil {
			t.Fatalf("non-protocol stdout: %q", line)
		}
	}
}

func TestExecutableName(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{goos: "windows", want: "threads-mcp.exe"},
		{goos: "linux", want: "threads-mcp"},
		{goos: "darwin", want: "threads-mcp"},
	}
	for _, test := range tests {
		t.Run(test.goos, func(t *testing.T) {
			if got := executableName(test.goos); got != test.want {
				t.Fatalf("executableName(%q) = %q, want %q", test.goos, got, test.want)
			}
		})
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), executableName(runtime.GOOS))
	command := exec.Command("go", "build", "-o", binary, "./cmd/threads-mcp")
	command.Dir = ".."
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	return binary
}

func executableName(goos string) string {
	if goos == "windows" {
		return "threads-mcp.exe"
	}
	return "threads-mcp"
}
