package releasetrust_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/releasetrust"
	"github.com/Hikyo-Org/hikyo/internal/releasetrust/testfixture"
)

func TestNightlyPlatformSelectionVerifiesExactNativeInventory(t *testing.T) {
	platforms := []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64", "windows/arm64"}
	for _, platform := range platforms {
		for _, mutation := range []string{"none", "missing binary", "missing metadata", "tampered", "extra binary", "extra package", "invalid signature", "unsupported", "ambiguous"} {
			t.Run(platform+"/"+mutation, func(t *testing.T) {
				compatibility := []byte("signed compatibility fixture")
				payloads := map[string][]byte{"upgrade-compatibility.json": compatibility, "binary-provenance.json": []byte("{}"), "checksums.txt": []byte("checksums")}
				artifacts := []releasetrust.Artifact{{Name: "upgrade-compatibility.json", Kind: "upgrade-compatibility"}, {Name: "binary-provenance.json", Kind: "binary-provenance"}, {Name: "checksums.txt", Kind: "checksum"}}
				for _, target := range platforms {
					name := strings.ReplaceAll(target, "/", "-") + ".archive"
					payloads[name] = []byte(target)
					artifacts = append(artifacts, releasetrust.Artifact{Name: name, Kind: "binary", Platform: target})
				}
				payloads["linux.deb"] = []byte("package")
				artifacts = append(artifacts, releasetrust.Artifact{Name: "linux.deb", Kind: "package", Platform: "linux/amd64", Format: "deb", Arch: "amd64"})
				if mutation == "ambiguous" {
					payloads["duplicate.archive"] = []byte("duplicate")
					artifacts = append(artifacts, releasetrust.Artifact{Name: "duplicate.archive", Kind: "binary", Platform: platform})
				}
				fixture, material, _ := testfixture.NightlyWithPayloads(t, compatibility, false, payloads, artifacts)
				snapshot := fixture.Snapshot(t)
				release, err := releasetrust.VerifyNightlyManifest(snapshot, material)
				if err != nil {
					t.Fatal(err)
				}
				selected, err := release.NightlyArtifacts(platform)
				if mutation == "ambiguous" {
					if err == nil {
						t.Fatal("ambiguous native archive accepted")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				keep := map[string]bool{}
				for _, artifact := range selected {
					keep[artifact.Name] = true
				}
				for name := range material.Artifacts {
					if !keep[name] {
						delete(material.Artifacts, name)
					}
				}
				material.Platform = platform
				native := strings.ReplaceAll(platform, "/", "-") + ".archive"
				switch mutation {
				case "missing binary":
					delete(material.Artifacts, native)
				case "missing metadata":
					delete(material.Artifacts, "checksums.txt")
				case "tampered":
					material.Artifacts[native] = strings.NewReader("tampered")
				case "extra binary":
					material.Artifacts["foreign.archive"] = strings.NewReader("foreign")
				case "extra package":
					material.Artifacts["linux.deb"] = bytes.NewReader(payloads["linux.deb"])
				case "invalid signature":
					material.Bundle = []byte("{}")
				case "unsupported":
					material.Platform = "linux/386"
				}
				_, err = releasetrust.VerifyNightly(snapshot, material)
				if mutation == "none" && err != nil {
					t.Fatal(err)
				}
				if mutation != "none" && err == nil {
					t.Fatal("invalid platform evidence accepted")
				}
			})
		}
	}
}
