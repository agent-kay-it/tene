package claudemd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_NewFile(t *testing.T) {
	dir := t.TempDir()
	gen := NewGenerator(dir)

	created, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if !created {
		t.Error("expected created = true")
	}

	content, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	if !strings.Contains(string(content), SectionHeader) {
		t.Error("file should contain section header")
	}
	if !strings.Contains(string(content), "tene get") {
		t.Error("file should contain tene usage")
	}
}

func TestGenerate_ExistingWithoutSection(t *testing.T) {
	dir := t.TempDir()
	existing := "# My Project\n\nSome existing content.\n"
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(existing), 0644)

	gen := NewGenerator(dir)
	created, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if !created {
		t.Error("expected created = true")
	}

	content, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if !strings.Contains(string(content), "# My Project") {
		t.Error("should preserve existing content")
	}
	if !strings.Contains(string(content), SectionHeader) {
		t.Error("should append tene section")
	}
}

func TestGenerate_ExistingWithSection(t *testing.T) {
	dir := t.TempDir()
	existing := "# My Project\n\n# Secrets Management\n\nAlready has tene.\n"
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(existing), 0644)

	gen := NewGenerator(dir)
	created, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if created {
		t.Error("expected created = false (should skip)")
	}
}

func TestHasTeneSection(t *testing.T) {
	gen := NewGenerator("")

	tests := []struct {
		content  string
		expected bool
	}{
		{"# Secrets Management", true},
		{"uses tene for secret management", true},
		{"# My Project", false},
		{"", false},
	}

	for _, tc := range tests {
		if got := gen.HasTeneSection(tc.content); got != tc.expected {
			t.Errorf("HasTeneSection(%q) = %v, want %v", tc.content, got, tc.expected)
		}
	}
}
