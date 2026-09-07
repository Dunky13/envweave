package crypto

import (
	"crypto/hkdf"
	"crypto/sha256"
	"fmt"
)

const custodyWrapLabel = "hikyo.dev/upgrade-custody/root-key-wrap/v1"

// CustodyWrapKey derives the secret that wraps the operator upgrade vault
// from the installation's root key. Everything the vault protects is already
// reachable by whoever can read that key, so a separate human-held secret
// added a prompt without adding a boundary. Domain-separated from every
// other root-key derivation by its label.
func CustodyWrapKey(root []byte) ([]byte, error) {
	if len(root) != KeySize {
		return nil, fmt.Errorf("crypto: custody wrap root is %d bytes, want %d", len(root), KeySize)
	}
	key, err := hkdf.Key(sha256.New, root, nil, custodyWrapLabel, KeySize)
	if err != nil {
		return nil, fmt.Errorf("crypto: derive custody wrap key: %w", err)
	}
	return key, nil
}
