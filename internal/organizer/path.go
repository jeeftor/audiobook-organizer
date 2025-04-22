package organizer

import (
	"runtime"
	"strings"
)

// Invalid characters per OS
var (
	windowsInvalidChars = []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	unixInvalidChars    = []string{"/"}
	// Additional problematic characters to sanitize for all platforms
	commonProblematicChars = []string{"<", ">", ":", "|", "?", "*", "'", "`", "\""}
)

// SanitizePath sanitizes a file path string by replacing invalid characters based on the current OS.
// On Windows, it replaces '<', '>', ':', '"', '/', '\', '|', '?', '*' with underscores.
// On Unix systems, it replaces '/' and other problematic characters with underscores.
// If ReplaceSpace is set, it also replaces spaces with the specified character.
func (o *Organizer) SanitizePath(s string) string {
	// First replace spaces if configured
	if o.config.ReplaceSpace != "" {
		s = strings.ReplaceAll(s, " ", o.config.ReplaceSpace)
	}

	// Then handle OS-specific invalid characters
	var invalidChars []string
	if runtime.GOOS == "windows" {
		invalidChars = windowsInvalidChars
	} else if runtime.GOOS == "darwin" {
		invalidChars = []string{":"}
	} else {
		// Linux/Unix: strict, replace both / and common problematic chars
		invalidChars = append(unixInvalidChars, commonProblematicChars...)
	}

	// Replace invalid characters with underscore
	for _, char := range invalidChars {
		s = strings.ReplaceAll(s, char, "_")
	}

	// Trim leading and trailing spaces and dots
	s = strings.Trim(s, " .")

	return s
}

func cleanSeriesName(series string) string {
	if idx := strings.LastIndex(series, " #"); idx != -1 {
		return strings.TrimSpace(series[:idx])
	}
	return series
}
