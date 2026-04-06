package keychain

import "errors"

var (
	// ErrKeyNotFound is returned when no stored key exists.
	ErrKeyNotFound = errors.New("keychain: master key not found")

	// ErrKeychainUnavailable is returned when the OS keychain is unavailable.
	ErrKeychainUnavailable = errors.New("keychain: OS keychain unavailable")
)
