package ci_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestStableSigningAndPublicationStaySeparate(t *testing.T) {
	type step struct {
		Run string `yaml:"run"`
	}
	type job struct {
		If          string            `yaml:"if"`
		Permissions map[string]string `yaml:"permissions"`
		Steps       []step            `yaml:"steps"`
	}
	type releaseWorkflow struct {
		On          map[string]yaml.Node `yaml:"on"`
		Permissions map[string]string    `yaml:"permissions"`
		Jobs        map[string]job       `yaml:"jobs"`
	}
	read := func(name string) releaseWorkflow {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(repositoryRoot(t), ".github", "workflows", name))
		if err != nil {
			t.Fatal(err)
		}
		var workflow releaseWorkflow
		if err := yaml.Unmarshal(raw, &workflow); err != nil {
			t.Fatal(err)
		}
		if len(workflow.Permissions) != 1 || workflow.Permissions["contents"] != "read" {
			t.Fatal("workflow default must have only contents read permission")
		}
		return workflow
	}
	build := read("release.yml")
	if len(build.On) != 1 {
		t.Fatal("stable signing must only be tag push triggered")
	}
	if _, ok := build.On["push"]; !ok {
		t.Fatal("stable signing push trigger is missing")
	}
	signer := build.Jobs["build-signed-draft"]
	if signer.Permissions["id-token"] != "write" {
		t.Fatal("tag build requires OIDC")
	}
	var buildRun strings.Builder
	for _, step := range signer.Steps {
		buildRun.WriteString(step.Run)
		buildRun.WriteByte('\n')
	}
	for _, required := range []string{"require-signed-tag.sh", "stable preflight", "create-build-provenance.sh", "sign-stable-draft.sh", "--draft", "stable verify"} {
		if !strings.Contains(buildRun.String(), required) {
			t.Fatalf("stable draft gate missing: %s", required)
		}
	}
	for _, forbidden := range []string{"--draft=false", "--key ", "COSIGN_PASSWORD", "ceremony.sh"} {
		if strings.Contains(buildRun.String(), forbidden) {
			t.Fatalf("tag build must not publish or use private keys: %s", forbidden)
		}
	}
	if _, err := os.Stat(filepath.Join(repositoryRoot(t), ".github", "workflows", "release-publish.yml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("publication requires the maintainer admin login, not an unsupported GITHUB_TOKEN workflow")
	}
	ceremony, err := os.ReadFile(filepath.Join(repositoryRoot(t), "scripts", "release", "stable-ceremony.sh"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(ceremony)
	start, end := strings.Index(text, "\n\tpublish)"), strings.Index(text, "\n\tsync-trust)")
	if start < 0 || end <= start {
		t.Fatal("local publication ceremony is missing")
	}
	publish := text[start:end]
	for _, required := range []string{"clean_main", "download_verify", "confirm ", "publish-stable-draft.sh", "HIKYO_RELEASE_TRUST_STATE"} {
		if !strings.Contains(publish, required) {
			t.Fatalf("local publication guard missing: %s", required)
		}
	}
	for _, forbidden := range []string{"workflow run", "sign-stable-draft.sh", "signer", "sign-blob"} {
		if strings.Contains(publish, forbidden) {
			t.Fatalf("local publication must not sign or delegate unsupported permissions: %s", forbidden)
		}
	}
}
