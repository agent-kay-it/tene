package cli

import (
	"testing"
)

func TestColorEnabled_NoColorFlag(t *testing.T) {
	flagNoColor = true
	defer func() { flagNoColor = false }()

	if colorEnabled() {
		t.Error("colorEnabled() should return false when --no-color is set")
	}
}

func TestColorEnabled_NoColorEnv(t *testing.T) {
	flagNoColor = false
	t.Setenv("NO_COLOR", "1")

	if colorEnabled() {
		t.Error("colorEnabled() should return false when NO_COLOR env is set")
	}
}

func TestColorize_Disabled(t *testing.T) {
	flagNoColor = true
	defer func() { flagNoColor = false }()

	result := colorize(colorRed, "hello")
	if result != "hello" {
		t.Errorf("colorize() = %q, want %q (no color)", result, "hello")
	}
}

func TestColorize_AllFunctions_Disabled(t *testing.T) {
	flagNoColor = true
	defer func() { flagNoColor = false }()

	text := "test message"
	if got := colorRed_(text); got != text {
		t.Errorf("colorRed_() = %q, want %q", got, text)
	}
	if got := colorGreen_(text); got != text {
		t.Errorf("colorGreen_() = %q, want %q", got, text)
	}
	if got := colorYellow_(text); got != text {
		t.Errorf("colorYellow_() = %q, want %q", got, text)
	}
	if got := colorBlue_(text); got != text {
		t.Errorf("colorBlue_() = %q, want %q", got, text)
	}
	if got := colorBold_(text); got != text {
		t.Errorf("colorBold_() = %q, want %q", got, text)
	}
	if got := colorDim_(text); got != text {
		t.Errorf("colorDim_() = %q, want %q", got, text)
	}
}
