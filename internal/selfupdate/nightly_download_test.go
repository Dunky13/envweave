package selfupdate

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/releaseidentity"
	"github.com/Hikyo-Org/hikyo/internal/upgradebundle"
)

func TestNightlyDownloadsOnlyNativeArchiveAndMetadata(t *testing.T) {
	installer, status, _, _, _, _ := preparedNightlyFixture(t, nil)
	native := mustArchiveName(t, status.LatestVersion)
	transport := installer.client.Transport
	requests := map[string]int{}
	installer.client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.Contains(request.URL.Path, "/releases/download/") {
			name := filepath.Base(request.URL.Path)
			requests[name]++
			if (strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".deb")) && name != native {
				t.Errorf("downloaded irrelevant payload %s on %s", name, nightlyPlatform())
			}
		}
		return transport.RoundTrip(request)
	})
	prepared, err := installer.PrepareNightly(t.Context(), status)
	if err != nil {
		t.Fatal(err)
	}
	if requests[native] != 1 || len(requests) != 8 {
		t.Fatalf("expected native archive and seven metadata files: %v", requests)
	}
	entries, err := os.ReadDir(prepared.Directory)
	if err != nil || len(entries) != len(requests) {
		t.Fatalf("cache differs from selected download inventory: %v %v", entries, err)
	}
	for _, entry := range entries {
		if requests[entry.Name()] != 1 {
			t.Fatalf("unexpected cached artifact: %s", entry.Name())
		}
	}
	if _, err := installer.PrepareNightly(t.Context(), status); err != nil {
		t.Fatal(err)
	}
	if requests[native] != 1 {
		t.Fatal("handoff downloaded the native archive again")
	}
}

func TestPlatformBundleRejectsInventoryModeSubstitution(t *testing.T) {
	installer, status, _, trust, _, _ := preparedNightlyFixture(t, nil)
	prepared, err := installer.PrepareNightly(t.Context(), status)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(prepared.BundleDirectory, "index.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"missing platform", "unsupported platform", "other platform", "downgrade"} {
		t.Run(mutation, func(t *testing.T) {
			var index upgradebundle.Index
			if err := json.Unmarshal(raw, &index); err != nil {
				t.Fatal(err)
			}
			switch mutation {
			case "missing platform":
				index.NightlyPlatform = ""
			case "unsupported platform":
				index.NightlyPlatform = "linux/386"
			case "other platform":
				index.NightlyPlatform = "windows/amd64"
				if nightlyPlatform() == index.NightlyPlatform {
					index.NightlyPlatform = "linux/amd64"
				}
			case "downgrade":
				index.Format, index.NightlyPlatform = upgradebundle.IndexFormat, ""
			}
			changed, err := json.Marshal(index)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, changed, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := upgradebundle.Load(t.Context(), prepared.BundleDirectory, trust.Pinned, releaseidentity.SnapshotFloor{}); err == nil {
				t.Fatal("substituted platform inventory accepted")
			}
		})
	}
}

func TestNightlyRejectsSignatureBeforeDownloadingExecutable(t *testing.T) {
	installer, status, _, _, _, responses := preparedNightlyFixture(t, nil)
	for index := range status.Assets {
		asset := &status.Assets[index]
		if asset.Name == "release-manifest.sigstore.json" {
			raw := []byte("{}")
			responses[asset.URL] = raw
			asset.Size, asset.Digest = int64(len(raw)), "sha256:"+string(releaseidentity.Hash(raw))
		}
	}
	transport := installer.client.Transport
	installer.client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, ".tar.gz") || strings.HasSuffix(request.URL.Path, ".zip") || strings.HasSuffix(request.URL.Path, ".deb") {
			t.Error("downloaded executable before authenticating manifest")
		}
		return transport.RoundTrip(request)
	})
	if _, err := installer.PrepareNightly(t.Context(), status); err == nil {
		t.Fatal("invalid manifest signature accepted")
	}
}
