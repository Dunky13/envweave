# Handoff: MCP best-practices audit and fixes (2026-09-10)

**Branch:** `t3code/audit-mcp-best-practices`
**Ask:** "Is our MCP implementation following best practices?" against the
2026-07-28 spec, the official security tutorial, Snyk, an unofficial mirror,
a YouTube video, and any other source no older than three months.

## What was produced

- `docs/research/mcp-best-practices-sources-2026-09-09.md`: 31 sources with
  dates, primary/secondary flags, every server-side MUST/SHOULD from the
  2026-07-28 spec, go-sdk release and advisory status, and a checklist.
- `docs/reports/mcp-best-practices-audit-2026-09-10.md`: practice-by-practice
  table (Met / Deliberate deviation / Gap) with code and test evidence, two
  findings, and their resolution.

## Verdict

Compliant, and stronger than the SDK default on Origin, Host, cursors, rate
limits, and output sanitisation. Two findings, both fixed in this branch.

## What changed in code

- **A2, uniform 401.** A `tools/call` whose bearer is not live (invalid,
  expired, revoked) now gets the same 401 with `WWW-Authenticate: Bearer` that
  a missing bearer gets. Before, it got HTTP 200 with a tool error. Live but
  refused artifacts (wrong class, no grant) keep the safe tool error.
  `internal/mcpserver/registry.go` marks `domain.ErrUnauthenticated`;
  `internal/mcpserver/handler.go` writes the response.
- **B1, `Mcp-Name` sentinel.** `handler.go` decodes `=?base64?...?=` before
  comparing `Mcp-Name` to the body and writes the decoded value back so the
  pinned SDK validator (which does not decode `Mcp-Name`) agrees.
- `internal/server/metrics.go` reads the tool label after the adapter runs so
  an encoded name lands on the right label.
- `scripts/mcp-public-smoke` now asserts the 401 disposition and has a
  negative test rejecting the old tool-error denial.
- ADR `docs/adr/mcp-server.md` amended by banner; user docs `mcp.mdx` updated.

## Verified

- `go test` for `internal/mcpserver`, `internal/server`, `internal/service`,
  `scripts/mcp-public-smoke`, and `internal/isolation -run MCP`: green.
- `scripts/ci/check-mcp-conformance.sh` (upstream 2026-07-28 suite, Node
  26.7.0): baseline passed.
- `go vet`, `go build ./...`, GOROOT `gofmt -l .`: clean.

## Open items

- File the `Mcp-Name` sentinel omission upstream against
  `modelcontextprotocol/go-sdk` (v1.7.0 decodes only `Mcp-Param-*`).
- When go-sdk v1.8.0 goes stable, re-run conformance and interop per the ADR
  upgrade rule; consider a non-zero `ttlMs` on the static catalog via
  `SetCacheable`.
- The YouTube source was not watched (metadata only). Nothing depends on it.
