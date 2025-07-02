// internal/organizer/path.go
package organizer

import (
	"fmt"
	"path/filepath"
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

// CleanSeriesName removes trailing series numbers (e.g., " #1") from series names.
// This is now public so it can be used throughout the package.
func CleanSeriesName(series string) string {
	if idx := strings.LastIndex(series, " #"); idx != -1 {
		return strings.TrimSpace(series[:idx])
	}
	return series
}

// IsSupportedAudioFile checks if a file extension represents a supported audio format.
// Uses a map for O(1) lookup performance instead of slice iteration.
func IsSupportedAudioFile(ext string) bool {
	return SupportedAudioExtensions[strings.ToLower(ext)]
}

// AddTrackPrefix adds a track number prefix to a filename if not already present.
// Returns the original filename if track number is 0 or prefix already exists.
func AddTrackPrefix(filename string, trackNumber int) string {
	if trackNumber <= 0 {
		return filename
	}

	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	prefix := fmt.Sprintf(TrackPrefixFormat, trackNumber)
	if strings.HasPrefix(baseName, prefix) {
		return filename
	}

	return fmt.Sprintf("%s%s%s", prefix, baseName, ext)
}

// HasTrackPrefix checks if a filename already has a track number prefix.
func HasTrackPrefix(filename string) bool {
	// Look for pattern like "01 - ", "02 - ", etc.
	if len(filename) < 5 {
		return false
	}

	// Check if it starts with digits followed by " - "
	if filename[2] == ' ' && filename[3] == '-' && filename[4] == ' ' {
		first := filename[0]
		second := filename[1]
		return first >= '0' && first <= '9' && second >= '0' && second <= '9'
	}

	return false
}

// ExtractTrackNumber extracts the track number from a filename prefix.
// Returns 0 if no track number prefix is found.
func ExtractTrackNumber(filename string) int {
	if !HasTrackPrefix(filename) {
		return 0
	}

	// Extract the two-digit number from the beginning
	trackStr := filename[:2]
	var trackNum int
	if _, err := fmt.Sscanf(trackStr, "%d", &trackNum); err == nil {
		return trackNum
	}

	return 0
}

// RemoveTrackPrefix removes the track number prefix from a filename if present.
func RemoveTrackPrefix(filename string) string {
	if !HasTrackPrefix(filename) {
		return filename
	}

	// Remove the "XX - " prefix (5 characters)
	return filename[5:]
}

// NormalizeFilename provides various filename normalization options.
type FilenameNormalizer struct {
	replaceSpaces    bool
	spaceReplacement string
	addTrackPrefix   bool
	trackNumber      int
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
func (fn *FilenameNormalizer) WithTrackPrefix(trackNumber int) *FilenameNormalizer {
	fn.addTrackPrefix = true
	fn.trackNumber = trackNumber
	return fn
}

// Normalize applies all configured normalizations to the filename.
func (fn *FilenameNormalizer) Normalize(filename string) string {
	result := filename

	// Add track prefix if configured
	if fn.addTrackPrefix {
		result = AddTrackPrefix(result, fn.trackNumber)
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
