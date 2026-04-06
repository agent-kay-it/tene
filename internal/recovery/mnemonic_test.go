package recovery

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateMnemonic(t *testing.T) {
	mnemonic, err := GenerateMnemonic()
	if err != nil {
		t.Fatalf("GenerateMnemonic() error: %v", err)
	}

	words := strings.Fields(mnemonic)
	if len(words) != 12 {
		t.Errorf("word count = %d, want 12", len(words))
	}

	if !ValidateMnemonic(mnemonic) {
		t.Error("generated mnemonic should be valid")
	}
}

func TestValidateMnemonic_Invalid(t *testing.T) {
	if ValidateMnemonic("not a valid mnemonic phrase at all") {
		t.Error("expected invalid mnemonic")
	}
}

func TestEncryptRecoverRoundtrip(t *testing.T) {
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	mnemonic, err := GenerateMnemonic()
	if err != nil {
		t.Fatalf("GenerateMnemonic() error: %v", err)
	}

	blob, err := EncryptMasterKey(masterKey, mnemonic)
	if err != nil {
		t.Fatalf("EncryptMasterKey() error: %v", err)
	}

	recovered, err := RecoverMasterKey(blob, mnemonic)
	if err != nil {
		t.Fatalf("RecoverMasterKey() error: %v", err)
	}

	if !bytes.Equal(masterKey, recovered) {
		t.Error("recovered key does not match original")
	}
}

func TestRecoverMasterKey_WrongMnemonic(t *testing.T) {
	masterKey := make([]byte, 32)

	mnemonic1, _ := GenerateMnemonic()
	mnemonic2, _ := GenerateMnemonic()

	blob, err := EncryptMasterKey(masterKey, mnemonic1)
	if err != nil {
		t.Fatalf("EncryptMasterKey() error: %v", err)
	}

	_, err = RecoverMasterKey(blob, mnemonic2)
	if err == nil {
		t.Error("expected error with wrong mnemonic")
	}
}

func TestEncryptMasterKey_InvalidMnemonic(t *testing.T) {
	_, err := EncryptMasterKey(make([]byte, 32), "invalid")
	if err == nil {
		t.Error("expected error for invalid mnemonic")
	}
}
