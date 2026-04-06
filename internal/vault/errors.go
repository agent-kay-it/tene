package vault

import "errors"

var (
	// ErrSecretNotFound is returned when the requested secret does not exist.
	ErrSecretNotFound = errors.New("vault: secret not found")

	// ErrMetaNotFound is returned when the requested metadata does not exist.
	ErrMetaNotFound = errors.New("vault: metadata not found")

	// ErrEnvironmentNotFound is returned when the requested environment does not exist.
	ErrEnvironmentNotFound = errors.New("vault: environment not found")

	// ErrEnvironmentExists is returned when trying to create an environment that already exists.
	ErrEnvironmentExists = errors.New("vault: environment already exists")

	// ErrVaultNotInitialized is returned when the vault is not initialized.
	ErrVaultNotInitialized = errors.New("vault: not initialized, run 'tene init' first")

	// ErrDatabaseCorrupted is returned when the database is corrupted.
	ErrDatabaseCorrupted = errors.New("vault: database corrupted")
)
