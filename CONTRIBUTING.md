# Contributing

Thanks for helping improve threads-mcp. Keep changes focused and explain the
user-visible reason for them.

## Before starting

Use the [public issue tracker](https://github.com/granitebps/threads-mcp/issues)
for bugs and feature proposals. For a larger change, open an issue before doing
substantial work so its scope and compatibility impact can be discussed.

Do not open a public issue for a suspected vulnerability. Follow the
[security policy](SECURITY.md) instead.

The v1 server is intentionally read-only, uses MCP over stdio, and reads only
content that Threads exposes without a login. Authentication, write operations,
historical pagination, and `get_profile_replies` are outside the v1 scope.

## Development setup

Install Go 1.26.6 or newer, then work from a clone or fork:

```bash
go mod download
go build -o ./bin/threads-mcp ./cmd/threads-mcp
go test -count=1 ./...
```

The executable is an MCP stdio server. Starting it directly will appear to wait
for input. Use the [README client configuration](README.md#client-configuration)
to connect an MCP client during manual testing.

## Make a change

- Add or update tests for behavior changes and bug fixes.
- Preserve the error envelope and keep diagnostic details off stdout.
- Do not add Threads credentials, cookies, login automation, proxy workarounds,
  or private content to code, tests, fixtures, issues, or commits.
- Keep generated files from `bin/` and `dist/` out of commits.
- If a dependency changes, explain why and review `go.mod`, `go.sum`, license
  obligations, and `THIRD_PARTY_NOTICES.md`.

Read the [v1 compatibility promise](docs/v1-compatibility.md) before changing a
tool name, input schema, output shape, error code, transport, or accepted input
form. Additive changes can still break clients if an existing value changes
meaning. Discuss any compatibility exception before implementation.

## Verify the change

Run the deterministic checks before opening a pull request:

```bash
test -z "$(gofmt -l .)"
go vet ./...
go build ./...
go test -race -count=1 ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1 run
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
```

Windows does not run the race detector in this repository's CI. On Windows,
run `go build ./...` and `go test -count=1 ./...`; GitHub Actions will run the
other platform jobs.

If imports or dependencies changed, also run `go mod tidy` and inspect the
resulting `go.mod` and `go.sum` diff. Do not include unrelated dependency churn.

### Live Threads check

Changes to the Threads provider, parsers, retry behavior, mappings, MCP data
tools, or live fixtures should also run:

```bash
go test -tags=live ./internal/provider/threadscli -run TestAnonymousCrawlerPublicReads -count=1 -v
```

This check calls an external public service. It is separate from deterministic
CI and can fail when Threads changes or limits anonymous access. Report the
actual result. Do not weaken assertions, add credentials, or hide a failure to
make it pass. Avoid rapid repeated runs.

Documentation-only changes do not need the Go or live test suites. Check local
links, code fences, shell examples, and `git diff --check` instead.

## Open a pull request

In the pull request description:

- explain the problem and the chosen scope;
- summarize user-visible and compatibility effects;
- list the checks you ran and their actual results;
- identify any live or platform checks you did not run;
- link the related issue when one exists.

Keep the pull request small enough to review. Leave unrelated cleanup for a
separate change. Maintainers may ask for updates when a change expands v1 scope,
alters the compatibility promise, or lacks verification.

Do not create release tags or publish artifacts as part of a contribution. The
maintainer follows the owner-approved process in [RELEASE.md](RELEASE.md).
