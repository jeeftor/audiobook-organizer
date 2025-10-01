// internal/organizer/album_detection.go
package organizer

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Variables to allow mocking in tests
var newAudioMetadataProviderFunc func(path string) MetadataProvider = func(path string) MetadataProvider {
	return NewAudioMetadataProvider(path)
}
var readDirFunc = os.ReadDir

// shouldProcessAsAlbum determines if a directory should be processed as a multi-file album
// based on the number and type of audio files it contains
func (o *Organizer) shouldProcessAsAlbum(dirPath string) bool {
	// Read all files in the directory
	entries, err := readDirFunc(dirPath)
	if err != nil {
		return false
	}

	// Count audio files and check if they have consistent metadata
	audioFiles := 0
	var albumTitle, albumArtist, albumSeries string
	var firstMetadata *Metadata

	// Track numbers to detect sequential tracks
	trackNumbers := make(map[int]bool)
	hasTrackNumbers := false

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		filePath := filepath.Join(dirPath, entry.Name())
		ext := strings.ToLower(filepath.Ext(filePath))

		// Check if this is an audio file
		if !IsSupportedAudioFile(ext) {
			continue
		}

		audioFiles++

		// Extract metadata to check for consistency
		provider := newAudioMetadataProviderFunc(filePath)
		metadata, err := provider.GetMetadata()
		if err != nil {
			continue
		}

		// Apply field mapping
		metadata.ApplyFieldMapping(o.config.FieldMapping)

		// Track sequential track numbers
		if metadata.TrackNumber > 0 {
			trackNumbers[metadata.TrackNumber] = true
			hasTrackNumbers = true
		}

		// Store first metadata for comparison
		if firstMetadata == nil {
			firstMetadata = &metadata
			albumTitle = metadata.Title
			albumSeries = metadata.GetValidSeries()
			if len(metadata.Authors) > 0 {
				albumArtist = metadata.Authors[0]
			}
		} else {
			// Check if this file belongs to the same album
			currentTitle := metadata.Title
			currentSeries := metadata.GetValidSeries()
			var currentArtist string
			if len(metadata.Authors) > 0 {
				currentArtist = metadata.Authors[0]
			}

			// Check for series consistency
			seriesMatch := (albumSeries == "" || currentSeries == "" || albumSeries == currentSeries)

			// If title or artist doesn't match, this might not be a cohesive album
			if currentTitle != albumTitle || (albumArtist != "" && currentArtist != "" && currentArtist != albumArtist) {
				// Titles don't match exactly, but check for patterns
				if !hasCommonPrefix(currentTitle, albumTitle) && !hasTrackNumberPattern(currentTitle, albumTitle) {
					// If series matches and we have track numbers, still consider it an album
					if !(seriesMatch && hasTrackNumbers && len(trackNumbers) > 1) {
						return false
					}
				}
			}
		}
	}

	// Check if track numbers are sequential or close to sequential
	if hasTrackNumbers && len(trackNumbers) > 1 {
		// If we have at least 2 sequential track numbers, it's likely an album
		for i := 1; i <= len(trackNumbers); i++ {
			if trackNumbers[i] && trackNumbers[i+1] {
				return true
			}
		}
	}

	// Consider it an album if there are multiple audio files with consistent metadata
	return audioFiles > 1 && firstMetadata != nil
}


// hasCommonPrefix checks if two strings have a common prefix that suggests they belong to the same album
// For example: "Album Name - Track 01" and "Album Name - Track 02"
func hasCommonPrefix(str1, str2 string) bool {
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

// hasTrackNumberPattern checks if two strings follow a track numbering pattern
// For example: "Book Title: Track 1" and "Book Title: Track 2"
// or "Book Title - Part 1" and "Book Title - Part 2"
func hasTrackNumberPattern(str1, str2 string) bool {
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

// stripNumbers removes all digits from a string
func stripNumbers(s string) string {
	var result strings.Builder
	for _, ch := range s {
		if !unicode.IsDigit(ch) {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// stringSimilarity calculates how similar two strings are (0.0 to 1.0)
// using a simple character-based approach
func stringSimilarity(s1, s2 string) float64 {
	// Convert to lowercase for comparison
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	// Get the shorter and longer string
	shorter, longer := s1, s2
	if len(s1) > len(s2) {
		shorter, longer = s2, s1
	}

	// Count matching characters
	matches := 0
	for i := 0; i < len(shorter); i++ {
		if shorter[i] == longer[i] {
			matches++
		}
	}

	// Calculate similarity as percentage of matching characters
	return float64(matches) / float64(len(longer))
}
