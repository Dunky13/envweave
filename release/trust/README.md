# Production trust bootstrap

This directory holds the committed public trust material: `root.json`,
`recovery-1.pub`, `primary-1.pub`, recovery-signed `metadata.json`
(`metadata.sigstore.json`) and the nightly `catalog.json`. The `*.example`
files are the fixture shapes used by `scripts/release/test-fixtures.sh`.

The keys behind this root were generated on 2026-09-06 for nightly activation
under the custody exception recorded in [BOOTSTRAP.md](BOOTSTRAP.md): on a
network-connected Mac, encrypted private keys retained locally, passphrases in
the macOS Keychain. That record, not `docs/release/signing.md`, describes how
these particular keys came to exist. Stable release activation must consider
that custody first; see `docs/release/release-plan-1.0.md`.

Never commit `*.key`, decrypted key material, or a signing passphrase. Release
CI refuses to run until the production bootstrap files exist and verify.
