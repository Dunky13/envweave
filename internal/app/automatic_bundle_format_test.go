//go:build darwin || linux

package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/upgradebundle"
)

func TestAutomaticBundleFormatRejectsHistoricalCandidateBeforeApplication(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		accepted     bool
	}{
		{"platform-aware", upgradebundle.IndexFormat + "\\n" + upgradebundle.PlatformIndexFormat, true},
		{"historical", upgradebundle.IndexFormat, false},
		{"unknown-command", "unknown command", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "candidate")
			script := "#!/bin/sh\n[ \"$1\" = --upgrade-bundle-formats ] || exit 2\nprintf '" + tc.output + "\\n'\n"
			if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			if err := checkAutomaticBundleFormat(t.Context(), binary); (err == nil) != tc.accepted {
				t.Fatalf("accepted=%t, err=%v", tc.accepted, err)
			}
		})
	}
}
