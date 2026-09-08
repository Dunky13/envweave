package cli_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/cli"
)

func TestParseVerbosityPreservesCommandArguments(t *testing.T) {
	for _, tc := range []struct {
		name       string
		args, want []string
		level      int
	}{
		{"leading", []string{"-vv", "upgrade"}, []string{"upgrade"}, 2},
		{"repeated", []string{"--verbose", "upgrade", "-v", "--verbose"}, []string{"upgrade"}, 3},
		{"saturates", []string{"-vvvv", "upgrade", "-vv"}, []string{"upgrade"}, 3},
		{"command level", []string{"upgrade", "--target", "v1.2.3", "-vvv"}, []string{"upgrade", "--target", "v1.2.3"}, 3},
		{"value looks verbose", []string{"values", "set", "--value", "-v"}, []string{"values", "set", "--value", "-v"}, 0},
		{"ambiguous option chain", []string{"values", "set", "--value", "--verbose", "-vv"}, []string{"values", "set", "--value", "--verbose", "-vv"}, 0},
		{"equals value", []string{"upgrade", "--target=-v", "-vv"}, []string{"upgrade", "--target=-v"}, 2},
		{"unknown short option", []string{"values", "-x", "-vv"}, []string{"values", "-x", "-vv"}, 0},
		{"mixed short group", []string{"upgrade", "-vx"}, []string{"upgrade", "-vx"}, 0},
		{"passthrough", []string{"run", "-v", "--", "command", "--verbose", "-vvv"}, []string{"run", "--", "command", "--verbose", "-vvv"}, 1},
		{"boundary after option", []string{"run", "--use-human-session", "--", "command", "-vv"}, []string{"run", "--use-human-session", "--", "command", "-vv"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := slices.Clone(tc.args)
			got, level := cli.ParseVerbosity(tc.args)
			if level != tc.level || !slices.Equal(got, tc.want) {
				t.Fatalf("arguments=%q level=%d, want %q level=%d", got, level, tc.want, tc.level)
			}
			if !slices.Equal(tc.args, original) {
				t.Fatal("mutated caller arguments")
			}
		})
	}
}

func TestVerboseRunKeepsJSONOnStdout(t *testing.T) {
	for _, level := range []string{"", "-v", "-vv", "-vvv"} {
		t.Run(level, func(t *testing.T) {
			ios, stdout, stderr := testIO(t, nil)
			args := []string{"context", "list", "-o", "json"}
			if level != "" {
				args = append([]string{level}, args...)
			}
			if code := cli.Run(t.Context(), ios, args); code != cli.ExitOK {
				t.Fatalf("exit=%d stderr=%s", code, stderr)
			}
			golden(t, "context-list-empty.json", stdout.Bytes())
			if level == "" && stderr.Len() != 0 {
				t.Fatalf("default diagnostics: %s", stderr)
			}
			if level != "" && !strings.Contains(stderr.String(), "starting context") {
				t.Fatalf("missing phase: %s", stderr)
			}
			if got := strings.Contains(stderr.String(), "elapsed="); got != (level == "-vvv") {
				t.Fatalf("timing visibility: %s", stderr)
			}
		})
	}
}

func TestVerboseHelpRemainsSideEffectFree(t *testing.T) {
	ios, stdout, stderr := testIO(t, nil)
	ios.Env = cli.Env{Getenv: func(string) string { t.Fatal("help accessed environment"); return "" }}
	if code := cli.Run(t.Context(), ios, []string{"-vvv", "update", "--help"}); code != cli.ExitOK {
		t.Fatalf("exit=%d", code)
	}
	if stderr.Len() != 0 || !strings.Contains(stdout.String(), "-vvv") {
		t.Fatalf("stdout=%s stderr=%s", stdout, stderr)
	}
}

func TestVerbositySurvivesProcessHandoff(t *testing.T) {
	args := []string{"backup", "upgrade-export", "--json", "--output", "-v"}
	for level := 0; level <= 3; level++ {
		forwarded := cli.WithVerbosityArguments(args, level)
		parsed, got := cli.ParseVerbosity(forwarded)
		if got != level || !slices.Equal(parsed, args) {
			t.Fatalf("handoff arguments=%q level=%d", parsed, got)
		}
	}
}
