package cli

import "testing"

func TestParameterFlagRefusesDuplicateAndMalformedInputs(t *testing.T) {
	values := map[string]string{}
	parse := parameterFlag(values)
	if err := parse("PR_NUMBER=123"); err != nil {
		t.Fatal(err)
	}
	if err := parse("PR_NUMBER=124"); err == nil {
		t.Fatal("duplicate silently replaced first parameter")
	}
	if values["PR_NUMBER"] != "123" {
		t.Fatal("duplicate mutated accepted value")
	}
	if err := parse("OTHER"); err == nil {
		t.Fatal("missing equals accepted")
	}
	if err := parse("bad=value"); err == nil {
		t.Fatal("invalid name accepted")
	}
}
