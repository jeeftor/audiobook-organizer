// internal/organizer/path.go
package organizer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Invalid characters per OS
var (
	windowsInvalidChars = []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	unixInvalidChars    = []string{"/"}
	// Additional problematic characters to sanitize for all platforms
	// Note: Removed apostrophe (') from this list to ensure consistent behavior across platforms
	commonProblematicChars = []string{"<", ">", ":", "|", "?", "*", "`", "\""}
)

// SupportedAudioExtensions as a map for O(1) lookup instead of slice iteration
var SupportedAudioExtensions = map[string]bool{
	".mp3":  true,
	".m4b":  true,
	".m4a":  true,
	".ogg":  true,
	".flac": true,
}

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
		invalidChars = append(windowsInvalidChars, commonProblematicChars...)
	} else if runtime.GOOS == "darwin" {
		invalidChars = []string{":"}
	} else {
		// Linux/Unix: only replace truly problematic characters
		// We're keeping apostrophes intact for consistent behavior with tests
		invalidChars = unixInvalidChars
		// Only add common problematic chars that aren't apostrophes
		for _, char := range commonProblematicChars {
			invalidChars = append(invalidChars, char)
		}
	}

	// Replace invalid characters with underscore
	for _, char := range invalidChars {
		s = strings.ReplaceAll(s, char, "_")
	}

	// Trim leading and trailing spaces and dots
	s = strings.Trim(s, " .")

	return s
}

// CleanSeriesName removes trailing series numbers (e.g., " #1") from series names.
// This is now public so it can be used throughout the package.
func CleanSeriesName(series string) string {
	if idx := strings.LastIndex(series, " #"); idx != -1 {
		return strings.TrimSpace(series[:idx])
	}
	return series
}

// ExtractSeriesNumber extracts the series number from a series string (e.g., "Mistborn #1" -> "1").
// Returns an empty string if no series number is found.
func ExtractSeriesNumber(series string) string {
	if idx := strings.LastIndex(series, " #"); idx != -1 {
		return strings.TrimSpace(series[idx+2:])
	}
	return ""
}

// GetSeriesNumberFromMetadata extracts the series number from metadata.
// It first checks RawData for series_index, then falls back to parsing the series string.
func GetSeriesNumberFromMetadata(metadata Metadata) string {
	// First try to get series_index from RawData
	if seriesIndex, ok := metadata.RawData["series_index"].(float64); ok && seriesIndex > 0 {
		// Format as integer if it's a whole number, otherwise with decimal
		if seriesIndex == float64(int(seriesIndex)) {
			return fmt.Sprintf("%d", int(seriesIndex))
		}
		return fmt.Sprintf("%.1f", seriesIndex)
	}

	// Fall back to extracting from series string
	if len(metadata.Series) > 0 && metadata.Series[0] != "" {
		return ExtractSeriesNumber(metadata.Series[0])
	}

	return ""
}

// IsSupportedAudioFile checks if a file extension represents a supported audio format.
// Uses a map for O(1) lookup performance instead of slice iteration.
func IsSupportedAudioFile(ext string) bool {
	return SupportedAudioExtensions[strings.ToLower(ext)]
}

// AddTrackPrefix adds a track number prefix to a filename if not already present.
// If a prefix exists but has wrong padding, it will be re-formatted with correct padding.
// The padding is automatically determined based on totalTracks (2 digits for <100, 3 for 100-999, etc.)
func AddTrackPrefix(filename string, trackNumber int, totalTracks int) string {
	if trackNumber <= 0 {
		return filename
	}

	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	// Calculate the required padding based on total tracks
	format := GetTrackPrefixFormat(totalTracks)
	correctPrefix := fmt.Sprintf(format, trackNumber)

	// Check if the filename already has a track prefix
	if HasTrackPrefix(baseName) {
		// Extract existing track number
		existingTrackNum := ExtractTrackNumber(baseName)

		// If the track number matches, remove old prefix and add new one with correct padding
		if existingTrackNum == trackNumber {
			// Remove the old prefix
			nameWithoutPrefix := RemoveTrackPrefix(baseName)
			return fmt.Sprintf("%s%s%s", correctPrefix, nameWithoutPrefix, ext)
		}

		// If track numbers don't match, keep original (something's wrong)
		return filename
	}

	// No existing prefix - add new one
	return fmt.Sprintf("%s%s%s", correctPrefix, baseName, ext)
}

// GetTrackPrefixFormat returns the appropriate format string based on the total number of tracks
// Examples:
//   - totalTracks < 100:  returns "%02d - " (e.g., "01 - ", "99 - ")
//   - totalTracks < 1000: returns "%03d - " (e.g., "001 - ", "999 - ")
//   - totalTracks >= 1000: returns "%04d - " (e.g., "0001 - ")
func GetTrackPrefixFormat(totalTracks int) string {
	if totalTracks < 100 {
		return "%02d - "
	} else if totalTracks < 1000 {
		return "%03d - "
	} else {
		return "%04d - "
	}
}

// HasTrackPrefix checks if a filename already has a track number prefix.
// Supports variable-length prefixes: "01 - ", "001 - ", "0001 - ", etc.
func HasTrackPrefix(filename string) bool {
	if len(filename) < 5 {
		return false
	}

	// Check for patterns: NN - , NNN - , NNNN - , etc.
	// Find the position of " - " separator
	for i := 2; i <= 4 && i < len(filename)-2; i++ {
		if i < len(filename) && filename[i] == ' ' &&
		   i+1 < len(filename) && filename[i+1] == '-' &&
		   i+2 < len(filename) && filename[i+2] == ' ' {
			// Check if all characters before the separator are digits
			allDigits := true
			for j := 0; j < i; j++ {
				if filename[j] < '0' || filename[j] > '9' {
					allDigits = false
					break
				}
			}
			return allDigits
		}
	}

	return false
}

// ExtractTrackNumber extracts the track number from a filename prefix.
// Returns 0 if no track number prefix is found.
// Supports variable-length prefixes: "01 - ", "001 - ", "0001 - ", etc.
func ExtractTrackNumber(filename string) int {
	if !HasTrackPrefix(filename) {
		return 0
	}

	// Find where " - " starts to determine the length of the track number
	for i := 2; i <= 4 && i < len(filename)-2; i++ {
		if filename[i] == ' ' && filename[i+1] == '-' && filename[i+2] == ' ' {
			trackStr := filename[:i]
			var trackNum int
			if _, err := fmt.Sscanf(trackStr, "%d", &trackNum); err == nil {
				return trackNum
			}
			break
		}
	}

	return 0
}

// RemoveTrackPrefix removes the track number prefix from a filename if present.
// Supports variable-length prefixes: "01 - ", "001 - ", "0001 - ", etc.
func RemoveTrackPrefix(filename string) string {
	if !HasTrackPrefix(filename) {
		return filename
	}

	// Find where " - " ends to determine how many characters to remove
	for i := 2; i <= 4 && i < len(filename)-2; i++ {
		if filename[i] == ' ' && filename[i+1] == '-' && filename[i+2] == ' ' {
			// Remove everything up to and including " - "
			return filename[i+3:]
		}
	}

	return filename
}

// NormalizeFilename provides various filename normalization options.
type FilenameNormalizer struct {
	replaceSpaces    bool
	spaceReplacement string
	addTrackPrefix   bool
	trackNumber      int
	totalTracks      int
}

// NewFilenameNormalizer creates a new filename normalizer with the given options.
func NewFilenameNormalizer() *FilenameNormalizer {
	return &FilenameNormalizer{}
}

// WithSpaceReplacement configures the normalizer to replace spaces.
func (fn *FilenameNormalizer) WithSpaceReplacement(replacement string) *FilenameNormalizer {
	fn.replaceSpaces = true
	fn.spaceReplacement = replacement
	return fn
}

// WithTrackPrefix configures the normalizer to add a track number prefix.
// totalTracks is used to determine the padding (e.g., 2 digits for <100 tracks, 3 for 100+)
func (fn *FilenameNormalizer) WithTrackPrefix(trackNumber int, totalTracks int) *FilenameNormalizer {
	fn.addTrackPrefix = true
	fn.trackNumber = trackNumber
	fn.totalTracks = totalTracks
	return fn
}

// Normalize applies all configured normalizations to the filename.
func (fn *FilenameNormalizer) Normalize(filename string) string {
	result := filename

	// Add track prefix if configured
	if fn.addTrackPrefix {
		result = AddTrackPrefix(result, fn.trackNumber, fn.totalTracks)
	}

	// Replace spaces if configured
	if fn.replaceSpaces && fn.spaceReplacement != "" {
		result = strings.ReplaceAll(result, " ", fn.spaceReplacement)
	}

	return result
}

// PathValidator provides validation for file and directory paths.
type PathValidator struct{}

// NewPathValidator creates a new path validator.
func NewPathValidator() *PathValidator {
	return &PathValidator{}
}

// IsValidPath checks if a path contains only valid characters for the current OS.
func (pv *PathValidator) IsValidPath(path string) bool {
	var invalidChars []string

	if runtime.GOOS == "windows" {
		invalidChars = windowsInvalidChars
	} else if runtime.GOOS == "darwin" {
		invalidChars = []string{":"}
	} else {
		invalidChars = append(unixInvalidChars, commonProblematicChars...)
	}

	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return false
		}
	}

	return true
}

// GetInvalidChars returns a list of characters that are invalid in the path.
func (pv *PathValidator) GetInvalidChars(path string) []string {
	var invalidChars []string
	var checkChars []string

	if runtime.GOOS == "windows" {
		checkChars = windowsInvalidChars
	} else if runtime.GOOS == "darwin" {
		checkChars = []string{":"}
	} else {
		checkChars = append(unixInvalidChars, commonProblematicChars...)
	}

	for _, char := range checkChars {
		if strings.Contains(path, char) {
			invalidChars = append(invalidChars, char)
		}
	}

	return invalidChars
}

// SuggestSanitizedPath suggests a sanitized version of the path.
func (pv *PathValidator) SuggestSanitizedPath(path string) string {
	result := path
	var invalidChars []string

	if runtime.GOOS == "windows" {
		invalidChars = windowsInvalidChars
	} else if runtime.GOOS == "darwin" {
		invalidChars = []string{":"}
	} else {
		invalidChars = append(unixInvalidChars, commonProblematicChars...)
	}

	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	return strings.Trim(result, " .")
}

// PathBuilder provides a fluent interface for building paths with various components.
type PathBuilder struct {
	parts     []string
	sanitizer func(string) string
}

// NewPathBuilder creates a new path builder.
func NewPathBuilder() *PathBuilder {
	return &PathBuilder{
		parts:     make([]string, 0),
		sanitizer: func(s string) string { return s }, // Default: no sanitization
	}
}

// WithSanitizer sets a custom sanitization function.
func (pb *PathBuilder) WithSanitizer(sanitizer func(string) string) *PathBuilder {
	pb.sanitizer = sanitizer
	return pb
}

// AddAuthor adds an author component to the path.
func (pb *PathBuilder) AddAuthor(author string) *PathBuilder {
	if author != "" {
		pb.parts = append(pb.parts, pb.sanitizer(author))
	}
	return pb
}

// AddSeries adds a series component to the path.
func (pb *PathBuilder) AddSeries(series string) *PathBuilder {
	if series != "" && series != InvalidSeriesValue {
		cleanSeries := CleanSeriesName(series)
		pb.parts = append(pb.parts, pb.sanitizer(cleanSeries))
	}
	return pb
}

// AddTitle adds a title component to the path.
func (pb *PathBuilder) AddTitle(title string) *PathBuilder {
	if title != "" {
		pb.parts = append(pb.parts, pb.sanitizer(title))
	}
	return pb
}

// AddCustom adds a custom component to the path.
func (pb *PathBuilder) AddCustom(component string) *PathBuilder {
	if component != "" {
		pb.parts = append(pb.parts, pb.sanitizer(component))
	}
	return pb
}

// Build constructs the final path from all added components.
func (pb *PathBuilder) Build(basePath string) string {
	if len(pb.parts) == 0 {
		return basePath
	}

	allParts := append([]string{basePath}, pb.parts...)
	return filepath.Join(allParts...)
}

// Reset clears all path components, allowing the builder to be reused.
func (pb *PathBuilder) Reset() *PathBuilder {
	pb.parts = pb.parts[:0]
	return pb
}

// GetComponents returns a copy of the current path components.
func (pb *PathBuilder) GetComponents() []string {
	components := make([]string, len(pb.parts))
	copy(components, pb.parts)
	return components
}

// GetSupportedFileTypes returns a list of all supported file extensions
func GetSupportedFileTypes() []string {
	types := make([]string, 0, len(SupportedAudioExtensions)+1)

	// Add audio extensions
	for ext := range SupportedAudioExtensions {
		types = append(types, ext)
	}

	// Add EPUB
	types = append(types, ".epub")

	return types
}

// Add these functions to path.go to centralize file type checking

// IsSupportedFileForFlatMode checks if a file extension is supported in flat mode
// This includes both audio files and EPUB files
func IsSupportedFileForFlatMode(ext string) bool {
	ext = strings.ToLower(ext)
	return SupportedAudioExtensions[ext] || ext == ".epub"
}

// IsSupportedFile checks if a file extension is supported by the organizer
// This is an alias for IsSupportedFileForFlatMode for clarity
func IsSupportedFile(ext string) bool {
	return IsSupportedFileForFlatMode(ext)
}

// GetSupportedExtensions returns a map of all supported extensions for O(1) lookup
func GetSupportedExtensions() map[string]bool {
	supported := make(map[string]bool)

	// Add audio extensions
	for ext := range SupportedAudioExtensions {
		supported[ext] = true
	}

	// Add EPUB
	supported[".epub"] = true

	return supported
}

// Alternative: Update your existing SupportedAudioExtensions to be more general
// You could rename it to SupportedExtensions and include EPUB:

var SupportedExtensions = map[string]bool{
	".mp3":  true,
	".m4b":  true,
	".m4a":  true,
	".ogg":  true,
	".flac": true,
	".epub": true,
}

// ApplyFilenamePattern applies a filename pattern with metadata variables.
// Supported variables:
//   - {track}: Track number with appropriate padding (e.g., "001", "01")
//   - {title}: Title from metadata
//   - {series}: Series name from metadata
//   - {author}: First author from metadata
//   - {album}: Album from metadata (if available in RawData)
//
// Example pattern: "{track} - {title}" with metadata title "My Book" and track 5
// would produce "05 - My Book" (padding depends on totalTracks)
func ApplyFilenamePattern(pattern string, metadata Metadata, totalTracks int) string {
	result := pattern

	// Replace {track} with padded track number
	if metadata.TrackNumber > 0 {
		format := GetTrackPrefixFormat(totalTracks)
		// Remove the " - " suffix from the format for track replacement
		trackFormat := strings.TrimSuffix(format, " - ")
		trackStr := fmt.Sprintf(trackFormat, metadata.TrackNumber)
		result = strings.ReplaceAll(result, "{track}", trackStr)
	} else {
		// If no track number, remove the {track} variable
		result = strings.ReplaceAll(result, "{track}", "")
	}

	// Replace {title}
	if metadata.Title != "" {
		result = strings.ReplaceAll(result, "{title}", metadata.Title)
	}

	// Replace {series}
	if len(metadata.Series) > 0 && metadata.Series[0] != "" {
		cleanSeries := CleanSeriesName(metadata.Series[0])
		result = strings.ReplaceAll(result, "{series}", cleanSeries)
	} else {
		result = strings.ReplaceAll(result, "{series}", "")
	}

	// Replace {author}
	if len(metadata.Authors) > 0 {
		result = strings.ReplaceAll(result, "{author}", metadata.Authors[0])
	} else {
		result = strings.ReplaceAll(result, "{author}", "")
	}

	// Replace {album} if available in RawData
	if album, ok := metadata.RawData["album"].(string); ok && album != "" {
		result = strings.ReplaceAll(result, "{album}", album)
	} else {
		result = strings.ReplaceAll(result, "{album}", "")
	}

	// Clean up any extra spaces or dashes that might result from empty variables
	result = strings.TrimSpace(result)
	// Clean up patterns like " - -" or "  -  " that result from missing variables
	result = regexp.MustCompile(`\s*-\s*-\s*`).ReplaceAllString(result, " - ")
	result = regexp.MustCompile(`^\s*-\s*|\s*-\s*$`).ReplaceAllString(result, "")
	// Clean up multiple spaces
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")

	return strings.TrimSpace(result)
}
