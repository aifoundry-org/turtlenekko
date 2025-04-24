package terminal

import (
	"os"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// Color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"
	Hidden    = "\033[8m"

	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// SupportsColor determines if the terminal supports color output
func SupportsColor() bool {
	// Check if NO_COLOR environment variable is set (standard for disabling color)
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}

	// Check if FORCE_COLOR environment variable is set (override detection)
	if _, exists := os.LookupEnv("FORCE_COLOR"); exists {
		return true
	}

	// Check if output is redirected to a file or pipe
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		// Check if we're in a CI environment which might support colors even when redirected
		if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("GITLAB_CI") != "" {
			return true
		}
		return false
	}

	// Get terminal type
	termType := os.Getenv("TERM")
	if termType == "" {
		// No TERM set, use OS-specific defaults
		if runtime.GOOS == "windows" {
			// Check for modern Windows terminals
			if _, exists := os.LookupEnv("WT_SESSION"); exists {
				return true // Windows Terminal
			}
			if _, exists := os.LookupEnv("WSLENV"); exists {
				return true // WSL
			}
			// ConEmu/Cmder sets ANSICON
			if _, exists := os.LookupEnv("ANSICON"); exists {
				return true
			}
			return false
		}
		// For Unix-like systems, default to true if we have a terminal
		return true
	}

	// Check for "dumb" terminal which doesn't support colors
	if termType == "dumb" {
		return false
	}

	// Check for common color-supporting terminal types
	colorTerms := []string{
		"xterm", "xterm-color", "xterm-256color",
		"screen", "screen-256color",
		"tmux", "tmux-256color",
		"rxvt", "rxvt-unicode", "rxvt-unicode-256color",
		"linux", "cygwin", "konsole",
		"vt100", "vt220", "ansi",
	}

	for _, t := range colorTerms {
		if strings.Contains(termType, t) {
			return true
		}
	}

	// Check for color terminal environment variable
	if colorterm := os.Getenv("COLORTERM"); colorterm != "" {
		return true
	}

	// Default to true for Linux and macOS if we have a terminal
	return runtime.GOOS == "linux" || runtime.GOOS == "darwin"
}

// Colorize applies color to text if supported
func Colorize(text string, color string) string {
	if SupportsColor() {
		return color + text + Reset
	}
	return text
}

// BoldText makes text bold if supported
func BoldText(text string) string {
	return Colorize(text, Bold)
}

// GreenText makes text green if supported
func GreenText(text string) string {
	return Colorize(text, Green)
}

// YellowText makes text yellow if supported
func YellowText(text string) string {
	return Colorize(text, Yellow)
}

// RedText makes text red if supported
func RedText(text string) string {
	return Colorize(text, Red)
}

// BlueText makes text blue if supported
func BlueText(text string) string {
	return Colorize(text, Blue)
}

// CyanText makes text cyan if supported
func CyanText(text string) string {
	return Colorize(text, Cyan)
}

// MagentaText makes text magenta if supported
func MagentaText(text string) string {
	return Colorize(text, Magenta)
}
