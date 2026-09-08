package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Hikyo-Org/hikyo/internal/releaseidentity"
	"github.com/Hikyo-Org/hikyo/internal/releasetrust"
)

func verifyPublishedOCI(snapshot releasetrust.Snapshot, release releasetrust.VerifiedRelease, trustedRoot []byte) error {
	if !snapshot.Valid() || !release.Valid() || release.SnapshotDigest() != snapshot.Digest() {
		return errors.New("published verification requires matching authenticated release and snapshot")
	}
	key, keyErr := snapshot.StableSigningPublicKey(release)
	policy, ok := snapshot.StablePolicy()
	if keyErr != nil {
		if !ok {
			return errors.New("published verification requires authenticated stable policy")
		}
		if releaseidentity.Hash(trustedRoot) != policy.TrustedRootSHA256 {
			return errors.New("published verification Sigstore root differs from authenticated policy")
		}
	} else {
		trustedRoot = key
	}
	f, err := os.CreateTemp("", "hikyo-stable-sigstore-root-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(trustedRoot)
	if err := errors.Join(err, f.Close()); err != nil {
		return err
	}
	cosign := os.Getenv("COSIGN_BIN")
	if cosign == "" {
		cosign = "cosign"
	}
	id := release.Identity()
	ref := "refs/tags/v" + id.Version
	for _, artifact := range release.Artifacts() {
		var subject string
		switch artifact.Kind {
		case "image":
			subject = artifact.Image + "@" + artifact.Digest
		case "chart-digest":
			subject = artifact.Chart + "@" + artifact.Digest
		default:
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		args := []string{"verify", "--trusted-root", f.Name(),
			"--certificate-identity", policy.RepositoryURI + "/" + policy.WorkflowPath + "@" + ref,
			"--certificate-oidc-issuer", policy.Issuer,
			"--certificate-github-workflow-sha", id.Commit,
			"--certificate-github-workflow-repository", strings.TrimPrefix(policy.RepositoryURI, "https://github.com/"),
			"--certificate-github-workflow-ref", ref, subject}
		if keyErr == nil {
			args = []string{"verify", "--insecure-ignore-tlog", "--key", f.Name(), subject}
		}
		cmd := exec.CommandContext(ctx, cosign, args...)
		// Cosign's output is public verification evidence; never log registry credentials.
		err := cmd.Run()
		cancel()
		if err != nil {
			return fmt.Errorf("published OCI signature for %s: %w", subject, err)
		}
	}
	return nil
}
