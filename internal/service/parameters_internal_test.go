package service

import (
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/parameters"
	"github.com/Hikyo-Org/hikyo/internal/store"
)

func TestParameterContractVersionCompatibility(t *testing.T) {
	for _, raw := range []string{`{}`, `{"declarations":{"NAME":".*"}}`, `{"version":1,"future_metadata":true,"declarations":{"NAME":".*"}}`} {
		if _, err := decodeParameterContract(raw); err != nil {
			t.Errorf("compatible contract %s: %v", raw, err)
		}
	}
	for _, raw := range []string{`{"version":2}`, `{"version":-1}`, `{"version":"1"}`, `{} {}`, `null`, `[]`} {
		if _, err := decodeParameterContract(raw); err == nil {
			t.Errorf("accepted incompatible contract %s", raw)
		}
	}
}

func TestParameterValidationPreservesLiteralSchemas(t *testing.T) {
	key := store.CatalogueKey{Name: "VALUE", Classification: "config", Declaration: `{"rule":{"type":"enum","members":["${VALUE}"]}}`}
	if err := validateValueWithParameters(key, "${VALUE}", nil); err != nil {
		t.Fatal("legacy literal refused", err)
	}
	if err := validateValueWithParameters(key, "${OTHER}", nil); err == nil {
		t.Fatal("literal bypassed schema")
	}
	if err := validateValueWithParameters(key, "$${VALUE}", map[string]string{"NAME": ".*"}); err != nil {
		t.Fatal("escaped literal schema refused", err)
	}
	if err := validateValueWithParameters(key, "$${OTHER}", map[string]string{"NAME": ".*"}); err == nil {
		t.Fatal("escaped-only schema validation deferred")
	}
}

func TestFrozenParameterContractRetainsEscapeSemantics(t *testing.T) {
	for _, tc := range []struct {
		version int
		want    string
	}{{0, "$123"}, {1, "${NUMBER}"}} {
		got, err := resolveConfig(parameters.Contract{Version: tc.version, Schemas: map[string]string{"VALUE": `{"rule":{"type":"string"}}`}}, map[string]string{"NUMBER": "123"}, "VALUE", "config", "$${NUMBER}")
		if err != nil || got != tc.want {
			t.Fatalf("v%d got %q, %v; want %q", tc.version, got, err, tc.want)
		}
	}
}
