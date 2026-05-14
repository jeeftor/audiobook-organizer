// Package organizer provides public APIs for audiobook metadata extraction and manipulation.
package organizer

import (
	"path/filepath"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// MetadataProvider wraps the internal metadata provider for public use
type MetadataProvider struct {
	provider *organizer.UnifiedMetadataProvider
}

// NewMetadataProvider creates a new metadata provider for the given file or directory.
// useEmbeddedOnly: if true, ignore metadata.json and use only embedded metadata from audio files
func NewMetadataProvider(path string, useEmbeddedOnly bool) *MetadataProvider {
	return &MetadataProvider{
		provider: organizer.NewMetadataProvider(path, useEmbeddedOnly),
	}
}

// GetMetadata extracts metadata from the file or directory
func (p *MetadataProvider) GetMetadata() (Metadata, error) {
	return p.provider.GetMetadata()
}

// ExtractMetadata is a convenience function to extract metadata from a file or directory
func ExtractMetadata(path string, useEmbeddedOnly bool) (Metadata, error) {
	provider := NewMetadataProvider(path, useEmbeddedOnly)
	return provider.GetMetadata()
}

// ExtractMetadataWithMapping extracts metadata and applies field mapping
func ExtractMetadataWithMapping(
	path string,
	useEmbeddedOnly bool,
	mapping FieldMapping,
) (Metadata, error) {
	metadata, err := ExtractMetadata(path, useEmbeddedOnly)
	if err != nil {
		return metadata, err
	}

	if !mapping.IsEmpty() {
		metadata.ApplyFieldMapping(mapping)
	}

	return metadata, nil
}

// FindAudioFileInDirectory finds the first audio file in a directory
func FindAudioFileInDirectory(dir string) (string, error) {
	return organizer.FindAudioFileInDirectory(dir)
}

// FindEPUBInDirectory finds the first EPUB file in a directory
func FindEPUBInDirectory(dir string) (string, error) {
	return organizer.FindEPUBInDirectory(dir)
}

// DetectFileType detects the type of file (audio, epub, json, unknown)
func DetectFileType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".mp3", ".m4b", ".m4a", ".ogg", ".flac":
		return "audio"
	case ".epub":
		return "epub"
	case ".json":
		return "json"
	default:
		return "unknown"
	}
}

// GetSeriesNumberFromMetadata extracts series number from metadata
func GetSeriesNumberFromMetadata(metadata Metadata) string {
	return organizer.GetSeriesNumberFromMetadata(metadata)
}

// CleanSeriesName removes series numbers from a series name
func CleanSeriesName(seriesName string) string {
	return organizer.CleanSeriesName(seriesName)
}

// FormatAuthorName formats an author name according to the specified format
func FormatAuthorName(name string, format AuthorFormat) string {
	formatter := organizer.NewAuthorFormatter(format)
	return formatter.FormatAuthor(name)
}

// Re-export AuthorFormat enum
type AuthorFormat = organizer.AuthorFormat

const (
	AuthorFormatFirstLast = organizer.AuthorFormatFirstLast
	AuthorFormatLastFirst = organizer.AuthorFormatLastFirst
	AuthorFormatPreserve  = organizer.AuthorFormatPreserve
)

// ConvertAuthorToFirstLast converts "Last, First" → "First Last"
func ConvertAuthorToFirstLast(authorName string) string {
	return organizer.ConvertToFirstLast(authorName)
}

// ConvertAuthorToLastFirst converts "First Last" → "Last, First"
func ConvertAuthorToLastFirst(authorName string) string {
	return organizer.ConvertToLastFirst(authorName)
}

// DetectAuthorFormat determines if name is "Last, First" or "First Last"
func DetectAuthorFormat(authorName string) AuthorFormat {
	return organizer.DetectFormat(authorName)
}

// ValidateMetadata checks if metadata has the minimum required fields
func ValidateMetadata(metadata Metadata) error {
	return metadata.Validate()
}

// IsMetadataValid checks if metadata is valid (has title and authors)
func IsMetadataValid(metadata Metadata) bool {
	return metadata.IsValid()
}
