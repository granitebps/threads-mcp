# threads-mcp

`threads-mcp` is a local, read-only MCP server for public Threads content. It provides keyword search, posts, replies beneath a post, profiles, and recent profile posts without a Threads account, Meta app, or API key.

## Important access notice

This project is unofficial and is not affiliated with Meta.

Profile and conversation reads use `threads-cli`'s anonymous crawler mode. Keyword search reads structured post data embedded in Threads' public search page. Both are crawler-facing web surfaces that Meta may change, restrict, or block without notice. This project does not establish that a particular use complies with Meta's terms or applicable law; users are responsible for how they collect and use public data.

The server never accepts Threads account credentials and does not use login cookies, private content, proxies, browser automation, CAPTCHA solving, or proxy rotation.

## Tools

| Tool | Purpose |
|---|---|
| `search_posts` | Search the public Threads search window by keyword. |
| `get_post` | Read one public post from its canonical URL. |
| `get_post_replies` | Read the visible public conversation beneath one post. |
| `get_profile` | Read one public profile. |
| `get_profile_posts` | Read the recent public posts exposed for one profile. |
| `get_server_info` | Report versions, access mode, supported inputs, and known limits. |

`get_profile_replies` intentionally does not exist. The project needs replies beneath a selected post, not a feed of replies written by one user.

Search, replies, and profile posts are public windows, not complete archives. Results report `complete`, `partial`, or `unknown`; most list results are `unknown` because Threads does not prove that the visible window is complete. Media URLs remain owned by Threads and may expire.

The [v1 compatibility promise](docs/v1-compatibility.md) defines the stable tool
names, inputs, output envelopes, field types, and structured error meanings that
start with v1.0.0.

## Build from a local repository

Requirements: Go 1.26.6 or newer.

```bash
git clone https://github.com/granitebps/threads-mcp.git
cd threads-mcp
go build -o ./bin/threads-mcp ./cmd/threads-mcp
```

Configure your MCP client with the absolute path to `bin/threads-mcp`.

On Windows PowerShell, build the `.exe` form instead:

```powershell
go build -o .\bin\threads-mcp.exe .\cmd\threads-mcp
```

Use the full path to `bin\threads-mcp.exe` as the MCP command. Do not omit
`.exe` when configuring a Windows client.

For local development, clients that support a working directory can run:

```bash
go run ./cmd/threads-mcp
```

## Install the command

After v1.0.0 is published and the repository is public, install that exact
version with Go:

```bash
go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.0
```

This command is expected to fail before the tag and public repository exist. Do
not use it as a pre-release test. After publication, run it from outside this
source checkout so the test cannot select local source.

Go writes the executable to `GOBIN`. If `GOBIN` is empty, it uses the `bin`
directory under the first `GOPATH` entry. Find the installed command on macOS or
Linux with:

```bash
go_bin=$(go env GOBIN)
if [ -z "$go_bin" ]; then go_bin="$(go env GOPATH)/bin"; fi
printf '%s\n' "$go_bin/threads-mcp"
```

On Windows PowerShell:

```powershell
$goBin = go env GOBIN
if (-not $goBin) { $goBin = Join-Path (go env GOPATH) "bin" }
Join-Path $goBin "threads-mcp.exe"
```

Use that absolute path as the MCP command. Installing with `@latest` is also
supported after publication, but client configurations should pin a tested
version when reproducibility matters.

GitHub Release archives are planned for Linux and macOS on `amd64` and `arm64`,
and Windows on `amd64`.

## Download a release archive

After v1.0.0 is published, open the
[v1.0.0 release](https://github.com/granitebps/threads-mcp/releases/tag/v1.0.0)
and download the archive for your computer:

| Operating system | CPU | Archive |
|---|---|---|
| macOS | Apple silicon (`arm64`) | `threads-mcp_1.0.0_darwin_arm64.tar.gz` |
| macOS | Intel (`amd64`) | `threads-mcp_1.0.0_darwin_amd64.tar.gz` |
| Linux | `arm64` | `threads-mcp_1.0.0_linux_arm64.tar.gz` |
| Linux | `amd64` | `threads-mcp_1.0.0_linux_amd64.tar.gz` |
| Windows | `amd64` | `threads-mcp_1.0.0_windows_amd64.zip` |

The archives also contain the license, README, and third-party notices. The
release includes a software bill of materials for each archive.

### Verify the download

Download your archive, `checksums.txt`, and `checksums.txt.bundle` from the same
release into one directory. Install
[Cosign](https://docs.sigstore.dev/cosign/system_config/installation/) v3.1.3 or
newer, then authenticate the checksum file against the exact GitHub Actions
workflow and release tag:

```bash
release_tag=v1.0.0
cosign verify-blob \
  --bundle checksums.txt.bundle \
  --certificate-identity "https://github.com/granitebps/threads-mcp/.github/workflows/release.yml@refs/tags/$release_tag" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  checksums.txt
```

On macOS, verify the selected archive against that authenticated checksum file:

```bash
archive=threads-mcp_1.0.0_darwin_arm64.tar.gz
expected=$(awk -v name="$archive" '$2 == name { print $1 }' checksums.txt)
actual=$(shasum -a 256 "$archive" | awk '{ print $1 }')
test -n "$expected" && test "$actual" = "$expected"
```

Change `archive` to the Intel filename on an Intel Mac. On Linux, use the same
commands with the correct Linux archive and replace `shasum -a 256` with
`sha256sum`.

On Windows PowerShell:

```powershell
$releaseTag = "v1.0.0"
$archive = "threads-mcp_1.0.0_windows_amd64.zip"
cosign verify-blob --bundle checksums.txt.bundle --certificate-identity "https://github.com/granitebps/threads-mcp/.github/workflows/release.yml@refs/tags/$releaseTag" --certificate-oidc-issuer "https://token.actions.githubusercontent.com" checksums.txt
$expected = Get-Content checksums.txt | Where-Object { $_ -match "  $([regex]::Escape($archive))$" } | ForEach-Object { ($_ -split '\s+')[0] }
if (-not $expected) { throw "Archive is missing from checksums.txt" }
$actual = (Get-FileHash -Algorithm SHA256 $archive).Hash.ToLowerInvariant()
if ($actual -ne $expected.ToLowerInvariant()) { throw "Checksum mismatch" }
```

Do not use the archive if either verification step fails. Cosign authenticates
the checksum file and its release-workflow identity. It does not Apple-notarize
the macOS executable or Authenticode-sign the Windows executable.

Extract the verified archive:

```bash
mkdir threads-mcp_1.0.0
tar -xzf "$archive" -C threads-mcp_1.0.0
```

On Windows PowerShell:

```powershell
Expand-Archive -Path $archive -DestinationPath .\threads-mcp_1.0.0
```

Configure your MCP client with the absolute path to the extracted `threads-mcp`
or `threads-mcp.exe`. Starting the executable in a terminal appears to wait for
input because it speaks MCP over standard input and output; use an MCP client to
test it.

Windows Scoop distribution is deferred and is not available for v1.0.0. Use
`go install`, a GitHub Release `.zip` after publication, or a local build.

## Client configuration

Choose either the local-repository executable or an installed executable. To use
the local repository, build it first:

```bash
go build -o ./bin/threads-mcp ./cmd/threads-mcp
```

Replace `/absolute/path/to/threads-mcp/bin/threads-mcp` below with the real
absolute path. Graphical applications may not inherit your shell `PATH`, so use
an absolute path. On Windows, use the `.exe` file. For example, configuration
files can use `C:/absolute/path/to/threads-mcp/bin/threads-mcp.exe`.

For a Go-installed command, replace the example path with the absolute location
reported in [Install the command](#install-the-command). Do not rely on a short
command name in a graphical client because it may use a different `PATH`.

### Codex

Add the local server from the command line:

```bash
codex mcp add threads -- /absolute/path/to/threads-mcp/bin/threads-mcp
```

You can configure the same server in `~/.codex/config.toml`, or in `.codex/config.toml` inside a trusted project:

```toml
[mcp_servers.threads]
command = "/absolute/path/to/threads-mcp/bin/threads-mcp"
```

In the Codex desktop app:

1. Open **Settings**, then select **MCP servers**.
2. Select **Add server**.
3. Enter `threads` as the name, choose **STDIO**, and enter the absolute executable path as the command.
4. Save the server, then select **Restart**.

On Windows, the equivalent Codex CLI command is:

```powershell
codex mcp add threads -- C:\absolute\path\to\threads-mcp\bin\threads-mcp.exe
```

The Codex desktop app, CLI, and IDE extension share MCP configuration on the same computer. Use `/mcp` to confirm that `threads` is connected. See the [official OpenAI MCP documentation](https://learn.chatgpt.com/docs/extend/mcp).

### Claude

Add the local server to Claude Code:

```bash
claude mcp add --transport stdio threads -- /absolute/path/to/threads-mcp/bin/threads-mcp
```

Claude Code uses local scope by default. Use its project or user scope flag if you want a different scope.

Claude Desktop reads MCP servers from `claude_desktop_config.json`. Add the following entry, then restart the app:

```json
{
  "mcpServers": {
    "threads": {
      "command": "/absolute/path/to/threads-mcp/bin/threads-mcp",
      "args": []
    }
  }
}
```

### OpenCode 2

Add the local server under `mcp.servers` in `opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "servers": {
      "threads": {
        "type": "local",
        "command": [
          "/absolute/path/to/threads-mcp/bin/threads-mcp"
        ]
      }
    }
  }
}
```

To run directly from a local repository, set the working directory:

```json
{
  "mcp": {
    "servers": {
      "threads-dev": {
        "type": "local",
        "command": ["go", "run", "./cmd/threads-mcp"],
        "cwd": "/absolute/path/to/threads-mcp"
      }
    }
  }
}
```

### Cursor

Add the server to `.cursor/mcp.json` in a project, or to `~/.cursor/mcp.json` for global use:

```json
{
  "mcpServers": {
    "threads": {
      "command": "/absolute/path/to/threads-mcp/bin/threads-mcp",
      "args": []
    }
  }
}
```

Clients that support environment variables can set `THREADS_MCP_TIMEOUT=60s`. Accepted values range from `5s` through `120s`.

## Configuration

The server uses MCP over stdio. It writes protocol messages only to stdout and diagnostics to stderr.

`THREADS_MCP_TIMEOUT` changes the per-tool deadline. It accepts durations from `5s` through `120s`; the default is `45s`.

No authentication, session, CSRF, proxy, or API-key setting is supported.

## Verify a checkout

```bash
go test ./...
go build -o ./bin/threads-mcp ./cmd/threads-mcp
```

On Windows PowerShell:

```powershell
go test ./...
go build -o .\bin\threads-mcp.exe .\cmd\threads-mcp
```

The live smoke test is separate because it depends on Threads:

```bash
go test -tags=live ./internal/provider/threadscli -run TestAnonymousCrawlerPublicReads -count=1 -v
```

## Maintainer releases

Use the [release guide](RELEASE.md) for preparation, approval, publishing, and
verification instructions for humans and AI agents. The
[v1.0.0 checklist](docs/release-v1.0.0.md) tracks outstanding work and evidence.

The owner approves each release before its version tag is pushed. A `v*` tag
push starts the release workflow, which requires CI and release-configuration
checks to succeed for the tagged commit before publishing. The workflow changes
still need verification on GitHub before launch. Normal code pushes do not
publish a release.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
