package selfupdate

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/releasetrust"
	"github.com/Hikyo-Org/hikyo/internal/releasetrust/testfixture"
	"github.com/Hikyo-Org/hikyo/internal/updatecheck"
	"github.com/gofrs/flock"
)

func TestStableWorkflowInstallerAuthenticatesBeforePersisting(t *testing.T) {
	for _, mutation := range []string{"none", "canonical old", "installer state", "legacy state", "installer lock", "missing policy", "wrong archive signature", "rollback"} {
		t.Run(mutation, func(t *testing.T) {
			f, m, stable, sign := testfixture.StableWorkflow(t)
			installer, status, _, responses := installerFixtureForVersion(t, "1.0.0", false, "")
			status.Channel, status.Immutable = updatecheck.ChannelStable, true
			installer.config = Config{StateDir: t.TempDir(), TrustRootBase64: base64.StdEncoding.EncodeToString(f.Pinned.Root), RecoveryKeyBase64: base64.StdEncoding.EncodeToString(f.Pinned.RecoveryPublicKey)}
			for name, raw := range map[string][]byte{"metadata.json": m.Metadata, "metadata.sigstore.json": m.MetadataSignature, "catalog.json": m.Catalog, "catalog.sigstore.json": m.CatalogSignature, "stable-policy.json": m.StablePolicy, "stable-policy.sigstore.json": m.StablePolicySignature, "stable-trusted-root.json": m.StableTrustedRoot, "primary.pub": f.PrimaryPublic} {
				responses[trustURL(name)] = raw
			}
			const archiveName = "hikyo_linux_arm64.tar.gz"
			archive := []byte("synthetic platform payload for 1.0.0")
			const base = "https://github.com/Hikyo-Org/Hikyo/releases/download/v1.0.0/"
			for name, raw := range map[string][]byte{"metadata.json": m.Metadata, "metadata.sigstore.json": m.MetadataSignature, "catalog.json": m.Catalog, "catalog.sigstore.json": m.CatalogSignature, "release-manifest.json": stable.Manifest, "release-manifest.sigstore.json": stable.ManifestSignature, "release-candidate.json": stable.Candidate, "upgrade-compatibility.json": stable.Compatibility, "build-provenance.json": stable.Provenance, "build-provenance.json.sigstore.json": stable.ProvenanceSignature, archiveName + ".sigstore.json": sign(archive, f.Metadata.SourceCommit, "v1.0.0")} {
				addAssetResponse(&status, responses, base, name, raw)
			}
			statePath := filepath.Join(installer.config.StateDir, "release-trust.json")
			switch mutation {
			case "installer state":
				snapshot, err := releasetrust.VerifySnapshot(f.Pinned, m, releasetrust.SnapshotFloor{})
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(statePath, testfixture.JSON(t, snapshot.Floor()), 0600); err != nil {
					t.Fatal(err)
				}
			case "legacy state":
				if err := updateVerificationState(statePath, f.Metadata, m.Metadata); err != nil {
					t.Fatal(err)
				}
			case "installer lock":
				lock := flock.New(releasetrust.VerificationLockPath(statePath))
				ok, err := lock.TryLock()
				if err != nil || !ok {
					t.Fatal("lock unavailable")
				}
				defer lock.Unlock()
			case "canonical old":
				f.Metadata.Sequence = 1
				f.Metadata.HighestRelease = nil
				f.Metadata.HighestReleaseSequence = nil
				f.Metadata.Releases = nil
				f.Metadata.SourceCommit = ""
				f.Metadata.ReleaseTag = ""
				f.Metadata.Event.SignedBy = f.Metadata.Recovery.ID
				f.Catalog.Sequence = 1
				baseline := f.Material(t)
				for name, raw := range map[string][]byte{"metadata.json": baseline.Metadata, "metadata.sigstore.json": baseline.MetadataSignature, "catalog.json": baseline.Catalog, "catalog.sigstore.json": baseline.CatalogSignature} {
					responses[trustURL(name)] = raw
				}
			case "missing policy":
				responses[trustURL("stable-policy.json")] = []byte("{}")
			case "wrong archive signature":
				responses[base+archiveName+".sigstore.json"] = sign(archive, f.Metadata.SourceCommit, "v1.0.1")
			case "rollback":
				newer := f.Metadata
				newer.Sequence = 3
				if err := updateVerificationState(statePath, newer, testfixture.JSON(t, newer)); err != nil {
					t.Fatal(err)
				}
			}
			err := installer.verifyStable(t.Context(), status, archiveName, archive)
			if mutation == "none" || mutation == "canonical old" || mutation == "installer state" || mutation == "legacy state" {
				if err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(statePath); err != nil {
					t.Fatal(err)
				}
			} else {
				if err == nil {
					t.Fatal("accepted invalid release")
				}
				if mutation != "rollback" {
					if _, err := os.Stat(statePath); !os.IsNotExist(err) {
						t.Fatalf("persisted failed evidence: %v", err)
					}
				}
			}
		})
	}
}
