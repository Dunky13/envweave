package parameters

import (
	"fmt"
	"strings"
	"testing"
)

func TestEscapedReferencesRemainLiteral(t *testing.T) {
	for _, tc := range []struct{ value, want string }{
		{"$${UNKNOWN}", "${UNKNOWN}"},
		{"$${", "${"},
		{"$${not a name}", "${not a name}"},
		{"$${NAME}/${NAME}", "${NAME}/123"},
		{"$$${NAME}", "$${NAME}"},
	} {
		t.Run(tc.value, func(t *testing.T) {
			declarations := map[string]string{"NAME": "[0-9]+"}
			if err := CheckReferences(tc.value, declarations); err != nil {
				t.Fatal(err)
			}
			got, err := Resolve(tc.value, map[string]string{"NAME": "123"}, 256)
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
	if _, err := Resolve("$${NAME}", nil, 6); err == nil {
		t.Fatal("escape bypassed output limit")
	}
	if err := CheckReferences("${unfinished", nil); err != nil {
		t.Fatal("unparameterized literal rejected", err)
	}
	if err := CheckReferences("${unfinished", map[string]string{"NAME": ".*"}); err == nil {
		t.Fatal("active malformed template accepted")
	}
}

func TestPatternCacheRemainsBoundedAndConcurrent(t *testing.T) {
	for n := 0; n < maxCachedPatterns+20; n++ {
		pattern := fmt.Sprintf("x{%d}", n)
		if err := Validate(map[string]string{"VALUE": pattern}, map[string]string{"VALUE": strings.Repeat("x", n)}); err != nil {
			t.Fatal(err)
		}
	}
	patternCache.Lock()
	size := len(patternCache.entries)
	patternCache.Unlock()
	if size > maxCachedPatterns {
		t.Fatalf("unbounded cache: %d", size)
	}
	for n := 0; n < 8; n++ {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 20; i++ {
				if err := Validate(map[string]string{"VALUE": "[0-9]+"}, map[string]string{"VALUE": "123"}); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
