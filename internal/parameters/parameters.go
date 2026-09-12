// Package parameters implements bounded, single-pass config substitution.
// Parameters are public configuration, never secret inputs or expressions.
package parameters

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	// MaxContractBytes bounds immutable schema metadata per publication.
	MaxContractBytes = 256 << 10
	MaxCount         = 32
	MaxValueBytes    = 256
	MaxPatternBytes  = 512
)

var namePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,63}$`)

// Contract is immutable snapshot metadata. Schemas records the declarations of
// templated config keys, so historical delivery never consults the live schema.
type Contract struct {
	Declarations map[string]string `json:"declarations,omitempty"`
	Schemas      map[string]string `json:"schemas,omitempty"`
}

func CheckDeclaration(name, pattern string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("parameter name %q must match %s", name, namePattern.String())
	}
	if len(pattern) == 0 || len(pattern) > MaxPatternBytes {
		return fmt.Errorf("parameter %s pattern must contain 1 to %d bytes", name, MaxPatternBytes)
	}
	if strings.ContainsRune(pattern, 0) {
		return fmt.Errorf("parameter %s pattern must not contain a NUL character", name)
	}
	if _, err := regexp.Compile("\\A(?:" + pattern + ")\\z"); err != nil {
		return fmt.Errorf("parameter %s has an invalid pattern: %w", name, err)
	}
	return nil
}

func CheckSupplied(values map[string]string) error {
	if len(values) > MaxCount {
		return fmt.Errorf("at most %d parameters are allowed", MaxCount)
	}
	for name, value := range values {
		if !namePattern.MatchString(name) {
			return fmt.Errorf("invalid parameter name %q", name)
		}
		if len(value) > MaxValueBytes || !utf8.ValidString(value) || strings.IndexFunc(value, unicode.IsControl) >= 0 {
			return fmt.Errorf("parameter %s must be valid text without control characters, at most %d bytes", name, MaxValueBytes)
		}
	}
	return nil
}

func Validate(declarations, supplied map[string]string) error {
	if err := CheckSupplied(supplied); err != nil {
		return err
	}
	for _, name := range names(supplied) {
		if _, ok := declarations[name]; !ok {
			return fmt.Errorf("undeclared parameter %s", name)
		}
	}
	for _, name := range names(declarations) {
		pattern := declarations[name]
		if err := CheckDeclaration(name, pattern); err != nil {
			return err
		}
		value, ok := supplied[name]
		if !ok {
			return fmt.Errorf("required parameter %s is missing", name)
		}
		re, err := regexp.Compile("\\A(?:" + pattern + ")\\z")
		if err != nil {
			return err
		}
		if !re.MatchString(value) {
			return fmt.Errorf("parameter %s does not match its validation pattern", name)
		}
	}
	return nil
}

// References accepts only ${NAME}; malformed expressions fail by key at callers.
func References(value string) ([]string, error) {
	var refs []string
	for {
		_, rest, found := strings.Cut(value, "${")
		if !found {
			return refs, nil
		}
		name, remaining, closed := strings.Cut(rest, "}")
		if !closed || !namePattern.MatchString(name) {
			return nil, fmt.Errorf("invalid parameter reference; expected ${NAME}")
		}
		refs = append(refs, name)
		value = remaining
	}
}

func CheckReferences(value string, declarations map[string]string) error {
	refs, err := References(value)
	if err != nil {
		return err
	}
	for _, name := range refs {
		if _, ok := declarations[name]; !ok {
			return fmt.Errorf("undeclared parameter %s", name)
		}
	}
	return nil
}

// Resolve substitutes once. Caller inputs are never scanned as new expressions.
func Resolve(value string, supplied map[string]string, maxBytes int) (string, error) {
	if _, err := References(value); err != nil {
		return "", err
	}
	var out strings.Builder
	for {
		prefix, rest, found := strings.Cut(value, "${")
		if !found {
			out.WriteString(value)
			break
		}
		out.WriteString(prefix)
		name, remaining, _ := strings.Cut(rest, "}")
		replacement, ok := supplied[name]
		if !ok {
			return "", fmt.Errorf("required parameter %s is missing", name)
		}
		if out.Len()+len(replacement) > maxBytes {
			return "", fmt.Errorf("resolved config exceeds %d bytes", maxBytes)
		}
		out.WriteString(replacement)
		value = remaining
	}
	if out.Len() > maxBytes {
		return "", fmt.Errorf("resolved config exceeds %d bytes", maxBytes)
	}
	return out.String(), nil
}

func Encode(values map[string]string) string {
	if len(values) == 0 {
		return "{}"
	}
	raw, _ := json.Marshal(values)
	return string(raw)
}

func names(values map[string]string) []string {
	out := make([]string, 0, len(values))
	for name := range values {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
