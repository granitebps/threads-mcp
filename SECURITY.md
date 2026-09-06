# Security policy

## Supported versions

Security fixes are released on the current v1 line. Upgrade to the latest v1
release to receive them. Development snapshots and versions older than v1.0.0
do not receive security updates.

The v1 support period begins when v1.0.0 is published. Until then, this project
has no supported public release.

## Report a vulnerability

Do not open a public issue or discussion for a suspected vulnerability.

Use GitHub's
[private vulnerability report](https://github.com/granitebps/threads-mcp/security/advisories/new).
This form becomes available after the repository is public and its maintainer
enables private vulnerability reporting. The maintainer must enable and test it
immediately after changing the repository to public, before announcing or
tagging v1.0.0. If the form is unavailable on the public repository, open an
issue that says only that private reporting is unavailable. Do not include
vulnerability details in that issue.

Include only the information needed to investigate:

- the affected version or commit;
- the affected component and MCP tool, if applicable;
- reproducible steps or a minimal proof of concept;
- the security impact and required conditions;
- any suggested fix or mitigation.

Remove credentials, cookies, tokens, private Threads content, and sensitive
local paths. This server does not require Threads credentials, so a report
should not contain them.

The maintainer will use the private advisory to investigate, coordinate a fix,
and publish an advisory when appropriate. Please keep the report private until
the maintainer and reporter agree that disclosure is safe.

## What belongs here

Report vulnerabilities in this repository's source, packaged executables,
dependencies, build or release process, and MCP behavior. Ordinary bugs,
Threads availability problems, incomplete public results, and feature requests
belong in the
[public issue tracker](https://github.com/granitebps/threads-mcp/issues).
