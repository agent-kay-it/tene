package vault

import "time"

// Secret represents an encrypted secret record.
type Secret struct {
	ID             int64
	Name           string
	EncryptedValue string // base64(nonce + ciphertext)
	Environment    string
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Environment represents a secret environment setting.
type Environment struct {
	Name      string
	IsActive  bool
	CreatedAt time.Time
}

// AuditEntry represents an audit log entry.
type AuditEntry struct {
	ID           int64
	Action       string // "secret.read", "secret.write", "secret.delete", "vault.init", "vault.passwd"
	ResourceName string
	Details      string
	Timestamp    time.Time
}
