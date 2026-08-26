# threads-mcp design

**Status:** Approved, revised provider policy

**Date:** 2026-08-26

## Purpose

`threads-mcp` is an open-source Model Context Protocol server that lets MCP clients read public Threads content without user credentials. It runs as a local stdio process and returns typed, machine-readable results.

The server covers post search, individual posts, replies beneath a post, public profiles, and recent posts by a profile. It does not expose a user's replies across other people's posts.

## Product constraints

- Read public Threads data only.
- Do not accept or use Threads access tokens, session cookies, or login credentials.
- Do not read private or login-restricted content.
- Do not add CAPTCHA handling, browser automation, proxy rotation, or other anti-bot workarounds.
- Use `threads-cli`'s documented anonymous crawler mode, including its crawler user agent, because Threads does not expose the required public reads to an ordinary application user agent.
- Disclose that this is an unofficial, crawler-facing web surface that Meta may change or restrict without notice. Do not describe it as an official API or guaranteed access.
- Do not silently fall back to credentials, proxies, CAPTCHA solving, or a hosted scraping provider when crawler-mode access fails.
- Run over MCP stdio. Reserve stdout for protocol messages and send diagnostics to stderr.
- Return unavailable values as omitted fields or JSON `null`. Never invent values.
- Report partial or uncertain results explicitly.

## V1 tool set

### `search_posts`

Search public Threads posts by keyword.

Input:

- `query`, required string containing 1 to 500 Unicode code points after trimming.
- `limit`, optional integer. The default is 10 and the maximum is 50.

Output includes matching public posts, the returned count, the requested limit, result completeness, and warnings.

### `get_post`

Read one public Threads post.

Input:

- `post`, required canonical Threads post URL or another identifier that the provider can fetch without guessing missing information.

Bare numeric IDs and shortcodes must not be advertised as fetchable until live tests prove that the provider can resolve them without an author handle.

### `get_post_replies`

Read replies beneath one public Threads post.

Input:

- `post`, required post identifier accepted by `get_post`.
- `limit`, optional integer. The default is 20 and the maximum is 100.

This tool returns the visible public conversation beneath the selected post. It does not return every reply written by a particular user. The result must say whether it is complete, partial, or unknown.

### `get_profile`

Read one public Threads profile.

Input:

- `username`, required Threads username. A leading `@` may be accepted and normalized.

### `get_profile_posts`

Read the recent public posts exposed for one profile.

Input:

- `username`, required Threads username.
- `limit`, optional integer. The default is 20 and the maximum is 100.

The documentation and result metadata must describe this as a recent public window, not a complete profile history.

### `get_server_info`

Return server, MCP protocol, SDK, and provider versions. Also report access mode, supported identifier forms, tool availability, known limits, and active release gates.

### Excluded tool

`get_profile_replies` is out of scope. The project needs replies beneath a selected post, not a feed of replies written by one user. The upstream implementation is also non-paginated and cannot reliably distinguish no replies from unavailable reply data.

## Result model

The server owns its public result types. Provider structs must not become the MCP contract because the provider is pre-1.0 and currently loses some missing-value information.

Common list results use this shape:

```json
{
  "items": [],
  "returned_count": 0,
  "requested_limit": 20,
  "completeness": "complete",
  "warnings": []
}
```

`completeness` is one of:

- `complete`, when the provider proves that no more matching public results exist within the requested operation.
- `partial`, when the provider reports an interruption after returning some records.
- `unknown`, when the provider cannot prove whether the returned window is complete.

Post records may contain:

- Stable ID and shortcode when available.
- Text.
- Author username and user ID.
- Canonical permalink.
- RFC 3339 UTC timestamp.
- Media type and public media URLs.
- Like, reply, repost, and quote counts only when presence is known.
- Reply and quote relationships only when exposed by the source.

Profile records may contain:

- Stable ID and username.
- Display name and biography.
- Profile image URL.
- Verification state only when presence is known.
- Follower and following counts only when presence is known.
- Canonical profile URL.

Media URLs remain owned by Threads and may expire. `threads-mcp` does not download, proxy, or promise durable storage for media.

Every successful MCP tool response includes `structuredContent` and an equivalent JSON text content item for clients that do not consume structured output.

## Error contract

Malformed MCP requests and unknown tools use MCP protocol errors. Expected tool failures return `isError: true` with a structured error object.

Stable error codes:

- `INVALID_INPUT`
- `NOT_FOUND`
- `ACCESS_RESTRICTED`
- `RATE_LIMITED`
- `NETWORK_FAILURE`
- `TIMEOUT`
- `UNSUPPORTED_OPERATION`
- `UPSTREAM_CHANGED`
- `INTERNAL_ERROR`

`ACCESS_RESTRICTED` means the content is unavailable through anonymous public access. The message may say that the content could be private or login-restricted, but it must not claim which case applies unless the provider proves it.

Errors may include a safe message, `retryable`, and `retry_after_seconds` when the provider supplies that value. They must not contain raw response bodies, headers, credentials, local paths, or stack traces.

## Architecture

The server is a Go application with these boundaries:

- `cmd/threads-mcp` owns process startup, version metadata, stdio transport, signal handling, and fatal stderr output.
- `internal/server` owns MCP setup, schemas, annotations, handlers, and conversion to MCP results.
- `internal/provider` defines the provider interface, capability report, page metadata, and provider errors.
- `internal/provider/threadscli` adapts the pinned `threads-cli` package for profiles, profile posts, posts, and replies, and owns a small public search-page reader for keyword search.
- `internal/domain` defines server-owned post, profile, reply, warning, completeness, and error models.
- `internal/identifier` validates and normalizes supported Threads identifiers.
- `internal/version` exposes build version, commit, MCP protocol version, SDK version, and provider version.

The request flow is:

1. The MCP SDK validates the tool input schema.
2. The handler applies semantic validation and a request deadline.
3. The handler calls the provider interface.
4. The provider adapter calls `threads-cli` in its documented anonymous crawler mode for profile and conversation reads, or fetches the public Threads search page for keyword search.
5. The adapter maps provider data and errors into server-owned types.
6. The handler returns structured content and equivalent JSON text.

The provider boundary permits a later optional hosted provider without changing the MCP tool contract. V1 does not expose provider selection or depend on a hosted service.

## Data provider decision

The preferred provider is a hybrid local adapter. It directly imports `github.com/tamnd/threads-cli/threads`, pinned to an exact release, for profiles, profile posts, posts, and replies. Keyword search instead reads the server-rendered JSON embedded in the public Threads search page because `threads-cli v0.1.1`'s persisted GraphQL search document is stale. A TypeScript or Go wrapper that spawns the `th` executable is rejected because it adds a second installation, process cancellation work, JSONL parsing, and Windows process handling without fixing the search failure.

This decision knowingly accepts `threads-cli`'s crawler user agent for both paths. The search reader performs one bounded HTTP GET per query, extracts only embedded `thread_items[].post` records, and does not run JavaScript or paginate beyond the public page. Hosted alternatives were not selected for V1 because they use the same or more invasive unofficial surfaces while adding API keys, usage cost, external availability, and disclosure of requested public targets to a third party. This acceptance does not establish compliance with Meta's terms; the README must state the access mechanism and operational risk plainly.

The adapter must force anonymous mode and must not read authentication variables from the environment.

The default MCP tool deadline is 45 seconds. `THREADS_MCP_TIMEOUT` may set a duration from 5 seconds through 120 seconds. The provider may make at most two attempts for one upstream request and must wait at least one second between outbound requests. Every request, retry wait, and pacing wait must honor context cancellation. If the pinned release cannot support those rules, the project needs an upstream fix, a narrowly documented local patch, or a different provider.

## Provider release gates

The first public release is blocked until all of these checks pass:

1. Public profile, post, post reply, and profile post reads work through `threads-cli`'s anonymous crawler mode, and search works through the public search page, all without credentials or proxies.
2. The adapter can distinguish missing optional values from real zero and false values, or it omits affected fields.
3. List operations expose complete, partial, or unknown status without converting pagination failures into clean success.
4. Empty search results can be distinguished from a changed or blocked search page, or they return `completeness: "unknown"` with a warning.
5. Retry and pacing waits stop promptly when the MCP request context is cancelled.

The preferred solution is to contribute the missing behavior upstream. A maintained fork is a fallback and requires an explicit decision because it increases maintenance and licensing work.

## MCP compatibility

The current final MCP specification is `2026-07-28`. The stable official Go SDK `v1.7.0` fully supports that protocol and preserves compatibility with `2025-11-25` and earlier peers.

Implementation pins `github.com/modelcontextprotocol/go-sdk v1.7.0`. The project must recheck current stable releases before a public release and update the pin through a separate reviewed dependency change when needed.

All tools use output schemas and read-only, non-destructive, idempotent, open-world annotations.

## Stdio rules

- Write only MCP protocol messages to stdout.
- Write logs, warnings, and fatal startup errors to stderr.
- Do not log full upstream payloads.
- Redact URLs if they could contain credentials, even though V1 accepts no credentials.
- Handle `SIGINT` and `SIGTERM` and close the MCP transport cleanly.
- Propagate request cancellation to the provider.

## Testing strategy

Deterministic pull request checks must not depend on live Threads access.

Required tests:

- Handler unit tests with a fake provider.
- Input schema and semantic validation tests.
- Identifier normalization tests.
- Provider-to-domain mapping tests, including missing optional fields.
- Error classification and redaction tests.
- Completeness and warning tests for interrupted pagination and ambiguous empty results.
- Structured-content and JSON-text equivalence tests.
- End-to-end stdio handshake and tool-call tests using an MCP client.
- A stdout test that rejects non-protocol output.
- Cancellation, timeout, pacing, and retry tests.
- Release artifact startup and MCP handshake tests.

Scheduled live smoke tests use a small set of stable public fixtures. They verify anonymous crawler-mode access and detect upstream page or query changes. Live failures must not be disguised as deterministic unit-test failures, but repeated failures must block a release.

CI covers Linux, macOS, and Windows builds. It runs formatting checks, unit tests, race tests where supported, `go vet`, static analysis, vulnerability scanning, and a release dry-run.

## Distribution

V1 distribution includes:

- GitHub Release archives for macOS and Linux on `amd64` and `arm64`, plus Windows on `amd64`.
- `go install github.com/granitebps/threads-mcp/cmd/threads-mcp@vX.Y.Z`.
- Local repository installation by building `./cmd/threads-mcp` and configuring clients with the absolute binary path.
- Local development through `go run ./cmd/threads-mcp` for clients that support a working directory.
- Scoop for Windows unless it is separately deferred before implementation.
- SHA-256 checksums, software bills of materials, and a Cosign bundle for the checksum file.

Homebrew and an npm launcher are out of V1 scope.

The official MCP Registry did not support Go modules at research time. Registry publication waits for native Go support or a separate, tested decision to publish an MCPB package. The project must not publish misleading `server.json` package metadata.

Documentation must include working configuration examples for Codex, Claude Code, Claude Desktop, Cursor, and OpenCode. GUI client examples use absolute binary paths because GUI applications may not inherit the shell `PATH`.

## Licensing

The recommended project license is Apache-2.0. Releases include:

- The project `LICENSE`.
- `THIRD_PARTY_NOTICES.md` naming the exact `threads-cli` version, repository, and Apache-2.0 license.
- Any notices required by copied or modified upstream files.
- Dependency metadata in the release software bill of materials.

No commit, push, pull request, package publication, or release is part of the initial implementation unless separately requested.

## Acceptance criteria

The project is ready for its first public release when:

- All six V1 tools have stable typed schemas and documented examples.
- A fresh install can start through stdio in every documented client configuration.
- stdout contains no logs or other non-protocol text.
- The supported public operations pass scheduled live checks through anonymous crawler mode.
- Partial and unavailable data is visible in the result contract.
- The server uses a stable official Go SDK compatible with MCP `2026-07-28`.
- Release archives, checksums, Cosign bundles, and software bills of materials pass artifact verification.
- The README states the public-data limits and does not promise complete Threads history.

## Reference material

- [MCP specification release 2026-07-28](https://blog.modelcontextprotocol.io/posts/2026-07-28/)
- [Official Go MCP SDK releases](https://github.com/modelcontextprotocol/go-sdk/releases)
- [`threads-cli` repository](https://github.com/tamnd/threads-cli)
- [`threads-cli` Go package](https://pkg.go.dev/github.com/tamnd/threads-cli/threads)
- [`twitter-mcp` reference repository](https://github.com/granitebps/twitter-mcp)
- [MCP Registry package types](https://github.com/modelcontextprotocol/registry/blob/main/docs/modelcontextprotocol-io/package-types.mdx)
