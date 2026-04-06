package cli

import (
	"fmt"
	"os"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// colorEnabled returns whether color output is enabled.
func colorEnabled() bool {
	// --no-color flag
	if flagNoColor {
		return false
	}
	// NO_COLOR environment variable (https://no-color.org/)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	// TTY detection
	return isTerminal()
}

// colorize applies an ANSI color to text.
// Returns the original text if color is disabled.
func colorize(color, text string) string {
	if !colorEnabled() {
		return text
	}
	return color + text + colorReset
}

// Convenience functions
func colorRed_(text string) string    { return colorize(colorRed, text) }
func colorGreen_(text string) string  { return colorize(colorGreen, text) }
func colorYellow_(text string) string { return colorize(colorYellow, text) }
func colorBlue_(text string) string   { return colorize(colorBlue, text) }
func colorBold_(text string) string   { return colorize(colorBold, text) }
func colorDim_(text string) string    { return colorize(colorDim, text) }

// printSuccess outputs a success message.
func printSuccess(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorGreen_(msg))
}

// printWarning outputs a warning message to stderr.
func printWarning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s\n", colorYellow_(msg))
}

// printError_ outputs an error message to stderr.
func printError_(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s\n", colorRed_(msg))
}
