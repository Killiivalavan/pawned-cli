package board

import (
	"os"
	"strings"
)

var unicodeSupported bool

func init() {
	unicodeSupported = detectUnicodeSupport()
}

// detectUnicodeSupport checks environment variables to determine if the terminal supports Unicode.
func detectUnicodeSupport() bool {
	// 1. Check LANG and LC_ALL for UTF-8
	for _, env := range []string{"LANG", "LC_ALL"} {
		if val := os.Getenv(env); val != "" {
			if strings.Contains(strings.ToUpper(val), "UTF-8") {
				return true
			}
		}
	}

	// 2. Check TERM_PROGRAM
	termProg := os.Getenv("TERM_PROGRAM")
	switch termProg {
	case "iTerm.app", "WezTerm", "kitty", "vscode":
		return true
	}

	// 3. Check TERM
	term := os.Getenv("TERM")
	for _, supported := range []string{"xterm", "alacritty", "rxvt-unicode"} {
		if strings.Contains(term, supported) {
			return true
		}
	}

	return false
}

// IsUnicodeSupported returns whether the current environment supports Unicode chess glyphs.
func IsUnicodeSupported() bool {
	return unicodeSupported
}
