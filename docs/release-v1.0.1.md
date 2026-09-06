# v1.0.1 recovery release checklist

Status: preparation in progress. Publishing is not authorized.

Follow the [release guide](../RELEASE.md) for the reusable procedure. This
checklist records the v1.0.1 recovery scope, evidence, approval, and public
verification. It does not authorize creating or pushing the tag.

## Why v1.0.1 is required

The immutable `v1.0.0` source tag is public and installable with Go, but
[publishing run 34016585356](https://github.com/granitebps/threads-mcp/actions/runs/34016585356)
failed before it created a GitHub Release. Cosign Installer v3.10.0 expected a
legacy signature asset that Cosign v3.1.3 does not publish. GitHub reruns use the
workflow from the tagged commit, so a correction on `main` cannot repair that
run. The v1.0.0 tag must not be moved or replaced.

Version v1.0.1 contains no MCP tool, schema, provider, or runtime behavior
change. It repairs publishing and adds earlier validation for future releases.

## Recovery changes

- [x] Pin Cosign Installer v4.1.2, which supports Cosign v3 release assets, while keeping Cosign v3.1.3.
- [x] Pin the Syft installer action to v0.24.2 so it uses GitHub's supported Node.js 24 action runtime, while keeping Syft v1.51.0.
- [x] Install the exact Cosign and Syft versions during the read-only release check and print both versions.
- [x] Sign and verify a temporary payload with Cosign v3 using an ephemeral local key, without OIDC credentials or transparency-log upload.
- [x] Run a complete GoReleaser snapshot with archives, checksums, and SBOMs on ordinary pushes and pull requests without signing or publishing.
- [x] Require a tag-specific `docs/release-notes-VERSION_TAG.md` file and pass it to GoReleaser.
- [x] Prepare the [v1.0.1 release notes](release-notes-v1.0.1.md).
- [x] Update public installation, archive, verification, limitation, and maintainer documentation for v1.0.1.
- [x] Keep the preflight read-only. Only the tag-only publishing job has `contents: write` and `id-token: write`.
- [x] Keep Homebrew, Scoop, and MCP Registry publishing disabled.

## Verify the candidate without publishing

Record evidence against the exact commit that would receive the tag. A new
change creates a new candidate and requires its checks again.

- [ ] Pass all Go tests, race tests, formatting, vet, build, lint, module-tidiness, and vulnerability checks.
- [ ] Pass actionlint and the workflow regression tests.
- [ ] Pass the release workflow's read-only toolchain preflight on GitHub, including Cosign installation, local sign/verify, Syft installation, GoReleaser configuration check, and complete snapshot build.
- [ ] Pass Linux, macOS, and Windows CI for the same commit with zero unexpected annotations.
- [ ] Pass all five live public Threads reads through MCP.
- [ ] Rehearse the complete snapshot in a clean full-history checkout and inspect archives, checksums, SBOMs, and the native packaged executable.
- [ ] Verify local-repository, local-installed, and public `v1.0.0` installed commands through real MCP sessions.
- [ ] Review the exact workflow diff, permissions, release-note path, action pins, and v1.0.1 public documentation.
- [ ] Confirm `v1.0.1` is absent locally, remotely, from the Go proxy, and from GitHub Releases.

The preflight validates the compatible installer, Cosign v3 commands, package
generation, and SBOM generation before tagging. The first v1.0.1 tag run remains
the first end-to-end check of GitHub OIDC signing and GitHub Release upload.

## Approval boundary

- Candidate identity: the commit containing the final checked version of this checklist.
- Approver: pending.
- Approval date: pending.
- Authorized actions: pending.

Before publishing, record explicit owner approval for the exact commit and for
creating and pushing `v1.0.1`, which starts public source and binary publication.
Approval for this recovery implementation, tests, commits, or pushes to `main`
is not tag or publication approval.

## Publish and verify only after approval

- [ ] Reconfirm a clean checkout, the exact approved commit, and absence of local and remote `v1.0.1` tags.
- [ ] Create and push the annotated `v1.0.1` tag from the approved commit.
- [ ] Confirm required CI and the release preflight succeed for the tagged commit.
- [ ] Confirm the publishing job succeeds and creates one GitHub Release for v1.0.1.
- [ ] Authenticate `checksums.txt` with its Cosign bundle, expected workflow identity, exact tag, and GitHub Actions issuer.
- [ ] Verify all five archives against the authenticated checksums and confirm five SPDX SBOMs are present.
- [ ] Extract the native archive and run a six-tool MCP session plus all five live data reads.
- [ ] Test public `go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.1` from a clean directory and verify server `1.0.1` and provider `v0.1.1`.
- [ ] Confirm the README release link, archive names, installation command, and verification commands work as written.

Do not call v1.0.1 complete until the public assets, signature, installed command,
and downloaded executable have been verified. Never move the tag or replace
published assets with different builds under the same version.
