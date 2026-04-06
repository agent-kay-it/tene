package encfile

import (
	"bytes"
	"testing"
)

func TestHeaderEncodeDecode(t *testing.T) {
	h := &Header{
		FormatVersion: FormatVersion,
		KDFAlgorithm:  KDFAlgArgon2id,
		KDFMemory:     65536,
		KDFIterations: 3,
		KDFParallel:   1,
	}
	// salt/nonce are zero-initialized

	encoded := h.Encode()
	if len(encoded) != HeaderSize {
		t.Errorf("header size = %d, want %d", len(encoded), HeaderSize)
	}

	// Verify magic bytes
	if !bytes.Equal(encoded[:4], MagicBytes[:]) {
		t.Errorf("magic = %x, want %x", encoded[:4], MagicBytes)
	}

	decoded, err := DecodeHeader(encoded)
	if err != nil {
		t.Fatalf("DecodeHeader() error: %v", err)
	}

	if decoded.FormatVersion != FormatVersion {
		t.Errorf("FormatVersion = %d, want %d", decoded.FormatVersion, FormatVersion)
	}
	if decoded.KDFMemory != 65536 {
		t.Errorf("KDFMemory = %d, want 65536", decoded.KDFMemory)
	}
	if decoded.KDFIterations != 3 {
		t.Errorf("KDFIterations = %d, want 3", decoded.KDFIterations)
	}
	if decoded.KDFParallel != 1 {
		t.Errorf("KDFParallel = %d, want 1", decoded.KDFParallel)
	}
}

func TestDecodeHeader_InvalidMagic(t *testing.T) {
	data := make([]byte, HeaderSize)
	copy(data[:4], []byte("FAKE"))

	_, err := DecodeHeader(data)
	if err != ErrInvalidMagic {
		t.Errorf("expected ErrInvalidMagic, got %v", err)
	}
}

func TestDecodeHeader_TooShort(t *testing.T) {
	data := make([]byte, 10)
	_, err := DecodeHeader(data)
	if err != ErrFileTooShort {
		t.Errorf("expected ErrFileTooShort, got %v", err)
	}
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	password := "testpassword123"
	plaintext := []byte(`{"exportVersion":1,"secrets":[]}`)

	encrypted, err := Encrypt(password, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Verify header
	header, err := DecodeHeader(encrypted)
	if err != nil {
		t.Fatalf("DecodeHeader() error: %v", err)
	}
	if header.FormatVersion != FormatVersion {
		t.Errorf("FormatVersion = %d, want %d", header.FormatVersion, FormatVersion)
	}

	decrypted, err := Decrypt(password, encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestDecrypt_WrongPassword(t *testing.T) {
	plaintext := []byte(`{"test":true}`)

	encrypted, _ := Encrypt("correct-password", plaintext)
	_, err := Decrypt("wrong-password", encrypted)
	if err == nil {
		t.Error("expected error for wrong password")
	}
}
