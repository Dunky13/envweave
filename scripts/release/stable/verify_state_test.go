package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/releaseidentity"
	"github.com/Hikyo-Org/hikyo/internal/releasetrust"
)

func TestDraftVerificationResumesBeforeCanonicalTrustPromotion(t *testing.T) {
	trust, directory := fullDraft(t)
	state := filepath.Join(t.TempDir(), "release-trust.json")
	args := []string{"verify", "--trust", trust, "--directory", directory, "--version", "1.0.0", "--commit", strings.Repeat("a", 40), "--state", state, "--latest"}
	if err := run(args, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(state)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := loadTrust(trust)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := releasetrust.DecodeVerificationFloor(first, canonical.snapshot.Floor())
	if err != nil {
		t.Fatal(err)
	}
	if persisted.MetadataSequence <= canonical.snapshot.Floor().MetadataSequence {
		t.Fatal("fixture must leave canonical trust behind signed draft")
	}
	if err := run(args, &bytes.Buffer{}); err != nil {
		t.Fatalf("same authenticated draft cannot resume before sync-trust: %v", err)
	}
	after, err := os.ReadFile(state)
	if err != nil || !bytes.Equal(first, after) {
		t.Fatal("resume changed trust state")
	}
	// A valid but older embedded snapshot cannot use the stale canonical state
	// as permission to lower the installer floor.
	for _, name := range []string{"metadata.json", "metadata.sigstore.json", "catalog.json", "catalog.sigstore.json"} {
		raw, err := os.ReadFile(filepath.Join(trust, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := run(args, &bytes.Buffer{}); err == nil {
		t.Fatal("accepted older snapshot after verified draft")
	}
	after, err = os.ReadFile(state)
	if err != nil || releaseidentity.Hash(first) != releaseidentity.Hash(after) {
		t.Fatal("rollback attempt changed persisted state")
	}
}
