# v1.0.2 patch release checklist

Status: published and verified on 2026-10-13.

Follow the [release guide](../RELEASE.md) for the reusable procedure. This
checklist records the v1.0.2 scope, evidence, approval, and public verification.

## Why v1.0.2 is required

Threads changed its public post pages around 2026-10-07. The scheduled
[live smoke run 37606582808](https://github.com/granitebps/threads-mcp/actions/runs/37606582808)
failed with `tool get_post returned no usable response`. Every `get_post` and
`get_post_replies` call returned `UPSTREAM_CHANGED` after five retries because
threads-cli v0.1.1 only reads `thread_items`, which the new pages no longer
contain.

The CI vulnerability check also started failing on Go 1.26.6 because of
[GO-2026-6617](https://pkg.go.dev/vuln/GO-2026-6617), fixed in Go 1.26.9. The
publishing job requires CI, so the Go bump is part of this release.

## Changes

- [x] Read post pages split into Relay fragments under `data.media`, taking the root post from the full fragment and `self_thread` and `direct_replies` from the fragment with the same id. Fall back to threads-cli when a page lacks that layout.
- [x] Add fixture tests for the split layout, nested reply parents, de-duplication, media, and unknown layouts.
- [x] Raise the `go.mod` Go version to 1.26.9 and update the documented minimum.
- [x] Prepare the [v1.0.2 release notes](release-notes-v1.0.2.md).
- [x] Update README installation, archive, and verification references to v1.0.2.
- [x] Confirm no MCP tool, schema, or error-code changes against the [v1 compatibility promise](v1-compatibility.md).

## Verify the candidate without publishing

Record evidence against the exact commit that would receive the tag. A new
change creates a new candidate and requires its checks again.

- [x] Pass all Go tests, race tests, formatting, vet, build, lint, module-tidiness, and vulnerability checks.
- [x] Pass Linux, macOS, and Windows CI for the same commit.
- [x] Pass the release workflow's read-only toolchain preflight on GitHub for the same commit.
- [x] Pass all five live public Threads reads through MCP.
- [x] Confirm `v1.0.2` is absent locally, remotely, and from GitHub Releases before tagging.

## Approval boundary

- Candidate identity: `dbe29204b2de7801f691ffb710480c7c3c381d0f`.
- Approver: repository owner.
- Approval date: 2026-10-13.
- Authorized actions: create and push annotated tag `v1.0.2` from the candidate commit, then allow the tag-triggered workflow to publish the GitHub Release.

The owner approved the exact commit and publishing actions before the tag was
created. The tag push consumed that approval; it does not authorize moving the
tag or replacing published files.

## Publish and verify only after approval

- [x] Reconfirm a clean checkout, the exact approved commit, and absence of local and remote `v1.0.2` tags.
- [x] Create and push the annotated `v1.0.2` tag from the approved commit.
- [x] Confirm required CI and the release preflight succeed for the tagged commit.
- [x] Confirm the publishing job succeeds and creates one GitHub Release for v1.0.2.
- [x] Authenticate `checksums.txt` with its Cosign bundle, expected workflow identity, exact tag, and GitHub Actions issuer.
- [x] Verify the archives against the authenticated checksums and confirm the SBOMs are present.
- [x] Test public `go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.2` from a clean directory and verify server `1.0.2` and provider `v0.1.1`.

## Verification evidence

- Pre-tag [CI run 38373720442](https://github.com/granitebps/threads-mcp/actions/runs/38373720442) passed Linux, macOS, Windows, lint, module-tidiness, and vulnerability checks for the candidate commit.
- Pre-tag [release preflight 38373720777](https://github.com/granitebps/threads-mcp/actions/runs/38373720777) passed the read-only toolchain and snapshot check.
- Locally on Go 1.26.9, formatting, vet, build, race tests, golangci-lint v2.13.1, and govulncheck v1.7.0 passed, and `go mod tidy` changed nothing. The live MCP probe passed all five data tools in 16.59 seconds.
- A local GoReleaser snapshot rehearsal was not run; the GitHub preflight built the snapshot for the candidate commit.
- [Tag release run 38374099635](https://github.com/granitebps/threads-mcp/actions/runs/38374099635) passed all required CI, preflight, and publishing jobs for the candidate commit.
- The published [v1.0.2 GitHub Release](https://github.com/granitebps/threads-mcp/releases/tag/v1.0.2) is neither a draft nor a prerelease and contains twelve assets: five archives, five SBOMs, `checksums.txt`, and `checksums.txt.bundle`.
- Cosign authenticated `checksums.txt` with certificate identity `https://github.com/granitebps/threads-mcp/.github/workflows/release.yml@refs/tags/v1.0.2` and issuer `https://token.actions.githubusercontent.com`. All ten checksum entries matched the downloaded files.
- The downloaded macOS ARM64 binary initialized through MCP, listed all six tools, reported server version `1.0.2` and provider version `v0.1.1`, and returned live replies from `get_post_replies`.
- A clean public `go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.2` succeeded through the Go proxy and passed the same MCP check.

Version v1.0.2 is complete. Never move the tag or replace published assets with
different builds under the same version.
