package upgradebundle

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/Hikyo-Org/hikyo/internal/releasetrust"
)

// VerifyNightlyDirectory authenticates a complete flat GitHub release download.
// Only the manifest and its signature are excluded from the signed inventory.
// Policy and Sigstore roots are payloads too; the recovery-signed snapshot must
// independently authorize their exact policy digest.
func VerifyNightlyDirectory(ctx context.Context, directory string, snapshot releasetrust.Snapshot) (releasetrust.VerifiedRelease, error) {
	return VerifyNightlyPlatformDirectory(ctx, directory, snapshot, "")
}

// VerifyNightlyPlatformDirectory verifies the closed selection for one platform.
func VerifyNightlyPlatformDirectory(ctx context.Context, directory string, snapshot releasetrust.Snapshot, platform string) (releasetrust.VerifiedRelease, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return releasetrust.VerifiedRelease{}, err
	}
	defer root.Close()
	r := documentReader{ctx: ctx, root: root}
	release, _, err := r.flatNightly(snapshot, platform)
	return release, err
}

func (r *documentReader) flatNightly(snapshot releasetrust.Snapshot, platform string) (releasetrust.VerifiedRelease, map[string][]byte, error) {
	documents := map[string][]byte{}
	for _, name := range []string{"release-manifest.json", "release-manifest.sigstore.json", "nightly-policy.json", "sigstore-trusted-root.json", "upgrade-compatibility.json"} {
		raw, err := r.read(name)
		if err != nil {
			return releasetrust.VerifiedRelease{}, nil, err
		}
		documents[name] = raw
	}
	assets, closers, err := r.payloads(".")
	if err != nil {
		return releasetrust.VerifiedRelease{}, nil, err
	}
	defer func() {
		for _, closer := range closers {
			closer.Close()
		}
	}()
	delete(assets, "release-manifest.json")
	delete(assets, "release-manifest.sigstore.json")
	release, err := releasetrust.VerifyNightly(snapshot, releasetrust.NightlyMaterial{
		Manifest: documents["release-manifest.json"], Bundle: documents["release-manifest.sigstore.json"],
		Policy: documents["nightly-policy.json"], TrustedRoot: documents["sigstore-trusted-root.json"],
		Compatibility: documents["upgrade-compatibility.json"], Artifacts: assets, Platform: platform,
	})
	return release, documents, err
}

// CopyNightlyRelease writes beneath a private assembler staging directory.
// Payloads on the same filesystem share storage with the download cache. They
// must be treated as immutable; removing either directory keeps the other valid.
// Each payload is rehashed through the verified envelope. The caller
// authenticates the complete staged bundle before atomic public publication.
func CopyNightlyRelease(ctx context.Context, directory, releasesDirectory string, snapshot releasetrust.Snapshot) (releasetrust.VerifiedRelease, error) {
	return CopyNightlyPlatformRelease(ctx, directory, releasesDirectory, snapshot, "")
}

// CopyNightlyPlatformRelease retains only the verified platform selection.
func CopyNightlyPlatformRelease(ctx context.Context, directory, releasesDirectory string, snapshot releasetrust.Snapshot, platform string) (releasetrust.VerifiedRelease, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return releasetrust.VerifiedRelease{}, err
	}
	defer root.Close()
	r := documentReader{ctx: ctx, root: root}
	release, documents, err := r.flatNightly(snapshot, platform)
	if err != nil {
		return releasetrust.VerifiedRelease{}, err
	}
	destination := filepath.Join(releasesDirectory, string(release.Identity().ManifestSHA256))
	if err := os.MkdirAll(releasesDirectory, 0700); err != nil {
		return releasetrust.VerifiedRelease{}, err
	}
	if err := os.Mkdir(destination, 0700); err != nil {
		return releasetrust.VerifiedRelease{}, err
	}
	for source, target := range map[string]string{
		"release-manifest.json": "manifest.json", "release-manifest.sigstore.json": "manifest.sigstore.json",
		"nightly-policy.json": "policy.json", "sigstore-trusted-root.json": "trusted-root.json", "upgrade-compatibility.json": "upgrade-compatibility.json",
	} {
		file, err := os.OpenFile(filepath.Join(destination, target), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return releasetrust.VerifiedRelease{}, err
		}
		_, err = file.Write(documents[source])
		if err := errors.Join(err, file.Sync(), file.Close()); err != nil {
			return releasetrust.VerifiedRelease{}, err
		}
	}
	if err := os.Mkdir(filepath.Join(destination, "payloads"), 0700); err != nil {
		return releasetrust.VerifiedRelease{}, err
	}
	payloadRoot, err := os.OpenRoot(filepath.Join(destination, "payloads"))
	if err != nil {
		return releasetrust.VerifiedRelease{}, err
	}
	defer payloadRoot.Close()
	var copied int64
	artifacts, err := release.NightlyArtifacts(platform)
	if err != nil {
		return releasetrust.VerifiedRelease{}, err
	}
	for _, artifact := range artifacts {
		// Link rather than duplicate the complete multi-platform inventory for
		// every route assembly. Load reopens and authenticates the staged files,
		// including rejecting symlinks or bytes changed since verification.
		target := filepath.Join(destination, "payloads", artifact.Name)
		if err := os.Link(filepath.Join(directory, artifact.Name), target); err == nil {
			input := &payloadReader{ctx: ctx, root: payloadRoot, name: artifact.Name, aggregate: &copied}
			if err := errors.Join(release.VerifyArtifact(artifact.Name, input), input.Close()); err != nil {
				return releasetrust.VerifiedRelease{}, err
			}
			file, err := openDocument(payloadRoot, artifact.Name)
			if err != nil {
				return releasetrust.VerifiedRelease{}, err
			}
			if err := errors.Join(file.Sync(), file.Close()); err != nil {
				return releasetrust.VerifiedRelease{}, err
			}
			continue
		} else if !errors.Is(err, syscall.EXDEV) {
			return releasetrust.VerifiedRelease{}, err
		}
		// Standalone offline assembly may use inputs on another filesystem.
		input := &payloadReader{ctx: ctx, root: root, name: artifact.Name, aggregate: &copied}
		output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return releasetrust.VerifiedRelease{}, err
		}
		err = release.VerifyArtifact(artifact.Name, io.TeeReader(input, output))
		err = errors.Join(err, input.Close(), output.Sync(), output.Close())
		if err != nil {
			return releasetrust.VerifiedRelease{}, err
		}
	}
	return release, nil
}
