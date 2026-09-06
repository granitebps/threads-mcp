# v1 compatibility promise

This document defines the public tool contract for threads-mcp v1. It starts
when v1.0.0 is published. Before that tag exists, the unreleased contract may
still change as part of release preparation.

The promise covers MCP tool names, accepted input shapes, successful output
shapes, and structured tool errors. It does not guarantee that Threads keeps
its public pages available or that a public result window is complete.

## Compatibility rules

Within v1.x, a release will not:

- remove or rename a v1 tool;
- rename an existing input or output field;
- change an existing field's JSON type;
- make an optional input required;
- change the documented meaning of an existing error code;
- replace a successful output envelope with a different envelope; or
- treat an accepted v1 input form as invalid without a security, legal, or upstream-access reason that is documented in the release notes.

A v1 minor or patch release may add:

- a new tool;
- a new optional input or output field;
- a new structured error code; or
- a new enum value when existing clients can safely treat an unknown value as unspecified.

Clients should ignore unknown output fields and handle unknown error codes as a
non-retryable error unless the response says `retryable: true`. The project will
use a new major version for an intentional breaking contract change. An urgent
security or upstream-access change may narrow behavior in v1; its release notes
must identify the exception and its effect.

This promise covers JSON structure and meaning, not exact prose. Error messages,
tool descriptions, warning text, result ordering, and omitted unavailable values
may change without a major version.

## Common behavior

All tools are read-only, idempotent, non-destructive, and access an external
public service. Inputs are JSON objects. Data-tool schemas reject undocumented
input properties; `get_server_info` accepts an empty object.

Every tool returns structured JSON and one text content item containing the same
JSON value. A successful data tool returns exactly one of the success envelopes
documented below. A tool failure sets the MCP result's `isError` value and
returns the `error` envelope instead of a success value.

## Tool inputs and envelopes

| Tool | Input | Default and limit | Success envelope |
| --- | --- | --- | --- |
| `search_posts` | Required `query` string. Leading and trailing whitespace is removed. The result must contain 1 to 500 Unicode code points. | Optional integer `limit`, default 10, range 1 to 50. | `{"result": Page<Post>}` |
| `get_post` | Required `post` string containing a public HTTPS Threads post URL on `threads.com` or `threads.net`, with or without `www`. | None. | `{"post": Post}` |
| `get_post_replies` | Required `post` string with the same rules as `get_post`. | Optional integer `limit`, default 20, range 1 to 100. | `{"result": Page<Reply>}` |
| `get_profile` | Required `username` string containing a bare handle, an `@handle`, or an HTTPS Threads profile URL. | None. | `{"profile": Profile}` |
| `get_profile_posts` | Required `username` string with the same rules as `get_profile`. | Optional integer `limit`, default 20, range 1 to 100. | `{"result": Page<Post>}` |
| `get_server_info` | Empty object. | None. | `ServerInfo` without a wrapper field. |

`get_profile_replies` is not a v1 tool. Use `get_post_replies` to read replies
beneath a selected post.

## Output types

### `Post`

`id` is a required string. These fields are optional because Threads may omit
them from its public response:

- String: `shortcode`, `text`, `username`, `user_id`, `permalink`, `timestamp`, `media_type`, `reply_to_id`, `quoted_post_id`.
- Array of strings: `media_urls`.
- Integer: `like_count`, `reply_count`, `repost_count`, `quote_count`.
- Boolean: `is_reply`, `is_quote_post`.

Missing optional fields are omitted rather than guessed. Media URLs remain owned
by Threads and may expire.

### `Reply`

`Reply` contains all `Post` fields. It may also contain string fields
`parent_id` and `root_id`.

### `Profile`

`id`, `username`, and `url` are strings. The provider requires a non-empty `id`
before returning success. These fields are optional:

- String: `name`, `biography`, `profile_pic_url`.
- Boolean: `is_verified`.
- Integer: `follower_count`, `following_count`.

### `Page<T>`

A page has:

- `items`: an array of `T`, which may be empty;
- `returned_count`: the number of returned items;
- `requested_limit`: the accepted request limit;
- `completeness`: `complete`, `partial`, or `unknown`; and
- `warnings`: an array of strings, which may be empty.

An empty `items` array is a successful response, not proof that matching public
content does not exist. Public search, replies, and profile posts are limited
windows. The provider normally reports `unknown`; it reports `partial` when it
returns some items and then encounters an upstream failure. `complete` is
reserved for a provider that can prove completeness.

The `limit` is a maximum returned-item count, not pagination. V1 does not promise
a stable order, cursor, offset, archive, or exhaustive result set.

### `ServerInfo`

`get_server_info` returns these fields:

- String: `server_name`, `server_version`, `commit`, `build_date`, `sdk_version`, `protocol_version`, `provider_name`, `provider_version`, `provider_mode`, `provider_user_agent`.
- Array of strings: `supported_tools`, `excluded_tools`, `supported_inputs`, `known_limitations`.

Build values may be `dev` or `unknown` for a local checkout. The provider version
identifies the provider dependency; it is not the server version.
`protocol_version` is the newest protocol version supported by the embedded SDK,
not the version negotiated for an individual client session. The SDK can
negotiate an older supported version with a client.

## Structured errors

The error envelope is:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "safe public message",
    "retryable": false,
    "retry_after_seconds": 30
  }
}
```

`code`, `message`, and `retryable` are always present. `retry_after_seconds` is
optional and is omitted when no delay is known. Internal causes and upstream raw
details are not part of the public response.

| Code | Meaning |
| --- | --- |
| `INVALID_INPUT` | The input does not match the tool contract or Threads rejected it as invalid. |
| `NOT_FOUND` | The requested public content was not found. |
| `ACCESS_RESTRICTED` | Threads did not allow anonymous public access. This does not prove whether the content is private, removed, or login-restricted. |
| `RATE_LIMITED` | Threads rate-limited the request. This is normally retryable. |
| `NETWORK_FAILURE` | The request to Threads failed at the network layer. This is normally retryable. |
| `TIMEOUT` | The request timed out or was cancelled. This is normally retryable. |
| `UNSUPPORTED_OPERATION` | The requested operation is recognized but unsupported. It is reserved for v1 use and is not normally produced by the current six tools. |
| `UPSTREAM_CHANGED` | Threads returned a response the provider could no longer understand, including a changed public page. |
| `INTERNAL_ERROR` | An unexpected internal error occurred. Details are intentionally hidden. |

`retryable: true` means another attempt may succeed. It does not guarantee
success, and clients should avoid immediate repeated requests. If
`retry_after_seconds` is present, clients should wait at least that many seconds.

## What is outside the compatibility promise

The following depend on Threads and are not stable API guarantees:

- availability of anonymous access;
- which public posts or replies Threads exposes;
- returned ordering and ranking;
- media URL lifetime;
- post, profile, or engagement values that Threads omits or changes; and
- exact timing, rate-limit thresholds, and upstream response latency.

The MCP transport remains stdio in v1. Adding another transport does not remove
stdio. Changes required by a newer compatible MCP protocol or SDK may occur in a
v1 release if the six tool contracts above remain compatible.
