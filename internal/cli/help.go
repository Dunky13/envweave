package cli

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// IsHelpFlag reports whether one argument asks for help.
func IsHelpFlag(arg string) bool {
	return arg == "--help" || arg == "-help" || arg == "-h"
}

// HelpRequested reports whether args ask for help anywhere before the "--"
// separator. Arguments after "--" belong to the child command of `hikyo run`
// and are never interpreted.
func HelpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if IsHelpFlag(arg) {
			return true
		}
	}
	return false
}

// CommandPath is the leading literal words of an invocation: everything up to
// the first flag or "--". `values set KEY --stdin --help` yields
// [values set KEY]; Help trims unmatched trailing words itself.
func CommandPath(args []string) []string {
	var path []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			break
		}
		path = append(path, arg)
	}
	return path
}

// Help writes the slice of the client help that applies to path: every
// entry whose leading words match, grouped under its section, followed by
// that section's notes. Trailing words that match nothing are dropped
// one at a time so `values set KEY --help` shows `values set`. It returns
// false when not even the first word is a known command.
func Help(w io.Writer, path []string) bool {
	if !HelpFromText(w, usageText, path) {
		return false
	}
	fmt.Fprint(w, verbosityHelp)
	return true
}

// HelpFromText is Help over an arbitrary usage text laid out like usageText.
func HelpFromText(w io.Writer, text string, path []string) bool {
	sections := parseUsage(text)
	for len(path) > 0 {
		if out := renderHelp(sections, path); out != "" {
			fmt.Fprint(w, out)
			return true
		}
		path = path[:len(path)-1]
	}
	return false
}

// A section is one heading of the usage text with its entries and notes.
type helpSection struct {
	heading string
	entries []helpEntry
	// notes are the section's prose paragraphs, each kept as its lines.
	notes [][]string
}

// An entry is one `hikyo ...` synopsis with its continuation lines.
type helpEntry struct {
	// words are the literal command words before the first placeholder,
	// flag, or description column: [values set], [project list|show|create].
	words []string
	lines []string
}

func (e helpEntry) matches(path []string) bool {
	if len(path) > len(e.words) {
		return false
	}
	for i, word := range path {
		if !slices.Contains(strings.Split(e.words[i], "|"), word) {
			return false
		}
	}
	return true
}

func renderHelp(sections []helpSection, path []string) string {
	var b strings.Builder
	for _, section := range sections {
		var matched []helpEntry
		for _, entry := range section.entries {
			if entry.matches(path) {
				matched = append(matched, entry)
			}
		}
		if len(matched) == 0 {
			continue
		}
		b.WriteString(section.heading + "\n")
		for _, entry := range matched {
			for _, line := range entry.lines {
				b.WriteString(line + "\n")
			}
		}
		for _, note := range section.notes {
			b.WriteString("\n")
			for _, line := range note {
				b.WriteString(line + "\n")
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func parseUsage(text string) []helpSection {
	var (
		sections []helpSection
		current  *helpSection
		entry    *helpEntry
		note     []string
		flush    = func() {
			if current == nil {
				return
			}
			if entry != nil {
				current.entries = append(current.entries, *entry)
				entry = nil
			}
			if note != nil {
				current.notes = append(current.notes, note)
				note = nil
			}
		}
	)
	for _, line := range strings.Split(text, "\n") {
		switch {
		case line == "":
			flush()
		case !strings.HasPrefix(line, " "):
			flush()
			if current != nil {
				sections = append(sections, *current)
			}
			current = &helpSection{heading: line}
		case current == nil:
			// The title line's neighbours before the first heading.
		case isEntryLine(line):
			flush()
			entry = &helpEntry{words: entryWords(line), lines: []string{line}}
		case entry != nil && strings.HasPrefix(line, "    "):
			entry.lines = append(entry.lines, line)
		default:
			if entry != nil {
				current.entries = append(current.entries, *entry)
				entry = nil
			}
			note = append(note, line)
		}
	}
	flush()
	if current != nil {
		sections = append(sections, *current)
	}
	return sections
}

func isEntryLine(line string) bool {
	return strings.HasPrefix(line, "  hikyo ") || strings.HasPrefix(line, "  sudo hikyo ")
}

// entryWords extracts the literal command words of a synopsis line. The
// description column is separated by three or more spaces; placeholders,
// flags and the "--" separator end the command words.
func entryWords(line string) []string {
	synopsis := strings.TrimSpace(line)
	if i := strings.Index(synopsis, "   "); i >= 0 {
		synopsis = synopsis[:i]
	}
	fields := strings.Fields(synopsis)
	if len(fields) > 0 && fields[0] == "sudo" {
		fields = fields[1:]
	}
	var words []string
	for _, field := range fields[1:] { // fields[0] is "hikyo"
		if strings.HasPrefix(field, "-") || strings.ContainsAny(field, "<[(") {
			break
		}
		words = append(words, field)
	}
	return words
}
