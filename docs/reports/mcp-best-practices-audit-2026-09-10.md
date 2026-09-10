# MCP implementation audit against current best practices

**Date:** 2026-09-10
**Scope:** `internal/mcpserver`, its wiring in `internal/app/generation.go` and
`internal/server`, the MCP admission service, deployment fixtures, and CI.
**Sources:** the primary-source catalogue in
[mcp-best-practices-sources-2026-09-09.md](../research/mcp-best-practices-sources-2026-09-09.md)
(source ids S1 to S31 below refer to that file). The user-supplied URLs are S13
(official tutorial), S29 (Snyk), S30 (modelcontextprotocol.info), and S31
(YouTube). Additional in-window sources (published on or after 2026-06-09) are
S14, S15, S16, S20 to S25.
**Method:** every checkable practice was classified as Met (code plus the test
that proves it), Deliberate deviation (locked by
[mcp-server ADR](../adr/mcp-server.md) with a tracking issue), or Gap. Only
Gaps are findings. Both findings were fixed in the same change.

## Verdict

The implementation meets or exceeds every normative MUST in the 2026-07-28
specification that applies to a stateless, tools-only server using its own
bearer scheme, and exceeds the non-normative guidance on cursors, rate
limiting, output sanitisation, and tool-list stability. Two findings were
identified and both were fixed in the same change as this report (options A2
and B1 below, chosen 2026-09-10). Neither was exploitable.

## Source caveats

- `modelcontextprotocol.info` (S30) is not the official site. It is undated,
  labels 2024-11-05 as current, and contains no transport-specific guidance. It
  was weighted low.
- The YouTube video (S31, Neon Postgres, published 2026-09-08) was not watched.
  Only its title, channel, date, and description were recorded. Its subject is
  tool ergonomics, not security.
- The Snyk article (S29) is dated 2025-05-27 and is outside the three-month
  window. It targets stdio JavaScript servers; its transferable items are
  covered below.
- In-window best-practice material beyond the specification itself is thin.
  The spec pages, the MCP project blog, the go-sdk releases, and a handful of
  vendor and incident posts are the whole set. Older material (CSA, OWASP,
  Anthropic, Snyk) is cited only where it is still the canonical reference and
  is labelled as out of window.

## Practice table

Evidence cites the code and the test that proves it. "ADR" means the practice
is fixed by the locked mcp-server ADR.

### Transport (S1, S2, S18)

| Practice | Source | Bucket | Evidence |
| --- | --- | --- | --- |
| Single POST endpoint; GET and DELETE return 405 with `Allow: POST` | S1 `#earlier-streamable-http-revisions` SHOULD | Met | `handler.go:157` delegates non-POST to the stateless SDK; `handler_test.go:260` asserts status and `Allow` |
| `Origin` validated, 403 on invalid | S1 `#security--endpoint` MUST | Met, exceeds | SDK v1.7.0 performs no Origin check by default (S18). Hikyo checks first: `handler.go:153` and `validOrigin` at `handler.go:472`; exact allowlist, `null` and `*` refused at construction `handler.go:107`; `handler_test.go:272`, `:343` |
| Host validated against configured authority, not learned from request | S18 (DNS rebinding, GHSA-xw59-hvm2-8pj6) | Met, exceeds | `validHost` at `handler.go:456` compares to `HIKYO_EXTERNAL_ORIGIN`; trusted-proxy forwarded authority must be single and exact; `handler_test.go:297`. SDK loopback protection left enabled |
| `MCP-Protocol-Version` header required; mismatch with `_meta` is 400 and -32020 | S1 `#protocol-version-header`, `#server-validation` MUST | Met | `handler.go:188`; probe confirmed a header-less request gets 400 -32020; `handler_test.go:357`. Stricter than the SDK default, which accepts a header-less request as 2025-03-26 |
| Unsupported version is 400 and -32022 with `supported` list | S1, S6 MUST | Met | `handler.go:220` and `:226`; `handler_test.go:558` sends `2025-11-25` |
| `Mcp-Method` and `Mcp-Name` required and must match body | S1 `#standard-request-headers` REQUIRED | Met | `handler.go:194`; `handler_test.go:357` |
| `=?base64?...?=` sentinel decoded before header comparison | S1 `#server-validation` MUST | Met (finding B, fixed) | `decodeHeaderSentinel` in `handler.go` decodes `Mcp-Name` and writes the decoded value back so the SDK validator (which does not decode `Mcp-Name`, `streamable_headers.go:380`) sees it; `handler_test.go` `TestBase64SentinelMcpNameIsDecodedBeforeComparison`. See finding B |
| `Mcp-Param-*` headers with invalid characters rejected | S1 `#server-behavior-for-custom-headers` MUST | Met | No tool marks `x-mcp-header`, so no param headers are expected; the SDK validator rejects malformed ones after Hikyo's checks pass |
| No `x-mcp-header` on sensitive parameters | S7 `#x-mcp-header` SHOULD NOT | Met | Not used anywhere in `tools.go` |
| Unknown method is 404 and -32601 | S1 MUST | Met | `handler.go:233`; `handler_test.go:558` |
| Notification returns 202 with no body; a malformed notification returns an HTTP error | S1 `#sending-messages` MUST | Met | `serveValidatedStaticNotification` at `handler.go:290`; `handler_test.go:525`. A `tools/call` without an id is refused with 400 before the bearer is read, so missing and presented bearers are identical there too (`TestToolCallNotificationIsRefusedBeforeBearerHandling`) |
| No `Mcp-Session-Id` minted, echoed, or accepted | S1 SHOULD, S12 (SEP-2567) | Met, exceeds | `Stateless: true`; presented session id refused at `handler.go:237`; `handler_test.go:160`, `:679` |
| `Content-Type` and `Accept` enforced | S1 MUST | Met | `handler.go:161` then SDK 415 and 400 paths |
| Request body bounded | S18 | Met | 256 KiB at `handler.go:26`, enforced at `handler.go:166` before parsing and again via SDK `MaxRequestBodyBytes` |
| Client disconnect cancels work | S1 `#cancellation` SHOULD | Met | `PropagateRequestCancellation: true` at `handler.go:128`; 30 s deadline at `handler.go:281`; `handler_test.go:636` proves cancellation reaches the operation |
| TLS in front; plain HTTP only for loopback in dev mode | S4 `#communication-security`, S26 | Met | ADR Decision section; `deploy/mcp/nginx.conf`; `scripts/ci/check-mcp-deployment.sh` |
| JSON messages are UTF-8, one message per POST, no batches | S2 | Met | `decodeOne` at `handler.go:424` rejects trailing values; batch arrays fail envelope decode |

### Authorization (S3, S4, S13, S19)

| Practice | Source | Bucket | Evidence |
| --- | --- | --- | --- |
| Token only in `Authorization: Bearer`; never query string, cookie, or body | S3 `#token-requirements` MUST | Met | `parseBearer` at `handler.go:404` is the only source; exactly one header; 4 KiB bound; `handler_test.go:473` |
| Only tokens issued for this server are accepted; no passthrough or transit | S3 `#token-handling`, S4, S13 `#token-passthrough` MUST | Met | Only `hikyo-token` service-account artifacts resolve; no outbound calls; `registry_test.go:116` proves no store, SQL, or crypto import; `Bearer` type redacts on format `registry.go:53`, `handler_test.go:695` |
| Token verified on every request; revocation immediate | S3 MUST | Met, exceeds | No caching; each page re-resolves and re-authorizes in a fresh transaction (`mcp_admission.go:37` then the page service); `mcp_e2e_test.go:428` proves revoked token is refused on the next call |
| Missing token is 401 with `WWW-Authenticate` | S3 `#error-handling` MUST | Met | `handler.go:276` |
| Invalid or expired token is 401 | S3 `#token-handling` MUST when the OAuth framework is adopted; otherwise REST parity and client ergonomics | Met (finding A, fixed) | Uniform 401 via `markUnauthenticated` in `registry.go` and `writeUnauthorized` in `handler.go`; `handler_test.go` `TestUnauthenticatedBearerUsesUniformHTTP401`; `mcp_e2e_test.go` asserts invalid, missing, and revoked are byte-identical 401s |
| Insufficient permission is 403 or an equivalent safe refusal | S3 `#error-handling` | Deliberate deviation | ADR "Threat-model controls": unauthorized is indistinguishable from nonexistent, so an ungranted principal receives the same safe tool error. A 403 would confirm scope existence. Accepted |
| RFC 9728 protected-resource metadata and `resource_metadata` challenge | S3, S5 MUST when the OAuth framework is adopted | Deliberate deviation | Authorization is OPTIONAL (S3 `#protocol-requirements`); Hikyo uses its own bearer scheme, which S9 `#auth` permits. OAuth necessity is tracked in #631. `docs/site/.../mcp.mdx:14` tells operators that OAuth buttons will not work |
| Tokens never logged | S4 `#token-theft` MUST, S21, S26 | Met | `metrics.go:368` logs method, tool, status, duration only; `audit.Context` carries IP and user agent, not the bearer |
| Audience binding | S3, S20 | Met by construction | Hikyo tokens are minted by Hikyo for Hikyo; there is no second audience |

### Tools (S7, S20, S22, S25, S28)

| Practice | Source | Bucket | Evidence |
| --- | --- | --- | --- |
| `tools` capability declared; nothing else advertised | S7 `#capabilities` MUST | Met | `handler.go:122`; `handler_test.go:160` asserts exactly one capability |
| `tools/list` deterministic, static, never varies with call count or client | S7 MUST NOT vary per connection; S22 rug-pull defence | Met, exceeds | Registry frozen at boot `registry.go:227`; sorted by name; `listChanged` not advertised; `handler_test.go:205`, `tools_test.go:444` pins the five rows |
| Non-null `inputSchema`; unknown fields rejected | S7 `#tool` | Met | `additionalProperties: false` inferred; `tools_test.go:271` sends an unknown field and gets an error; `tools_test.go:481` asserts the schema text |
| Inputs validated server-side; no shell or SQL concatenation | S7 `#security-considerations` MUST, S25 | Met | SDK validates against schema before the handler; ids are typed `domain.Scope`; page size re-checked in `normalizePageSize` at `tools.go:233`; store uses sqlc queries |
| Annotations reflect real behaviour | S7, S26, S28 | Met | `readOnlyHint`, `idempotentHint` true, `destructiveHint`, `openWorldHint` false at `registry.go:190`; `Register` refuses any authorization operation that is not an audited-none read `registry.go:145` |
| Descriptions factual, no model instructions, no secrets | S25, S20, S22 | Met | Each description in `tools.go` states operation, scope, and what is never returned |
| `structuredContent` conforms to `outputSchema` | S7 `#output-schema` MUST | Met | SDK validates output against the pinned schema; result bound at `registry.go:216` |
| Serialized JSON mirrored into a `TextContent` block | S7 `#structured-content` SHOULD | Deliberate deviation | ADR "Pagination and output bounds": 4 KiB summary only, never a duplicate row set. Rationale is token economy and a single bound. Pre-SEP-2106 clients cannot recover data from text; all named clients in `mcp.mdx` read `structuredContent` |
| Tool execution errors carry `isError: true` with actionable text; protocol errors use JSON-RPC codes; internals not leaked | S7 `#error-handling`, S28, S30 | Met | Named safe errors (`invalid_cursor`, `invalid_argument`, `traversal_limit_reached`, `result_item_too_large`) at `cursor.go:33`; everything else collapses to `Hikyo operation refused` at `registry.go:213`; `handler_test.go:595` |
| Outputs sanitised: no secrets, env, tokens, paths | S7 MUST, S23 CVE-2026-67357, S24 | Met, exceeds | `mapConfiguration` at `tools.go:498` and `mapPending` keep plaintext only for `config` classification; canary secret fixture in `mcp_e2e_test.go:167`, `mcp_pagination_test.go:190`, `tools_test.go:284` |
| Rate limiting on `tools/call` | S7 MUST, S21, S26 | Met, exceeds | Datastore-coordinated token bucket 60/min, burst 20, concurrency 4 per principal, 8 per org, 64 per instance (`mcp_admission.go:62`); uniform 429 with `Retry-After` (`handler.go:336`); per-IP admission on discovery (`handler.go:246`); `tools_test.go:245`, `handler_test.go:411`, `:615` |
| Rate limiting charged only after authorization, so guessed ids cannot occupy tenant budgets | ADR, S21 | Met | `MCPAdmission.Acquire` authorizes before claiming capacity; `mcp_e2e_test.go:305` proves an unauthorized principal consumes zero buckets |
| Pre-authentication flood on `tools/call` with a bogus bearer | S7, S21 | Met, parity with REST | Bounded by the 64 instance slots (`handler.go:268`), the 30 s deadline, and the global 512-request cap. Token lookup is a hash-verifier read, not a KDF (`machine.go:66`), so the cost per probe is one indexed read. REST applies its per-IP `Enter` limiter only to password and factor paths, so MCP matches the REST bearer path exactly |
| Cross-call handles are random, opaque, bound, expiring | S13 `#state-handle-hijacking`, S7 `#stateful-tools` | Met, exceeds | Cursors are AEAD-sealed under a keyring-derived key, bound to tool, scope ids, and page size, expire in 15 minutes without renewal, and carry no bearer or principal (`cursor.go:88`); `tools_test.go:193`, `:210`, `:223`, `cursor_test.go` |
| Human in the loop is a client obligation; server keeps read-only tools genuinely read-only | S7 `#user-interaction-model` | Met | Registry refuses non-read operations at construction |
| Untrusted tenant text in outputs stays typed data, never instructions | S25, S20 | Met | ADR "Secret and model-context boundary": descriptions and notes are structured fields, never concatenated into `instructions`; no `instructions` field is served |

### Pagination and caching (S8, S11, S24)

| Practice | Source | Bucket | Evidence |
| --- | --- | --- | --- |
| Cursors opaque, stable, integrity-protected, never a path or raw offset | S8 SHOULD, S24 | Met, exceeds | See cursor row above; keyset position not offset (`paginate.go`) |
| Invalid cursor handled gracefully | S8 SHOULD | Met | One named `invalid_cursor` tool error for tampered, expired, or wrong-scope, without saying which |
| `ttlMs >= 0` and `cacheScope` on `server/discover` and `tools/list` | S11 MUST | Met | Probe confirmed `ttlMs: 0`, `cacheScope: "public"` on both. Public is correct because the catalog is tenant-free and identical for every token |
| Chain bounds prevent bulk extraction through small pages | ADR | Met, exceeds | 10 pages, 1000 items, 1 MiB per chain (`cursor.go:22`); `tools_test.go:339`, `:402` |

### Protocol hygiene (S6, S9, S10)

| Practice | Source | Bucket | Evidence |
| --- | --- | --- | --- |
| `server/discover` implemented with `supportedVersions` | S10 MUST | Met | `handler.go:355` forces the pinned list; `handler_test.go:160` |
| Missing `_meta` fields produce 400 and -32602 | S9 `#_meta` MUST | Met | `handler.go:213` hands the SDK the schema check; `handler_test.go:396` |
| `resultType` and `serverInfo` in every result | S9 MUST and SHOULD | Met | Probe confirmed both on discover, list, and call |
| No emission of -32002, -32042, or private codes outside the defined set | S9 `#error-codes` MUST NOT | Met | Hikyo emits only -32020, -32022, -32601; SDK v1.7.0 maps resource-not-found to -32602 by default |
| JSON Schema 2020-12; no network `$ref`; bounded complexity | S9 | Met | Schemas are inferred from Go structs at boot and contain no `$ref` |
| JSON nesting depth bounded | S16 (v1.8.0-pre.1 adds a 1000-level cap) | Met by other means | 256 KiB body bound plus Go's `encoding/json` 10 000-level limit. Watch item, see follow-ups |

### Supply chain and operations (S16, S17, S26, S29)

| Practice | Source | Bucket | Evidence |
| --- | --- | --- | --- |
| go-sdk at or above v1.4.1; all four GHSAs patched | S17 | Met | v1.7.0 pinned exactly; `registry_test.go:137` fails CI on drift. No go-sdk advisory since 2026-06-09 |
| Dependency vulnerability scanning | S29, S26 | Met | `govulncheck` on the release-shaped binary in `ci.yml:873` and `:961`; local run today: zero reachable vulnerabilities in `internal/mcpserver` (five unreachable in `x/crypto` and `x/mod`) |
| SBOM | S29, S26 | Met | `anchore/sbom-action` SPDX in `ci.yml:230` |
| Dependency updates | S26 | Met | Dependabot is active (PRs #715, #716 on this branch's history) |
| Upstream conformance suite in CI | S15 (conformance testing priority) | Met, exceeds | `scripts/ci/check-mcp-conformance.sh` runs the 2026-07-28 upstream suite against a fixture that still uses Hikyo's real host, origin, and header adapters (`conformance.go`) |
| Named-client interoperability | ADR evidence item 5 | Met | `client_interop_test.go`; Claude Code, Codex CLI, and others in `mcp.mdx` |
| Structured request logging with identity, tool, outcome, duration; arguments redacted | S21, S25, S26 | Met | Metrics and debug log per call; audit events `grant.denied` and `auth.artifact_class_refused` carry `origin=mcp` and the operation; arguments are ids only and are not logged |
| Container packaging; digest-pinned image | S29, S26 | Met | `deploy/mcp/compose.yaml`; `check-mcp-deployment.sh` |

### Tool design ergonomics (S28, S31; out of window and non-security)

Anthropic's guidance (2025-09) and the Neon video description (2026-09-08)
both favour fewer workflow-shaped tools and search over list-all. Hikyo ships
five list tools with no filter argument. For a read-only phase-1 surface this
is fine and keeps the closed-registry proof simple. If clients are observed
paging through definitions to find one key, a `key_name` filter on
`hikyo_inspect_configuration` would be the smallest ergonomic step. Not a
finding.

## Findings

### A. Invalid or expired bearer returns 200 with a tool error instead of 401

**What happens.** A `tools/call` with a missing bearer gets 401 with
`WWW-Authenticate: Bearer` (`handler.go:276`). A `tools/call` with a present
but invalid, expired, or revoked bearer reaches the service, fails
`domain.ErrUnauthenticated` inside the transaction, and is collapsed to the
tool-level `Hikyo operation refused` with HTTP 200 and `isError: true`
(`registry.go:213`). This is deliberate: `mcp_e2e_test.go:428` asserts that a
revoked token and a garbage token produce byte-identical bodies, and the ADR
"Threat-model controls" section groups invalid, expired, revoked, and
unauthorized under one non-disclosing disposition.

**Why it is worth a decision.** Three pressures point the other way.

- The MCP authorization spec says invalid or expired tokens MUST receive 401
  (S3 `#token-handling`). That section is conditional on adopting the OAuth
  framework, which Hikyo has not, so this is guidance rather than a violation.
- REST parity: the REST surface maps `ErrUnauthenticated` to 401
  `authentication required` (`errors.go:71`, `:225`) with no audit event. The
  ADR sentence "retain the existing silent, non-enumerating
  authentication-failure disposition" names that REST behaviour.
- Client ergonomics: named clients (Claude Code, Codex CLI) surface a 401 as an
  authentication problem to the user. A 200 with `isError` is handed to the
  model as a retryable tool failure, so a rotated or expired token produces
  retries and a confusing transcript instead of a clear prompt to fix the
  credential.

**Counter-argument.** A 401 tells whoever holds a leaked token whether it is
still live. The uniform tool error does not. REST already discloses this, so
parity does not add exposure, but it is the one property the current
behaviour protects and Marc should weigh it explicitly.

**Not affected either way.** Authorization denial for a live token (no grant,
or wrong artifact class) stays the safe tool error. That is the
unauthorized-equals-nonexistent property and it is correct.

**Resolution (2026-09-10): A2 implemented.** `domain.ErrUnauthenticated` from
a tool handler now sets the call state and the transport writes the same
uniform 401 a missing bearer receives. Revoked, expired, unknown, and missing
are byte-identical (the e2e suite mints an actually expired credential). ADR amended by banner. The public smoke probe
(`scripts/mcp-public-smoke`) and the e2e suite assert the new disposition and
the probe has a negative test that rejects the old tool-error denial.

**Options considered.**

- A1: keep as is; record the parity divergence and the client-ergonomics cost
  in the ADR so it is a documented choice.
- A2 (recommended): map `domain.ErrUnauthenticated` from the tool handler to a
  uniform 401 with `WWW-Authenticate: Bearer` and no body detail, using the
  same `callState` mechanism that already turns `ErrRateLimited` into a
  uniform 429 (`handler.go:74`, `:336`). Revoked and invalid stay
  byte-identical, now both 401. Requires an ADR amendment sentence and updates
  to the two e2e assertions.

### B. `Mcp-Name` base64 sentinel is not decoded before comparison

**What happens.** The spec allows a client to wrap any header value in
`=?base64?...?=` and requires servers to decode it before comparing to the body
(S1 `#server-validation` MUST). Hikyo's mirror check at `handler.go:201`
compares the raw header string. The pinned SDK does the same for `Mcp-Name`
(`streamable_headers.go:380`) and decodes only `Mcp-Param-*` values
(`decodeHeaderValue` at `streamable_headers.go:490`), so removing Hikyo's
check would not close the gap. An encoded `Mcp-Name` is refused with 400 and
-32020 by both layers.

**Why it is small.** Hikyo tool names are `[A-Za-z0-9_.-]`, so no compliant
client needs to encode them; the official Go client encodes only values with
non-token characters or surrounding whitespace. The failure is a clean 400,
not a bypass.

**Resolution (2026-09-10): B1 implemented.** `decodeHeaderSentinel` in
`handler.go` mirrors the SDK's `decodeHeaderValue` rule and the decoded name is
written back to the request header before the SDK validator runs. Upstream
fixed the same omission in go-sdk #1242 (2026-09-06) and #1246 (2026-09-07),
after v1.8.0-pre.2; no release carries it yet, so nothing needs filing. The
adapter keeps its own decode because its mirror check runs before the SDK.

**Options considered.**

- B1 (recommended): decode the sentinel in `handler.go` before the `Mcp-Name`
  comparison, mirroring the SDK's `decodeHeaderValue` rule (strict base64,
  reject on decode failure), and add one test with an encoded name. Roughly
  ten lines.
- B2: accept and track; revisit when go-sdk v1.8.0 stabilises and check
  whether the upstream validator gained the decode (it did, on `main`, in
  #1242).

## Follow-ups that are not findings

- **go-sdk v1.8.0.** Pre-releases from 2026-09-04 add a JSON nesting cap,
  `ServerOptions.SupportedProtocolVersions`, `SetCacheable` for non-zero
  `ttlMs`, and remove four `MCPGODEBUG` flags. None is required for
  compliance. When it goes stable, the ADR's upgrade rule applies: reviewed
  changelog plus the full conformance, security, and interop suite.
- **`ttlMs: 0` on a static catalog.** Compliant, but a non-zero TTL would let
  clients cache the five-tool list. Optional after v1.8.0 exposes the hook.
- **Text mirroring.** If a pre-SEP-2106 client is ever named as supported, the
  summary-only deviation would need revisiting. No such client is named today.

## What this audit did not do

- Production code changed only for findings A and B; no other change.
- The YouTube video was not watched.
- The OWASP guide PDF body and the three August CVE records were not fetched;
  their data comes from the Adversa digest (S23).
- Spec section anchors in the source catalogue were inferred from headings by
  a summarising fetch and should be spot-checked before being quoted in an
  ADR amendment.
