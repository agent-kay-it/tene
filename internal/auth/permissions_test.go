// Tests for the declarative permission tier model.
//
// Coverage strategy:
//   - String/RequiresUnlock — round-trip every defined PermLevel constant.
//   - CommandTier table — assert byte-by-byte the 26 entries from
//     docs/sprints/cli-ux-permission-model/design.md §1.1. A drift between
//     the design doc and this map is a G4 violation and must fail loud.
//   - TierFor — happy path + unknown path returns ok=false.
//   - Validate — synthetic cobra trees exercise the success path and the
//     missing-tier path so we don't depend on the real rootCmd here
//     (root_test.go covers the production tree).
package auth

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestPermLevel_String — every declared level produces a distinct,
// non-empty lowercase token. The string is part of the cli.<tier>.<verb>
// audit prefix (F4) so reordering or renaming here is a breaking change.
func TestPermLevel_String(t *testing.T) {
	cases := []struct {
		level PermLevel
		want  string
	}{
		{PermMetaRead, "metaread"},
		{PermSecretWrite, "secretwrite"},
		{PermSecretRead, "secretread"},
	}

	seen := make(map[string]struct{}, len(cases))
	for _, c := range cases {
		got := c.level.String()
		if got != c.want {
			t.Errorf("PermLevel(%d).String() = %q, want %q", int(c.level), got, c.want)
		}
		if got == "" {
			t.Errorf("PermLevel(%d).String() is empty", int(c.level))
		}
		if _, dup := seen[got]; dup {
			t.Errorf("PermLevel(%d).String() = %q duplicates an earlier level", int(c.level), got)
		}
		seen[got] = struct{}{}
	}
}

// TestPermLevel_String_Unknown — a stray int outside the defined constants
// still returns a non-empty diagnostic string instead of crashing. This
// guards against e.g. a future fourth tier being added but its String()
// branch being forgotten.
func TestPermLevel_String_Unknown(t *testing.T) {
	got := PermLevel(99).String()
	if got == "" {
		t.Errorf("PermLevel(99).String() is empty; want diagnostic")
	}
	if !strings.Contains(got, "unknown") {
		t.Errorf("PermLevel(99).String() = %q; want it to contain 'unknown'", got)
	}
}

// TestPermLevel_RequiresUnlock — PermMetaRead must be the ONLY tier that
// returns false. This is the gate that prevents `tene list`-style verbs
// from prompting for a password.
func TestPermLevel_RequiresUnlock(t *testing.T) {
	cases := []struct {
		level PermLevel
		want  bool
	}{
		{PermMetaRead, false},
		{PermSecretWrite, true},
		{PermSecretRead, true},
	}

	for _, c := range cases {
		if got := c.level.RequiresUnlock(); got != c.want {
			t.Errorf("PermLevel(%d).RequiresUnlock() = %v, want %v", int(c.level), got, c.want)
		}
	}
}

// TestCommandTier_HasAllExpected — exhaustive table assertion. Every
// entry from design.md §1.1 is verified; the total count is checked
// separately so adding an extra rogue entry without updating the design
// doc still fails the test.
//
// 26 entries: 16 PermMetaRead + 5 PermSecretWrite + 5 PermSecretRead.
func TestCommandTier_HasAllExpected(t *testing.T) {
	expected := map[string]PermLevel{
		// PermMetaRead (16)
		"list":        PermMetaRead,
		"env":         PermMetaRead,
		"env list":    PermMetaRead,
		"env create":  PermMetaRead,
		"env delete":  PermMetaRead,
		"permissions": PermMetaRead,
		"whoami":      PermMetaRead,
		"version":     PermMetaRead,
		"update":      PermMetaRead,
		"completion":  PermMetaRead,
		"logout":      PermMetaRead,
		"audit":       PermMetaRead,
		"audit tail":  PermMetaRead,
		"audit show":  PermMetaRead,
		"config":      PermMetaRead,
		"migrate":     PermMetaRead,

		// PermSecretWrite (5)
		"set":         PermSecretWrite,
		"import":      PermSecretWrite,
		"delete":      PermSecretWrite,
		"init":        PermSecretWrite,
		"audit prune": PermSecretWrite,

		// PermSecretRead (5)
		"get":     PermSecretRead,
		"export":  PermSecretRead,
		"run":     PermSecretRead,
		"passwd":  PermSecretRead,
		"recover": PermSecretRead,
	}

	if len(CommandTier) != len(expected) {
		t.Errorf("CommandTier has %d entries, expected %d", len(CommandTier), len(expected))
	}

	// Every expected entry must be present with the right tier.
	for path, want := range expected {
		got, ok := CommandTier[path]
		if !ok {
			t.Errorf("CommandTier[%q] missing", path)
			continue
		}
		if got != want {
			t.Errorf("CommandTier[%q] = %v, want %v", path, got, want)
		}
	}

	// And no extra rogue entries.
	for path := range CommandTier {
		if _, ok := expected[path]; !ok {
			t.Errorf("CommandTier[%q] is undeclared in expected table — update design.md §1.1", path)
		}
	}
}

// TestCommandTier_Counts — sanity check on the documented 16/5/5 split.
// Catches the case where a refactor accidentally re-tiers a verb (e.g.
// promoting `list` from MetaRead to SecretRead) which would silently
// re-introduce the password prompt user complaint.
func TestCommandTier_Counts(t *testing.T) {
	counts := map[PermLevel]int{}
	for _, tier := range CommandTier {
		counts[tier]++
	}

	expected := map[PermLevel]int{
		PermMetaRead:    16,
		PermSecretWrite: 5,
		PermSecretRead:  5,
	}

	for tier, want := range expected {
		if got := counts[tier]; got != want {
			t.Errorf("CommandTier count for %s = %d, want %d", tier, got, want)
		}
	}
}

// TestTierFor_Known — happy path: a known path returns its tier and ok=true.
func TestTierFor_Known(t *testing.T) {
	tier, ok := TierFor("list")
	if !ok {
		t.Fatal("TierFor(\"list\") ok = false, want true")
	}
	if tier != PermMetaRead {
		t.Errorf("TierFor(\"list\") = %v, want PermMetaRead", tier)
	}
}

// TestTierFor_KnownSubcommand — exercises a space-joined path so the
// design's "env list" / "audit prune" convention is validated.
func TestTierFor_KnownSubcommand(t *testing.T) {
	tier, ok := TierFor("audit prune")
	if !ok {
		t.Fatal("TierFor(\"audit prune\") ok = false, want true")
	}
	if tier != PermSecretWrite {
		t.Errorf("TierFor(\"audit prune\") = %v, want PermSecretWrite", tier)
	}
}

// TestTierFor_Unknown — a bogus path returns ok=false so callers can fail
// closed. The dispatcher in root.go uses this to refuse to run an
// undeclared command rather than silently defaulting to PermMetaRead.
func TestTierFor_Unknown(t *testing.T) {
	_, ok := TierFor("nonexistent-verb")
	if ok {
		t.Errorf("TierFor(\"nonexistent-verb\") ok = true, want false")
	}
}

// TestValidate_NilRoot — defensive: a nil root surfaces a clear error
// instead of nil-deref.
func TestValidate_NilRoot(t *testing.T) {
	if err := Validate(nil); err == nil {
		t.Errorf("Validate(nil) = nil, want error")
	}
}

// TestValidate_Pass — a synthetic cobra tree whose every leaf is in
// CommandTier returns nil. We reuse real entries ("list", "set", "get")
// so the test doesn't have to coordinate with arbitrary fake names.
func TestValidate_Pass(t *testing.T) {
	root := &cobra.Command{Use: "tene"}
	root.AddCommand(&cobra.Command{Use: "list", Run: func(*cobra.Command, []string) {}})
	root.AddCommand(&cobra.Command{Use: "set", Run: func(*cobra.Command, []string) {}})
	root.AddCommand(&cobra.Command{Use: "get", Run: func(*cobra.Command, []string) {}})

	if err := Validate(root); err != nil {
		t.Errorf("Validate(synthetic tree of all-tier-declared verbs) = %v, want nil", err)
	}
}

// TestValidate_FailMissing — adding an unregistered command surfaces an
// error wrapped around ErrMissingTier whose message names the missing
// path so an operator can locate it. This is the test that, when run by
// CI, catches a developer who adds AddCommand without updating
// CommandTier (G4 enforcement).
func TestValidate_FailMissing(t *testing.T) {
	root := &cobra.Command{Use: "tene"}
	root.AddCommand(&cobra.Command{Use: "list", Run: func(*cobra.Command, []string) {}})
	root.AddCommand(&cobra.Command{Use: "rogue-undeclared-cmd", Run: func(*cobra.Command, []string) {}})

	err := Validate(root)
	if err == nil {
		t.Fatal("Validate(tree with undeclared verb) = nil, want error")
	}
	if !errors.Is(err, ErrMissingTier) {
		t.Errorf("Validate error = %v, want errors.Is(..., ErrMissingTier)", err)
	}
	if !strings.Contains(err.Error(), "rogue-undeclared-cmd") {
		t.Errorf("Validate error %q does not name the missing path", err.Error())
	}
}

// TestValidate_FailMissingSubcommand — same check at one level deeper.
// Verifies that the walk recurses into command groups.
func TestValidate_FailMissingSubcommand(t *testing.T) {
	root := &cobra.Command{Use: "tene"}
	group := &cobra.Command{Use: "audit", Run: func(*cobra.Command, []string) {}}
	group.AddCommand(&cobra.Command{Use: "rogue-sub", Run: func(*cobra.Command, []string) {}})
	root.AddCommand(group)

	err := Validate(root)
	if err == nil {
		t.Fatal("Validate(tree with undeclared sub-verb) = nil, want error")
	}
	if !strings.Contains(err.Error(), "audit rogue-sub") {
		t.Errorf("Validate error %q does not name the missing sub-path", err.Error())
	}
}

// TestValidate_IgnoresHelpCommand — cobra auto-adds a "help" command.
// Our walk must skip it so a real root tree (which never declares "help"
// in CommandTier) validates cleanly.
func TestValidate_IgnoresHelpCommand(t *testing.T) {
	root := &cobra.Command{Use: "tene"}
	root.AddCommand(&cobra.Command{Use: "list", Run: func(*cobra.Command, []string) {}})

	// Force cobra to materialize its help machinery by calling InitDefaultHelpCmd
	// — this mirrors what happens once rootCmd.Execute() runs.
	root.InitDefaultHelpCmd()

	if err := Validate(root); err != nil {
		t.Errorf("Validate(tree with cobra auto help cmd) = %v, want nil — help must be skipped", err)
	}
}
