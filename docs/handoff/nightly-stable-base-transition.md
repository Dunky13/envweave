# Nightly base versions and stable transition eligibility

Nightlies now retain the latest stable major/minor/patch version instead of
incrementing its minor version. A nightly may contain migrations added after
that base. Consequently, `1.1.0-nightly...` must never update to stable `1.1.0`;
stable `1.1.1` or higher is eligible only with the existing compatibility proofs.

`releaseidentity.CompareUpdateVersions` supplies shared update ordering for
release selection and binary installation. Within a matching version core, an
explicit nightly sorts after stable. The binary installer checks the installed
and selected versions again before downloading, rather than trusting the
availability flag. Recovery bridge verification independently rejects a stable
target at or below the nightly base even if its signed sequence increases.

Existing published next-minor nightly identities remain unchanged and receive
the same conservative eligibility rule. This can delay their next update until
a release exceeds that historical base; no identity or migration history is
rewritten. Initial pre-stable nightlies retain their existing `0.0.1` base.

The automatic server upgrade command still supports only nightly targets.
This change does not add stable server automation or relax authenticated routes,
operator attestation, migration sequence checks, or backup/restore proof.

Regression coverage includes both channels, same-base and older stable refusals,
build metadata, higher patch/minor/major stable eligibility, nightly selection
after its stable base, installer refusal before network/file access, and signed
bridge refusal despite increasing release sequence. The release shell fixture
checks stable-base naming including nonzero patch versions.

Validation passed: Go tests for `releaseidentity`, `updatecheck`, `selfupdate`,
`releasetrust`, `upgradecompat`, `upgradebundle`, `cli`, `app`, `store/upgrade`,
`upgradegate`, `scripts/release/nightly`, and `scripts/release/assemble-upgrade`;
Go vet for the four changed production packages; nightly release shell fixtures;
ShellCheck; nightly workflow actionlint; docs site check; and diff whitespace.

Cross-provider review was skipped at the user's request. The user subsequently
authorized signed commit, push, PR creation, and merge after green CI.
