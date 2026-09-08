# Release signing ceremonies in security-focused projects

Researched 2026-09-08. Scope: simplify Hikyo stable releases while the maintainer's Mac stays online, without routine USB handling. The operator uses self-hosted Vaultwarden.

## Finding

The examined projects demonstrate several credible online release models. An offline computer and multiple removable copies of signing keys are not a universal requirement for software that handles secrets. Artifact authentication, build provenance, release authorization, key isolation, and disaster recovery are separate controls; none substitutes for the others.

This is a comparison of published procedures and source code, not a security audit, certification, or numerical ranking. Repository settings, internal approvals, hardware custody, and incident procedures cannot be inferred from workflow YAML alone. The workflow snapshots below describe the inspected default branches; they are not proof that every historical release used those exact workflows. Release assets were inventoried, not cryptographically verified during this research.

## Vaultwarden: online CI build attestations

At commit `b7667e27bf3500a2446d39446b1a7b10b8b25991`, the release workflow runs for main-branch pushes and release tags. Its jobs use a GitHub `release` environment, restricted job permissions, `id-token: write`, and `attestations: write`. The workflow uses `actions/attest` for built binaries and container image digests, publishing container attestations to the configured registries. It obtains the signing identity through GitHub OIDC/Sigstore rather than referencing a maintainer-managed artifact-signing private key. Registry publishing credentials remain separate secrets. [V1]

This makes the GitHub repository, workflow identity, runner execution, and registry authorization part of the trust boundary. The existence of an environment named `release` does not establish that it requires a second person's approval. The inspected workflow does not document an offline signing step. Its latest stable release, `1.37.2`, has no attached GitHub release assets; the binary workflow artifacts and container attestations must not be described as downloadable stable-release assets on that basis. [V1, V2]

## SOPS: maintainer-approved release plus keyless signing and provenance

At commit `ba5d2731a3ab154c92837d41278a067b53673e30`, the maintainer procedure requires a release-preparation PR, approval from at least one other maintainer, green main CI, and a signed Git tag. Pushing that tag starts the automated release. [S1]

GoReleaser builds binaries, packages and SBOMs. Cosign signs the checksum manifest and container images using GitHub OIDC. Separate pinned SLSA generator workflows produce artifact and container provenance. The release templates provide commands that verify the checksum signature against the expected workflow identity and issuer, verify artifact hashes, and verify provenance. [S2, S3]

This combines human release authorization with CI identity and provenance. No persistent Cosign release private key needs to be recovered from a lost maintainer laptop. The human Git-tag signing identity still needs its own custody and recovery. The filenames of the SLSA generator workflows are not, by themselves, evidence of an independently audited project-wide SLSA level. The documented procedure does not require disconnecting the maintainer's machine. [S1, S2]

Latest stable release `v3.13.3` contains a checksum file, Sigstore bundle, and in-toto provenance among its assets. Their presence was checked through the release API; their signatures were not verified in this investigation. [S4]

## OpenBao: online GPG and Cosign signing, then human draft review

At commit `1859768bdb897d92723c68f8457128c9fd707a98`, the documented release manager starts a release workflow that stages a draft. The manager checks that binaries work and GPG/Cosign signatures verify, attaches release notes, and publishes the draft. Publication triggers the container and package-repository workflows, followed by further checks of container functionality and signatures. [O1]

The artifact release workflow imports `GPG_PRIVATE_KEY` and `GPG_PASSWORD` from GitHub secrets into its runner. Its signing script creates detached GPG signatures and Cosign bundles. The signing jobs grant `id-token: write`; Cosign signing does not pass a persistent key argument. Container publication likewise signs using Cosign in CI. The Debian repository workflow separately imports the GPG key and passphrase from CI secrets. [O2, O3, O4, O5]

This is direct public evidence of online signing with a long-lived software GPG key available to CI, alongside GitHub identity-based Cosign signing. It is not evidence that the GPG key is held in an offline HSM. A compromised authorized signing runner may be able to use or extract the imported key. Human review of a draft and verification of published artifacts remain valuable controls, but two signatures produced within the same compromised CI boundary are not two independent approvals. [O2, O3]

The inspected checklist does not require multiple USB devices or an offline maintainer machine. Latest stable release `v2.6.2` exposes 251 assets, including GPG signatures and Sigstore bundles. Product variants labelled HSM refer to OpenBao runtime features; they do not establish HSM custody for the project's release signing key. [O1, O6]

## Bitwarden: CI identity and platform-specific cloud signing

At server commit `4a58a9fa2e1ca364f3d6669babcfcd390189fe20`, non-PR publish-branch builds sign container image digests using `cosign sign --yes` with GitHub OIDC capability. No persistent Cosign key argument is supplied. Registry credentials fetched from Azure Key Vault are separate from the image-signing identity. [B1]

At clients commit `25a9a01a2c74de62dd95d205a1c4398fe80eefbc`, the Windows desktop signing integration invokes AzureSignTool using a signing-vault URL, client identity, client secret, and certificate name. This demonstrates remote Azure Key Vault certificate signing. It does not establish that the deployed signing key is HSM-backed or disclose its private backup arrangements. [B2]

The macOS CLI build shows a different arrangement: it exports Developer ID certificate material as a `.p12` from Azure Key Vault, imports it into a temporary runner Keychain, signs with Apple's Developer ID, and notarizes. This is explicit online CI signing with certificate material present on the runner, not evidence that every Bitwarden signing key is non-exportable. [B3]

The CLI publishing workflow also uses a central reusable workflow with an environment boundary, a deployment-bot actor restriction, release-commit checkout, promotion of an already built/tested artifact, and npm OIDC publishing with provenance. The repository YAML does not establish the required number of human release approvers. [B4]

## age: an online hardware signer plus public transparency

At commit `b74dce4cdbe35b5e5f66c06d9612b72f89028758`, age publishes an actual maintainer release playbook. It starts a Tillitis TKey hardware SSH agent with a user-supplied secret, retrieves that secret through `passage`, downloads the GitHub-built release artifacts, submits signatures to Sigsum through the hardware agent, and uploads the resulting proof files. Verification binds the artifact to an age public key and a public append-only transparency log under the chosen Sigsum policy. [A1]

This is a concrete online hardware-signing design. The signing device is not a USB backup stick, and the procedure does not require air-gapping or multiple removable storage copies. It still introduces hardware availability and recovery considerations, which are not documented in this playbook. The build workflow separately produces GitHub build-provenance attestations. [A1, A2]

The playbook explicitly says that reproducing artifacts locally and signing those instead of the GitHub Actions builds is future work. Therefore the inspected procedure must not be described as independently rebuilding and comparing all released binaries. It also does not establish a multiple-maintainer approval quorum. [A1]

## What the security choices mean

| Approach | Principal benefit | Remaining exposure |
| --- | --- | --- |
| Local encrypted keys with Keychain | Simple operation; persistent release keys stay outside CI | A compromised unlocked Mac may use or extract signing credentials |
| CI keyless signing with OIDC | No long-lived artifact-signing key to lose; signed workflow identity and transparency evidence | An attacker controlling an authorized workflow may produce validly signed malicious output |
| Managed or hardware signing | Depending on the mechanism, private-key export can be prevented | Authorized malicious signing requests remain possible; review and access controls still matter |
| Offline signing with separated custodians | Can isolate keys from the online build/publish environment | More handling and recovery complexity; signing malicious input remains possible without independent verification |

These are dimensions of protection, not a weakest-to-strongest certification ladder. A signature authenticates an artifact or statement. It does not establish that the code is safe. A checksum generated beside an artifact does not independently establish source-to-binary correspondence.

## Recommendation for Hikyo

**Decision update, 2026-09-08:** The owner chose to implement the SOPS-style
keyless migration now, with OpenBao-style reviewed draft publication, while
1.0 waits for design work. The earlier staged recommendation below is retained
as research history. The [online guide](../release/online-signing.md) and
[ADR](../adr/stable-workflow-signing.md) now define the selected implementation;
no stable policy has been activated as part of this work.

For 1.0, a supported online ceremony using the existing encrypted primary and recovery keys with macOS Keychain is a reasonable proposed compromise. It preserves the already distributed trust identities while removing network-disconnection and USB-transfer steps. Hikyo's checked-in bootstrap record already documents this local online custody for nightly activation, but explicitly says it must be considered separately before stable activation. This research does not activate stable signing or modify that policy. [H1]

Retain the final candidate checks, signed commits/tags, exact manifest binding, signature verification, immutable published versions, explicit publication confirmation, and verification of downloaded published bytes. A resumable script can perform most of the mechanics. Keeping primary and recovery roles does not create independent security domains if both are accessible from the same Mac.

For subsequent releases, SOPS is a useful model: an approved candidate, signed tag, keyless release signatures, and build provenance. Hikyo already documents OIDC-signed nightlies, but extending that policy to stable releases still requires an explicit verifier/trust-policy migration. Replacing a pinned stable key with CI identity is not merely changing the ceremony's command-line flags. [H1, S1, S2]

Vaultwarden can store recovery material, but the current installation is not automatically a disaster-recovery copy. Store the encrypted signing keys, the information needed to unlock them, public-key fingerprints, and restoration instructions. Ensure the recovery information is not available only in the Keychain of the lost Mac. Preserve an additional encrypted backup outside both that Mac and the Vaultwarden host, with a way to unlock it if Vaultwarden is unavailable.

If keys are stored as Vaultwarden attachments, back up the attachments as well as a consistent database backup. Vaultwarden's own documentation explicitly identifies attachments as separate from the database and calls their backup required. It also recommends a remote copy and periodically testing restoration. This research did not inspect the operator's running Vaultwarden instance or establish that its backups currently meet those conditions. [V3]

## Primary sources

- [V1] [Vaultwarden release workflow](https://github.com/dani-garcia/vaultwarden/blob/b7667e27bf3500a2446d39446b1a7b10b8b25991/.github/workflows/release.yml), especially environment/permissions and binary/container attestation steps.
- [V2] [Vaultwarden 1.37.2 release API](https://api.github.com/repos/dani-garcia/vaultwarden/releases/tags/1.37.2).
- [V3] [Vaultwarden: Backing up your vault](https://github.com/dani-garcia/vaultwarden/wiki/Backing-up-your-vault), retrieved 2026-09-08; wiki content is mutable.
- [S1] [SOPS release procedure](https://github.com/getsops/sops/blob/ba5d2731a3ab154c92837d41278a067b53673e30/docs/release.md).
- [S2] [SOPS release workflow](https://github.com/getsops/sops/blob/ba5d2731a3ab154c92837d41278a067b53673e30/.github/workflows/release.yml).
- [S3] [SOPS GoReleaser configuration and verification instructions](https://github.com/getsops/sops/blob/ba5d2731a3ab154c92837d41278a067b53673e30/.goreleaser.yaml).
- [S4] [SOPS v3.13.3 release API](https://api.github.com/repos/getsops/sops/releases/tags/v3.13.3).
- [O1] [OpenBao release checklist](https://github.com/openbao/openbao/blob/1859768bdb897d92723c68f8457128c9fd707a98/website/content/community/policies/release.mdx).
- [O2] [OpenBao release workflow](https://github.com/openbao/openbao/blob/1859768bdb897d92723c68f8457128c9fd707a98/.github/workflows/release.yml).
- [O3] [OpenBao artifact signing script](https://github.com/openbao/openbao/blob/1859768bdb897d92723c68f8457128c9fd707a98/scripts/release/sign.sh).
- [O4] [OpenBao container release workflow](https://github.com/openbao/openbao/blob/1859768bdb897d92723c68f8457128c9fd707a98/.github/workflows/release-images.yml).
- [O5] [OpenBao package publication workflow](https://github.com/openbao/openbao/blob/1859768bdb897d92723c68f8457128c9fd707a98/.github/workflows/release-packages.yml).
- [O6] [OpenBao v2.6.2 release API](https://api.github.com/repos/openbao/openbao/releases/tags/v2.6.2).
- [B1] [Bitwarden server build and image signing](https://github.com/bitwarden/server/blob/4a58a9fa2e1ca364f3d6669babcfcd390189fe20/.github/workflows/build.yml#L358).
- [B2] [Bitwarden Windows signing integration](https://github.com/bitwarden/clients/blob/25a9a01a2c74de62dd95d205a1c4398fe80eefbc/apps/desktop/sign.js#L14) and [desktop workflow](https://github.com/bitwarden/clients/blob/25a9a01a2c74de62dd95d205a1c4398fe80eefbc/.github/workflows/build-desktop.yml#L552).
- [B3] [Bitwarden macOS CLI certificate import and signing](https://github.com/bitwarden/clients/blob/25a9a01a2c74de62dd95d205a1c4398fe80eefbc/.github/workflows/build-cli.yml#L206) and [notarization](https://github.com/bitwarden/clients/blob/25a9a01a2c74de62dd95d205a1c4398fe80eefbc/.github/workflows/build-cli.yml#L276).
- [B4] [Bitwarden central CLI publishing workflow](https://github.com/bitwarden/gh-actions/blob/b05a37717d558616c348e76af8b9077cbbb8aff1/.github/workflows/_publish-cli-npm.yml).
- [A1] [age Sigsum verification and maintainer release playbook](https://github.com/FiloSottile/age/blob/b74dce4cdbe35b5e5f66c06d9612b72f89028758/SIGSUM.md).
- [A2] [age build workflow and provenance](https://github.com/FiloSottile/age/blob/b74dce4cdbe35b5e5f66c06d9612b72f89028758/.github/workflows/build.yml).
- [H1] [Hikyo's existing nightly bootstrap and custody record](../../release/trust/BOOTSTRAP.md).
