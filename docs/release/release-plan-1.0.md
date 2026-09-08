# Hikyo 1.0 release plan

Written 2026-09-07 during the autonomous release run (decision log:
[release-run-2026-09-07.html](../reports/1.0/release-run-2026-09-07.html)).
This is the step-by-step order of operations from the current `main` to a
published, signed, verifiable `v1.0.0`. Steps are split into what an agent
can do without keys or external credentials and what only the owner can do.
The [acceptance ledger](acceptance-1.0.md) stays the authority for every
criterion's status; this plan only sequences the remaining work.

## Where things stand

- 42 of 49 acceptance criteria are INTERNAL-PASS at the tested candidate.
  Nothing code-side is red.
- Still open: S1 freeze (needs the tag), O3 signing ceremony, O4 public
  disclosure endpoint check, M4/GH-E2E/GH-CONTRACT live GitHub run, and the
  MCP named-client proof (#651). A7 social sign-in is POST-1.0 by owner
  decision.
- The trust root in `release/trust/` is real and recovery-signed, with
  `pending_release` 1.0.0 at sequence 1. It was generated for nightly
  activation under the online-custody exception in
  `release/trust/BOOTSTRAP.md`, not by the offline ceremony in
  `signing.md`. Step 10 below decides what to do with that.
- The six GitHub security advisories for the 2026-09-05 fork merge are
  published (2026-09-08); CVE assignment is pending with GitHub.
- Since the last acceptance refresh (main `549726d9`), 28 further PRs merged
  (upgrade lane, self-configuration, privacy controls, UI audit follow-ups).
  The candidate that gets tagged must be re-accepted at its own commit;
  ordinary CI covers most of that automatically on every PR.

## Phase A: land the release-run PR (agent, done or in flight)

1. Merge PR #710 (self-config refusal names the key). Auto-merge is armed.
2. Merge the release-run PR #711: repository hygiene, dead-code and
   duplication sweep, status-ledger corrections, this plan.
3. Confirm `main` CI is green after both merges (the `ci-required` aggregate
   plus the isolation, race and fuzz shards).

## Phase B: candidate freeze preparation (agent, no keys needed)

4. Regenerate the CLI goldens one last time and review every diff as a spec
   change. Done 2026-09-07 at `8ecf1b9c`: `go test ./internal/cli -run
   'Golden|Frozen' -update` produced no diff. Repeat at the final candidate
   if any CLI change merges after it.
5. `x-extensible-enum` audit on `api/openapi.yaml`. Done 2026-09-07: twelve
   enums are declared extensible, including the two the handoff flagged
   (`AdapterProvider`, `SamlProviderWarning.code`). Everything else closes at
   the tag and can never grow.
6. Re-walk `docs/spec/self-hoster-checklist.md` against a fresh
   `go build -tags ui` binary, the chart and the compose file at the
   candidate commit. Append the date and commit to
   `docs/release/self-hoster-candidate.md`.
7. Refresh the acceptance ledger header to the candidate commit once its full
   CI run is green, linking that run. Re-run the four dedicated lanes
   (`floor-acceptance`, `floor-bench`, `operator-floor`, `ops-floor`) on the
   candidate via `workflow_dispatch` and link them.
8. Update `docs/site/src/content/docs/docs/roadmap.mdx` and `docs/status/`
   so nothing still describes 1.0 as future-tense once the tag exists. Keep
   `CAP-PUBLIC-RELEASE` partial until the release is public.

## Phase C: owner-only steps before the tag

9. **Publish the six draft advisories.** Done 2026-09-08 07:39 UTC on the
   owner's instruction: GHSA-8h6m-jpwj-v83p, GHSA-2qjg-h73x-x9j4,
   GHSA-xv87-mmx6-hw25, GHSA-hqjh-jg7p-qr2m, GHSA-wxjf-x869-pfh4 and
   GHSA-p29g-pjgc-jqmv are published with affected range
   `<= 0.0.1-nightly.20260905.23.g907e2f41` and patched version
   `0.0.1-nightly.20260906.24.g90b4ca6a` (the first nightly containing fix
   commit `3700a0ef`). CVE IDs were requested for all six; GitHub assigns
   them asynchronously. The 1.0.0 release notes should list them.
10. **Decide the signing root custody (Q5 in the decision log).**
    Options: (a) accept the existing `recovery-1` / `primary-1` root, generated
    online under the documented exception, for 1.0; (b) run the offline
    ceremony in `signing.md` for a fresh primary and rotate to it under the
    existing recovery root; (c) replace the recovery root too, which breaks
    every nightly installation's pinned trust and the legacy bridges.
    Recommendation: **(b)**. The recovery root is already deployed to nightly
    users and bridges, so replacing it costs real installations; a fresh
    offline primary is cheap and restores the ADR posture for the key that
    actually signs release bytes. Record the choice in `BOOTSTRAP.md`.
11. **GitHub live provider run** for M4 / GH-E2E / GH-CONTRACT. Needs seven
    fine-grained PATs and the dedicated fixture repository, organisation and
    protected environment (see `internal/adapter/githubactions/contract_external_test.go`
    lines 40 to 60 for the exact token floors). Run with
    `HIKYO_GITHUB_CONTRACT=1` and the `HIKYO_GITHUB_CONTRACT_*` variables;
    commit the redacted result under `docs/reports/1.0/evidence/` and flip the
    three ledger rows.
12. **MCP named-client proof (#651).** Deploy the candidate image behind
    public HTTPS with two PostgreSQL replicas, run
    `scripts/mcp-public-smoke`, then exercise Inspector, the Go client, Codex
    CLI and Claude Code with a scoped production tool. Commit the redacted
    artifact; set each profile in the operator support table to supported,
    unsupported with reason, or not yet verified. Unverified profiles stay
    honestly marked; they do not block the tag.
13. **Quarterly PVR notification receipt.** Self-report through the private
    vulnerability channel and the fallback address, record the acknowledgement
    time in `release/repository/`, and refresh
    `fallback-channel-test.json` if it is older than 93 days at tag time.
14. **Public docs endpoint check** at the candidate:
    `scripts/ci/check-docs-live.sh https://hikyo.app security@developwent.io`
    and `scripts/ci/check-oss-policy.sh`. Both must pass from outside CI too.

## Phase D: tag and build (owner, mechanical)

15. Confirm `release/trust/metadata.json` still carries
    `pending_release` 1.0.0 at sequence 1 (the bootstrap metadata is already
    the candidate for the first release; no pre-tag metadata bump is needed).
    If step 10 chose a fresh primary, that rotation metadata must be
    recovery-signed and merged first.
16. Create the tag on the merged candidate commit: `git tag -s v1.0.0 <sha>`
    and push it. Tag creation is admin-only by ruleset; `check-tag.sh`
    refuses a reused version or a commit not reachable from `main`.
17. `release-build` runs: trust bootstrap verification, both-engine
    compatibility generation, GoReleaser archives, native packages, the
    distroless multi-arch image, the digest-pinned chart, SBOMs, provenance,
    the rendered installer and an **unsigned draft release**. Wait for it.
18. The `freeze-guard` CI job wakes up on its own once `v1.0.0` exists: it
    diffs `api/openapi.yaml` against the tag on every later PR that touches
    the API. No wiring change is needed; it is already in the required-job
    registry.

## Phase E: offline signing ceremony (owner, per `signing.md`)

19. Download every draft asset online; recompute `checksums.txt`; compare the
    GHCR image and chart index digests with the digest files; confirm the
    manifest matches `release-candidate.json`.
20. Offline, recovery key only: `scripts/release/bind-manifest.sh` binds the
    manifest into `metadata.bound.json` (sequence 2, `highest_release` 1.0.0),
    recovery-sign it, commit it as `release/trust/metadata.json` plus its
    signature, merge to `main`.
21. Offline, primary key only: `scripts/release/sign-bundle.sh` over every
    asset and both OCI payloads.
22. Online, no private key mounted: `scripts/release/publish-oci-signatures.sh`,
    upload the manifest and every `*.sigstore.json` to the draft.
23. Redownload the complete draft and run
    `verify-bundle.sh --published --state "$XDG_STATE_HOME/hikyo/release-trust.json"`.
    Only then publish the release. GitHub immutable releases lock the assets.
24. Homebrew: the ceremony renders `Casks/hikyo.rb` and opens the tap PR;
    review its CI and merge it separately.

## Phase F: after publication (agent can do most of it)

25. Flip `CAP-PUBLIC-RELEASE` to implemented in `docs/status/ledger.json`
    with the release URL and manifest hash as evidence; regenerate the README
    block with `scripts/ci/check-doc-status.mjs --write`.
26. Fill the "Official release" and "Freeze" rows of the acceptance ledger
    with the tag, manifest SHA-256, image and chart digests
    (`PENDING_IMAGE_DIGEST` / `PENDING_CHART_DIGEST` placeholders).
27. Close #79 and #41 with a comment linking the release, the ledger and the
    decision log. Update #527 so the roadmap's "1.0 critical path" section
    reads as history and "first after 1.0" (#605 to #613) becomes the active
    lane.
28. Nightly bridges: confirm the next nightly's `catalog.json` authorises an
    upgrade path from the last pre-1.0 nightly into 1.0.0 (`BRIDGES.md`), and
    that `hikyo upgrade` on a systemd host reaches 1.0.0 from a nightly.
29. Announce: docs site landing page, README badge, release notes naming the
    six advisories, the SOPS/KMS binary size decision and the post-1.0 lanes.

## What is deliberately not in 1.0

- Social sign-in and open registration (#605 to #615, #589): post-1.0 by owner
  decision on 2026-09-03; its breaking pre-freeze slice (#617) is merged.
- Kubernetes condition reporting to the server (#683): the UI shows the
  status as unknown, which is the honest posture until the reporting channel
  is designed.
- MCP human delegation / OAuth (#631): decision lane, no 1.0 claim.
- The #619 architecture leftovers (hand-written store SQL, TxAuthorizer
  pass-throughs, authz derivation, MachineAccess split): locked-decision
  material for an ADR, not pre-freeze refactoring.
- New delivery adapters, PKI, SSH certificates, Transit/KMS, repository
  scanning, temporary access, scheduled rotation (#148 to #164).
