# Preview automation: #721, #722, #723, #725, #730

This branch implements the five issues as one coordinated security change.
The originating tickets are [#721](https://github.com/Hikyo-Org/Hikyo/issues/721),
[#722](https://github.com/Hikyo-Org/Hikyo/issues/722),
[#723](https://github.com/Hikyo-Org/Hikyo/issues/723),
[#725](https://github.com/Hikyo-Org/Hikyo/issues/725), and
[#730](https://github.com/Hikyo-Org/Hikyo/issues/730).

## Viability and dependency order

The lifecycle request fits the existing automation capability model. Opening
machine transport admission does not grant capabilities. Environment creation
and deletion remain governed by project-wide `definitions-edit`; a `pr-*` name
is not an authorization boundary. Hikyo has no production-tier concept.
Dedicated preview projects provide the intended isolation. A machine cannot
delete a protected environment, including through a definitions plan.

Address #730 before the executable #722 guide. #721 and #725 are independent
integration prerequisites when the deployment uses cluster-CA federation or
native Kubernetes consumers. #723 is the separate alternative for previews
whose differences are public configuration only, with explicit amendments to
the flat, revision and API models. It does not provide per-preview secrets.

## Delivered behavior

- #721: instance administrators can set or clear a per-issuer PEM CA bundle.
  It replaces system roots for that issuer's discovery/JWKS fetches, including
  redirects. HTTPS, hostname verification, bounded responses and explicit
  private-network egress policy remain effective. Static JWKS rejects a bundle.
  Cached keys and the final authorization transaction are bound to CA changes.
- #730: machine credentials can reach environment metadata/create/clone/delete,
  staging/clear, copy, publish, pin-list and revision-metadata routes. Live grants,
  machine-reveal opt-in, protected publication and approval policies still bind.
  Copying secrets requires source reveal and destination reveal/publish.
- #730: machine CLI export uses delivery rather than the human historical-export
  endpoint. No reveal requests config-only projection; reveal refuses before
  output if any set secret is withheld. Machines cannot select an arbitrary
  `--revision`. The response identifies the actual selected snapshot without a
  second metadata lookup. Workload history remains pin-bound.
- #722: the discoverable preview guide resolves names to IDs, uses stable
  per-preview secrets, explicitly publishes returned draft versions, stops on
  approval requests, exports through private files, and covers teardown,
  credential revocation, source-copy omissions and capacity limits.
- #723 and #725: see the detailed
  [parameter handoff](723-environment-parameters.md) and
  [native Secret handoff](725-native-secret-types.md).

Environment teardown also releases only the deleted environment's scoped grants
and their origins, invalidating affected principals in the same transaction.
Without this, existing grant foreign keys could refuse cleanup with a conflict.
Both direct deletion and definitions-plan deletion use this cleanup.

## Parameter boundaries

Only declared `${NAME}` references in config values are substituted, once.
Secret values remain literal. Supplied values are bounded public inputs and
appear in disclosure audit records. Missing/invalid inputs or invalid resolved
config refuse the entire fetch. Inputs bind delivery cursors and operator stamps.

Declarations activate on the next publication. Each snapshot preserves its
contract and relevant config schemas, so edits cannot reinterpret historical
delivery. Contracts are bounded to 256 KiB and charged to snapshot storage
budgets. Hidden secret bytes do not contribute to a caller-visible render limit.

Parameter declaration management is CLI/API, with a narrowly documented UI
parity exception. Definitions bundles do not carry these declarations. A
parameter-template source cannot be cloned into an undeclared destination;
that operation refuses atomically. Use the parameter delivery workflow or a
concrete clone source, as described in the guide.

## Verification and delivery state

Dedicated isolation regressions cover the real machine HTTP lifecycle, foreign
scope refusal, credential revocation, protected deletion, config-only and revealed
delivery, human-only historical export, environment grant cleanup, CA changes,
snapshot parameter history, parameter audit, atomic schema refusal and bounded
contracts. CLI tests cover real command routing and secure-output refusal.

Local validation completed on 2026-09-12:

- Default Go package coverage completed with failed checks fixed and rerun.
  PostgreSQL-heavy packages were serialized after the initial combined run
  exposed checkpoint contention. The service package passed 563 tests and app
  passed 299, including actual SQLite/PostgreSQL restore and upgrade drills.
- Isolation covered 379 top-level tests across three sequential SQLite/PostgreSQL
  shards, plus one new static regression. Shards 0 and 1 passed outright. Shard 2
  passed every database case; its sole failure was an old static capability
  checker that omitted the existing conditional machine-reveal opt-in. The
  corrected checker, new negative cases, and API pin-bound-history guard passed
  separately. No runtime code changed for that final correction.
- `go vet ./...`, repository-wide gofmt, diff checks, focused race suites,
  regenerated Go/CRD drift checks, chart structural/mutation checks, generated
  TypeScript verification (20 tests), web typecheck and 959 web tests passed.
  Docs verification built 61 pages and passed navigation, policy, CSP and browser
  offline checks.
- Ordinary Standards and Spec reviews completed without unresolved code findings.
  The cross-provider review was explicitly skipped by the user.
- Existing external Forgejo, frozen-client/server and MCP deployment fixtures
  were not configured. The new native Kubernetes acceptance fixture compiles
  and is wired into CI, but local kind bootstrap failed before its tests ran.
  See the #725 handoff for exact diagnostics and the shared Docker memory limit.
  Both owned clusters were removed; the existing cluster subsequently reported
  node Ready and API `/readyz` healthy. No native consumer pass is claimed.

No push, PR, merge, deployment or external issue closure was performed.

Migrations 51 and 52 apply to SQLite and PostgreSQL. The generated development
compatibility manifest includes both. Generated Go, TypeScript and CRD artifacts
belong to this change; regenerate them from their source definitions.
