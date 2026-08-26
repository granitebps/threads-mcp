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

## Build from a local repository

Requirements: Go 1.26 or newer.

```bash
git clone https://github.com/granitebps/threads-mcp.git
cd threads-mcp
go build -o ./bin/threads-mcp ./cmd/threads-mcp
```

Configure your MCP client with the absolute path to `bin/threads-mcp`.

For local development, clients that support a working directory can run:

```bash
go run ./cmd/threads-mcp
```

## Other installation options

Go:

```bash
go install github.com/granitebps/threads-mcp/cmd/threads-mcp@latest
```

GitHub Release archives are planned for Linux and macOS on `amd64` and `arm64`, and Windows on `amd64`.

Windows Scoop distribution is planned:

```powershell
scoop bucket add granitebps https://github.com/granitebps/scoop-bucket
scoop install threads-mcp
```

## Client configuration

Build the local executable first:

```bash
go build -o ./bin/threads-mcp ./cmd/threads-mcp
```

Replace `/absolute/path/to/threads-mcp/bin/threads-mcp` below with the real absolute path. Graphical applications may not inherit your shell `PATH`, so use an absolute path.

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

The live smoke test is separate because it depends on Threads:

```bash
go test -tags=live ./internal/provider/threadscli -run TestAnonymousCrawlerPublicReads -count=1 -v
```

## Maintainer release checklist

```text
[ ] Deterministic CI passes.
[ ] Anonymous crawler-mode live smoke passes.
[ ] Official Go SDK pin is a stable release supporting MCP 2026-07-28.
[ ] threads-cli pin and Apache-2.0 notice are current.
[ ] Snapshot archives, checksums, Cosign bundles, and SBOMs verify.
[ ] Client configuration smoke checks are recorded for Codex, Claude, Cursor, and OpenCode.
```

Publishing, signing, or pushing a release is not performed automatically from a local checkout.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
