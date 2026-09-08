package selfupdate

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/releaseidentity"
)

func TestPruneNightlyCacheKeepsRetainedIdentityAndTrustState(t *testing.T) {
	state := t.TempDir()
	keep := releaseidentity.Identity{Profile: releaseidentity.NightlyV1, Version: "1.1.0-nightly.3", Sequence: 3, Commit: strings.Repeat("b", 40), CompatibilitySHA256: releaseidentity.Hash([]byte("claim")), ManifestSHA256: releaseidentity.Hash([]byte("keep"))}
	stale := releaseidentity.Hash([]byte("stale"))
	for _, dir := range []string{"nightly-" + string(keep.ManifestSHA256), "nightly-" + string(stale), "bundle-route-snapshot", ".nightly-download-abandoned", ".nightly-bundle-inputs-abandoned"} {
		if err := os.MkdirAll(filepath.Join(state, dir, "inner"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range []string{"executable-" + string(keep.ManifestSHA256) + "-linux-amd64", "executable-" + string(stale) + "-linux-amd64", "nightly-trust.json", "nightly-trust.lock", "unrelated"} {
		if err := os.WriteFile(filepath.Join(state, file), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var progress bytes.Buffer
	installer := newInstaller(nil, nil)
	installer.config = Config{StateDir: state, Progress: &progress}
	if err := installer.PruneNightlyCache(keep); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(state)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	want := []string{"executable-" + string(keep.ManifestSHA256) + "-linux-amd64", "nightly-" + string(keep.ManifestSHA256), "nightly-trust.json", "nightly-trust.lock", "unrelated"}
	slices.Sort(want)
	if !slices.Equal(names, want) {
		t.Fatalf("after prune: %v, want %v", names, want)
	}
	if !strings.Contains(progress.String(), "Removed 5 cached nightly artifacts") {
		t.Fatalf("progress: %q", progress.String())
	}
	if err := installer.PruneNightlyCache(releaseidentity.Identity{}); err == nil {
		t.Fatal("accepted an invalid retained identity")
	}
}
