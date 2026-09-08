package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/releaseidentity"
	"github.com/Hikyo-Org/hikyo/internal/releasetrust"
)

func TestNativePackageNamesMatchReleasePackaging(t *testing.T) {
	for _, version := range []string{"1.0.0", "1.2.3-rc.1", "1.2.3-alpha-test.2"} {
		for _, format := range []string{"deb", "rpm", "apk", "archlinux"} {
			for _, arch := range []string{"amd64", "arm64"} {
				got, err := nativePackageName(version, format, arch)
				if err != nil {
					t.Fatal(err)
				}
				expected, err := exec.Command("sh", "-c", `. "$1"; package_file_name "$2" "$3" "$4"`, "package-name-fixture", "../../lib/release.sh", version, format, arch).Output()
				if err != nil {
					t.Fatal(err)
				}
				if got != strings.TrimSpace(string(expected)) {
					t.Fatalf("package identity drift for %s %s %s: %s != %s", version, format, arch, got, expected)
				}
			}
		}
	}
	for _, args := range [][3]string{{"1.0.0+build", "deb", "amd64"}, {"1.0.0", "deb", "386"}, {"1.0.0", "exe", "amd64"}} {
		if _, err := nativePackageName(args[0], args[1], args[2]); err == nil {
			t.Fatal("accepted unsupported native package identity")
		}
	}
}

func TestBinaryProvenanceAndChecksumSemantics(t *testing.T) {
	_, directory := fullDraft(t)
	id := releaseidentity.Identity{Version: "1.0.0", Commit: strings.Repeat("a", 40)}
	provenance, err := os.ReadFile(filepath.Join(directory, "binary-provenance.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest releasetrust.Manifest
	fixtureReadJSON(t, filepath.Join(directory, "release-manifest.json"), &manifest)
	checksums, err := os.ReadFile(filepath.Join(directory, "checksums.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyBinaryProvenance(directory, id); err != nil {
		t.Fatal(err)
	}
	if err := verifyChecksums(directory, manifest.Artifacts); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []struct{ name, old, new string }{
		{"source commit", strings.Repeat("a", 40), strings.Repeat("b", 40)},
		{"archive OCI divergence", `"path":"image-root/amd64/hikyo","sha256":"` + strings.Repeat("e", 64), `"path":"image-root/amd64/hikyo","sha256":"` + strings.Repeat("f", 64)},
		{"duplicate architecture", `"goarch":"arm64"`, `"goarch":"amd64"`},
		{"wrong producer", `"name":"goreleaser"`, `"name":"other"`},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			modified := strings.Replace(string(provenance), mutation.old, mutation.new, 1)
			if modified == string(provenance) {
				t.Fatal("mutation did not change fixture")
			}
			fixtureWrite(t, directory, "binary-provenance.json", []byte(modified))
			if err := verifyBinaryProvenance(directory, id); err == nil {
				t.Fatal("accepted false binary provenance")
			}
		})
	}
	fixtureWrite(t, directory, "binary-provenance.json", provenance)
	lines := strings.Split(strings.TrimSuffix(string(checksums), "\n"), "\n")
	for _, mutation := range []struct{ name, raw string }{
		{"missing archive", strings.Join(lines[:len(lines)-1], "\n") + "\n"},
		{"duplicate archive", strings.Join(append(append([]string{}, lines[:len(lines)-1]...), lines[0]), "\n") + "\n"},
		{"wrong archive hash", strings.Repeat("0", 64) + string(checksums)[64:]},
		{"unrelated file", string(checksums) + strings.Repeat("a", 64) + "  install.sh\n"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			fixtureWrite(t, directory, "checksums.txt", []byte(mutation.raw))
			if err := verifyChecksums(directory, manifest.Artifacts); err == nil {
				t.Fatal("accepted false archive checksums")
			}
		})
	}
}
