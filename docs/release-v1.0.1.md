# v1.0.1 recovery release checklist

Status: published and verified on 2026-09-06.

Follow the [release guide](../RELEASE.md) for the reusable procedure. This
checklist records the v1.0.1 recovery scope, evidence, approval, and public
verification.

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

- [x] Pass all Go tests, race tests, formatting, vet, build, lint, module-tidiness, and vulnerability checks.
- [x] Pass actionlint and the workflow regression tests.
- [x] Pass the release workflow's read-only toolchain preflight on GitHub, including Cosign installation, local sign/verify, Syft installation, GoReleaser configuration check, and complete snapshot build.
- [x] Pass Linux, macOS, and Windows CI for the same commit with zero unexpected annotations.
- [x] Pass all five live public Threads reads through MCP.
- [x] Rehearse the complete snapshot in a clean full-history checkout and inspect archives, checksums, SBOMs, and the native packaged executable.
- [x] Verify local-repository, local-installed, and public `v1.0.0` installed commands through real MCP sessions.
- [x] Review the exact workflow diff, permissions, release-note path, action pins, and v1.0.1 public documentation.
- [x] Confirm `v1.0.1` was absent locally, remotely, from the Go proxy, and from GitHub Releases before tagging.

The preflight validates the compatible installer, Cosign v3 commands, package
generation, and SBOM generation before tagging. The first v1.0.1 tag run remains
the first end-to-end check of GitHub OIDC signing and GitHub Release upload.

## Approval boundary

- Candidate identity: `c8ef81561e799fe5194d3c04947f70ae0d635069`.
- Approver: repository owner.
- Approval date: 2026-09-06.
- Authorized actions: create and push annotated tag `v1.0.1` from the candidate commit, then allow the tag-triggered workflow to publish the GitHub Release.

The owner gave explicit approval for the exact commit and publishing actions
before the tag was created. The tag push consumed that approval; it does not
authorize moving the tag or replacing published files.

## Publish and verify only after approval

- [x] Reconfirm a clean checkout, the exact approved commit, and absence of local and remote `v1.0.1` tags.
- [x] Create and push the annotated `v1.0.1` tag from the approved commit.
- [x] Confirm required CI and the release preflight succeed for the tagged commit.
- [x] Confirm the publishing job succeeds and creates one GitHub Release for v1.0.1.
- [x] Authenticate `checksums.txt` with its Cosign bundle, expected workflow identity, exact tag, and GitHub Actions issuer.
- [x] Verify all five archives against the authenticated checksums and confirm five SPDX SBOMs are present.
- [x] Extract the native archive and run a six-tool MCP session plus all five live data reads.
- [x] Test public `go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.1` from a clean directory and verify server `1.0.1` and provider `v0.1.1`.
- [x] Confirm the README release link, archive names, installation command, and verification commands work as written.

## Verification evidence

- Final pre-tag [CI run 34025103997](https://github.com/granitebps/threads-mcp/actions/runs/34025103997) passed Linux, macOS, Windows, lint, module-tidiness, and vulnerability checks for the candidate commit.
- Final pre-tag [release preflight 34025104095](https://github.com/granitebps/threads-mcp/actions/runs/34025104095) installed Cosign v3.1.3 and Syft v1.51.0, completed the local-key Cosign test, validated GoReleaser configuration, and generated the full snapshot. The candidate had zero GitHub annotations.
- The local live MCP probe passed all five data tools in 16.47 seconds. The snapshot contained five archives, five SPDX 2.3 SBOMs, and ten valid checksum entries. The native archive passed a six-tool MCP session.
- [Tag release run 34025410031](https://github.com/granitebps/threads-mcp/actions/runs/34025410031) passed all required CI, preflight, and publishing jobs with zero annotations.
- The published [v1.0.1 GitHub Release](https://github.com/granitebps/threads-mcp/releases/tag/v1.0.1) is neither a draft nor a prerelease and contains twelve assets: five archives, five SBOMs, `checksums.txt`, and `checksums.txt.bundle`.
- Cosign authenticated `checksums.txt` with certificate identity `https://github.com/granitebps/threads-mcp/.github/workflows/release.yml@refs/tags/v1.0.1` and issuer `https://token.actions.githubusercontent.com`.
- All ten authenticated checksums and all twelve GitHub asset digests matched. Every archive contained the expected binary, license, README, and third-party notices. The five SBOMs identify SPDX version 2.3 and contain package data.
- The downloaded macOS ARM64 binary initialized through MCP, listed all six tools, reported server version `1.0.1` and provider version `v0.1.1`, and passed all five live public Threads reads.
- A clean public `go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.1` succeeded through the Go proxy. The installed binary passed the same MCP metadata check. The proxy maps v1.0.1 to candidate commit `c8ef81561e799fe5194d3c04947f70ae0d635069`.

Version v1.0.1 is complete. Never move the tag or replace published assets with
different builds under the same version.
