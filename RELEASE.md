# Release guide

Use this guide to prepare and publish a threads-mcp release. It is written for
maintainers and AI agents. Use the checklist for the version being prepared.
The [v1.0.1 checklist](docs/release-v1.0.1.md) records the completed recovery
release, and the [v1.0.0 checklist](docs/release-v1.0.0.md) records the first tag
and its incomplete binary release.

Reading this guide or completing its checks does not authorize publication.
Obtain explicit approval for the version, exact commit, and publishing actions.

## Shared release policy

Use the same policy across the MCP projects, with package-specific build steps:

1. Normal code changes run checks without publishing.
2. Prepare and verify the exact release candidate, including installation and live MCP checks.
3. The owner approves the version and commit before its release tag is pushed.
4. The tag starts automation. Required checks must pass before the publishing job starts.
5. Verify the public installation and record the result.

For threads-mcp, publish Go binaries through GoReleaser and support versioned
`go install`. A Node.js MCP can follow the same policy while publishing to npm.
Do not copy Go commands into a Node.js release process. Adopting this policy in
another repository requires a separate approved change.

A public Go version tag can make source available to Go tools before a GitHub
binary release finishes. Approval must therefore come before the tag push, not
just before publishing the release page. Never move a published version tag to
different code. See the [Go publishing guidance](https://go.dev/doc/modules/publishing).

## Current implementation and scope

The release workflow separates read-only preparation from tag-only publishing:

- [CI](.github/workflows/ci.yml) runs on main-branch pushes and pull requests and can be called by the release workflow. It checks Linux, macOS, Windows, formatting, build, tests, dependencies, lint, and vulnerabilities.
- [Release automation](.github/workflows/release.yml) installs and exercises the pinned Cosign and Syft tools, checks GoReleaser configuration, and builds a complete unsigned snapshot on main-branch pushes, pull requests, and `v*` tag pushes. Tag pushes also call the same-commit CI workflow. The publishing job requires both CI and the release preflight to succeed. CI and release checkouts use the triggering commit ID, not a moving branch or tag. External actions are pinned to verified commit IDs with their release versions in comments.
- [Live smoke testing](.github/workflows/live-smoke.yml) is separate from CI. A scheduled run on another commit is not evidence for the release candidate.
- [GoReleaser configuration](.goreleaser.yaml) defines downloads, checksums, dependency inventories called SBOMs, and Cosign signing. Configuration alone does not prove that published files or signatures work.
- Scoop is deferred. The release command explicitly skips Scoop and does not pass its token, so an existing repository secret cannot enable upload. Enabling Scoop later requires a separately approved workflow change.
- Homebrew and MCP Registry publishing are not configured. Ordinary Go modules and binary archives are not listed as supported Registry package types. Registry publication needs a separate packaging decision. Check the [current package types](https://modelcontextprotocol.io/registry/package-types) before adding it.

The tag-only path still provides the first real test of GitHub OIDC signing and
release upload for a version. Deferred distribution channels do not block a
release, provided they stay disabled and are not advertised as available.

Stop if this guide and the checked-out workflows disagree. Resolve the difference
before publishing; do not treat the guide as proof that a safeguard exists.

## 1. Prepare the candidate

Before running checks:

- Read the repository instructions and version-specific checklist.
- Confirm `docs/release-notes-VERSION_TAG.md` exists for the exact tag, and review its version, supported platforms, installation instructions, license notices, and known limitations.
- Compare tool schemas and error behavior with the [v1 compatibility promise](docs/v1-compatibility.md). Treat an incompatible change as a release blocker unless the approved version and migration plan account for it.
- Review tracked files and history for secrets or unintended personal information before making the repository public. Do not copy discovered secrets into logs or release notes.
- Use a clean checkout of the intended commit. Preserve unrelated work; do not reset or discard it to make checks pass.
- Use the Go version required by `go.mod`, currently Go 1.26.6 or newer. Match the tool versions in the workflows. The current release pins are GoReleaser v2.18.0, Cosign Installer v4.1.2, Cosign v3.1.3, and Syft v1.51.0.
- Review publishing permissions and credentials. The workflows default to read-only repository contents. Only the publishing job has `contents: write` for GitHub Release uploads and `id-token: write` for keyless Cosign signing. It receives the built-in `GITHUB_TOKEN`; Scoop credentials are not passed. External actions use verified commit IDs. See [GitHub's security guidance](https://docs.github.com/en/actions/reference/security/secure-use).

The command examples below use a POSIX shell and start in the repository root.
They are instructions for an approved work session, not commands an agent should
run merely because it read this file.

```bash
git status --short
git rev-parse HEAD
```

Record the full commit ID. A release candidate must have no uncommitted changes.
If a check requires a code or configuration change, obtain approval, make the
change, commit through the approved process, and verify the new candidate.

## 2. Verify without publishing

### Code and dependency checks

Run these checks on macOS or Linux. Windows CI uses `go test -count=1 ./...`
instead of the race-enabled command.

```bash
gofmt -l .
go vet ./...
go build ./...
go test -race -count=1 ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1 run
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
```

Formatting must list no files. All other checks must succeed. Use the workflow
pins if they have changed since this guide was written.

The dependency check runs `go mod tidy` and then checks for changes to `go.mod`
and `go.sum`. It can modify files, so run it in an approved disposable checkout
or use the candidate's CI result. Any resulting dependency change needs review
and a new candidate; do not publish from that changed checkout.

Confirm that Linux, macOS, and Windows CI passed for the exact candidate commit.
Record workflow links and commit IDs. Cross-compilation is not a substitute for
executing tests on the target operating system.

### Live MCP checks

```bash
go test -tags=live ./internal/provider/threadscli -run TestAnonymousCrawlerPublicReads -count=1 -v
```

The probe checks all five data-reading tools through MCP. Error responses must
fail; valid empty lists are allowed. Do not weaken assertions to work around
Threads access failures. Record the failure and stop release preparation until
it is understood and resolved or explicitly reviewed by the owner.

### Publishing-toolchain and package rehearsal

Use a fresh disposable checkout of the candidate, with full Git history and
the pinned GoReleaser and Syft tools installed. GoReleaser's pre-build hook runs
`go mod tidy`, so check afterward that tracked files did not change.

The workflow's read-only `check` job must install the exact Cosign release,
print the Cosign and Syft versions, sign and verify a temporary payload with an
ephemeral local key, and run the complete snapshot build. This detects installer,
command-line, packaging, and SBOM failures before a release tag exists. It does
not request an OIDC token or upload the temporary signature to a transparency
log.

```bash
goreleaser check
goreleaser release --snapshot --clean --skip=sign,scoop
git diff --exit-code
```

Snapshot mode does not publish. Signing and Scoop are explicitly skipped here.
Do not substitute a normal release command. `--clean` deletes only the disposable
checkout's generated `dist` directory before rebuilding it.

Inspect the generated files, not just the command's exit status:

- Confirm archives for Linux and macOS on amd64 and arm64, plus Windows amd64.
- Confirm the license, README, third-party notices, checksums, and SBOMs are present as configured.
- Verify checksums and extract the archive for the current operating system into a clean directory.
- Connect an MCP client to the extracted executable using its absolute path. Windows uses `threads-mcp.exe`. Maintainers can set `THREADS_MCP_TEST_BINARY` to that absolute path and `THREADS_MCP_TEST_VERSION` to the expected build version, then run `go test -count=1 ./tests -run TestPackagedBinaryHandshakeAndTools -v`.
- List all six tools and call `get_server_info`. Verify the reported version against the snapshot's generated metadata and the provider version against the pinned dependency.
- Test local-repository use and an installed executable separately, following the [client setup instructions](README.md#client-configuration). Verify successful data-reading calls and no non-protocol startup output on stdout.

Record which platforms were executed and which were only built. An unsigned
snapshot does not verify signing. Public `go install ...@VERSION` and published
signature verification remain pending until publication.

## 3. Record approval before publishing

In the version-specific checklist, record:

- Release version and full candidate commit ID.
- Verification date, operating systems, workflow links, and actual results.
- Remaining limitations and any explicitly accepted exceptions.
- Approver, approval date, and exact authorized actions.

Authorization must name tag creation and push, automated publication, and any
required visibility or repository-setting changes. Approval to prepare files,
run tests, or commit changes is not publishing approval. For AI agents, stop and
ask the owner when the approved scope is missing or unclear. Confirm immediately
before an irreversible publishing action.

For the first public release, review the whole repository before changing its
visibility. Recheck that the candidate and workflow safeguards still match the
approval. Do not launch while the blockers above remain open.

## 4. Start the approved release

Only after approval, set the approved tag and full commit ID. These placeholders
must be replaced with the values recorded in the checklist:

```bash
release_tag=APPROVED_VERSION_TAG
candidate_commit=APPROVED_FULL_COMMIT_ID
git status --short
git rev-parse HEAD
git tag --list "$release_tag"
git ls-remote --tags origin "refs/tags/$release_tag" "refs/tags/$release_tag^{}"
```

The checkout must be clean, HEAD must equal the approved commit, and the local
and remote tag checks must return no matching tag. A network or authentication
error does not mean a tag is absent. Confirm the origin repository is correct.
Stop if any result differs from the approved state.

Then, with the owner's authorization reconfirmed:

```bash
git tag -a "$release_tag" "$candidate_commit" -m "Release $release_tag"
git push origin "refs/tags/$release_tag"
```

The push starts the release workflow. Watch the run for this tag and commit.
Do not also run GoReleaser publishing locally or create a second release path.
The required checks must pass before the workflow builds and publishes the
release downloads. A failed, cancelled, or skipped required check must prevent
publishing. Never bypass this dependency to work around a failed check.

The workflow loads the GitHub Release body from
`docs/release-notes-VERSION_TAG.md`. A missing version-specific file must fail
before GoReleaser runs.

## 5. Verify the public release

- Confirm the workflow succeeded and the release tag resolves to the approved commit.
- Download the expected archives, `checksums.txt`, its Cosign bundle, and SBOMs from that release.
- Verify the checksum signature using the expected workflow identity and issuer, then verify each archive against the authenticated checksums. Do not accept an arbitrary signing identity supplied alongside a downloaded file.
- For the current tag-triggered workflow, the expected identity is `https://github.com/granitebps/threads-mcp/.github/workflows/release.yml@refs/tags/VERSION_TAG`, with `VERSION_TAG` replaced by the approved tag. The issuer is `https://token.actions.githubusercontent.com`. Recheck these expectations if the workflow changes.
- Inspect the SBOMs and compare their dependencies with the release's source and build metadata.
- Repeat the extracted-binary MCP checks from a clean directory.

From outside the source checkout, use a temporary installation directory to
avoid replacing an existing installation. The tag must already be public:

```bash
release_install_dir=$(mktemp -d)
GOBIN="$release_install_dir" go install "github.com/granitebps/threads-mcp/cmd/threads-mcp@$release_tag"
```

Connect an MCP client to the installed executable and check `get_server_info`.
The server version must match the tag without its leading `v`; the provider
version must identify the dependency, not the server. Repeat the live tool and
installation checks and verify the README download links.

Record the results before calling the release complete. See
[Go's executable installation guidance](https://go.dev/doc/go-get-install-deprecation)
and [GoReleaser signing documentation](https://goreleaser.com/customization/sign/sign/).

## 6. Handle a failed or partial release

Stop before retrying. A timeout or failed workflow can leave a tag, some uploaded
files, or a published package behind. Inspect the remote state first and record
what succeeded. Do not print credentials or sensitive output in the evidence.

Never force-push or move a published version tag, and do not replace published
files with different builds under the same version. A retry that only finishes
missing work on the exact same release requires inspection and explicit approval.
If code or published content needs correction, prepare a new patch version and
obtain approval for that release. Document the affected version and user impact.

If the tag is public but publishing stops before creating a GitHub Release, the
Go source version is already immutable. Fix the workflow on a new candidate and
publish a new patch version; do not reuse or move the failed tag. Normal branch
checks should reproduce the corrected installer and package path before the new
tag is approved.

Do not publish a Registry entry before its underlying package is available and
verified. Registry distribution is currently outside this project's configured
release process; see the [MCP publishing guide](https://modelcontextprotocol.io/registry/github-actions)
if that scope is approved later.
