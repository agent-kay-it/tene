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

// HashPath returns a short hex hash (first 12 chars) of the given path.
// Exported for callers (CLI sentinel paths) that need a stable per-project
// hash without re-implementing the convention. The 12-char length is wider
// than hashPath's 16-hex (=8 bytes) used by ServiceName, but derives from
// the same sha256 input so collisions across the two consumers do not
// matter — they live in different namespaces (OS keychain vs filesystem
// sentinel).
func HashPath(path string) string {
	h := sha256.Sum256([]byte(path))
	return hex.EncodeToString(h[:])[:12]
}

// FallbackInfo describes why the active KeyStore is (or is not) the file
// fallback. Used by the CLI to render a one-time notice; the keychain
// package itself never prints anything.
type FallbackInfo struct {
	// Used is true iff NewStoreWithStatus returned a FileStore instead of
	// a KeyringStore. False for the happy-path macOS/Linux/Windows native
	// keychain case.
	Used bool

	// Reason is a short machine-readable tag explaining why fallback
	// activated. Possible values:
	//   - ""                        : keychain available, no fallback
	//   - "env_override"            : TENE_KEYCHAIN_FALLBACK=file forced it
	//   - "keychain_unavailable"    : the test Set() call to the OS
	//                                 keychain failed (CI, Docker,
	//                                 headless box, libsecret missing)
	Reason string

	// Path is the absolute filesystem path of the FileStore key file
	// (empty string when Used == false). The CLI displays this to the
	// user in the notice text so they know exactly where their key lives.
	Path string
}

// NewStore returns the appropriate KeyStore based on the environment.
//
// This is the legacy entry point that discards FallbackInfo. New callers
// that want to emit a fallback notice should call NewStoreWithStatus
// directly. Kept for backward compatibility with existing call sites
// (loadApp prior to F6, tests).
func NewStore(projectPath string) KeyStore {
	ks, _ := NewStoreWithStatus(projectPath)
	return ks
}

// NewStoreWithStatus returns the appropriate KeyStore plus a FallbackInfo
// describing whether the file-based fallback was used and why.
//
// Selection precedence:
//  1. TENE_KEYCHAIN_FALLBACK=file env override -> FileStore
//     (Reason="env_override")
//  2. OS keychain probe: a no-op Set() to the project-scoped service.
//     If it succeeds, we use the OS keychain (Reason="").
//     If it fails (CI, Docker, headless, libsecret missing), we fall
//     back to FileStore (Reason="keychain_unavailable").
//
// The fallback file path is always ~/.tene/keyfile (per-user, not
// per-project — keychain.NewFileStore historically used a single path,
// preserved here for backward compatibility with existing installs).
// Per-project isolation, when fallback is active, comes from the master
// key being derived per-vault rather than from the key file path.
func NewStoreWithStatus(projectPath string) (KeyStore, FallbackInfo) {
	home, _ := os.UserHomeDir()
	keyfilePath := filepath.Join(home, ".tene", "keyfile")

	if os.Getenv("TENE_KEYCHAIN_FALLBACK") == "file" {
		return NewFileStore(keyfilePath), FallbackInfo{
			Used:   true,
			Reason: "env_override",
			Path:   keyfilePath,
		}
	}

	service := ServiceName + "-" + hashPath(projectPath)
	ks := NewKeyringStore(service)

	// Test keychain availability via a small Set/Delete round-trip.
	testKey := "keychain-test"
	if err := keyring.Set(service, testKey, "test"); err != nil {
		// Keychain unavailable -> file fallback
		return NewFileStore(keyfilePath), FallbackInfo{
			Used:   true,
			Reason: "keychain_unavailable",
			Path:   keyfilePath,
		}
	}
	_ = keyring.Delete(service, testKey)

	return ks, FallbackInfo{Used: false}
}
