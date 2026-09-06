package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	cosignInstaller = "sigstore/cosign-installer@6f9f17788090df1f26f669e9d70d6ae9567deba6 # v4.1.2"
	syftInstaller   = "anchore/sbom-action/download-syft@3ad7283483fc7af8ff2b4ea19663c2d5ca935e26 # v0.24.2"
)

func TestReleaseWorkflowUsesCompatiblePinnedToolchain(t *testing.T) {
	workflow := readRepositoryFile(t, ".github/workflows/release.yml")
	for _, required := range []string{
		cosignInstaller,
		"cosign-release: v3.1.3",
		syftInstaller,
		"syft-version: v1.51.0",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("release workflow is missing %q", required)
		}
	}
}

func TestReleaseWorkflowPreflightsPublishingToolchain(t *testing.T) {
	workflow := readRepositoryFile(t, ".github/workflows/release.yml")
	checkJob := textBetween(t, workflow, "\n  check:\n", "\n  release:\n")
	for _, required := range []string{
		cosignInstaller,
		"cosign version",
		syftInstaller,
		"syft version",
		"cosign generate-key-pair",
		"cosign signing-config create",
		"cosign sign-blob",
		"--signing-config",
		"cosign verify-blob",
		"--insecure-ignore-tlog",
		"release --snapshot --clean --skip=sign,scoop",
	} {
		if !strings.Contains(checkJob, required) {
			t.Errorf("release preflight is missing %q", required)
		}
	}
	if strings.Contains(checkJob, "id-token: write") || strings.Contains(checkJob, "contents: write") {
		t.Error("release preflight must remain read-only")
	}
}

func TestReleaseWorkflowUsesVersionSpecificReleaseNotes(t *testing.T) {
	workflow := readRepositoryFile(t, ".github/workflows/release.yml")
	releaseJob := textBetween(t, workflow, "\n  release:\n", "")
	for _, required := range []string{
		"contents: write",
		"id-token: write",
		"test -f \"docs/release-notes-${GITHUB_REF_NAME}.md\"",
		"--release-notes=docs/release-notes-${{ github.ref_name }}.md",
	} {
		if !strings.Contains(releaseJob, required) {
			t.Errorf("release job is missing %q", required)
		}
	}
}

func TestV101RecoveryReleaseNotesExist(t *testing.T) {
	notes := readRepositoryFile(t, "docs/release-notes-v1.0.1.md")
	for _, required := range []string{"# threads-mcp v1.0.1", "v1.0.0", "Cosign"} {
		if !strings.Contains(notes, required) {
			t.Errorf("v1.0.1 release notes are missing %q", required)
		}
	}
}

func TestReadmeHasOneCodexServerHeader(t *testing.T) {
	readme := readRepositoryFile(t, "README.md")
	if count := strings.Count(readme, "[mcp_servers.threads]"); count != 1 {
		t.Fatalf("README has %d Codex server headers, want 1", count)
	}
}

func TestTextBetweenHandlesWindowsNewlines(t *testing.T) {
	text := "jobs:\r\n  check:\r\n    runs-on: ubuntu-latest\r\n  release:\r\n"
	got := textBetween(t, text, "\n  check:\n", "\n  release:\n")
	if !strings.Contains(got, "runs-on: ubuntu-latest") {
		t.Fatalf("check job = %q", got)
	}
}

func readRepositoryFile(t *testing.T, name string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("..", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func textBetween(t *testing.T, text, start, end string) string {
	t.Helper()
	text = strings.ReplaceAll(text, "\r\n", "\n")
	startIndex := strings.Index(text, start)
	if startIndex < 0 {
		t.Fatalf("missing section start %q", start)
	}
	text = text[startIndex+len(start):]
	if end == "" {
		return text
	}
	endIndex := strings.Index(text, end)
	if endIndex < 0 {
		t.Fatalf("missing section end %q", end)
	}
	return text[:endIndex]
}
