package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSetGet_Basic(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Set
	_, _, err := env.run("set", "API_KEY", "test-value-123", "--overwrite")
	if err != nil {
		t.Fatalf("set error: %v", err)
	}

	// Get
	stdout, _, err := env.run("get", "API_KEY")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}

	got := strings.TrimSpace(stdout)
	if got != "test-value-123" {
		t.Errorf("get = %q, want %q", got, "test-value-123")
	}
}

func TestSet_InvalidKeyName(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, err := env.run("set", "invalid-key", "value")
	if err == nil {
		t.Error("expected error for invalid key name")
	}
}

func TestSet_DuplicateWithoutOverwrite(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, _ = env.run("set", "MY_KEY", "value1")
	_, _, err := env.run("set", "MY_KEY", "value2")
	if err == nil {
		t.Error("expected error for duplicate key without --overwrite")
	}
}

func TestSet_OverwriteExisting(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, _ = env.run("set", "MY_KEY", "value1")
	_, _, err := env.run("set", "MY_KEY", "value2", "--overwrite")
	if err != nil {
		t.Fatalf("overwrite error: %v", err)
	}

	stdout, _, _ := env.run("get", "MY_KEY")
	if strings.TrimSpace(stdout) != "value2" {
		t.Errorf("get after overwrite = %q, want %q", strings.TrimSpace(stdout), "value2")
	}
}

func TestGet_SecretNotFound(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, err := env.run("get", "NONEXISTENT")
	if err == nil {
		t.Error("expected error for nonexistent secret")
	}
}

func TestGet_JSON(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	_, _, _ = env.run("set", "TEST_KEY", "test-val")

	stdout, _, err := env.runJSON("get", "TEST_KEY")
	if err != nil {
		t.Fatalf("get --json error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}
	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["value"] != "test-val" {
		t.Errorf("value = %v, want test-val", result["value"])
	}
}

func TestSetGet_TableDriven(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{"basic ASCII", "MY_SECRET", "hello123", false},
		{"unicode value", "UNICODE_VAL", "korean", false},
		{"long value", "LONG_VAL", strings.Repeat("x", 1000), false},
		{"special chars", "SPECIAL", "pa$$w0rd!@#", false},
		{"newline in value", "MULTILINE", "line1\nline2", false},
		{"empty value", "EMPTY", "", true},
		{"invalid key lowercase", "lowercase", "val", true},
		{"invalid key dash", "MY_DASH_KEY", "val", false},
		{"reserved key PATH", "PATH", "val", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := env.run("set", tt.key, tt.value, "--overwrite")
			if (err != nil) != tt.wantErr {
				t.Errorf("set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				stdout, _, getErr := env.run("get", tt.key)
				if getErr != nil {
					t.Errorf("get(%q) error = %v", tt.key, getErr)
					return
				}
				got := strings.TrimSpace(stdout)
				if got != tt.value {
					t.Errorf("get(%q) = %q, want %q", tt.key, got, tt.value)
				}
			}
		})
	}
}
