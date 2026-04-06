package crypto

import (
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// DeriveSubKey derives a sub-key from the master key using HKDF-SHA256.
func DeriveSubKey(masterKey []byte, purpose string, length int) ([]byte, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("%w: master key must be 32 bytes", ErrInvalidKeyLength)
	}

	hkdfReader := hkdf.New(sha256.New, masterKey, nil, []byte(purpose))
	subKey := make([]byte, length)
	if _, err := io.ReadFull(hkdfReader, subKey); err != nil {
		return nil, fmt.Errorf("crypto: HKDF expand failed: %w", err)
	}
	return subKey, nil
}
