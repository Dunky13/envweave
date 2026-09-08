# Stable workflow signing handoff

Owner request: implement the SOPS/OpenBao release combination while 1.0 waits for
design work. Work branch: `t3code/guide-release-1-0-online`, based on
`119c0833234c8c54784c20d2cebb3d957e074db1`. Implementation and the operator runbook are delivered together on this branch.
The owner authorized commit and push; merging and activation remain separate.
No stable tag, production signature, trust activation, live key migration,
repository-setting mutation or release publication occurred.

## Entry points

- [Operator guide](../release/online-signing.md), published by the docs build at
  `/release/online-signing/` and linked from `/release/signing/` and installation.
- [ADR amendment](../adr/stable-workflow-signing.md) records recovery delegation,
  publication authority, accepted online custody and first-party provenance limits.
- `scripts/release/ceremony.sh` defaults to the online helper. `--offline` keeps
  the legacy flow explicit. `--help` is mutation-free.
- `scripts/release/stable` prepares public policy/candidates/metadata and verifies
  complete releases. It never reads private signing keys.

## Final behavior

One recovery-signed stable policy delegates routine release signatures to the
exact GitHub workflow/tag/commit identity and pins numeric repository/owner IDs,
issuer, runner, SCT, Fulcio/Rekor material and existing authority inventories.
No actual `release/trust/stable-policy*` files were generated or installed.

The tag workflow signs a complete draft. The maintainer reviews its manifest hash
and publishes using the helper from clean current main, with their existing `gh`
login. GitHub's immutable-release settings API requires Administration:read;
the built-in Actions token lacks that permission. Therefore publication is local,
not an unattended workflow requiring another PAT. The live read-only check
returned `enabled: true, enforced_by_owner: false` on 2026-09-08. This value is
time-specific and the helper rechecks it before publication.

Publication validates tag signature and exact-commit CI, checks signed inventory
and real OCI subjects, rechecks draft inventory before publication, and verifies
public immutable bytes afterward. An already public retry verifies without edits.
Outstanding drafts and unsynchronized published trust block the next candidate.

Trust/runtime, self-update and offline bundle assembly support keyless stable
evidence, explicit authorized historical releases and current revocation. The
CLI and runtime share verification floor encoding, old-state migration and lock
paths. Retrying a verified draft works before canonical trust promotion.

The installer pins standalone verifier hashes before execution. Public-key
filenames from unverified metadata cannot replace bootstrap root/verifier/trust
paths; `.pub` namespace and duplicate/recovery-name refusal are regression tested.

## Verification evidence

- Go suites passed for `internal/releasetrust`, `internal/selfupdate`,
  `internal/upgradebundle`, `internal/upgradeassembly`, `scripts/release/stable`
  and `scripts/ci`; related `scripts/release`, `internal/buildcompat`,
  `internal/app`, `internal/upgradegate`, `internal/upgradecompat` and
  `internal/hostupgrade` suites passed. Scoped `go vet` passed.
- Synthetic Sigstore tests exercise actual certificate, SCT, SET, Merkle proof
  and checkpoint verification. They reject identity substitution, rollback,
  policy escalation, false provenance and current revocations. Complete CLI
  round trips include keyless and legacy historical bundles. OCI subprocess
  tests assert exact Cosign arguments and pinned public bytes using a fake
  registry command; they are not live stable-release proof.
- Real Cosign legacy fixture suite passed with temporary test-only keys. Shell
  fixtures cover ceremony refusal/retries, manifest production, installer
  bootstrap attacks, workflow signing/publication gates, Homebrew branch safety
  and nightly regressions. Actionlint, ShellCheck and Helm checks passed.
- Standalone verifier cross-builds passed for Linux/macOS/Windows on amd64/arm64.
  Docs `pnpm --dir docs/site run verify` passed: zero diagnostics, 47 routes,
  OSS/PWA checks and browser offline-route test.

The initial broader Go command included two nonexistent package names
(`internal/upgradeplan`, `internal/upgradejob`), so that invocation exited 1
despite its actual package suites passing. It was corrected to the real
`upgradegate`, `upgradecompat` and `hostupgrade` packages, which passed.

## Review disposition

Three review rounds completed with CLEAN / SOUND verdicts for trust/runtime,
installer/bootstrap and ceremony/publication. Fixed findings include shared
state migration/locking, draft retry floors, historical authorization, installer
bootstrap overwrite, unsafe fixed-branch retries and impossible workflow-token
administration access. The final guide distinguishes publication's settings/tag/CI
checks from the build's tag-immutability probe. Review really performs remote
Cosign verification: `--published` reaches the Go verifier's subprocess path.

## Resume and activation

Review and merge the implementation before using the operator commands. Then
verify recovery custody and independent Vaultwarden backups, run `ceremony.sh
setup`, inspect/sign the proposed delegation, and merge its public trust PR.
`ceremony.sh status` must authenticate the activated policy before a release tag.
Vaultwarden storage and restore have not been inspected or modified.

Keep 1.0 paused until designs and final-candidate acceptance are approved. Earlier
OSS/live-provider evidence does not automatically cover the eventual release
commit. Publishing is a separate deliberate owner action; do not run setup,
build or publish merely to demonstrate this implementation.
