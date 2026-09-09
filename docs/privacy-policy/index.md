---
title: Privacy policy
permalink: /privacy-policy/
---

# Privacy policy for threads-mcp

Effective date: September 9, 2026

GBPS operates and maintains `threads-mcp`, an open-source, locally run tool for
accessing Threads content through the Model Context Protocol. This policy
explains how `threads-mcp` processes information when a user connects a Threads
account or requests Threads content.

Current public releases use anonymous access to public Threads content and do
not accept access tokens. This policy also covers official Threads API access
enabled in a development build or a future release.

## Information processed

When official Threads API access is enabled and configured by a user,
`threads-mcp` may process:

- Threads account information made available by the Threads API, such as a
  Threads user ID, username, display name, and profile information;
- public Threads posts, replies, profile information, and keyword-search
  results;
- search terms and other requests entered by the user; and
- a Threads access token supplied through the user's local configuration.

`threads-mcp` does not ask for or process a user's Threads password.

## How information is used

`threads-mcp` uses this information only to:

- authenticate requests that the user makes to the official Threads API;
- retrieve Threads content requested by the user;
- return the requested content to the user's MCP client; and
- report errors and operational information needed to run the tool.

GBPS does not use Threads data for advertising, sell it, or use it to build
profiles about users.

## Local processing and storage

`threads-mcp` runs on the user's computer. GBPS does not operate a hosted
`threads-mcp` service and does not receive or store access tokens, search terms,
or Threads API responses produced by a user's local installation.

An access token configured for `threads-mcp` remains in the user's local
environment or local application storage. Requests containing that token are
sent to Meta's Threads API. Results are returned to the MCP client selected by
the user. That MCP client may keep conversation history or logs under its own
privacy and retention practices.

## Sharing

`threads-mcp` sends requests to Meta when it accesses the official Threads API.
Meta processes those requests under its own terms and privacy policy. The tool
also provides results to the MCP client chosen by the user.

GBPS does not sell or rent personal information. Because GBPS does not operate
a hosted service for `threads-mcp`, GBPS does not disclose locally processed
Threads data to advertisers or data brokers.

## Retention and deletion

GBPS does not retain data produced by a local `threads-mcp` installation.
Locally stored tokens, configuration, logs, or MCP conversation history remain
until the user deletes them or the software that stores them removes them.

To stop `threads-mcp` from accessing a Threads account, the user can:

1. revoke the application's access from the app and website permissions in the
   user's Threads or Meta account settings;
2. remove the Threads access token and related `threads-mcp` configuration from
   the user's computer; and
3. delete any results or conversation history retained by the user's MCP
   client.

If a user believes GBPS has received personal information through a support
request or another direct communication, the user may request access or
deletion by emailing [granitebagas28@gmail.com](mailto:granitebagas28@gmail.com).

## Security

Users are responsible for protecting Threads access tokens and the computers
and MCP clients where they run `threads-mcp`. Access tokens should not be added
to source control, issue reports, logs, or other public locations.

## Children's privacy

`threads-mcp` is not directed to children under 13, and GBPS does not knowingly
collect personal information from children through a hosted `threads-mcp`
service.

## Changes to this policy

GBPS may update this policy when `threads-mcp` changes or when legal or platform
requirements change. The effective date at the top of this page will identify
the latest revision.

## Contact

For privacy questions or deletion requests, contact GBPS at
[granitebagas28@gmail.com](mailto:granitebagas28@gmail.com).
