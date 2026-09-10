# MCP server best practices: primary-source catalogue (2026-09-09)

Purpose: checkable practices for auditing a Go MCP server (official go-sdk v1.7.0, Streamable HTTP, stateless, JSON responses, protocol 2026-07-28, bearer-token auth, tools only, read-only). External sources only; repo code was not read.

Freshness window: sources published on or after 2026-06-09 are marked "yes". Everything else is marked "no" and is NOT relabelled as new.

RFC keywords (MUST / SHOULD / MAY) are quoted only where the source is normative. Section anchors are given as `#anchor` on the cited page.

## 1. Source table

| # | Source | URL | Publisher | Publication date | Primary / secondary | In window (>= 2026-06-09)? |
|---|---|---|---|---|---|---|
| S1 | MCP spec 2026-07-28: Streamable HTTP transport | https://modelcontextprotocol.io/specification/2026-07-28/basic/transports/streamable-http | MCP project (modelcontextprotocol.io) | 2026-07-28 (revision date) | primary, normative | yes |
| S2 | MCP spec 2026-07-28: Transports overview | https://modelcontextprotocol.io/specification/2026-07-28/basic/transports | MCP project | 2026-07-28 | primary, normative | yes |
| S3 | MCP spec 2026-07-28: Authorization | https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization | MCP project | 2026-07-28 | primary, normative | yes |
| S4 | MCP spec 2026-07-28: Authorization Security Considerations | https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization/security-considerations | MCP project | 2026-07-28 | primary, normative | yes |
| S5 | MCP spec 2026-07-28: Authorization Server Discovery | https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization/authorization-server-discovery | MCP project | 2026-07-28 | primary, normative | yes |
| S6 | MCP spec 2026-07-28: Versioning and Compatibility (the `basic/lifecycle` URL resolves here) | https://modelcontextprotocol.io/specification/2026-07-28/basic/versioning | MCP project | 2026-07-28 | primary, normative | yes |
| S7 | MCP spec 2026-07-28: Tools | https://modelcontextprotocol.io/specification/2026-07-28/server/tools | MCP project | 2026-07-28 | primary, normative | yes |
| S8 | MCP spec 2026-07-28: Pagination | https://modelcontextprotocol.io/specification/2026-07-28/server/utilities/pagination | MCP project | 2026-07-28 | primary, normative | yes |
| S9 | MCP spec 2026-07-28: Base protocol overview (statelessness, `_meta`, error codes, JSON Schema) | https://modelcontextprotocol.io/specification/2026-07-28/basic/index | MCP project | 2026-07-28 | primary, normative | yes |
| S10 | MCP spec 2026-07-28: Discovery (`server/discover`) | https://modelcontextprotocol.io/specification/2026-07-28/server/discover | MCP project | 2026-07-28 | primary, normative | yes |
| S11 | MCP spec 2026-07-28: Caching (`ttlMs`, `cacheScope`) | https://modelcontextprotocol.io/specification/2026-07-28/server/utilities/caching | MCP project | 2026-07-28 | primary, normative | yes |
| S12 | MCP spec 2026-07-28: Changelog | https://modelcontextprotocol.io/specification/2026-07-28/changelog | MCP project | 2026-07-28 | primary | yes |
| S13 | Security Best Practices (tutorial page; the `specification/.../basic/security_best_practices` URL serves the identical document) | https://modelcontextprotocol.io/docs/2026-07-28/tutorials/security/security_best_practices | MCP project | 2026-07-28 (versioned with the spec; no separate date shown) | primary; normative standing unclear, see caveats | yes |
| S14 | MCP blog: The 2026-07-28 Specification | https://blog.modelcontextprotocol.io/posts/2026-07-28/ | MCP project (D. Soria Parra, D. Delimarsky) | 2026-07-28 | primary (project blog) | yes |
| S15 | MCP blog: The New MCP Roadmap | https://blog.modelcontextprotocol.io/posts/mcp-roadmap/ | MCP project (D. Soria Parra, D. Delimarsky) | 2026-08-22 | primary (project blog) | yes |
| S16 | go-sdk releases page | https://github.com/modelcontextprotocol/go-sdk/releases | MCP project / Google | v1.7.0 on 2026-07-28; v1.8.0-pre.1 and pre.2 on 2026-09-04 | primary | yes |
| S17 | go-sdk security advisories | https://github.com/modelcontextprotocol/go-sdk/security/advisories | MCP project | latest advisory 2026-03-30 | primary | no (all four advisories predate the window) |
| S18 | go-sdk v1.7.0 `mcp/streamable.go` (source) | https://raw.githubusercontent.com/modelcontextprotocol/go-sdk/v1.7.0/mcp/streamable.go | MCP project | 2026-07-28 (tag date) | primary | yes |
| S19 | go-sdk v1.7.0 `auth/auth.go` (source) | https://raw.githubusercontent.com/modelcontextprotocol/go-sdk/v1.7.0/auth/auth.go | MCP project | 2026-07-28 (tag date) | primary | yes |
| S20 | Microsoft Security Blog: The state of MCP security in 2026 | https://techcommunity.microsoft.com/blog/microsoft-security-blog/the-state-of-mcp-security-in-2026/4531327 | Microsoft (J. Thakur, S. Pradhan) | 2026-06-26 (updated 2026-06-30) | secondary | yes |
| S21 | Cloudflare blog: MCP security updates | https://blog.cloudflare.com/mcp-security-updates/ | Cloudflare (K. Johnson) | 2026-08-14 (modified 2026-08-18) | secondary | yes |
| S22 | Pillar Security: Deadbugz, currently active MCP supply chain campaign | https://www.pillar.security/blog/deadbugz-currently-active-mcp-supply-chain-campaign | Pillar Security (A. Fogel) | 2026-08-12 | secondary (incident primary source) | yes |
| S23 | Adversa AI: Top MCP security resources, September 2026 | https://adversa.ai/blog/top-mcp-security-resources-september-2026/ | Adversa AI (S. Malenkovich) | 2026-09-07 | secondary (digest) | yes |
| S24 | arXiv 2608.00150: Exposed by Design (internet-facing MCP audit) | https://arxiv.org/abs/2608.00150 | N. Padilla | 2026-07-31 (v1) | primary (research) | yes |
| S25 | mcpservers.org: MCP Security Best Practices, practical guide for 2026 | https://blog.mcpservers.org/posts/mcp-security-best-practices | mcpservers.org (M. Webb) | 2026-07-10 | secondary | yes |
| S26 | Cloud Security Alliance: Agentic MCP Security Best Practices Guide v1 (marked draft) | https://labs.cloudsecurityalliance.org/agentic/agentic-mcp-security-best-practices-v1/ | CSA | 2026-03-27 | secondary | no |
| S27 | OWASP GenAI: A Practical Guide for Secure MCP Server Development | https://genai.owasp.org/resource/a-practical-guide-for-secure-mcp-server-development/ | OWASP GenAI Security Project | 2026-02-16 | secondary | no |
| S28 | Anthropic Engineering: Writing effective tools for agents | https://www.anthropic.com/engineering/writing-tools-for-agents | Anthropic | 2025-09-11 | secondary | no |
| S29 | Snyk: 5 best practices for building MCP servers | https://snyk.io/articles/5-best-practices-for-building-mcp-servers/ | Snyk (L. Tal) | 2025-05-27 | secondary | no |
| S30 | modelcontextprotocol.info: MCP Best Practices (unofficial mirror site) | https://modelcontextprotocol.info/docs/best-practices/ | ModelContextProtocol.Info (third party) | undated (footer "2024") | secondary, unofficial | no |
| S31 | YouTube: MCP Just Got a Whole Lot Better | https://youtu.be/BqRhBq-_kgE | Neon Postgres (YouTube channel) | 2026-09-08 | secondary; video NOT watched | yes (metadata only) |

Sources that were searched for but not fetched or not found: an official MCP blog post specifically on security in the window (none exists; the blog posts in window are S14 and S15). OWASP has no MCP-specific release in the window (their 2026-09-02 announcement covers LLM Top 10 2026 and an Agent Control Standard, not MCP servers: https://www.prnewswire.com/news-releases/owasp-genai-security-project-releases-2026-top-10-for-llm-applications-debuts-agent-control-standard-and-new-resources-for-securing-generative-and-agentic-ai-302867085.html).

## 2. Normative spec requirements (2026-07-28), server side

### 2.1 Statelessness and `_meta` (S9)

- Servers MUST NOT rely on prior requests over the same connection to establish context (capabilities, protocol version, client identity). `#statelessness`
- State spanning requests MUST be referenced by an explicit identifier the client passes on each request. `#statelessness`
- Every request MUST carry `_meta["io.modelcontextprotocol/protocolVersion"]` and `_meta["io.modelcontextprotocol/clientCapabilities"]`; a request missing either is malformed and the server MUST reject it with JSON-RPC `-32602`; on HTTP the status MUST be `400`. `#_meta`
- Server MUST NOT rely on capabilities the client has not declared; if a capability is required and absent, return `MissingRequiredClientCapabilityError` (`-32021`) with HTTP `400`. `#_meta`
- Servers SHOULD include `_meta["io.modelcontextprotocol/serverInfo"]` in every result. `#_meta`
- `clientInfo`/`serverInfo` are self-reported; implementations SHOULD NOT use them for behaviour or security decisions. `#_meta`
- Every result MUST include `resultType` (`"complete"` for ordinary results). `#result-responses`
- Error codes: `-32020` HeaderMismatch, `-32021` MissingRequiredClientCapability, `-32022` UnsupportedProtocolVersion. Implementations MUST NOT emit codes in `-32020..-32099` other than those defined; MUST NOT emit `-32002` (resource not found is now `-32602`) or `-32042`; new implementations SHOULD NOT use `-32000..-32019`. `#error-codes`
- JSON Schema: MUST support 2020-12 as default dialect; MUST NOT auto-dereference network `$ref`s (opt-in only, disabled by default, allowlist or reject loopback/link-local/private ranges, timeouts, size limits, logging); SHOULD bound schema depth / subschema count / validation time to prevent DoS. `#json-schema-usage`, `#ref-resolution`, `#composition-keyword-resource-use`

### 2.2 Versioning and discovery (S6, S10)

- No handshake; every request declares its version. Unsupported version: server MUST respond with `UnsupportedProtocolVersionError` (`-32022`) listing `supported` versions. `#protocol-version-negotiation`
- Servers MUST implement `server/discover`. Result includes `supportedVersions`, `capabilities`, optional `instructions`; servers SHOULD include `serverInfo` in `_meta`. (S10 `#response`, `#discoverresult`)
- A modern-only server SHOULD name the protocol versions it supports in any error it returns to a legacy `initialize` request. (S6 `#backward-compatibility-with-initialization-based-versions`)
- Legacy client hitting a modern HTTP server: request lacks required headers and is rejected with `400` per server validation. (S6 compatibility matrix)

### 2.3 Streamable HTTP transport (S1, S2)

Security and endpoint (`#security--endpoint`):
- Server MUST provide a single MCP endpoint path that supports POST.
- Servers MUST validate the `Origin` header on all incoming connections to prevent DNS rebinding. If `Origin` is present and invalid, servers MUST respond `403 Forbidden` (body MAY be an id-less JSON-RPC error).
- When running locally, servers SHOULD bind only to `127.0.0.1`, not `0.0.0.0`.
- Servers SHOULD implement proper authentication for all connections.

Sending messages (`#sending-messages`):
- Client MUST send `Accept` listing both `application/json` and `text/event-stream`.
- Body MUST be a single JSON-RPC request or notification; no batches, no client-sent responses.
- Notification accepted: server MUST return `202 Accepted` with no body. Not accepted: MUST return an HTTP error status (e.g. `400`).
- Request: server MUST return `Content-Type: application/json` (single JSON object) or `Content-Type: text/event-stream`. Either is compliant; a JSON-only server is allowed.
- No client-to-server notifications are defined in this revision on HTTP; `notifications/cancelled` is stdio only.

Receiving / SSE (`#receiving-messages`): only relevant if SSE is used. Notifications on an SSE stream MUST relate to the originating request; server MUST NOT send independent JSON-RPC requests; final response SHOULD terminate the stream; SHOULD send `X-Accel-Buffering: no`; `Last-Event-ID` resumability is not supported.

Cancellation (`#cancellation`): closing the response stream MUST be treated as cancellation; server SHOULD stop work as soon as practical and MUST NOT send further messages for it.

Request metadata (`#request-metadata`, `#protocol-version-header`, `#standard-request-headers`, `#server-validation`):
- Every POST MUST include `MCP-Protocol-Version`; the value MUST match `_meta` protocolVersion; on mismatch the server MUST reject with `400` and `-32020 HeaderMismatch`.
- Unsupported version: MUST respond `400` with `UnsupportedProtocolVersionError`.
- Unknown RPC method: MUST respond `404 Not Found` with JSON-RPC `-32601`.
- Missing `MCP-Protocol-Version`: server MAY treat as `2025-03-26` only if it supports pre-2025-06-18 clients; otherwise MUST reject per server validation.
- `Mcp-Method` (all requests) and `Mcp-Name` (`tools/call`, `resources/read`, `prompts/get`) are REQUIRED for compliance.
- Servers that process the body MUST reject requests where header values do not match body values, returning `400` and JSON-RPC `-32020`. Failure conditions: required standard header missing; header/body mismatch (after Base64 sentinel decoding for `Mcp-Name` and `Mcp-Param-*`); invalid characters.
- Servers MUST decode `=?base64?...?=` sentinel values before comparison. Integer `Mcp-Param-*` values SHOULD be compared numerically.
- Header names MUST be compared case-insensitively; values are case-sensitive. `#case-sensitivity`
- `x-mcp-header` is optional for servers (MAY). If used: value MUST be non-empty, token syntax, no CR/LF, case-insensitively unique, primitive types only (no `number`), statically reachable via `properties` chain. Server developers SHOULD NOT mark sensitive parameters (passwords, keys, tokens, PII) with `x-mcp-header` (S7 `#x-mcp-header`).
- Servers MUST reject requests with a recognized `Mcp-Param-*` header containing invalid characters. `#server-behavior-for-custom-headers`

Backward compatibility (`#earlier-streamable-http-revisions`): a server supporting only this revision SHOULD respond to GET or DELETE on the MCP endpoint with `405 Method Not Allowed`; SHOULD ignore `Mcp-Session-Id` and never mint or echo session ids; SHOULD ignore `Last-Event-ID`. HTTP+SSE (2024-11-05) is Deprecated; new implementations SHOULD NOT adopt it.

Transports overview (S2): JSON-RPC messages MUST be UTF-8. Body is the source of truth; bindings that mirror metadata define how mismatches are rejected. `#messages`, `#request-metadata`

### 2.4 Authorization (S3, S4, S5)

Scope note for the audit: "Authorization is OPTIONAL for MCP implementations. When supported: implementations using an HTTP-based transport SHOULD conform to this specification." (S3 `#protocol-requirements`). The framework is OAuth 2.1 resource-server behaviour. A server that authenticates with a pre-shared static bearer token is not implementing the OAuth framework, so the RFC 9728 / discovery MUSTs below are conditional on adopting it. Servers and clients "MAY negotiate their own custom authentication and authorization strategies" (S9 `#auth`). The audience-binding and no-passthrough rules are the ones that transfer directly to any bearer scheme.

If the OAuth framework is adopted:
- MCP servers MUST implement OAuth 2.0 Protected Resource Metadata (RFC 9728). (S3 `#overview` item 4)
- PRM document MUST include `authorization_servers` with at least one entry. (S5 `#authorization-server-location`)
- Servers MUST implement at least one discovery mechanism: `WWW-Authenticate: Bearer resource_metadata="..."` on `401`, or the well-known URI (path-aware `/.well-known/oauth-protected-resource/<path>` or root). (S5 `#protected-resource-metadata-discovery-requirements`)
- Servers SHOULD include `scope` in the `WWW-Authenticate` challenge. (S3 `#scope-selection-strategy`)
- Servers SHOULD NOT include `offline_access` in `WWW-Authenticate` scope or `scopes_supported`. (S3 `#refresh-tokens`)

Applies to any bearer-token server:
- Access token MUST be in the `Authorization: Bearer` header on every request; MUST NOT be in the URI query string. (S3 `#token-requirements`)
- MCP servers MUST validate access tokens per OAuth 2.1 section 5.2; MUST validate the token was issued specifically for them as audience (RFC 8707); invalid or expired tokens MUST receive `401`. (S3 `#token-handling`)
- MCP servers MUST only accept tokens valid for their own resources and MUST NOT accept or transit any other tokens. (S3 `#token-handling`; S4 `#access-token-privilege-restriction`; S13 `#token-passthrough`)
- If the server calls upstream APIs it MUST NOT pass through the token it received from the client. (S4 `#access-token-privilege-restriction`)
- Servers MUST return `401` (authorization required or token invalid), `403` (invalid scopes / insufficient permissions), `400` (malformed authorization request). (S3 `#error-handling`)
- Insufficient scope at runtime: server SHOULD respond `403` with `WWW-Authenticate: Bearer error="insufficient_scope", scope="...", resource_metadata="..."`, optional `error_description`; SHOULD include all scopes required for the operation in a single challenge; SHOULD be consistent. (S3 `#runtime-insufficient-scope-errors`)
- Servers MUST account for scope hierarchies when deciding token sufficiency. (S3 `#step-up-authorization-flow`)
- Clients and servers MUST implement secure token storage; tokens cached or logged on the server are a theft vector. (S4 `#token-theft`)
- All authorization server endpoints MUST be HTTPS. (S4 `#communication-security`)
- Implementors MUST follow OAuth 2.1 section 7 security considerations. (S4 intro)

### 2.5 Tools (S7)

- Servers that support tools MUST declare the `tools` capability. `#capabilities`
- `tools/list` MUST return the set currently available to the requesting client; MUST NOT vary per-connection or as a side effect of other requests; MAY vary by the authorization presented on the request. `#capabilities`
- Servers SHOULD return tools in deterministic order. `#capabilities`
- Every request MUST include required `_meta` fields (page note).
- `inputSchema` MUST be a valid JSON Schema object (not `null`). For no-parameter tools the recommended form is `{"type":"object","additionalProperties":false}`. `#tool`
- Clients MUST treat annotations as untrusted unless the server is trusted (so annotations are hints only; the server's actual behaviour must match). `#tool`
- Tool names SHOULD be 1..128 chars, case-sensitive, only `A-Za-z0-9_-.`, SHOULD NOT contain spaces/commas, SHOULD be unique within the server. `#tool-names`
- Structured content: if `outputSchema` is provided, servers MUST return `structuredContent` conforming to it; a tool returning structured content SHOULD also return the serialized JSON in a `TextContent` block. `#structured-content`, `#output-schema`
- Errors: protocol errors (unknown tool, malformed request, server error) are JSON-RPC errors (`-32602` for unknown tool); tool execution errors (API failure, input validation, business logic) go in the result with `isError: true` and actionable text. `#error-handling`
- Security Considerations, servers MUST: validate all tool inputs; implement proper access controls; rate limit tool invocations; sanitize tool outputs. `#security-considerations`
- Human in the loop: there SHOULD always be a human with the ability to deny tool invocations (client-side obligation, but the server should not assume auto-approval). `#user-interaction-model`
- Stateful tools (non-normative): handle is a name, not a capability; validate caller authorization against the handle on every call; opaque, high-entropy, bounded lifetime; state retention policy in tool description; expired handle returns a tool execution error. `#stateful-tools`

### 2.6 Pagination (S8)

- Cursor is an opaque string; page size is server-determined. `#pagination-model`
- Servers SHOULD provide stable cursors and handle invalid cursors gracefully. `#implementation-guidelines`
- Invalid cursors SHOULD return `-32602` (Invalid params). `#error-handling`
- `tools/list` is a paginated operation. `#operations-supporting-pagination`
- Related security data point: arXiv 2608.00150 (S24) found "path traversal via cursor manipulation" in production servers, so cursors that encode file paths or offsets without integrity protection are a real-world bug class.

### 2.7 Caching (S11)

- Servers MUST include `ttlMs` (>= 0) and `cacheScope` on `resultType: "complete"` results of `server/discover`, `tools/list`, `prompts/list`, `resources/list`, `resources/templates/list`, `resources/read`. `#cacheable-results`, `#time-to-live-ttl-field`
- `cacheScope` MUST be the same across all pages of one list request. `#interaction-with-pagination`
- Security: `"public"` results may be shared across authorization contexts even from an authenticated endpoint. Servers MUST apply per-primitive access controls and MUST NOT rely on `cacheScope` alone. If `tools/list` output varies by token, it must be `"private"`. `#security-considerations`

### 2.8 Security Best Practices document (S13), server-relevant items

- Token passthrough: MCP servers MUST NOT accept any tokens not explicitly issued for the MCP server. `#token-passthrough`
- State handle hijacking: servers that implement authorization MUST verify all inbound requests; MUST NOT treat possession of a state handle as authentication; SHOULD use secure random non-deterministic handles; SHOULD bind handles server-side to the authenticated user (`<user_id>:<handle>`, user id from the verified token). `#state-handle-hijacking`
- Local servers over HTTP SHOULD require an authorization token or use unix sockets / IPC. `#local-mcp-server-compromise`
- Scope minimization: minimal initial scope set; precise scope challenges; log elevation events with correlation ids; do not treat claimed scopes as sufficient without server-side authorization logic. `#scope-minimization`
- Not applicable to a tools-only, read-only, non-proxy server: confused deputy consent flow (`#confused-deputy-problem`), SSRF during OAuth discovery (client obligation, `#server-side-request-forgery-ssrf`), OAuth authorization URL validation (client), stdio proxy escalation, mix-up attacks (client), CIMD trust policies (authorization server).

### 2.9 Changelog highlights relevant to the audit (S12)

- Sessions and `Mcp-Session-Id` removed (SEP-2567); `initialize` handshake removed, `_meta` per request (SEP-2575); `server/discover` MUST (SEP-2575); GET endpoint replaced by `subscriptions/listen` (SEP-2575); `ping`, `logging/setLevel` removed; servers MUST NOT emit `notifications/message` for requests without `logLevel` in `_meta`; SSE resumability removed; `resultType` required (SEP-2322); `Mcp-Method`/`Mcp-Name` required and `x-mcp-header` added (SEP-2243); `ttlMs`/`cacheScope` required (SEP-2549); resource-not-found `-32002` to `-32602`; error code partition with renumbering to `-32020/-32021/-32022`; Roots, Sampling, Logging deprecated (SEP-2577); HTTP+SSE deprecated; DCR deprecated in favour of CIMD.

## 3. Per-source practices (non-normative sources)

### S14 MCP blog, The 2026-07-28 Specification (2026-07-28)
- Streamable HTTP requests must include `Mcp-Method`, `Mcp-Name`, and `MCP-Protocol-Version: 2026-07-28`; gateways can route and rate-limit on headers rather than bodies.
- Cross-call state: mint an explicit handle from a tool and have the model pass it back.
- All four Tier 1 SDKs (TypeScript, Python, Go, C#) support 2026-07-28 as of release day; expect migration effort for servers that depended on session ids.
- DCR deprecated in favour of CIMD; issuer validation (RFC 9207, SEP-2468); issuer-bound client credentials (SEP-2352).

### S15 MCP blog, The New MCP Roadmap (2026-08-22)
- Priority areas: agentic messaging primitives, HTTP-native transport unification, agent identity and enterprise security (DPoP RFC 9449 finalisation, workload identity federation via spec PR #1933, token exchange), improved primitives (one contract for `tools/call` result forms, progressive discovery), SDK DX and conformance testing.
- No next revision date given. Nothing here changes 2026-07-28 server obligations; DPoP is forward-looking.

### S20 Microsoft, The state of MCP security in 2026 (2026-06-26)
- Treat tool descriptions and tool outputs as untrusted input.
- Treat MCP servers as OAuth 2.0 resource servers; audience-bound tokens; identity-aware gateway in front of every server rejecting calls lacking a valid audience-bound token; example validates issuer, audience, expiry.
- Least privilege per resource, narrow scopes, short-lived tokens; a governable identity per agent.
- Approval cannot be a one-time event: pin tool definitions, alert on drift (rug-pull tripwire), registry as known-good baseline.
- Sandbox local stdio servers; block outbound by default; validate and sanitize every input and output; never pass raw shell commands or unsanitized paths; keep SDKs patched (called one of the largest MCP vulnerability classes this year).

### S21 Cloudflare, MCP security updates (2026-08-14)
- `MCP-Protocol-Version` is the strong network fingerprint; 2026-07-28 traffic is trivially identifiable by inspecting proxies. Absence proves nothing.
- Three control points: client hooks, network gateway, the server itself (richest context, last chance to deny).
- Server-side middleware should authorize the caller per tool, rate limit, inspect arguments, and record outcomes before invoking the handler.
- Tool risk tiers (read-only through critical); reads pass, writes get attribution plus audit event, critical blocked before handler.
- Tool arguments are the most sensitive element of a call for logging purposes; logging only after execution explains but cannot prevent.
- Origins should reject direct requests that bypass a gateway (Access policy, source-IP restriction, or enterprise authorization).
- DCR deprecation means many providers require pre-registered clients with fixed credentials.

### S22 Pillar Security, Deadbugz (2026-08-12) and S23 Adversa digest (2026-09-07)
- Runtime-gated metadata poisoning: server serves benign tool descriptions until the third `tools/call`, then swaps `tools/list` / `prompts/get` content to exfiltrate credentials. Advertises `tools.listChanged` so clients refresh.
- Defences: tool descriptions and schemas are a security boundary; fingerprint tool definitions at approval and diff on every reconnect; treat any change to an approved server's definitions as a security event requiring re-approval; keep sensitive actions policy-enforced, not metadata-driven.
- For a server author the checkable mirror image: tool list and descriptions must be deterministic, static, and not vary on call count or client (this is also the spec MUST NOT in S7 `#capabilities`).
- August 2026 CVEs cited by S23: CVE-2026-73498 Atlassian MCP (client-supplied path into `open()`, arbitrary file read, CVSS 7.7, fixed 0.22.0); CVE-2026-67357 ArcadeDB MCP (settings tool returned HA cluster token in cleartext, CVSS 7.7, fixed 26.7.3); CVE-2026-19956 facebook-ads-mcp-server (SSRF in `fetch_pagination_url`, CVSS 5.3). All three are classic web flaws, none model-specific. Primary records: https://nvd.nist.gov/vuln/detail/CVE-2026-73498 , https://www.cve.org/CVERecord?id=CVE-2026-67357 , https://www.cve.org/CVERecord?id=CVE-2026-19956 (CVE pages not fetched; data as reported by S23).
- S23 guidance: watch tool metadata for post-approval drift; reconstruct full agent action chains; put a token-validating, scoped authorization layer in front of MCP servers; red-team it.

### S24 arXiv 2608.00150, Exposed by Design (2026-07-31, N. Padilla)
- 21,000+ MCP instances visible; 640 confirmed production servers; 414 audited; 68 reportable vulnerabilities (SQL injection, SSRF to cloud metadata, prompt template injection, path traversal via cursor manipulation); 91.8% lack OAuth authentication; 687 tool instances expose shell execution without access control; 41.6% churn within three days.
- Checkable takeaways: pagination cursors must not be trusted as paths or offsets; tool outputs that echo secrets (config, env, tokens) are a recurring finding; unauthenticated exposure is the norm, so bearer auth is already above baseline.

### S25 mcpservers.org, MCP Security Best Practices (2026-07-10)
- Implement OAuth 2.1; do not build a custom auth scheme.
- No token passthrough; validate audience; mint or exchange audience-scoped downstream credentials.
- Session/handle ids: cryptographically random; bind to user identity; never treat an id as authentication.
- Tool descriptions narrow, factual, free of anything that reads like an instruction to the model.
- Wrap untrusted external content in tool output with clear delimiters, labelled as data (reduces, does not solve, prompt injection).
- Treat every tool argument as untrusted; validate types, ranges, formats server-side; never concatenate into shell or SQL.
- Start clients on minimal read-only scopes; short-lived scoped tokens.
- Log tool invocations, arguments, caller identity, outcome; note that logging full bodies leaks tokens (no explicit redaction guidance given).
- Does not discuss the HTTP `Origin` header. References `/specification/latest`, not a dated revision.

### S26 CSA, Agentic MCP Security Best Practices Guide (2026-03-27, marked draft; out of window)
- TLS 1.2+ mandatory, 1.3 preferred; CA-issued certificates; no plaintext HTTP; beware tunnel subdomain reuse (ngrok style).
- OAuth 2.1 with PKCE; validate RFC 8414 metadata; scopes at tool level; per-invocation authorization checks.
- Access tokens <= 1 hour; never store tokens in plaintext files or env vars; fetch from secret managers.
- Sign or hash tool descriptions; treat definitions as versioned immutable artifacts; surface changes visibly.
- Logging: agent identity, tool name, parameters with sensitive values redacted, timestamp, status, duration; central SIEM; >= 90 days retention; append-only audit log at top maturity.
- Supply chain: curated registry, SBOM, continuous dependency monitoring, remediation windows (30 days High/Critical; 72 hours CVSS >= 9.0).
- Least privilege: minimal service account; read-only tools must not carry write/delete scopes; tenant id as mandatory filter on data access.
- Rate limiting: per-execution CPU/memory quotas; hard timeouts; gateway-enforced rate limits with alerting.
- Sandboxing: containers or micro-VMs; filesystem and outbound network deny by default.
- Maps to OWASP ASI01..ASI06, CSA AICM controls, MITRE ATLAS technique ids.

### S27 OWASP GenAI, A Practical Guide for Secure MCP Server Development (2026-02-16; out of window)
- Only the listing page was fetched; the guide body (PDF) was not. Listing summary: MCP servers run with delegated user permissions, dynamic tool-based architectures, and chained tool calls, raising the impact of a single compromise. Companion cheat sheet for third-party servers dated 2025-11-04: https://genai.owasp.org/resource/cheatsheet-a-practical-guide-for-securely-using-third-party-mcp-servers-1-0/

### S28 Anthropic, Writing effective tools for agents (2025-09-11; out of window)
- Fewer, workflow-shaped tools over one-tool-per-endpoint; prefer search tools over list-all tools.
- Namespace with service/resource prefixes; test prefix vs suffix with evals.
- Return high-signal fields; resolve opaque ids to names; choose output format by evaluation.
- Token efficiency: `response_format` enum (concise/detailed), pagination, filtering, truncation with guidance; Claude Code caps tool responses at 25,000 tokens by default.
- Descriptions written as for a new hire; unambiguous parameter names (`user_id`, not `user`); strict data models; use MCP annotations to flag open-world or destructive tools.
- Error messages: specific, actionable, show correct input format instead of tracebacks.
- Build evals from realistic multi-step tasks; read raw transcripts.

### S29 Snyk, 5 best practices (2025-05-27; out of window)
- Tool names: no spaces, dots, parentheses; prefer snake_case.
- Logging: never `console.log` to stdout on stdio; log to file. (HTTP logging not covered.)
- Do not lead search-tool results with "not found" text; the model anchors on it.
- Scan own code (command injection, path traversal) and dependencies; SBOM.
- Package as a container.
- Examples are JavaScript with zod; no protocol version cited.

### S30 modelcontextprotocol.info, MCP Best Practices (undated; unofficial; out of window)
- Not the official site (official is modelcontextprotocol.io). Footer says "2024"; navigation labels 2024-11-05 as "Current"; no protocol version cited in the body; code samples are pseudo-code with undefined helpers and no SDK.
- Content is generic distributed-systems advice: single responsibility per server; bind to localhost; JWT auth; capability-based authorization; strict input schema validation; output sanitization; circuit breakers; token-bucket rate limiting; structured error categories with generic `INTERNAL_ERROR` to callers and full stack traces only in logs; connection pooling; multi-level caching; Prometheus metrics; structured JSON logs with method, duration, client id, result size; health and readiness probes; Kubernetes rolling updates; contract, load, and chaos tests.
- Nothing MCP-transport-specific (no Origin, session, header, or tool-schema guidance). Treat as low-weight.

### S31 YouTube, MCP Just Got a Whole Lot Better (Neon Postgres, 2026-09-08)
- Title: "MCP Just Got a Whole Lot Better". Channel: Neon Postgres (https://www.youtube.com/@neondatabase). Published 2026-09-08 per page metadata; ~4.1K views at fetch time.
- Description (paraphrased from the page): modern MCP clients use progressive tool discovery and code mode, so servers should expose ergonomic workflow tools instead of one tool per endpoint; covers layered discovery (search, inspect, execute), programmatic tool calling, and Neon's `createWithCompute`-style consolidation of three API calls into one tool. Chapters run 00:00 to 05:14.
- The video content itself was NOT reviewed (yt-dlp unavailable); only page metadata was fetched. No claims about its content beyond the description are made here.

## 4. go-sdk status (github.com/modelcontextprotocol/go-sdk)

- Latest stable: v1.7.0, released 2026-07-28, the first release with full 2026-07-28 protocol support (S16). It is still marked "Latest" on GitHub.
- Pre-releases: v1.8.0-pre.1 and v1.8.0-pre.2, both 2026-09-04 (S16). No new protocol revision; hardening and leak fixes:
  - JSON nesting capped at 1000 levels; `MaxEventSize` on client transports; `StdioTransport.MaxLineLength`; OAuth DCR responses bounded to 1 MB; OAuth discovery metadata validated.
  - New `ServerOptions.SupportedProtocolVersions` (narrow only; unknown versions panic at construction).
  - Stateful handler receiving a 2026-07-28 request now returns JSON-RPC `-32022` instead of plain-text 400.
  - New `ServerOptions.SetCacheable` hook for `ttlMs`/`cacheScope` per result.
  - pre.2 removes MCPGODEBUG flags `seterroroverwrite`, `enableoriginverification`, `disablecontenttypecheck`, `disablelocalhostprotection`.
  - Fixes: ServerSession leaks on stateful servers, `Client.Connect` leaks, `subscriptions/listen` close deadlock, ListTools panic on malformed tools, tolerate 404 on `subscriptions/listen` in stateless mode.
- v1.7.0 behaviour confirmed from source (S18, S19), with spec comparison:
  - `Stateless: true` is required to accept 2026-07-28; stateful sessions negotiate down to 2025-11-25 (S16). Stateless mode: GET and DELETE return `405` with `Allow: POST` (matches S1 SHOULD); `Mcp-Session-Id` ignored unless `MCPGODEBUG=allowsessionsinstateless=1`.
  - Origin validation: `CrossOriginProtection` is nil by default, meaning NO `Origin` check runs. The field is deprecated; auto-population only happens with `MCPGODEBUG=enableoriginverification=1`, and that flag is deleted in v1.8.0-pre.2. Documented replacement: wrap the handler with `http.NewCrossOriginProtection().Handler(h)` (Go 1.25+). Check before shipping: that type rejects requests whose `Origin` host differs from the request `Host`, so browser-hosted MCP clients with a foreign `Origin` would get 403 unless trusted origins are registered (verify the trusted-origin method in the Go 1.25 `net/http` docs; not fetched here). Non-browser clients send no `Origin` and pass. Spec S1 says servers MUST validate `Origin` and return `403`. This is the same vulnerability class as GHSA-89xv-2j6f-qhc8 / CVE-2026-33252 (cross-site tool execution, fixed in v1.4.1 by adding Content-Type checking and a configurable origin check).
  - DNS rebinding Host check: on by default. A request arriving on a loopback address with a non-loopback `Host` header gets `403` unless `DisableLocalhostProtection` is set. This is a Host check, not an Origin check; do not conflate the two.
  - Content-Type: POST must be `application/json` (parameters tolerated) else `415`. Accept: POST must list both JSON and SSE else `400`.
  - Body limit: `MaxRequestBodyBytes` default 4 MiB (`DefaultMaxRequestBodyBytes`), enforced with `http.MaxBytesReader`, `413` on overflow; negative disables (documented as unsafe for untrusted clients).
  - `MCP-Protocol-Version` header: if present and unsupported, `400` listing supported versions. If absent, the SDK defaults to `2025-03-26` (legacy accept path). A 2026-07-28 body with `_meta` protocolVersion but no header is rejected with `400` and `-32020`. Spec S1 says a server that does not support pre-2025-06-18 clients MUST reject a header-less request; the SDK's default is the permissive MAY branch. Policy choice for the auditor.
  - Header/body validation: `validateMcpHeaders` runs on single non-batch messages; mismatch returns HTTP `400` with JSON-RPC `-32020`, echoing the request id (matches S1 MUST).
  - `JSONResponse: true` returns `Content-Type: application/json`; `subscriptions/listen` always forces SSE; notification-only POSTs return `202` with no body (matches S1).
  - `PropagateRequestCancellation` defaults to false, so a tool handler keeps running after the client closes the connection. Spec S1 says the server SHOULD stop work as soon as practical. Recommend enabling for stateless 2026-07-28 servers.
  - Rate limiting: nothing in the SDK. Spec S7 says servers MUST rate limit tool invocations; it has to be middleware.
  - Error code: `CodeResourceNotFound` now equals `-32602`; the old `-32002` returns only with `MCPGODEBUG=customresnotfounderrcode=1` (spec S9 says implementations MUST NOT emit `-32002`).
  - `auth.RequireBearerToken` (S19): parses `Authorization` with `strings.Fields`, requires exactly two fields with `bearer` (case-insensitive). Missing or malformed header: `401` "no bearer token". Verifier error unwrapping to `ErrInvalidToken`: `401`; `ErrOAuth`: `400`; anything else: `500`. Missing expiration (unless `AllowMissingExpiration`): `401`; expired beyond `ClockSkew`: `401`. Scope check runs before expiry: any required scope missing gives `403` "insufficient scope". `WWW-Authenticate: Bearer resource_metadata="...", scope="..."` is emitted on 401 and 403 only if `opts` has `ResourceMetadataURL` and/or `Scopes` set. There is no `error="insufficient_scope"` parameter and `scope=` lists all required scopes, not the missing ones. Spec S3 `#runtime-insufficient-scope-errors` is SHOULD-level, so this is a gap, not a violation. `TokenInfo` is available via `TokenInfoFromContext`.
  - `auth.ProtectedResourceMetadataHandler` serves RFC 9728 JSON with `Access-Control-Allow-Origin: *`, no field validation.
  - Things the release notes claim but that should be verified at runtime, not assumed: `ttlMs`/`cacheScope` are emitted on `tools/list` and `server/discover`; `structuredContent` is mirrored into a `TextContent` block; tool arguments are validated against `inputSchema` before the handler runs; `tools/list` ordering is deterministic; `ReadOnlyHint`/`IdempotentHint` are always serialized (v1.7.0 release note).
- Advisories (S17): four published, all High, all before the window, all patched at or before v1.4.1:
  - GHSA-wvj2-96wp-fq3f / CVE-2026-27896, 2026-02-25: case-insensitive `encoding/json` key matching (CWE-178/436), fixed 1.3.1. https://github.com/advisories/GHSA-wvj2-96wp-fq3f
  - GHSA-89xv-2j6f-qhc8 / CVE-2026-33252, 2026-03-18: cross-site tool execution for HTTP servers without authorization (CWE-352, CVSS 7.1), affected <= 1.4.0, fixed 1.4.1. https://github.com/modelcontextprotocol/go-sdk/security/advisories/GHSA-89xv-2j6f-qhc8
  - GHSA-q382-vc8q-7jhj (no CVE), 2026-03-18: null-Unicode key folding in `segmentio/encoding`, affected <= 1.4.0, fixed 1.4.1. https://github.com/modelcontextprotocol/go-sdk/security/advisories/GHSA-q382-vc8q-7jhj
  - GHSA-xw59-hvm2-8pj6 / CVE-2026-34742, 2026-03-30: DNS rebinding protection disabled by default for localhost servers (CWE-1188), fixed 1.4.0. https://github.com/modelcontextprotocol/go-sdk/security/advisories/GHSA-xw59-hvm2-8pj6
  - No go-sdk advisory has been published since 2026-06-09. v1.7.0 is not affected by any of the four. The v1.8.0-pre.1 hardening (JSON depth cap, size limits, DCR response bound) shipped without an advisory. Go vuln DB entry for the first: https://osv.dev/vulnerability/GO-2026-4569

## 5. Consolidated audit checklist (each item cites its source)

Transport
- [ ] Single POST endpoint; GET/DELETE return 405 with `Allow: POST` (S1 `#earlier-streamable-http-revisions`; SDK does this in stateless mode, S18).
- [ ] `Origin` header validated with 403 on invalid (S1 `#security--endpoint`). SDK default is NO check; wrap with `http.NewCrossOriginProtection` (S18).
- [ ] Host-header DNS rebinding check left enabled (`DisableLocalhostProtection` false) (S18; GHSA-xw59-hvm2-8pj6).
- [ ] Decide policy on absent `MCP-Protocol-Version`: SDK accepts as 2025-03-26; spec MUST reject if legacy clients are unsupported (S1 `#protocol-version-header`).
- [ ] `Mcp-Method`/`Mcp-Name`/`Mcp-Param-*` header-body mismatch yields 400 / -32020 (S1 `#server-validation`; SDK does this, S18).
- [ ] No sensitive parameter marked `x-mcp-header` (S7 `#x-mcp-header`).
- [ ] Content-Type 415 and Accept 400 checks not disabled via MCPGODEBUG (S18).
- [ ] `MaxRequestBodyBytes` left at default or explicitly set; never negative (S18).
- [ ] `PropagateRequestCancellation: true` so handlers stop on disconnect (S1 `#cancellation`; S18).
- [ ] No `Mcp-Session-Id` minted or echoed; `allowsessionsinstateless` not set (S1; S18).
- [ ] TLS terminated somewhere in front; no plaintext HTTP outside loopback (S4 `#communication-security`; S26).
- [ ] If bound to 0.0.0.0, that is deliberate and auth is enforced (S1 `#security--endpoint`).

Authorization
- [ ] Bearer token only in `Authorization` header; never in query string (S3 `#token-requirements`).
- [ ] Token verified on every request; audience bound to this server; 401 on invalid/expired (S3 `#token-handling`; S4).
- [ ] Token never forwarded upstream (S4 `#access-token-privilege-restriction`; S13 `#token-passthrough`).
- [ ] 401 vs 403 vs 400 mapping correct (S3 `#error-handling`; SDK mapping in S19).
- [ ] If OAuth framework adopted: RFC 9728 metadata served, `WWW-Authenticate` carries `resource_metadata` and `scope` (S3, S5). If static bearer: document that OAuth discovery is intentionally not implemented (S3 `#protocol-requirements`).
- [ ] Tokens never logged; log redaction for `Authorization` and tool arguments (S4 `#token-theft`; S21; S26).
- [ ] Missing-expiration and clock-skew options set deliberately (S19).

Tools
- [ ] `tools` capability declared; `tools/list` deterministic, static per token, never varies with call count or connection (S7 `#capabilities`; S22 rug-pull).
- [ ] Every tool has non-null `inputSchema`; no-arg tools use `{"type":"object","additionalProperties":false}` (S7 `#tool`).
- [ ] Inputs validated server-side against schema (types, ranges, formats); no shell or SQL concatenation (S7 `#security-considerations`; S25; S20).
- [ ] `readOnlyHint`/`idempotentHint`/`destructiveHint`/`openWorldHint` reflect real behaviour; read-only tools carry no write capability (S7 `#tool`; S26; S28).
- [ ] Descriptions factual, no instructions to the model, no embedded secrets (S25; S20; S22).
- [ ] `structuredContent` conforms to `outputSchema` and is mirrored into a text block (S7 `#structured-content`).
- [ ] Tool execution errors use `isError: true` with actionable text; protocol errors use JSON-RPC codes; internal details not leaked (S7 `#error-handling`; S28; S30).
- [ ] Outputs sanitized: no config, env, tokens, or internal paths echoed (S7 `#security-considerations`; S23 CVE-2026-67357; S24).
- [ ] Untrusted external content in outputs delimited and labelled as data (S25; S20).
- [ ] Rate limiting on `tools/call` via middleware (S7 `#security-considerations`; S21; S26).
- [ ] Any cross-call handle is random, opaque, bound to the verified user, expiring (S13 `#state-handle-hijacking`; S7 `#stateful-tools`).

Pagination and caching
- [ ] Cursors opaque, stable, integrity-checked, never a path or raw offset; invalid cursor returns -32602 (S8; S24).
- [ ] `ttlMs >= 0` and `cacheScope` on `tools/list` and `server/discover`; `"private"` if the list varies by token (S11).

Protocol hygiene
- [ ] `server/discover` implemented and reachable (S10).
- [ ] Missing `_meta` fields produce 400 / -32602; unsupported version produces 400 / -32022 with `supported` list (S9 `#_meta`; S6).
- [ ] No emission of -32002 or -32042 (S9 `#error-codes`).
- [ ] `serverInfo` in result `_meta` (S9; S10).
- [ ] JSON Schema 2020-12 default; no network `$ref` dereference; schema complexity bounded (S9).

Supply chain and ops
- [ ] go-sdk >= v1.4.1 (all four GHSAs); on v1.7.0 now; track v1.8.0 for the removed MCPGODEBUG flags and `SupportedProtocolVersions` (S16, S17).
- [ ] Dependency scanning and SBOM in CI (S29; S26).
- [ ] Structured request logging with caller identity, tool name, outcome, duration; arguments redacted (S21; S25; S26).
- [ ] Human-in-the-loop is a client obligation; server should not assume it and should keep read-only tools genuinely read-only (S7 `#user-interaction-model`).

## 6. Caveats

- Unofficial mirror: modelcontextprotocol.info (S30) is a third-party site, undated, references 2024-11-05 as "current", and contains no MCP-transport-specific guidance. Weighted low.
- Unwatched video: S31 (youtu.be/BqRhBq-_kgE, Neon Postgres, 2026-09-08). Only page metadata and the description were fetched; the video content was not reviewed and nothing in this file is attributed to it beyond the description. Its topic (workflow-shaped tools, progressive discovery) is tool-design ergonomics, not security.
- Normative standing of S13: the `specification/2026-07-28/basic/security_best_practices` URL serves the same document as the tutorial URL, and the authorization spec links to it under `/docs/.../tutorials/`. The binding auth security requirements are in `basic/authorization/security-considerations` (S4). Treat S13 MUSTs as guidance the spec incorporates by reference, not as an independent normative section.
- URL resolution: `basic/lifecycle` resolves to "Versioning and Compatibility" (S6); `basic/transports` is only an overview, the Streamable HTTP requirements are on the sub-page (S1); pagination lives at `server/utilities/pagination` (S8).
- Undated sources: only S30. S26 is dated but marked "draft". All spec pages carry the revision date, not a page date.
- Out-of-window sources (S17 advisories, S26, S27, S28, S29, S30) are included because the task named them or because they remain the canonical reference; they are not presented as new. Few sources in the window are specifically about server best practice; the in-window material is the spec itself (S1..S13), the project blog (S14, S15), the go-sdk releases (S16), and vendor or incident posts (S20..S25).
- Not fetched: the OWASP MCP guide PDF body (S27), the three August CVE records (data via S23), the SEP documents themselves (SEP-2575, 2567, 2243, 2549, 2322 are cited via the changelog S12 and blog S14, not read directly), the Go 1.25 `net/http` CrossOriginProtection docs, the Microsoft post body was obtained via a text-render proxy (r.jina.ai) after the direct fetch returned only a title.
- Secondary WebFetch summarisation: spec text was extracted by a summarising fetch tool; quoted MUST/SHOULD wording was checked against the returned page text but anchors were inferred from headings and should be spot-checked when cited in the audit.
