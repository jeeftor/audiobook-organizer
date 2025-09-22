// internal/organizer/album_utils.go
package organizer

import (
	"sort"
	"strings"
)

// HasCommonPrefix checks if two strings have a common prefix that suggests they belong to the same album
// For example: "Album Name - Track 01" and "Album Name - Track 02"
func HasCommonPrefix(str1, str2 string) bool {
	// Find common prefix
	minLen := len(str1)
	if len(str2) < minLen {
		minLen = len(str2)
	}

	// For very short strings, check if they're identical up to a separator
	if minLen <= 5 {
		// Check if the strings start the same and contain a common separator
		commonSeparators := []string{" - ", ": ", ", "}
		for _, sep := range commonSeparators {
			if strings.Contains(str1, sep) && strings.Contains(str2, sep) {
				// Get the part before the separator
				parts1 := strings.Split(str1, sep)
				parts2 := strings.Split(str2, sep)
				if parts1[0] == parts2[0] {
					return true
				}
			}
		}
	}

	// Look for common prefix ending with a separator like " - " or ": "
	for i := minLen; i > 3; i-- {
		prefix := str1[:i]
		if strings.HasPrefix(str2, prefix) {
			// Check if the prefix ends with a common separator
			if strings.HasSuffix(prefix, " - ") ||
			   strings.HasSuffix(prefix, ": ") ||
			   strings.HasSuffix(prefix, ", ") {
				return true
			}
		}
	}

	return false
}

// HasTrackNumberPattern checks if two strings follow a track numbering pattern
// For example: "Book Title: Track 1" and "Book Title: Track 2"
// or "Book Title - Part 1" and "Book Title - Part 2"
func HasTrackNumberPattern(str1, str2 string) bool {
	// Common track number patterns
	patterns := []string{
		"Track", "Tr", "Part", "Chapter", "Disc", "CD", "Episode", "Ep", "Section", "Vol", "Volume",
	}

	// Check if both strings contain the same pattern word
	for _, pattern := range patterns {
		pattern = strings.ToLower(pattern)
		str1Lower := strings.ToLower(str1)
		str2Lower := strings.ToLower(str2)

		if strings.Contains(str1Lower, pattern) && strings.Contains(str2Lower, pattern) {
			// If both contain the pattern, they're likely part of the same album
			return true
		}
	}

	// Check for common base name with different numbers
	// Strip numbers from both strings and compare
	str1NoNum := stripNumbers(str1)
	str2NoNum := stripNumbers(str2)

	// If the strings are very similar without numbers, they likely belong to the same album
	return stringSimilarity(str1NoNum, str2NoNum) > 0.7
}

// SortFilesByTrackNumber sorts files by their track number metadata
func SortFilesByTrackNumber(files interface{}) {
	// Use type assertion to handle different file info types
	switch f := files.(type) {
	case []struct {
		path     string
		metadata Metadata
		dir      string
	}:
		sort.Slice(f, func(i, j int) bool {
			// If track numbers are available, sort by them
			if f[i].metadata.TrackNumber > 0 && f[j].metadata.TrackNumber > 0 {
				return f[i].metadata.TrackNumber < f[j].metadata.TrackNumber
			}

			// Otherwise sort by filename
			return f[i].path < f[j].path
		})
	}
}
