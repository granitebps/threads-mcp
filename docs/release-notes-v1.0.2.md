# threads-mcp v1.0.2

This patch release restores `get_post` and `get_post_replies`. It works without
a Threads account, Meta app, or API key.

## Fixes

- Threads changed its public post pages around 2026-10-07. The post and its
  replies now arrive as separate fragments, and threads-cli v0.1.1 found no
  posts in them. `get_post` and `get_post_replies` failed with a retryable
  `UPSTREAM_CHANGED` error on every post. Both tools now read the new layout and
  fall back to the previous parser when a page still uses the old one.
- The minimum Go version is now 1.26.9. Go 1.26.6 is affected by
  [GO-2026-6617](https://pkg.go.dev/vuln/GO-2026-6617) in `net/http`.

There are no MCP tool, input, output, or error-code changes. `get_post_replies`
returns the author's own thread continuation first, then each visible reply
chain, with `parent_id` pointing at the post each reply answers.

## Included tools

- `search_posts` searches the public Threads search window by keyword.
- `get_post` reads one public post from its canonical URL.
- `get_post_replies` reads the visible public conversation beneath a post.
- `get_profile` reads one public profile.
- `get_profile_posts` reads the recent public posts exposed for a profile.
- `get_server_info` reports server, provider, protocol, and access details.

`get_profile_replies` is not included. Use `get_post_replies` for replies
beneath a selected post.

## Install

Install the exact version with Go 1.26.9 or newer:

```bash
go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.2
```

Release archives are available for macOS on Apple silicon and Intel, Linux on
`arm64` and `amd64`, and Windows on `amd64`. Configure your MCP client with the
absolute path to `threads-mcp`, or `threads-mcp.exe` on Windows. The
[client setup guide](https://github.com/granitebps/threads-mcp#client-configuration)
has examples for Codex, Claude, Cursor, and OpenCode.

Homebrew and Scoop packages are not available in v1.0.2.

## Verify a download

Each release archive has a SHA-256 entry in `checksums.txt` and a software bill
of materials. The release also includes `checksums.txt.bundle`, which contains
the keyless Cosign signature material for the checksum file.

Follow the
[download verification instructions](https://github.com/granitebps/threads-mcp#verify-the-download)
before running a downloaded executable. The binaries are not Apple-notarized or
Windows Authenticode-signed, so operating-system warnings are possible even
after checksum and Cosign verification succeed.

## Access limits

This project is unofficial and is not affiliated with Meta. It reads public,
anonymous Threads web pages that Meta may change, limit, or block without
notice.

- Private or login-only content is not supported.
- Search, replies, and profile posts are limited public windows, not complete
  archives. There is no historical pagination in v1.0.2.
- `get_post_replies` returns only the replies embedded in the public post page.
- Anonymous search returns results for common keywords and phrases. Many longer
  or less common multi-word queries, and all quoted phrases, return no results
  from Threads itself.
- An empty result does not prove that matching content does not exist.
- Threads sometimes returns an incomplete page with HTTP 200. The server retries
  affected reads, then returns a retryable `UPSTREAM_CHANGED` error if every
  attempt is incomplete.
- Threads may omit fields, and its media URLs may expire.

Read the
[access notice](https://github.com/granitebps/threads-mcp#important-access-notice)
before using the server.

## Compatibility and support

The [v1 compatibility promise](https://github.com/granitebps/threads-mcp/blob/v1.0.2/docs/v1-compatibility.md)
covers the six tool names, accepted inputs, output envelopes, field types, and
structured error meanings throughout v1.x.

Use the [issue tracker](https://github.com/granitebps/threads-mcp/issues) for
ordinary bugs. Report suspected vulnerabilities through the
[private security route](https://github.com/granitebps/threads-mcp/security/advisories/new)
without opening a public issue.
