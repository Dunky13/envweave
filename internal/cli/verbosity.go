package cli

import "strings"

// ParseVerbosity removes repeatable -v/--verbose options and -vv/-vvv groups.
// Options may precede the command or follow command words. A non-verbosity
// option without '=' and its next argument remain untouched because the next
// argument may be a value, including a literal '-v'. Place verbosity before
// other options to avoid ambiguity after boolean flags. Nothing after '--' is
// interpreted, including the child command of hikyo run.
func ParseVerbosity(args []string) ([]string, int) {
	remaining := make([]string, 0, len(args))
	level := 0
	previousMayTakeValue := false
	for i, arg := range args {
		if arg == "--" {
			remaining = append(remaining, args[i:]...)
			break
		}
		count := 0
		if arg == "--verbose" {
			count = 1
		} else if strings.HasPrefix(arg, "-v") && strings.Trim(arg[1:], "v") == "" {
			count = len(arg) - 1
		}
		if count > 0 && !previousMayTakeValue {
			level = min(3, level+count)
			continue
		}
		remaining = append(remaining, arg)
		// Inspect every preserved token, including possible option values.
		// A boolean followed by a value-taking option must not hide that
		// second option and cause its '-v' value to be consumed.
		previousMayTakeValue = strings.HasPrefix(arg, "-") && arg != "-" && !strings.Contains(arg, "=")
	}
	return remaining, level
}

// WithVerbosityArguments carries the diagnostic level across process handoffs.
// Leading options cannot be confused with values owned by the child command.
func WithVerbosityArguments(args []string, level int) []string {
	if level <= 0 {
		return args
	}
	return append([]string{"-" + strings.Repeat("v", min(3, level))}, args...)
}

const verbosityHelp = `global diagnostics (stderr only):
  -v, --verbose                        phases and progress; repeat for more detail
  -vv                                  artifact and HTTP request details
  -vvv                                 timings as well (maximum level)
  put verbosity before other options; arguments after -- are untouched
`

// VerbosityHelp describes global diagnostics for host and client help.
func VerbosityHelp() string { return verbosityHelp }
