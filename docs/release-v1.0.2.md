# v1.0.2 patch release checklist

Status: candidate in preparation.

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

- [ ] Pass all Go tests, race tests, formatting, vet, build, lint, module-tidiness, and vulnerability checks.
- [ ] Pass Linux, macOS, and Windows CI for the same commit.
- [ ] Pass the release workflow's read-only toolchain preflight on GitHub for the same commit.
- [ ] Pass all five live public Threads reads through MCP.
- [ ] Confirm `v1.0.2` is absent locally, remotely, and from GitHub Releases before tagging.

## Approval boundary

- Candidate identity: pending.
- Approver: repository owner.
- Approval date: pending.
- Authorized actions: pending.

## Publish and verify only after approval

- [ ] Reconfirm a clean checkout, the exact approved commit, and absence of local and remote `v1.0.2` tags.
- [ ] Create and push the annotated `v1.0.2` tag from the approved commit.
- [ ] Confirm required CI and the release preflight succeed for the tagged commit.
- [ ] Confirm the publishing job succeeds and creates one GitHub Release for v1.0.2.
- [ ] Authenticate `checksums.txt` with its Cosign bundle, expected workflow identity, exact tag, and GitHub Actions issuer.
- [ ] Verify the archives against the authenticated checksums and confirm the SBOMs are present.
- [ ] Test public `go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.2` from a clean directory and verify server `1.0.2` and provider `v0.1.1`.

## Verification evidence

Pending.
