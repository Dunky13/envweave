//go:build !windows

package filedurability

import (
	"errors"
	"os"
)

// SyncDirectory persists directory entries and reports sync or close failures.
func SyncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	return errors.Join(directory.Sync(), directory.Close())
}
