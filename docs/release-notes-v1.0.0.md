<!--
Draft release body. Do not publish it until the v1.0.0 checklist identifies an
approved, fully verified candidate commit.
-->

# threads-mcp v1.0.0

This is the first public release of threads-mcp, a local, read-only MCP server
for public Threads content. It works without a Threads account, Meta app, or API
key.

## Included tools

- `search_posts` searches the public Threads search window by keyword.
- `get_post` reads one public post from its canonical URL.
- `get_post_replies` reads the visible public conversation beneath a post.
- `get_profile` reads one public profile.
- `get_profile_posts` reads the recent public posts exposed for a profile.
- `get_server_info` reports server, provider, protocol, and access details.

All tools are read-only and use MCP over stdio. Data tools return structured
results and structured errors. Search, replies, and profile-post results also
report their completeness status and warnings.

`get_profile_replies` is not included. Use `get_post_replies` when you need the
replies beneath a selected post.

## Install

Install the exact version with Go 1.26.6 or newer:

```bash
go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.0
```

Release archives are available for:

- macOS on Apple silicon and Intel;
- Linux on `arm64` and `amd64`;
- Windows on `amd64`.

You can also clone the repository and build the executable locally. Configure
your MCP client with the absolute path to `threads-mcp`, or `threads-mcp.exe` on
Windows. See the
[client setup guide](https://github.com/granitebps/threads-mcp#client-configuration)
for Codex, Claude, Cursor, and OpenCode examples.

Homebrew and Scoop packages are not available in v1.0.0.

## Verify a download

Each release archive has a SHA-256 entry in `checksums.txt` and a software bill
of materials. The release also includes `checksums.txt.bundle`, which contains
the keyless Cosign signature material for the checksum file.

Follow the
[download verification instructions](https://github.com/granitebps/threads-mcp#verify-the-download)
before running a downloaded executable. The binaries are not Apple-notarized or
Windows Authenticode-signed, so operating-system warnings are possible even
after the checksum and Cosign verification succeed.

## Access limits

This project is unofficial and is not affiliated with Meta. It reads public,
anonymous Threads web pages that Meta may change, limit, or block without
notice.

- Private or login-only content is not supported.
- Search, replies, and profile posts are limited public windows, not complete
  archives. There is no historical pagination in v1.0.0.
- An empty result does not prove that matching content does not exist.
- Threads sometimes returns an incomplete page with HTTP 200. The server retries
  affected reads with fresh requests, then returns a retryable
  `UPSTREAM_CHANGED` error if every attempt is incomplete.
- Threads may omit fields, and its media URLs may expire.

Read the
[access notice](https://github.com/granitebps/threads-mcp#important-access-notice)
before using the server.

## Compatibility and support

The [v1 compatibility promise](https://github.com/granitebps/threads-mcp/blob/v1.0.0/docs/v1-compatibility.md)
covers the six tool names, accepted inputs, output envelopes, field types, and
structured error meanings throughout v1.x.

This is the first stable release, so there is no earlier stable version to
migrate from. Users of a development checkout can install v1.0.0 or build the
tag, then keep the same absolute executable-path configuration in their MCP
client.

Use the [issue tracker](https://github.com/granitebps/threads-mcp/issues) for
ordinary bugs. Report suspected vulnerabilities through the
[private security route](https://github.com/granitebps/threads-mcp/security/advisories/new)
without opening a public issue.
