package claudemd

import (
	"os"
	"path/filepath"
	"strings"
)

// Generator creates and manages CLAUDE.md files.
type Generator struct {
	projectDir string
}

// NewGenerator creates a new Generator.
func NewGenerator(projectDir string) *Generator {
	return &Generator{projectDir: projectDir}
}

// Generate creates CLAUDE.md or appends the tene section to an existing file.
// Returns true if the file was created/modified, false if skipped.
func (g *Generator) Generate() (bool, error) {
	path := filepath.Join(g.projectDir, "CLAUDE.md")

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new file
			return true, os.WriteFile(path, []byte(SecretsMdTemplate), 0644)
		}
		return false, err
	}

	// Skip if tene section already exists
	if g.HasTeneSection(string(content)) {
		return false, nil
	}

	// Append section to existing file
	separator := "\n\n"
	if !strings.HasSuffix(string(content), "\n") {
		separator = "\n\n"
	}
	updated := string(content) + separator + SecretsMdTemplate
	return true, os.WriteFile(path, []byte(updated), 0644)
}

// HasTeneSection checks if the content already contains the tene Secrets Management section.
func (g *Generator) HasTeneSection(content string) bool {
	return strings.Contains(content, SectionHeader) ||
		(strings.Contains(content, "tene") && strings.Contains(content, "secret management"))
}

// FilePath returns the absolute path to CLAUDE.md.
func (g *Generator) FilePath() string {
	return filepath.Join(g.projectDir, "CLAUDE.md")
}
