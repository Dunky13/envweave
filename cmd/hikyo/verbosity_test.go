package main

import (
	"os"
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/upgradebundle"
)

func TestGlobalVerbosityKeepsVersionAndHelpContracts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		args   []string
		code   int
		stdout string
		stderr string
	}{
		{"machine version", []string{"-vvv", "--version"}, 0, version + "\n", ""},
		{"bundle formats", []string{"--upgrade-bundle-formats"}, 0, upgradebundle.IndexFormat + "\n" + upgradebundle.PlatformIndexFormat + "\n", ""},
		{"overview", []string{"-v", "--verbose", "--help"}, 0, "global diagnostics (stderr only):", ""},
		{"missing command", []string{"-vvv"}, 2, "", ""},
		{"server help", []string{"-vvv", "server", "--help"}, 0, "", "Usage of server:"},
		{"migrate help", []string{"-vvv", "migrate", "--help"}, 0, "", "Usage of migrate:"},
		{"backup help", []string{"-vvv", "backup", "export", "--help"}, 0, "backup", ""},
		{"import help", []string{"-vvv", "import", "--help"}, 0, "import", ""},
		{"upgrade help", []string{"upgrade", "-vvv", "--help"}, 0, "upgrade", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer stdout.Close()
			stderr, err := os.CreateTemp(t.TempDir(), "stderr")
			if err != nil {
				t.Fatal(err)
			}
			defer stderr.Close()
			oldArgs, oldOut, oldErr := os.Args, os.Stdout, os.Stderr
			t.Cleanup(func() { os.Args, os.Stdout, os.Stderr = oldArgs, oldOut, oldErr })
			os.Args, os.Stdout, os.Stderr = append([]string{"hikyo"}, tc.args...), stdout, stderr
			if code := run(); code != tc.code {
				t.Fatalf("exit=%d want=%d", code, tc.code)
			}
			out, err := os.ReadFile(stdout.Name())
			if err != nil {
				t.Fatal(err)
			}
			errors, err := os.ReadFile(stderr.Name())
			if err != nil {
				t.Fatal(err)
			}
			if tc.name == "machine version" && string(out) != tc.stdout {
				t.Fatalf("machine output=%q", out)
			}
			if !strings.Contains(string(out), tc.stdout) {
				t.Fatalf("stdout=%q", out)
			}
			if tc.code == 0 {
				if tc.stderr == "" && len(errors) != 0 {
					t.Fatalf("unexpected diagnostic output=%q", errors)
				}
				if !strings.Contains(string(errors), tc.stderr) {
					t.Fatalf("missing help on stderr=%q", errors)
				}
				if strings.Contains(string(errors), "hikyo:") || strings.Contains(string(errors), "elapsed=") {
					t.Fatalf("help emitted diagnostics=%q", errors)
				}
			}
			if tc.code != 0 && !strings.Contains(string(errors), "global diagnostics") {
				t.Fatalf("missing usage=%q", errors)
			}
		})
	}
}
