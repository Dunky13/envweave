//go:build darwin || linux

package upgradeassembly

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/buildcompat"
	"github.com/Hikyo-Org/hikyo/internal/releaseidentity"
	"github.com/Hikyo-Org/hikyo/internal/releasetrust"
	"github.com/Hikyo-Org/hikyo/internal/releasetrust/testfixture"
	"github.com/Hikyo-Org/hikyo/internal/upgradebundle"
)

func TestStableWorkflowOfflineBundleRoundTrip(t *testing.T) {
	_, declaration, err := buildcompat.Development()
	if err != nil {
		t.Fatal(err)
	}
	declaration.Profile = releaseidentity.StableV1
	declaration.Version = "1.0.0"
	declaration.Sequence = 1
	declaration.Commit = strings.Repeat("a", 40)
	f, m, release, _ := testfixture.StableWorkflowWithCompatibility(t, testfixture.JSON(t, declaration))
	base := t.TempDir()
	snapshotDir := filepath.Join(base, "snapshot")
	keysDir := filepath.Join(base, "keys")
	releaseDir := filepath.Join(base, "release")
	write := func(dir string, files map[string][]byte) {
		t.Helper()
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		for name, raw := range files {
			if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	write(snapshotDir, map[string][]byte{"metadata.json": m.Metadata, "metadata.sigstore.json": m.MetadataSignature, "catalog.json": m.Catalog, "catalog.sigstore.json": m.CatalogSignature, "stable-policy.json": m.StablePolicy, "stable-policy.sigstore.json": m.StablePolicySignature, "stable-trusted-root.json": m.StableTrustedRoot})
	write(keysDir, map[string][]byte{"primary.pub": f.PrimaryPublic})
	write(releaseDir, map[string][]byte{"release-manifest.json": release.Manifest, "release-manifest.sigstore.json": release.ManifestSignature, "release-candidate.json": release.Candidate, "upgrade-compatibility.json": release.Compatibility, "build-provenance.json": release.Provenance, "build-provenance.json.sigstore.json": release.ProvenanceSignature})
	output := filepath.Join(base, "output")
	if err := Assemble(t.Context(), Options{Pinned: f.Pinned, SnapshotDirectory: snapshotDir, KeysDirectory: keysDir, OutputDirectory: output, Releases: []string{releaseDir}}); err != nil {
		t.Fatal(err)
	}
	loaded, err := upgradebundle.Load(t.Context(), output, f.Pinned, releasetrust.SnapshotFloor{})
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Snapshot().StableKeylessEnabled() {
		t.Fatal("lost stable delegation during assembly")
	}
	if _, err := loaded.MatchBuild(release.Compatibility); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(output, "stable-policy.sigstore.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := upgradebundle.Load(t.Context(), output, f.Pinned, releasetrust.SnapshotFloor{}); err == nil {
		t.Fatal("accepted unsigned offline delegation")
	}
}
