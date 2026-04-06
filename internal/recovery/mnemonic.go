package recovery

import (
	"fmt"

	"github.com/tyler-smith/go-bip39"
)

// GenerateMnemonic generates a 12-word BIP-39 mnemonic from 128-bit entropy.
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("recovery: failed to generate entropy: %w", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("recovery: failed to generate mnemonic: %w", err)
	}
	return mnemonic, nil
}

// ValidateMnemonic validates a BIP-39 mnemonic phrase.
func ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}
