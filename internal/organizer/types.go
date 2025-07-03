// internal/organizer/types.go
package organizer

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FieldMapping defines how fields map to our final fields
type FieldMapping struct {
	TitleField   string   `json:"title_field,omitempty"`   // "title", "album", "series"
	SeriesField  string   `json:"series_field,omitempty"`  // "series", "album"
	AuthorFields []string `json:"author_fields,omitempty"` // ["artist", "album_artist"] or ["authors"]
	TrackField   string   `json:"track_field,omitempty"`   // "track", "track_number"
}

// IsEmpty returns true if the field mapping is empty
func (fm FieldMapping) IsEmpty() bool {
	return fm.TitleField == "" && fm.SeriesField == "" && len(fm.AuthorFields) == 0 && fm.TrackField == ""
}

// DefaultFieldMapping returns the default field mapping
func DefaultFieldMapping() FieldMapping {
	return FieldMapping{
		TitleField:   "title",
		SeriesField:  "series",
		AuthorFields: []string{"authors"},
		TrackField:   "track",
	}
}

// AudioFieldMapping returns field mapping for audio files
func AudioFieldMapping() FieldMapping {
	return FieldMapping{
		TitleField:   "title",
		SeriesField:  "album",
		AuthorFields: []string{"artist", "album_artist"},
		TrackField:   "track",
	}
}

// EpubFieldMapping returns field mapping for EPUB files
func EpubFieldMapping() FieldMapping {
	return FieldMapping{
		TitleField:   "title",
		SeriesField:  "series",
		AuthorFields: []string{"authors"},
		TrackField:   "track",
	}
}

// Metadata contains the essential fields we need for audiobook organization
type Metadata struct {
	// Core identification fields
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Series      []string `json:"series"`
	TrackNumber int      `json:"track_number,omitempty"`

	// Additional core fields
	Album      string `json:"album,omitempty"`
	TrackTitle string `json:"track_title,omitempty"`

	// Source information
	SourceType string `json:"source_type"` // "epub", "audio", "json"
	SourcePath string `json:"source_path"`

	// Raw data from the source for field mapping and advanced use
	RawData map[string]interface{} `json:"raw_data,omitempty"`

	// Field mapping configuration (moved from embedded to separate processor)
	fieldMapping FieldMapping
}

// NewMetadata creates a new Metadata instance
func NewMetadata() Metadata {
	return Metadata{
		RawData: make(map[string]interface{}),
	}
}

// GetFirstAuthor returns the first author or a default value if no authors exist
func (m *Metadata) GetFirstAuthor(defaultValue string) string {
	if len(m.Authors) > 0 && m.Authors[0] != "" {
		return m.Authors[0]
	}
	return defaultValue
}

// GetValidSeries returns the first valid series name, cleaning it of series numbers
func (m *Metadata) GetValidSeries() string {
	if len(m.Series) > 0 && m.Series[0] != "" && m.Series[0] != InvalidSeriesValue {
		return CleanSeriesName(m.Series[0])
	}
	return ""
}

// IsValid checks if metadata contains the minimum required fields
func (m *Metadata) IsValid() bool {
	return m.Title != "" && len(m.Authors) > 0 && m.Authors[0] != ""
}

// Validate ensures that essential metadata fields (title and authors) are present
func (m *Metadata) Validate() error {
	if len(m.Authors) == 0 || m.Authors[0] == "" {
		return fmt.Errorf("missing author information")
	}

	if m.Title == "" {
		return fmt.Errorf("missing title information")
	}

	return nil
}

// ApplyFieldMapping applies the field mapping configuration to set the final fields
func (m *Metadata) ApplyFieldMapping(mapping FieldMapping) {
	m.fieldMapping = mapping

	// Store original title for potential use in series mapping
	originalTitle := m.Title

	// Apply title field mapping
	if mapping.TitleField != "" {
		switch mapping.TitleField {
		case "title":
			// Keep original title
		case "series":
			if len(m.Series) > 0 {
				m.Title = m.Series[0]
			}
		case "album":
			if m.Album != "" {
				m.Title = m.Album
			}
		case "track_title":
			if m.TrackTitle != "" {
				m.Title = m.TrackTitle
			}
		default:
			if val := m.getRawValue(mapping.TitleField); val != "" {
				m.Title = val
			}
		}
	}

	// Apply series field mapping
	if mapping.SeriesField != "" {
		switch mapping.SeriesField {
		case "series":
			// Keep original series
		case "title":
			if originalTitle != "" {
				m.Series = []string{originalTitle}
			}
		default:
			if val := m.getRawValue(mapping.SeriesField); val != "" {
				m.Series = []string{val}
			}
		}
	}

	// Apply author field mapping
	if len(mapping.AuthorFields) > 0 {
		var allAuthors []string
		for _, field := range mapping.AuthorFields {
			if val := m.getRawValue(field); val != "" {
				// Split authors by common delimiters if needed
				authors := splitAuthors(val)
				for _, author := range authors {
					if !contains(allAuthors, author) {
						allAuthors = append(allAuthors, author)
					}
				}
			}
		}
		if len(allAuthors) > 0 {
			m.Authors = allAuthors
		}
	}

	// Apply track field mapping
	if mapping.TrackField != "" {
		switch mapping.TrackField {
		case "track":
			// Keep original track number
		default:
			if val, ok := m.RawData[mapping.TrackField]; ok {
				switch v := val.(type) {
				case int:
					m.TrackNumber = v
				case float64:
					m.TrackNumber = int(v)
				case string:
					// Try to parse string as int
					if num, err := strconv.Atoi(v); err == nil {
						m.TrackNumber = num
					}
				}
			}
		}
	}
}

// FormatFieldMappingAndValues returns a formatted string showing the current field mapping and values
func (m *Metadata) FormatFieldMappingAndValues() string {
	var sb strings.Builder

	sb.WriteString("Field Mappings:\n")
	sb.WriteString(fmt.Sprintf("  Title Field: %s\n", m.fieldMapping.TitleField))
	sb.WriteString(fmt.Sprintf("  Series Field: %s\n", m.fieldMapping.SeriesField))
	sb.WriteString(fmt.Sprintf("  Author Fields: %v\n", m.fieldMapping.AuthorFields))
	sb.WriteString(fmt.Sprintf("  Track Field: %s\n", m.fieldMapping.TrackField))

	sb.WriteString("\nCurrent Values:\n")
	sb.WriteString(fmt.Sprintf("  Title: %s\n", m.Title))
	if len(m.Series) > 0 {
		sb.WriteString(fmt.Sprintf("  Series: %v\n", m.Series))
	}
	if len(m.Authors) > 0 {
		sb.WriteString(fmt.Sprintf("  Authors: %v\n", m.Authors))
	}
	if m.TrackNumber > 0 {
		sb.WriteString(fmt.Sprintf("  Track Number: %d\n", m.TrackNumber))
	}

	return sb.String()
}

// getRawValue safely extracts string values from raw data
func (m *Metadata) getRawValue(field string) string {
	if val, ok := m.RawData[field]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
		// Handle []string for authors
		if sliceVal, ok := val.([]string); ok && len(sliceVal) > 0 {
			return strings.Join(sliceVal, ", ")
		}
	}

	// Handle special cases for built-in fields
	switch field {
	case "title":
		return m.Title
	case "album":
		return m.Album
	case "series":
		if len(m.Series) > 0 {
			return m.Series[0]
		}
	case "authors":
		return strings.Join(m.Authors, ", ")
	}

	return ""
}

// splitAuthors splits a string containing multiple authors into a slice of individual authors
// It handles common delimiters like semicolons, commas, and slashes
func splitAuthors(authorsStr string) []string {
	// Common delimiters: semicolon, comma, slash, newline, or multiple spaces
	delimiters := []string{";", ",", "/", "\n", "  "}

	// Replace all delimiters with a consistent delimiter
	replaced := authorsStr
	for _, delim := range delimiters[1:] {
		replaced = strings.ReplaceAll(replaced, delim, delimiters[0])
	}

	// Split by the first delimiter
	authors := strings.Split(replaced, delimiters[0])

	// Clean up each author name
	for i, author := range authors {
		authors[i] = strings.TrimSpace(author)
	}

	// Remove any empty strings
	var result []string
	for _, author := range authors {
		if author != "" {
			result = append(result, author)
		}
	}

	return result
}

// Helper function to check if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Support types
type LogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	SourcePath string    `json:"source_path"`
	TargetPath string    `json:"target_path"`
	Files      []string  `json:"files"`
}

type Summary struct {
	MetadataFound    []string
	MetadataMissing  []string
	Moves            []MoveSummary
	EmptyDirsRemoved []string
}

type MoveSummary struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type MetadataProvider interface {
	GetMetadata() (Metadata, error)
}

// SetVerboseMode is a placeholder function for setting verbose mode in metadata providers
func SetVerboseMode(verbose bool) {
	// This can be implemented later if needed
}
