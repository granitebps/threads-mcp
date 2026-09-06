# v1.0.0 publishing checklist

Status: preparation in progress. Publishing is not authorized.

This checklist records the work needed to prepare and publish the first public
release. Creating this file does not authorize implementation, commits, pushes,
repository-setting changes, tags, or publication. Obtain approval for each work
batch before making changes.

Follow the [release guide](../RELEASE.md) for the reusable procedure. This file
tracks v1.0.0 work and evidence; it is not a second set of release instructions.

## Agreed release policy

The owner decides when to release. After checks and explicit approval for the
version and exact commit, a version-tag push starts automated publishing.
Normal code changes run checks only. Verify the public installation afterward.

Approval must precede the tag push because a public Go version tag can already
make the source installable. The publishing job must also require successful
checks for that same commit. This safeguard is implemented in the local workflow
files; GitHub execution remains unverified.

This policy is agreed; launch approval is still not granted. Applying the same
policy to twitter-mcp is a separate repository change, not completed here.

## Release scope

Keep the existing local stdio MCP server and its six tools:

- `search_posts`
- `get_post`
- `get_post_replies`
- `get_profile`
- `get_profile_posts`
- `get_server_info`

Support use from a local repository and as an installed executable launched by an
MCP client. Interactive CLI subcommands, Homebrew, authentication, and
`get_profile_replies` remain outside this release's scope. Unofficial Threads
access does not guarantee availability or complete results.

## Prepare while the repository is private

- [x] Fix the live smoke test to reject MCP results marked `IsError` and validate expected result data. Allow legitimate empty lists without accepting an error response as success.
- [x] Fix server and provider version reporting, including binaries installed with `go install` without GoReleaser build flags.
- [x] Document the shared release policy and add a human/AI release guide linked from the README.
- [x] Require CI checks and GoReleaser configuration validation to pass for the exact tagged commit before the release workflow can publish. Keep checkouts pinned to that commit and preserve checks on normal code changes.
- [ ] Validate the workflow's success, failure, and skipped-check behavior without publishing. Record what was verified locally and what still requires GitHub execution.
- [x] Review publishing permissions and pin third-party actions to verified commit IDs. Keep check jobs read-only.
- [x] Document the V1 compatibility promise for tool names, inputs, outputs, and error codes.
- [x] Update the README with the correct Go minimum and Windows `.exe` instructions.
- [x] Verify the README covers local-repository and installed-command setup.
- [x] Add version-pinned installation instructions using `@v1.0.0`, clearly marked as available only after publication.
- [x] Add user-facing release download, checksum, and signature verification instructions to the README. The maintainer procedure is in RELEASE.md.
- [x] Document known limitations, troubleshooting, and support.
- [x] Add `SECURITY.md` with a private vulnerability-reporting route.
- [x] Add `CONTRIBUTING.md`.
- [x] Add dependency-update configuration for Go modules and GitHub Actions.
- [x] Review tracked files and Git history for secrets or unintended personal information. Do not put discovered secrets in this checklist or logs.
- [x] Prepare the [`v1.0.0` release notes](release-notes-v1.0.0.md).
- [x] Record that Scoop is deferred and do not advertise it as available.
- [x] Ensure Scoop upload stays disabled unless separately approved, including when a bucket token is present.
- [x] Confirm the Go SDK supports the documented MCP protocol version and review the threads-cli version and license notices.

## Verify the release candidate without publishing

Check items only against the final candidate commit. Earlier successful runs are
background evidence, not proof that a changed candidate passes. Record dates,
platforms, commands or workflow links, and actual results below.

- [ ] Pass Linux, macOS, and Windows CI.
- [ ] Pass lint, formatting, dependency checks, and a fresh vulnerability scan.
- [ ] Pass the strengthened live Threads smoke test.
- [ ] Generate snapshot archives, checksums, and SBOMs without publishing or uploading to Scoop.
- [ ] Extract and test packaged binaries in clean locations. Record which platforms were executed and which were only cross-compiled.
- [ ] Verify installed-command and local-repository use through an MCP client. Record client checks for Codex, Claude, Cursor, and OpenCode, or obtain an explicit scope decision for any untested client.
- [ ] Record the exact candidate commit and verification results.
- [ ] Review the signing configuration. Leave actual signature verification pending until signing has run; an unsigned snapshot does not verify signing.
- [ ] Review remaining failures, limitations, or deferred items before requesting launch approval.

## Cross-project follow-up

- [ ] In a separately approved twitter-mcp task, adapt its release workflow and guide to the same owner-approved, tag-triggered policy using npm publishing. Record that repository's verification separately.

This follow-up does not authorize edits to twitter-mcp and is not a threads-mcp
release verification result.

## Launch approval boundary

Stop here until the owner explicitly authorizes public visibility, repository
settings, and the release tag/publishing actions. Record the approved scope and
candidate commit. Do not infer launch approval from passing tests.

- Approval date: not granted.
- Approved by: not recorded.
- Approved candidate commit: not selected.
- Approved actions: none.

## Publish only after explicit approval

- [ ] Change repository visibility to public.
- [ ] Configure repository description, topics, branch protection, and available security controls. Enable and test private vulnerability reporting and security alerts.
- [ ] Confirm `v1.0.0` does not already exist and the candidate commit still matches the verified commit.
- [ ] Confirm the required CI dependency is implemented and the owner's approval still covers this exact tag, commit, and publishing actions before pushing the tag.
- [ ] Create and push `v1.0.0` from the verified commit. This triggers the publishing workflow.
- [ ] Confirm the release workflow succeeds.
- [ ] Verify published archives, checksums, SBOMs, and the Cosign bundle, including the expected signing identity and issuer.
- [ ] Test public `go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.0` from a clean environment.
- [ ] Download a published binary and repeat the real MCP test.
- [ ] Confirm the release page and README installation links work.

Do not move or overwrite a published `v1.0.0` tag to correct a defect. Record the
failure and obtain approval for a corrective release.

## Verification evidence

### Preparation batch 1

Verified on 2026-09-04 with Go 1.26.6 on macOS ARM64, against base commit
`5e78dc9f4bb2d8b94ac5f7585765abf871319edc` plus the uncommitted batch 1 changes.
These results do not replace final-candidate verification.

- Reproduced the weak live-response assertions and incorrect provider version with failing regression tests before fixing them.
- `go test -race -count=1 ./...` passed. This includes offline live-response validation, version metadata fixtures, and real stdio MCP tests for local and release-style builds.
- `go test -tags=live ./internal/provider/threadscli -run TestAnonymousCrawlerPublicReads -count=1 -v` passed in 16.57 seconds. All five data-reading MCP tools returned successful responses with valid payloads. Legitimate empty lists are covered by offline tests.
- `go vet ./...`, `go build ./...`, formatting checks, and `git diff --check` passed.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1 run` passed with zero issues. A fresh vulnerability scan remains part of final-candidate verification.
- `GOOS=windows GOARCH=amd64 go build ./...` passed. Windows binaries were not executed; cross-platform CI remains pending for these changes.
- Local builds report server version `dev` and provider version `v0.1.1`. Release-style builds preserve the supplied server version, commit, and build date while reporting the dependency version separately.
- Versioned-install metadata is covered by unit fixtures. Public `go install ...@v1.0.0` remains untested until the release exists. Missing build dates remain `unknown`; a commit timestamp is not a build timestamp.
- No commits, pushes, tags, publication, or repository visibility changes were performed.

### Documentation batch

- Added `RELEASE.md`, linked it and this checklist from the README, and recorded the agreed release policy and remaining automation work.
- Checked local Markdown link targets, balanced code fences, and shell-example syntax. Checked snapshot options against the installed GoReleaser v2.18.0 help.
- No release commands were executed. This documentation review does not verify CI execution, packaged binaries, signatures, or publication.
- No workflow changes, twitter-mcp edits, commits, pushes, tags, or publishing were performed in this batch.

### Release workflow preparation

Reviewed on 2026-09-04 against base commit
`5e78dc9f4bb2d8b94ac5f7585765abf871319edc` plus the uncommitted changes.

- Updated the CI workflow to allow reuse by the tag-triggered release workflow. Normal main-branch and pull-request CI triggers remain unchanged.
- Added required CI and GoReleaser configuration-check dependencies to publishing. Both results must be successful, and publishing is restricted to version-tag push events.
- Pinned every checkout in these two workflows to `github.sha`. No action versions, publishing permissions, signing settings, or distribution targets changed in this batch.
- Local structured workflow checks detected the missing dependencies, skipped tag validation, and unpinned checkouts before editing; the same checks passed afterward. These are configuration checks, not GitHub execution tests.
- `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12` passed with no diagnostics. `goreleaser check` validated the release configuration. `git diff --check` passed.
- Reviewed the publishing condition: a tag push with both checks successful is eligible; failed, cancelled, or skipped checks are not; branch pushes and pull requests cannot publish. Actual GitHub scheduling and failure/skip behavior remain unverified, so the workflow-execution checklist item stays open.
- No commits, pushes, tags, publishing runs, or twitter-mcp changes were performed. These results apply to the uncommitted workflow changes, not a final release candidate.

### Workflow security preparation

Reviewed on 2026-09-05 against base commit
`5e78dc9f4bb2d8b94ac5f7585765abf871319edc` plus the uncommitted changes.

- Resolved each configured action version from its official GitHub repository and replaced the moving version tag with that commit ID. Kept the version in a comment for update tools and reviewers.
- Pinned checkout v5 to `fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09`, setup-go v6 to `924ae3a1cded613372ab5595356fb5720e22ba16`, golangci-lint-action v8 to `4afd733a84b1f43292c63897423277bb7f4313a9`, goreleaser-action v6 to `e435ccd777264be153ace6237001ef4d979d3a7a`, cosign-installer v3.10.0 to `d7543c93d881b35a8faa02e8e3605f69b7a1ce62`, and sbom-action v0.20.6 to `f8bdd1d8ac5e901a77a92f111440fdb1b593736b`.
- Confirmed CI, configuration checks, and live smoke remain `contents: read`. Only the publishing job has `contents: write` and `id-token: write`, required for GitHub Release uploads and keyless Cosign signing. The job receives the built-in `GITHUB_TOKEN` and no Scoop token.
- Added `--skip=scoop` to the automated release command and removed the Scoop credential from its environment. This prevents an existing secret from enabling the deferred upload.
- The workflow-policy check failed before the edit because actions used moving tags. The same check passed afterward, confirming the approved action commits, limited publishing permissions, and disabled Scoop upload. Actionlint and `goreleaser check` also passed. GitHub execution remains pending.

### GitHub workflow execution and Node.js runtime review

Reviewed on 2026-09-06 against commit
`d6e4f16e45b71436528d49beba954771ee4ea42d` plus the uncommitted action-pin
update.

- GitHub CI run [34010686450](https://github.com/granitebps/threads-mcp/actions/runs/34010686450) passed on the pushed commit. Linux and macOS passed formatting, vet, build, and race-enabled tests. Windows passed build and tests. The separate lint, module-tidiness, and fresh vulnerability-scan jobs also passed.
- GitHub Release run [34010686609](https://github.com/granitebps/threads-mcp/actions/runs/34010686609) passed its GoReleaser configuration check. Its reusable CI and publishing jobs were skipped for the `main` branch push, and no release was created.
- These runs verify successful CI and skipped publishing for a branch push. They do not exercise GitHub's behavior when a required dependency fails, so the full workflow-behavior checklist item remains open.
- The CI run warned that `golangci/golangci-lint-action` v8 declares the deprecated Node.js 20 action runtime. GitHub forced that action onto Node.js 24 for the run.
- Updated the pinned action to official release v9.3.0 at commit `ba0d7d2ec06a0ea1cb5fa41b2e4a3ab91d21278a`. Its action metadata declares Node.js 24, and its compatibility guidance supports the configured golangci-lint v2 series. The lint binary remains pinned to v2.13.1.
- GitHub has not executed the updated action pin yet. A later push must pass CI without the Node.js 20 warning before this update becomes final-candidate evidence.

### Compatibility documentation

Reviewed on 2026-09-05 against base commit
`5e78dc9f4bb2d8b94ac5f7585765abf871319edc` plus the uncommitted changes.

- Added `docs/v1-compatibility.md` from the implemented MCP registrations, JSON schemas, domain models, error mapping, and server tests. Linked it from the README and release guide.
- The promise starts when v1.0.0 is published. It protects existing tool names, accepted input forms, success envelopes, field types, and error meanings throughout v1.x.
- The document permits additive optional fields, tools, error codes, and safely handled enum values. It excludes upstream Threads availability, completeness, ordering, omitted values, media lifetime, and exact error prose.
- Local Markdown links and code fences passed validation. `go test -count=1 ./internal/server ./internal/domain ./internal/identifier` passed, and `git diff --check` reported no errors. No tool behavior changed in this batch.

### Go and Windows documentation

Reviewed on 2026-09-05 against base commit
`5e78dc9f4bb2d8b94ac5f7585765abf871319edc` plus the uncommitted changes.

- Corrected the README minimum from Go 1.26 to the `go.mod` requirement, Go 1.26.6.
- Added Windows PowerShell build and verification commands, required `.exe` guidance, a Windows configuration-path example, and a Codex CLI example.
- Replaced the unavailable Scoop installation example with an explicit v1.0.0 deferral. Local builds, `go install`, and future GitHub Release archives remain the supported alternatives.
- The README requirement was checked against `go.mod`; the Scoop-installation example is absent; local links, code fences, and `git diff --check` passed. A Windows amd64 `.exe` cross-build passed. The executable was not run on Windows, and no client configuration, workflow, or distribution behavior changed in this batch.

### Local and installed command documentation

Reviewed on 2026-09-05 against base commit
`5e78dc9f4bb2d8b94ac5f7585765abf871319edc` plus the uncommitted changes.

- Split local-repository building from command installation and explained how to locate the Go-installed executable on macOS, Linux, and Windows.
- Added the exact `go install github.com/granitebps/threads-mcp/cmd/threads-mcp@v1.0.0` command, marked as available only after the tag and public repository exist. Documented `@latest` as an optional, less reproducible alternative.
- Clarified that graphical clients should use the absolute local-build or installed-command path.
- Built the repository command and ran `GOBIN=<temporary-directory> go install ./cmd/threads-mcp` into separate temporary locations. Both executables passed a real stdio MCP initialization, returned all six expected tools, and reported server version `dev` plus provider version `v0.1.1` through `get_server_info`.
- README shell examples passed syntax checks; local links, code fences, and `git diff --check` passed. PowerShell was unavailable, so the Windows commands were inspected but not executed in this batch.
- Public version-pinned installation remains a post-publication check and stays unchecked in the launch section. These local startup checks do not satisfy the final-candidate client matrix item.

### Release download and verification documentation

Reviewed on 2026-09-05 against base commit
`5e78dc9f4bb2d8b94ac5f7585765abf871319edc` plus the uncommitted changes.

- Added the v1.0.0 archive names for every configured operating system and architecture, extraction instructions, and a warning that the executable is an MCP stdio server rather than an interactive command.
- Added checksum verification for macOS, Linux, and Windows. The instructions first authenticate `checksums.txt` with its Cosign bundle and pin verification to this repository's release workflow, the exact `v1.0.0` tag, and GitHub Actions' OIDC issuer.
- Required Cosign v3.1.3 or newer. The release workflow already pins v3.1.3,
  which contains the [upstream fix](https://github.com/sigstore/cosign/security/advisories/GHSA-fx35-mq7g-6g98)
  for the legacy-bundle identity-verification vulnerability affecting earlier v3 releases.
- Checked archive names against the GoReleaser v2 default template and the repository's generated snapshot artifacts. Checked the `verify-blob` flags against Cosign v3.1.3 help and current Sigstore documentation.
- Ran the documented POSIX checksum comparison against the existing macOS ARM64 snapshot archive and extracted it into a new temporary directory. The checksum matched, and the archive contained the executable, license, README, and third-party notices.
- The README now states that Cosign verification does not provide Apple notarization or Windows Authenticode signing. Platform-native signing remains a known limitation to document and review.
- Published assets and signatures do not exist yet, so the commands could not be run end to end. That verification remains in the launch checklist.
- No commits, pushes, tags, publishing, or repository visibility changes were performed.

### Crawler response resilience

Verified on 2026-09-05 with Go 1.26.6 on macOS ARM64, against base commit
`5e78dc9f4bb2d8b94ac5f7585765abf871319edc` plus the uncommitted changes.

- The pre-commit live smoke exposed intermittent incomplete Threads responses. Identical crawler requests returned HTTP 200, but five of six sampled profile pages omitted the structured profile data. The pinned `threads-cli` dependency consequently returned a false `not found` result. A separate three-attempt CLI sample succeeded twice and failed once.
- Added fresh semantic retries for incomplete profile, profile-post, post, post-reply, and search-page responses. Normal successful reads still stop after one attempt. Real HTTP not-found responses are not retried. Exhausted incomplete scalar and search reads return retryable `UPSTREAM_CHANGED` instead of `NOT_FOUND`.
- Disabled the dependency's pre-parse HTML cache for these reads. Otherwise an incomplete HTTP 200 response could be reused for the full cache lifetime and defeat every retry.
- Added regression coverage for recovery, exhausted retries, real not-found behavior, empty profile-post windows, post replies, search results, and cache bypass. The tests failed against the old behavior before the implementation and passed afterward.
- Updated the live smoke to exercise the resilient provider and MCP paths rather than calling the raw dependency before the provider. Three consecutive live runs then passed all five data-reading MCP tools in 18.97, 14.41, and 13.89 seconds.
- Threads can still return incomplete responses after every attempt. Repeat the live smoke against the final candidate and keep this risk in the public limitations.

### Public limitations, troubleshooting, and support

Reviewed on 2026-09-05 against commit
`a1e2901204656b524bdf12da037cf0c470c11f12` plus the uncommitted documentation
changes.

- Added README guidance for unofficial anonymous access, incomplete-response retries, limited result windows, unsupported private content, omitted fields, expiring media URLs, platform-native signing, and deferred package managers.
- Added startup checks for executable paths, Windows `.exe`, execute permissions, MCP stdio behavior, client restarts, and operating-system warnings.
- Documented every v1 structured error code with a concrete user action. The guidance treats each response's `retryable` field as authoritative and explains that a successful empty window does not prove absence.
- Added a GitHub Issues support route and a diagnostic checklist. The README tells users not to disclose credentials, cookies, tokens, private content, or sensitive paths.
- The README sends suspected vulnerabilities to the security policy instead of the public issue tracker.
- Checked the guidance against the implemented configuration, server error serialization, compatibility promise, and crawler retry behavior.

### Security policy

Reviewed on 2026-09-05 against commit
`a1e2901204656b524bdf12da037cf0c470c11f12` plus the uncommitted documentation
changes.

- Added `SECURITY.md` with the v1 support policy, report contents, disclosure expectations, and a boundary between security reports and ordinary support.
- Routed suspected vulnerabilities to GitHub's private vulnerability-reporting form. The repository is still private, so enabling and testing that form remains an explicit public-launch checklist item.
- Linked the policy from the README and removed the temporary pre-launch wording.
- No response-time guarantee or email address was invented. If GitHub's private form is unexpectedly unavailable after launch, the policy permits a detail-free public issue that reports only the broken private route.

### Contributor guide

Reviewed on 2026-09-05 against commit
`a1e2901204656b524bdf12da037cf0c470c11f12` plus the uncommitted documentation
changes.

- Added `CONTRIBUTING.md` with the supported v1 scope, Go setup, deterministic checks, live-test boundary, compatibility review, dependency hygiene, and pull request expectations.
- Matched commands and pinned tool versions to the current CI and release guide. Windows guidance uses the repository's non-race test path.
- Kept live Threads access outside deterministic CI and told contributors to report, not hide, upstream failures.
- Linked the guide from the README. Security reports remain routed to `SECURITY.md`.
- Did not invent a code of conduct, contributor license agreement, commit-signing rule, response-time promise, or release authority for contributors.

### Dependency updates

Reviewed on 2026-09-05 against commit
`a1e2901204656b524bdf12da037cf0c470c11f12` plus the uncommitted preparation
changes.

- Added `.github/dependabot.yml` with weekly version checks for the root Go module and GitHub Actions.
- Kept dependency updates in separate pull requests so each upgrade has its own CI result and review. Dependabot's default limit of five open version-update pull requests per ecosystem remains unchanged.
- Used `chore(deps)`-style commit and pull request titles to match the repository's existing commit style.
- Added no auto-merge, assignee, private registry, credential, ignore rule, or target-branch configuration.
- The current workflow actions remain pinned to commit IDs with same-line version comments, a format GitHub documents as supported by Dependabot.
- Parsed and checked the file locally. GitHub has not processed the configuration yet.

### Secret and personal-information review

Reviewed on 2026-09-06 against commit
`a1e2901204656b524bdf12da037cf0c470c11f12` plus the uncommitted preparation
changes.

- Ran a metadata-only pattern scan over all 63 reachable historical text blobs, all three commit messages, and 50 tracked or untracked non-ignored worktree files. No historical binary blobs were skipped.
- Found no private-key headers, common provider token formats, assigned credentials, authenticated URLs, authorization values, JWTs, home-directory paths, or macOS temporary paths. Project-authored content contains no email addresses. The complete third-party license bundle contains two upstream attribution email addresses copied verbatim from required license text; both match their sources and are intentional.
- Confirmed that the configured origin URL contains no embedded credential.
- Found one distinct non-GitHub-noreply commit identity, used by all three commits. The owner explicitly accepted that identity for the future public history on 2026-09-05. Its name and email address are not copied into this checklist.
- Pattern matching cannot prove that no secret exists. Repeat the review against the final release candidate and enable the available GitHub security controls at launch.

### SDK, provider dependency, and license review

Reviewed on 2026-09-06 against commit
`a1e2901204656b524bdf12da037cf0c470c11f12` plus the uncommitted preparation
changes.

- Confirmed that `github.com/modelcontextprotocol/go-sdk v1.7.0` is the latest stable release. The only newer versions are v1.8.0 prereleases, so no SDK upgrade was made.
- Confirmed from the SDK release notes and source that v1.7.0 supports MCP protocol version `2026-07-28` and can negotiate supported older versions. The compatibility document now distinguishes the SDK's newest supported version from a session's negotiated version.
- Confirmed that `github.com/tamnd/threads-cli v0.1.1` is the latest tag. The only newer upstream commit changes release-workflow action pins, not provider behavior, so the project remains on the stable tag instead of using a pseudo-version.
- Verified that threads-cli v0.1.1 uses Apache License 2.0, including its exact upstream copyright line, and has no root `NOTICE` file.
- Replaced the abbreviated third-party notice with exact upstream license text for all 19 modules compiled into at least one supported release target: Linux and macOS on `amd64` and `arm64`, plus Windows on `amd64`, with `CGO_ENABLED=0`.
- Verified every bundled license section byte-for-byte against the corresponding module cache file. Every compiled module had a root license file.
- Confirmed that GoReleaser already includes `THIRD_PARTY_NOTICES.md` in every archive, so no release configuration change was needed.

### Release notes

Reviewed on 2026-09-05 against commit
`a1e2901204656b524bdf12da037cf0c470c11f12` plus the uncommitted preparation
changes.

- Added `docs/release-notes-v1.0.0.md` as the canonical GitHub release-body draft.
- Covered the six tools, version-pinned Go installation, supported release archives, client setup, download verification, access limits, v1 compatibility, support, and private security reporting.
- Described v1.0.0 as the first stable release instead of inventing migration steps from an earlier release.
- Kept Homebrew, Scoop, authentication, private content, historical pagination, and `get_profile_replies` outside the advertised release.
- Added no candidate commit, test result, publication date, or success claim. Those facts remain pending final-candidate verification and launch approval.

### Final candidate and launch

- Candidate commit: not selected.
- Final verification date: not recorded.
- Scoop decision: deferred unless separately approved and configured.

| Check | Commit or artifact | Date and environment | Evidence and result |
| --- | --- | --- | --- |
| Cross-platform CI | Pending | Pending | Pending |
| Lint, dependency, and security checks | Pending | Pending | Pending |
| Strengthened live smoke | Pending | Pending | Pending |
| Snapshot archives, checksums, and SBOMs | Pending | Pending | Pending |
| Clean packaged-binary tests | Pending | Pending | Pending |
| Local-repository MCP test | Pending | Pending | Pending |
| Installed-command MCP test | Pending | Pending | Pending |
| Published signatures and artifacts | Pending launch approval | Pending | Not performed |
| Public Go installation and download test | Pending launch approval | Pending | Not performed |

## Deferred items and remaining risks

Record each deferred item with its reason and whether it blocks launch. Do not
mark an unverified check complete because it was deferred.

- Threads may restrict anonymous access or change its public pages after release.
- Batch 1 strengthened the live assertions and passed locally. Repeat the smoke test against the final candidate before launch.
- Intermittent incomplete HTTP 200 responses are retried with fresh reads. Exhaustion can still prevent a public read, so this mitigation needs final-candidate and post-release monitoring.
- Tagged signing and public installation require verification during launch.
