package releasetrust_test

import (
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/releaseidentity"
	"github.com/Hikyo-Org/hikyo/internal/releasetrust"
	"github.com/Hikyo-Org/hikyo/internal/releasetrust/testfixture"
)

func TestVerificationFloorPreservesInstallerAndRuntimeState(t *testing.T) {
	current := releasetrust.SnapshotFloor{MetadataSequence: 2, MetadataSHA256: releaseidentity.Hash([]byte("metadata")), CatalogSequence: 3, CatalogSHA256: releaseidentity.Hash([]byte("catalog")), HighestReleaseSequence: 1}
	for _, raw := range [][]byte{testfixture.JSON(t, current), []byte(`{"trust_sequence":2,"metadata_sha256":"` + string(current.MetadataSHA256) + `","highest_release_sequence":1,"highest_release":"1.0.0"}`), []byte(`{"trust_sequence":2,"metadata_sha256":"` + string(current.MetadataSHA256) + `","highest_release_sequence":1,"highest_release":"1.0.0","catalog_sequence":3,"catalog_sha256":"` + string(current.CatalogSHA256) + `"}`)} {
		got, err := releasetrust.DecodeVerificationFloor(raw, current)
		if err != nil {
			t.Fatal(err)
		}
		if got != current {
			t.Fatalf("changed trust floor: %#v", got)
		}
	}
	// An old canonical snapshot must not lower the persisted metadata counter.
	old := current
	old.MetadataSequence = 1
	raw := []byte(`{"trust_sequence":4,"metadata_sha256":"` + strings.Repeat("a", 64) + `","highest_release_sequence":2,"highest_release":"1.1.0"}`)
	got, err := releasetrust.DecodeVerificationFloor(raw, old)
	if err != nil {
		t.Fatal(err)
	}
	if got.MetadataSequence != 4 || got.HighestReleaseSequence != 2 {
		t.Fatal("legacy migration lowered state")
	}
	if err := got.Advance(current); err == nil {
		t.Fatal("accepted rollback after migration")
	}
	if releasetrust.VerificationLockPath("/tmp/release-trust.json") != "/tmp/release-trust.lock" {
		t.Fatal("installer and runtime lock paths differ")
	}
}
