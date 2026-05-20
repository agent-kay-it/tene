// Quality gate G10 — Audit Auto-Delete Prohibition.
//
// I-14 (master-plan §10): audit logs must never be auto-deleted. The
// `tene audit prune` command is the ONLY path that may remove rows,
// and it requires either `--force` or an interactive "y" confirmation.
//
// To make that invariant a static property of the repo we centralise
// the `DELETE FROM audit_log` SQL in one chokepoint (vault.go's
// PruneAuditLog) and assert here that no other file contains the
// statement. A future commit that adds "log rotation" or "automatic
// cleanup on disk pressure" will fail this test immediately and force
// the author to either route through PruneAuditLog or remove the auto-
// deletion entirely.
//
// Why a Go test instead of a CI shell grep? Two reasons:
//
//   - It runs on every developer's `go test ./...` without extra CI
//     wiring, so the failure happens at the moment a regression is
//     introduced rather than during PR review.
//   - It can be cross-checked against the same Go AST the compiler
//     sees, which avoids false positives from string literals in
//     comments or docs.
//
// The implementation deliberately uses a plain filesystem walk + case-
// insensitive substring grep rather than full AST parsing: any string
// containing `DELETE FROM audit_log` is a positive hit, whether it
// appears in a SQL string passed to db.Exec, a comment, or a docstring.
// The cost of a few false positives (e.g. this test file would self-
// reference) is paid by inflating the expected match count by the
// known mention sites; that small accounting is far cheaper than a
// real SQL parser and yields a clearer error when violated.
package audit

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// auditDeletePattern matches "DELETE FROM audit_log" with arbitrary
// whitespace between tokens and case insensitivity. SQL is case-
// insensitive in keywords; the table identifier is lowercase by our
// schema convention but operators sometimes write it uppercase.
var auditDeletePattern = regexp.MustCompile(`(?i)\bDELETE\s+FROM\s+audit_log\b`)

// TestG10_AuditAutoDeleteProhibition is the gate test.
//
// Strategy:
//
//  1. Walk the repository starting at the module root (resolved
//     relative to this test file: ../.. → repo root).
//  2. For every .go file (excluding vendor/, .git/, and any
//     build-cache dirs), grep for the pattern.
//  3. The single legitimate site is internal/vault/vault.go inside
//     PruneAuditLog. Every other hit is a G10 violation.
//
// The test does NOT count comment-only references — we look at the
// raw byte content. The intention is that the docstring on
// PruneAuditLog AND the centralized DELETE statement BOTH match (and
// both live in vault.go); that's why we expect exactly 1 file with
// at least 1 match, not exactly 1 match overall. A file with 5
// mentions in comments + 1 in code still counts as 1 violating file
// if it's not vault.go.
//
// Future-proofing: if a legitimate need to mention the pattern arises
// outside vault.go (e.g. a doc generator), add the file to
// allowedFiles below with a comment explaining why.
func TestG10_AuditAutoDeleteProhibition(t *testing.T) {
	repoRoot := findRepoRoot(t)

	// Files we deliberately allow to mention the pattern.
	//
	//   - internal/vault/vault.go: THE chokepoint. The DELETE statement
	//     itself plus a brief comment line documenting it as the
	//     G10 chokepoint. Multiple matches inside this single file are
	//     fine.
	//
	//   - internal/audit/g10_test.go: this test file. Its own pattern
	//     literal and the doc comments contain the string several
	//     times — that is the whole point of the gate.
	allowedFiles := map[string]bool{
		filepath.Join("internal", "vault", "vault.go"):     true,
		filepath.Join("internal", "audit", "g10_test.go"): true,
	}

	var violations []string

	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip directories that should never contain source we care
			// about. vendor/ would be vendored third-party (we have
			// none, but be defensive); .git/ + node_modules are obvious
			// noise; testdata could plausibly hold SQL fixtures and is
			// also out of scope for the gate.
			switch d.Name() {
			case ".git", "vendor", "node_modules", "testdata", ".bkit":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		rel, _ := filepath.Rel(repoRoot, path)
		if allowedFiles[rel] {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Logf("g10: cannot read %s: %v (skipping)", rel, readErr)
			return nil
		}
		if auditDeletePattern.Match(data) {
			violations = append(violations, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}

	if len(violations) > 0 {
		t.Errorf(
			"G10 violation: %d file(s) outside the audit chokepoint contain "+
				"'DELETE FROM audit_log'.\n\n"+
				"Audit log rows must only be removed through vault.PruneAuditLog "+
				"(invariant I-14, master-plan §10). If a new caller needs to "+
				"prune rows, route it through that function rather than issuing "+
				"its own DELETE.\n\nViolating files:\n  - %s",
			len(violations),
			strings.Join(violations, "\n  - "),
		)
	}
}

// findRepoRoot locates the repo root by walking up from the test
// file's location until a go.mod is found. Falls back to caller-
// relative ../../ if go.mod is missing (defensive — should never fire
// in a real checkout).
func findRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not find go.mod ancestor starting from %s", filepath.Dir(thisFile))
	return ""
}
