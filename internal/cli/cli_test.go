package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDelete_Basic(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, _ = env.run("set", "DEL_KEY", "value1")

	// Delete with --force (skip confirmation)
	_, _, err := env.run("delete", "DEL_KEY", "--force")
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}

	// Get should fail
	_, _, err = env.run("get", "DEL_KEY")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestDelete_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, err := env.run("delete", "NONEXISTENT", "--force")
	if err == nil {
		t.Error("expected error for nonexistent secret")
	}
}

func TestList_Basic(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, _ = env.run("set", "KEY_A", "val1")
	_, _, _ = env.run("set", "KEY_B", "val2")

	stdout, _, err := env.run("list")
	if err != nil {
		t.Fatalf("list error: %v", err)
	}

	if !strings.Contains(stdout, "KEY_A") {
		t.Error("list should contain KEY_A")
	}
	if !strings.Contains(stdout, "KEY_B") {
		t.Error("list should contain KEY_B")
	}
}

func TestList_JSON(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, _ = env.run("set", "JSON_KEY", "json-val")

	stdout, _, err := env.runJSON("list")
	if err != nil {
		t.Fatalf("list --json error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse error: %v\nstdout: %s", err, stdout)
	}
	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	count, ok := result["count"].(float64)
	if !ok || count < 1 {
		t.Errorf("count = %v, want >= 1", result["count"])
	}
}

func TestEnvCreate_Basic(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, err := env.run("env", "create", "staging")
	if err != nil {
		t.Fatalf("env create error: %v", err)
	}

	// List should contain staging
	stdout, _, err := env.runJSON("env", "list")
	if err != nil {
		t.Fatalf("env list error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	envs, ok := result["environments"].([]any)
	if !ok {
		t.Fatal("expected environments array")
	}

	found := false
	for _, e := range envs {
		if m, ok := e.(map[string]any); ok && m["name"] == "staging" {
			found = true
		}
	}
	if !found {
		t.Error("staging environment not found in list")
	}
}

func TestEnvCreate_Duplicate(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, _ = env.run("env", "create", "staging")
	_, _, err := env.run("env", "create", "staging")
	if err == nil {
		t.Error("expected error for duplicate environment")
	}
}

func TestEnvSwitch(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, _ = env.run("env", "create", "staging")
	_, _, err := env.run("env", "staging")
	if err != nil {
		t.Fatalf("env switch error: %v", err)
	}
}

func TestEnvSwitch_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, err := env.run("env", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent environment")
	}
}

func TestImportExport_DotEnv(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Create a .env file
	envFile := filepath.Join(env.Dir, "test.env")
	content := "API_KEY=secret123\nDB_HOST=localhost\n"
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	// Import
	_, _, err := env.run("import", envFile)
	if err != nil {
		t.Fatalf("import error: %v", err)
	}

	// Verify
	stdout, _, err := env.run("get", "API_KEY")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if strings.TrimSpace(stdout) != "secret123" {
		t.Errorf("get API_KEY = %q, want %q", strings.TrimSpace(stdout), "secret123")
	}

	stdout, _, err = env.run("get", "DB_HOST")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if strings.TrimSpace(stdout) != "localhost" {
		t.Errorf("get DB_HOST = %q, want %q", strings.TrimSpace(stdout), "localhost")
	}

	// Export to file
	exportFile := filepath.Join(env.Dir, "export.env")
	_, _, err = env.run("export", "--file", exportFile)
	if err != nil {
		t.Fatalf("export error: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportFile); os.IsNotExist(err) {
		t.Error("export file not created")
	}

	data, _ := os.ReadFile(exportFile)
	exported := string(data)
	if !strings.Contains(exported, "API_KEY=secret123") {
		t.Error("export should contain API_KEY=secret123")
	}
}

func TestImport_FileNotFound(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, err := env.run("import", "/nonexistent/file.env")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestVersion(t *testing.T) {
	env := setupTestEnv(t)

	stdout, _, err := env.run("version")
	if err != nil {
		t.Fatalf("version error: %v", err)
	}
	if !strings.Contains(stdout, "tene") {
		t.Errorf("version output = %q, want to contain 'tene'", stdout)
	}
}

func TestWhoami(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	stdout, _, err := env.run("whoami")
	if err != nil {
		t.Fatalf("whoami error: %v", err)
	}
	if !strings.Contains(stdout, "test-project") {
		t.Errorf("whoami should contain project name, got: %s", stdout)
	}
}

func TestWhoami_JSON(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	stdout, _, err := env.runJSON("whoami")
	if err != nil {
		t.Fatalf("whoami error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}
	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["project"] != "test-project" {
		t.Errorf("project = %v, want test-project", result["project"])
	}
}
