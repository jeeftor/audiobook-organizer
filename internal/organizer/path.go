package organizer

import (
	"runtime"
	"strings"
)

// Invalid characters per OS
var (
	windowsInvalidChars = []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	unixInvalidChars    = []string{"/"}
)

// SanitizePath sanitizes a file path string by replacing invalid characters based on the current OS.
// On Windows, it replaces '<', '>', ':', '"', '/', '\', '|', '?', '*' with underscores.
// On Unix systems, it only replaces '/' with underscore.
// If replaceSpace is set, it also replaces spaces with the specified character.
func (o *Organizer) SanitizePath(s string) string {
	// First replace spaces if configured
	if o.replaceSpace != "" {
		s = strings.ReplaceAll(s, " ", o.replaceSpace)
	}

	// Then handle OS-specific invalid characters
	var invalidChars []string
	if runtime.GOOS == "windows" {
		invalidChars = windowsInvalidChars
	} else {
		invalidChars = unixInvalidChars
	}

	// Replace invalid characters with underscore
	for _, char := range invalidChars {
		s = strings.ReplaceAll(s, char, "_")
	}

	return s
}

func cleanSeriesName(series string) string {
	if idx := strings.LastIndex(series, " #"); idx != -1 {
		return strings.TrimSpace(series[:idx])
	}
	return series
}
