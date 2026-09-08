package selfupdate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hikyo-Org/hikyo/internal/filedurability"
	"github.com/Hikyo-Org/hikyo/internal/releaseidentity"
	"github.com/gofrs/flock"
)

// PruneNightlyCache removes verified nightly directories, assembled bundles,
// extracted executables and abandoned staging directories that no kept
// identity references. Every full nightly is several hundred MiB, and a route
// with intermediate hops multiplies that, so the cache is trimmed after each
// completed upgrade. The trust state and lock files are never touched: the
// state carries the rollback floor and the highest observed release.
func (i *Installer) PruneNightlyCache(keep ...releaseidentity.Identity) (err error) {
	if i == nil || i.config.StateDir == "" {
		return errors.New("selfupdate: installer is not configured")
	}
	if err := realNightlyDirectory(i.config.StateDir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	lock := flock.New(filepath.Join(i.config.StateDir, "nightly-trust.lock"))
	locked, err := lock.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("selfupdate: another nightly verification is running")
	}
	defer func() { err = errors.Join(err, lock.Unlock()) }()
	retained := map[string]bool{}
	for _, identity := range keep {
		if identity.Validate() != nil {
			return errors.New("selfupdate: retained nightly identity is invalid")
		}
		retained["nightly-"+string(identity.ManifestSHA256)] = true
		retained["executable-"+string(identity.ManifestSHA256)+"-"] = true
	}
	entries, err := os.ReadDir(i.config.StateDir)
	if err != nil {
		return err
	}
	removed := 0
	for _, entry := range entries {
		name := entry.Name()
		generated := strings.HasPrefix(name, "nightly-") || strings.HasPrefix(name, "bundle-") || strings.HasPrefix(name, "executable-") || strings.HasPrefix(name, ".nightly-")
		if !generated || name == "nightly-trust.json" || name == "nightly-trust.lock" {
			continue
		}
		if retained[name] || retainedExecutable(retained, name) {
			continue
		}
		if err := os.RemoveAll(filepath.Join(i.config.StateDir, name)); err != nil {
			return err
		}
		removed++
	}
	if removed == 0 {
		return nil
	}
	i.progress("  Removed %d cached nightly artifacts no longer needed.", removed)
	return filedurability.SyncDirectory(i.config.StateDir)
}

// retainedExecutable matches executable-<manifest>-<os>-<arch> against the
// retained executable-<manifest>- prefixes.
func retainedExecutable(retained map[string]bool, name string) bool {
	for prefix := range retained {
		if strings.HasPrefix(prefix, "executable-") && strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
