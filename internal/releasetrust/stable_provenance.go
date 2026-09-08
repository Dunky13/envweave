package releasetrust

import (
	"bytes"
	"errors"
	"regexp"

	"github.com/Hikyo-Org/hikyo/internal/releaseidentity"
)

type StableProvenance struct {
	Type          string              `json:"_type"`
	Subject       []ProvenanceSubject `json:"subject"`
	PredicateType string              `json:"predicateType"`
	Predicate     struct {
		BuildDefinition struct {
			BuildType          string `json:"buildType"`
			ExternalParameters struct {
				Repository string `json:"repository"`
				Ref        string `json:"ref"`
			} `json:"externalParameters"`
			ResolvedDependencies []struct {
				URI    string            `json:"uri"`
				Digest map[string]string `json:"digest"`
			} `json:"resolvedDependencies"`
		} `json:"buildDefinition"`
		RunDetails struct {
			Builder struct {
				ID string `json:"id"`
			} `json:"builder"`
			Metadata struct {
				InvocationID string `json:"invocationId"`
			} `json:"metadata"`
		} `json:"runDetails"`
	} `json:"predicate"`
}

type ProvenanceSubject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

// VerifyStableProvenance verifies both the independent workflow signature and
// the complete artifact subject set. This is provenance from the build workflow,
// not a claim of an independently isolated SLSA builder.
func VerifyStableProvenance(snapshot Snapshot, release VerifiedRelease, signature, raw []byte) error {
	if !snapshot.StableKeylessEnabled() || !release.Valid() || release.state.signingKeyID != StableWorkflowSigner {
		return errors.New("stable provenance requires delegated workflow release")
	}
	if len(raw) == 0 || len(raw) > MaxDocumentBytes {
		return errors.New("missing or oversized stable provenance")
	}
	if err := VerifyStableArtifactSignature(snapshot, release, signature, raw); err != nil {
		return err
	}
	if err := release.VerifyArtifact("build-provenance.json", bytes.NewReader(raw)); err != nil {
		return err
	}
	var statement StableProvenance
	if err := decodeDocument(raw, &statement); err != nil {
		return err
	}
	policy := snapshot.state.workflowPolicy
	identity := release.Identity()
	ref := "refs/tags/v" + identity.Version
	definition := statement.Predicate.BuildDefinition
	workflow := policy.RepositoryURI + "/" + policy.WorkflowPath + "@" + ref
	if statement.Type != "https://in-toto.io/Statement/v1" || statement.PredicateType != "https://slsa.dev/provenance/v1" || definition.BuildType != "https://hikyo.dev/build/github-actions/v1" || definition.ExternalParameters.Repository != policy.RepositoryURI || definition.ExternalParameters.Ref != ref || statement.Predicate.RunDetails.Builder.ID != workflow {
		return errors.New("stable provenance build identity mismatch")
	}
	if len(definition.ResolvedDependencies) != 1 || definition.ResolvedDependencies[0].URI != "git+"+policy.RepositoryURI+"@"+ref || len(definition.ResolvedDependencies[0].Digest) != 1 || definition.ResolvedDependencies[0].Digest["gitCommit"] != identity.Commit {
		return errors.New("stable provenance source commit mismatch")
	}
	if !regexp.MustCompile("^" + regexp.QuoteMeta(policy.RepositoryURI) + `/actions/runs/[1-9][0-9]*/attempts/[1-9][0-9]*$`).MatchString(statement.Predicate.RunDetails.Metadata.InvocationID) {
		return errors.New("stable provenance invocation identity invalid")
	}
	expected := map[string]string{}
	for _, artifact := range release.Artifacts() {
		if artifact.Kind != "build-provenance" {
			expected[artifact.Name] = artifact.SHA256
		} else if artifact.Name != "build-provenance.json" {
			return errors.New("unexpected stable provenance artifact")
		}
	}
	if len(statement.Subject) != len(expected) {
		return errors.New("stable provenance subject inventory incomplete")
	}
	for _, subject := range statement.Subject {
		digest, exists := expected[subject.Name]
		if !exists || len(subject.Digest) != 1 || subject.Digest["sha256"] != digest || releaseidentity.Digest(digest).Validate() != nil {
			return errors.New("stable provenance subject substituted or repeated")
		}
		delete(expected, subject.Name)
	}
	return nil
}
