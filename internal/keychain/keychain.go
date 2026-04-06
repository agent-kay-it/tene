package keychain

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	ServiceName = "tene"
	AccountName = "master-key"
)

// KeyStore is the interface for securely storing and loading the Master Key.
type KeyStore interface {
	// Store saves the Master Key.
	Store(key []byte) error

	// Load retrieves the stored Master Key.
	// Returns ErrKeyNotFound if no key is stored.
	Load() ([]byte, error)

	// Delete removes the stored Master Key.
	Delete() error

	// Exists checks if a Master Key is stored.
	Exists() bool
}

// KeyringStore uses the OS keychain via go-keyring.
type KeyringStore struct {
	service string
}

// NewKeyringStore creates a new OS keychain-based KeyStore.
func NewKeyringStore(service string) *KeyringStore {
	return &KeyringStore{service: service}
}

func (k *KeyringStore) Store(key []byte) error {
	encoded := base64.StdEncoding.EncodeToString(key)
	return keyring.Set(k.service, AccountName, encoded)
}

func (k *KeyringStore) Load() ([]byte, error) {
	encoded, err := keyring.Get(k.service, AccountName)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("keychain: failed to load key: %w", err)
	}
	return base64.StdEncoding.DecodeString(encoded)
}

func (k *KeyringStore) Delete() error {
	err := keyring.Delete(k.service, AccountName)
	if err == keyring.ErrNotFound {
		return nil
	}
	return err
}

func (k *KeyringStore) Exists() bool {
	_, err := keyring.Get(k.service, AccountName)
	return err == nil
}

// hashPath returns a short hash of the project path for unique service names.
func hashPath(path string) string {
	h := sha256.Sum256([]byte(path))
	return hex.EncodeToString(h[:8])
}

// NewStore returns the appropriate KeyStore based on the environment.
func NewStore(projectPath string) KeyStore {
	if os.Getenv("TENE_KEYCHAIN_FALLBACK") == "file" {
		home, _ := os.UserHomeDir()
		return NewFileStore(filepath.Join(home, ".tene", "keyfile"))
	}

	service := ServiceName + "-" + hashPath(projectPath)
	ks := NewKeyringStore(service)

	// Test keychain availability
	testKey := "keychain-test"
	if err := keyring.Set(service, testKey, "test"); err != nil {
		// Keychain unavailable -> file fallback
		home, _ := os.UserHomeDir()
		return NewFileStore(filepath.Join(home, ".tene", "keyfile"))
	}
	_ = keyring.Delete(service, testKey)

	return ks
}
