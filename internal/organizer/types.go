// internal/organizer/types.go
package organizer

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
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

// OutputFieldMapping defines how fields map to our final fields (alias for backward compatibility)
type OutputFieldMapping = FieldMapping

// AudioFieldMapping returns field mapping for audio files
func AudioFieldMapping() OutputFieldMapping {
	return OutputFieldMapping{
		TitleField:   "title",
		SeriesField:  "album",
		AuthorFields: []string{"artist", "album_artist"},
		TrackField:   "track",
	}
}

// EpubFieldMapping returns field mapping for EPUB files
func EpubFieldMapping() OutputFieldMapping {
	return OutputFieldMapping{
		TitleField:   "title",
		SeriesField:  "series",
		AuthorFields: []string{"authors"},
		TrackField:   "track",
	}
}

// Metadata contains the essential fields we need for audiobook organization
type Metadata struct {
	// Field mapping configuration
	FieldMapping OutputFieldMapping `json:"field_mapping,omitempty"`

	// Final fields (set by providers after applying field mapping)
	Title   string   `json:"title"`   // Title (after field mapping)
	Authors []string `json:"authors"` // Authors (after field mapping)
	Series  []string `json:"series"`  // Series (after field mapping) - Changed to []string

	// Additional metadata fields
	Album       string `json:"album,omitempty"`        // Album name
	TrackTitle  string `json:"track_title,omitempty"`  // Track title (might be different from main title)
	TrackNumber int    `json:"track_number,omitempty"` // Track number

	// Raw metadata for custom field mapping
	RawMetadata map[string]interface{} `json:"raw_metadata,omitempty"`

	// Source information
	SourceType string `json:"source_type"` // "epub" or "audio"
	SourcePath string `json:"source_path"` // Path to source file
}

// Clone creates a deep copy of the Metadata object
func (m *Metadata) Clone() *Metadata {
	clone := &Metadata{
		Title:       m.Title,
		Authors:     make([]string, len(m.Authors)),
		Series:      make([]string, len(m.Series)),
		TrackNumber: m.TrackNumber,
		Album:       m.Album,
		SourceType:  m.SourceType,
		SourcePath:  m.SourcePath,
		RawMetadata: make(map[string]interface{}),
	}

	// Copy slices
	copy(clone.Authors, m.Authors)
	copy(clone.Series, m.Series)

	// Copy map
	for k, v := range m.RawMetadata {
		clone.RawMetadata[k] = v
	}

	// Don't copy field mapping - we want the raw metadata without mapping applied

	return clone
}

// MetadataSourceType defines the supported metadata source types
type MetadataSourceType string

const (
	SourceTypeAudio MetadataSourceType = "audio"
	SourceTypeEPUB  MetadataSourceType = "epub"

	DefaultSourceType = SourceTypeAudio
)

// NewMetadataWithSourceType creates a new Metadata instance
func NewMetadataWithSourceType(sourceType string) Metadata {
	metadata := Metadata{
		SourceType:  sourceType,
		RawMetadata: make(map[string]interface{}),
	}

	// Set default field mapping
	switch sourceType {
	case "audio", "mp3", "m4b", "m4a":
		metadata.SourceType = string(SourceTypeAudio)
		metadata.FieldMapping = AudioFieldMapping()
	case "epub", "book":
		metadata.SourceType = string(SourceTypeEPUB)
		metadata.FieldMapping = EpubFieldMapping()
	default:
		metadata.SourceType = string(DefaultSourceType)
		metadata.FieldMapping = AudioFieldMapping()
	}

	return metadata
}

// NewMetadata creates a new Metadata instance with default field mapping
func NewMetadata() Metadata {
	return NewMetadataWithSourceType("")
}

// ApplyFieldMapping applies the field mapping configuration to set the final fields
func (m *Metadata) ApplyFieldMapping() {
	// Apply title field mapping
	if m.FieldMapping.TitleField != "" {
		switch m.FieldMapping.TitleField {
		case "title":
			// Already set, do nothing
		case "album":
			if m.Album != "" {
				m.Title = m.Album
			}
		case "series":
			if len(m.Series) > 0 && m.Series[0] != "" {
				m.Title = m.Series[0]
			}
		case "track_title":
			if m.TrackTitle != "" {
				m.Title = m.TrackTitle
			}
		default:
			// Check raw metadata
			if val, ok := m.RawMetadata[m.FieldMapping.TitleField]; ok {
				if strVal, ok := val.(string); ok {
					m.Title = strVal
				}
			}
		}
	}

	// Apply series field mapping
	if m.FieldMapping.SeriesField != "" {
		switch m.FieldMapping.SeriesField {
		case "series":
			// Already set, do nothing
		case "album":
			if m.Album != "" {
				m.Series = []string{m.Album}
			}
		case "title":
			if m.Title != "" {
				m.Series = []string{m.Title}
			}
		default:
			// Check raw metadata
			if val, ok := m.RawMetadata[m.FieldMapping.SeriesField]; ok {
				if strVal, ok := val.(string); ok {
					m.Series = []string{strVal}
				}
			}
		}
	}

	// Apply author field mapping
	if len(m.FieldMapping.AuthorFields) > 0 {
		var allAuthors []string

		// First check if we already have authors from the default field
		if field := "authors"; contains(m.FieldMapping.AuthorFields, field) {
			if len(m.Authors) > 0 {
				allAuthors = append(allAuthors, m.Authors...)
			}
		}

		// Then check all other specified fields
		for _, field := range m.FieldMapping.AuthorFields {
			if field == "authors" {
				continue // Already handled above
			}

			// Check raw metadata
			if val, ok := m.RawMetadata[field]; ok {
				if strVal, ok := val.(string); ok && strVal != "" && !contains(allAuthors, strVal) {
					allAuthors = append(allAuthors, strVal)
				}
				if sliceVal, ok := val.([]string); ok && len(sliceVal) > 0 {
					for _, author := range sliceVal {
						if !contains(allAuthors, author) {
							allAuthors = append(allAuthors, author)
						}
					}
				}
			}
		}

		// Only update if we found additional authors
		if len(allAuthors) > 0 {
			m.Authors = allAuthors
		}
	}
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

// FormatMetadataWithMapping returns a beautifully formatted string showing the metadata
func (m *Metadata) FormatMetadataWithMapping() string {
	var sb strings.Builder

	// Get file type icon and color
	fileIcon, fileType := m.getFileTypeDisplay()

	// Header with file type
	sb.WriteString(fmt.Sprintf("%s %s Metadata\n", fileIcon, fileType))
	sb.WriteString(strings.Repeat("─", len(fileType)+10) + "\n")

	// Add filename (just the base name, not the full path)
	if m.SourcePath != "" {
		filename := filepath.Base(m.SourcePath)
		sb.WriteString(fmt.Sprintf("%s Filename: %s\n", color.CyanString("📂"), color.WhiteString(filename)))
	}

	// Core fields that every file type should have
	m.formatCoreFields(&sb)

	// File-type specific fields
	switch m.SourceType {
	case "audio":
		m.formatAudioFields(&sb)
	case "epub":
		m.formatEPUBFields(&sb)
	}

	// Show combined field mapping and values
	if !m.FieldMapping.IsEmpty() {
		sb.WriteString(m.FormatFieldMappingAndValues())
	}

	return sb.String()
}

func (m *Metadata) getFileTypeDisplay() (string, string) {
	switch m.SourceType {
	case "audio":
		ext := strings.ToLower(filepath.Ext(m.SourcePath))
		switch ext {
		case ".mp3":
			return color.GreenString("🎵"), color.GreenString("MP3 Audio")
		case ".m4b":
			return color.CyanString("🎧"), color.CyanString("M4B Audiobook")
		case ".m4a":
			return color.YellowString("🔊"), color.YellowString("M4A Audio")
		default:
			return color.MagentaString("🎶"), color.MagentaString("Audio")
		}
	case "epub":
		return color.BlueString("📚"), color.BlueString("EPUB Book")
	default:
		return color.WhiteString("📄"), color.WhiteString("Metadata")
	}
}

func (m *Metadata) formatCoreFields(sb *strings.Builder) {
	// Title
	if m.Title != "" {
		sb.WriteString(fmt.Sprintf("%s Title: %s\n", color.CyanString("📖"), color.WhiteString(m.Title)))
	}

	// Authors
	if len(m.Authors) > 0 {
		sb.WriteString(fmt.Sprintf("%s Authors: %s\n", color.CyanString("👥"), color.WhiteString(strings.Join(m.Authors, ", "))))
	}

	// Series
	if len(m.Series) > 0 {
		seriesName := m.Series[0]

		// Check if we have a series index
		if seriesIndex, ok := m.RawMetadata["series_index"].(float64); ok && seriesIndex > 0 {
			sb.WriteString(fmt.Sprintf("%s Series: %s (#%.1f)\n", color.CyanString("📚"), color.WhiteString(seriesName), seriesIndex))
		} else {
			sb.WriteString(fmt.Sprintf("%s Series: %s\n", color.CyanString("📚"), color.WhiteString(seriesName)))
		}
	}

	// Track number for audio files
	if m.TrackNumber > 0 {
		sb.WriteString(fmt.Sprintf("%s Track: %d\n", color.CyanString("🔢"), m.TrackNumber))
	}
}

func (m *Metadata) formatAudioFields(sb *strings.Builder) {
	//sb.WriteString(fmt.Sprintf("\n%s Audio Specific:\n", color.CyanString("🎵")))

	// Album
	if album, ok := m.RawMetadata["album"].(string); ok && album != "" {
		sb.WriteString(fmt.Sprintf("%s Album: %s\n", color.CyanString("💿"), color.WhiteString(album)))
	}

	// Artist vs Album Artist
	if artist, ok := m.RawMetadata["artist"].(string); ok && artist != "" {
		sb.WriteString(fmt.Sprintf("%s Artist: %s\n", color.CyanString("🎤"), color.WhiteString(artist)))
	}
	if albumArtist, ok := m.RawMetadata["album_artist"].(string); ok && albumArtist != "" {
		sb.WriteString(fmt.Sprintf("%s Album Artist: %s\n", color.CyanString("🎭"), color.WhiteString(albumArtist)))
	}

	// Composer
	if composer, ok := m.RawMetadata["composer"].(string); ok && composer != "" {
		sb.WriteString(fmt.Sprintf("%s Composer: %s\n", color.CyanString("🎼"), color.WhiteString(composer)))
	}

	// Narrator (if found)
	if narrator, ok := m.RawMetadata["narrator"].(string); ok && narrator != "" {
		sb.WriteString(fmt.Sprintf("%s Narrator: %s\n", color.CyanString("🗣️"), color.WhiteString(narrator)))
	}

	// Track details
	if trackTotal, ok := m.RawMetadata["track_total"].(int); ok && trackTotal > 0 {
		sb.WriteString(fmt.Sprintf("%s Track: %d of %d\n", color.CyanString("📊"), m.TrackNumber, trackTotal))
	}

	// Disc details
	if discNum, ok := m.RawMetadata["disc_number"].(int); ok && discNum > 0 {
		if discTotal, ok := m.RawMetadata["disc_total"].(int); ok && discTotal > 0 {
			sb.WriteString(fmt.Sprintf("%s Disc: %d of %d\n", color.CyanString("💽"), discNum, discTotal))
		} else {
			sb.WriteString(fmt.Sprintf("%s Disc: %d\n", color.CyanString("💽"), discNum))
		}
	}

	// Genre
	if genre, ok := m.RawMetadata["genre"].(string); ok && genre != "" {
		sb.WriteString(fmt.Sprintf("%s Genre: %s\n", color.CyanString("🎭"), color.WhiteString(genre)))
	}

	// Year
	if year, ok := m.RawMetadata["year"].(int); ok && year > 0 {
		sb.WriteString(fmt.Sprintf("%s Year: %d\n", color.CyanString("📅"), year))
	}

	// Content Group (might contain series info)
	if contentGroup, ok := m.RawMetadata["content_group"].(string); ok && contentGroup != "" {
		sb.WriteString(fmt.Sprintf("%s Content Group: %s\n", color.CyanString("📑"), color.WhiteString(contentGroup)))
	}

	// Track Title (if different from main title)
	if trackTitle, ok := m.RawMetadata["track_title"].(string); ok && trackTitle != "" && trackTitle != m.Title {
		sb.WriteString(fmt.Sprintf("%s Track Title: %s\n", color.CyanString("🎵"), color.WhiteString(trackTitle)))
	}

	// Comment (if present)
	if comment, ok := m.RawMetadata["comment"].(string); ok && comment != "" {
		if len(comment) > 100 {
			comment = comment[:97] + "..."
		}
		sb.WriteString(fmt.Sprintf("%s Comment: %s\n", color.CyanString("💬"), color.WhiteString(comment)))
	}
}

func (m *Metadata) formatEPUBFields(sb *strings.Builder) {
	//sb.WriteString(fmt.Sprintf("\n%s EPUB Specific:\n", color.CyanString("📚")))

	// Publisher
	if publisher, ok := m.RawMetadata["publisher"].(string); ok && publisher != "" {
		sb.WriteString(fmt.Sprintf("%s Publisher: %s\n", color.CyanString("🏢"), color.WhiteString(publisher)))
	}

	// Language
	if language, ok := m.RawMetadata["language"].(string); ok && language != "" {
		sb.WriteString(fmt.Sprintf("%s Language: %s\n", color.CyanString("🌍"), color.WhiteString(language)))
	}

	// Identifier (ISBN, etc.)
	if identifier, ok := m.RawMetadata["identifier"].(string); ok && identifier != "" {
		sb.WriteString(fmt.Sprintf("%s Identifier: %s\n", color.CyanString("🆔"), color.WhiteString(identifier)))
	}

	// Subjects/Tags
	if subjects, ok := m.RawMetadata["subjects"].([]string); ok && len(subjects) > 0 {
		if len(subjects) <= 3 {
			sb.WriteString(fmt.Sprintf("%s Subjects: %s\n", color.CyanString("🏷️"), color.WhiteString(strings.Join(subjects, ", "))))
		} else {
			sb.WriteString(fmt.Sprintf("%s Subjects: %s, ... (%d total)\n",
				color.CyanString("🏷️"), color.WhiteString(strings.Join(subjects[:3], ", ")), len(subjects)))
		}
	}
}

func (m *Metadata) formatFieldMappingInfo(sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("\n%s Field Mapping:\n", color.CyanString("🔧")))

	if m.FieldMapping.TitleField != "" && m.FieldMapping.TitleField != "title" {
		sb.WriteString(fmt.Sprintf("   %s Title from: %s\n", color.CyanString("📖"), color.WhiteString(m.FieldMapping.TitleField)))
	}

	if m.FieldMapping.SeriesField != "" && m.FieldMapping.SeriesField != "series" {
		sb.WriteString(fmt.Sprintf("   %s Series from: %s\n", color.CyanString("📚"), color.WhiteString(m.FieldMapping.SeriesField)))
	}

	if len(m.FieldMapping.AuthorFields) > 0 {
		if len(m.FieldMapping.AuthorFields) == 1 {
			sb.WriteString(fmt.Sprintf("   %s Author from: %s\n", color.CyanString("👤"), color.WhiteString(m.FieldMapping.AuthorFields[0])))
		} else if len(m.FieldMapping.AuthorFields) > 1 {
			sb.WriteString(fmt.Sprintf("   %s Authors from: %s\n", color.CyanString("👥"), color.WhiteString(strings.Join(m.FieldMapping.AuthorFields, ", "))))
		}
	}

	if m.FieldMapping.TrackField != "" && m.FieldMapping.TrackField != "track" {
		sb.WriteString(fmt.Sprintf("   %s Track from: %s\n", color.CyanString("🔢"), color.WhiteString(m.FieldMapping.TrackField)))
	}
}

func (m *Metadata) FormatFieldMappingAndValues() string {
	var sb strings.Builder

	// Define consistent colors for each component
	titleColor := color.New(color.FgHiBlue).SprintFunc()
	seriesColor := color.New(color.FgGreen).SprintFunc()
	authorColor := color.New(color.FgMagenta).SprintFunc()

	sb.WriteString(fmt.Sprintf("\n%s Field Mapping:\n", color.CyanString("🔧")))

	// Title field and value
	titleField := "title"
	if m.FieldMapping.TitleField != "" {
		titleField = m.FieldMapping.TitleField
	}
	sb.WriteString(fmt.Sprintf("   %s %s: %s → %s\n",
		color.CyanString("📖"),
		color.CyanString("Title"),
		color.YellowString(titleField),
		titleColor(m.Title)))

	// Series field and value
	seriesField := "series"
	if m.FieldMapping.SeriesField != "" {
		seriesField = m.FieldMapping.SeriesField
	}
	if len(m.Series) > 0 {
		seriesValue := strings.Join(m.Series, ", ")
		sb.WriteString(fmt.Sprintf("   %s %s: %s → %s\n",
			color.CyanString("📚"),
			color.CyanString("Series"),
			color.YellowString(seriesField),
			seriesColor(seriesValue)))
	}

	// Author fields and values
	authorFields := []string{"authors"}
	if len(m.FieldMapping.AuthorFields) > 0 {
		authorFields = m.FieldMapping.AuthorFields
	}
	if len(m.Authors) > 0 {
		authorsValue := strings.Join(m.Authors, ", ")
		sb.WriteString(fmt.Sprintf("   %s %s: %s → %s\n",
			color.CyanString("👥"),
			color.CyanString("Authors"),
			color.YellowString(strings.Join(authorFields, ", ")),
			authorColor(authorsValue)))
	}

	// Track field and value
	trackField := "track"
	if m.FieldMapping.TrackField != "" {
		trackField = m.FieldMapping.TrackField
	}
	if m.TrackNumber > 0 {
		sb.WriteString(fmt.Sprintf("   %s %s: %s → %d\n",
			color.CyanString("🔢"),
			color.CyanString("Track"),
			color.YellowString(trackField),
			m.TrackNumber))
	}

	// Add layout style display
	if layout, ok := m.RawMetadata["layout"].(string); ok && layout != "" {
		sb.WriteString(fmt.Sprintf("   %s %s: ",
			color.CyanString("📁"),
			color.CyanString("Layout")))

		// Display the layout with appropriate colors
		switch layout {
		case "author-series-title":
			sb.WriteString(fmt.Sprintf("%s/%s/%s/\n",
				authorColor("Author"),
				seriesColor("Series"),
				titleColor("Title")))
		case "author-title":
			sb.WriteString(fmt.Sprintf("%s/%s/\n",
				authorColor("Author"),
				titleColor("Title")))
		case "author-only":
			sb.WriteString(fmt.Sprintf("%s/\n",
				authorColor("Author")))
		default:
			sb.WriteString(fmt.Sprintf("%s\n", color.YellowString(layout)))
		}
	}

	return sb.String()
}

func (m *Metadata) FormatMetadata() string {
	var sb strings.Builder

	// Get file type display
	icon, typeStr := m.getFileTypeDisplay()
	sb.WriteString(fmt.Sprintf("%s %s Metadata\n", icon, typeStr))
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	// Define consistent colors for each component
	titleColor := color.New(color.FgHiBlue).SprintFunc()
	seriesColor := color.New(color.FgGreen).SprintFunc()
	authorColor := color.New(color.FgMagenta).SprintFunc()
	trackColor := color.New(color.FgYellow).SprintFunc()

	// Filename
	if m.SourcePath != "" {
		sb.WriteString(fmt.Sprintf("%s Filename: %s\n", color.CyanString("📂"), color.WhiteString(filepath.Base(m.SourcePath))))
	}

	// Title from raw metadata
	if title, ok := m.RawMetadata["title"].(string); ok && title != "" {
		sb.WriteString(fmt.Sprintf("%s Title: %s\n", color.CyanString("📖"), titleColor(title)))
	}

	// Authors from raw metadata
	if authors, ok := m.RawMetadata["authors"].([]string); ok && len(authors) > 0 {
		sb.WriteString(fmt.Sprintf("%s Authors: %s\n", color.CyanString("👥"), authorColor(strings.Join(authors, ", "))))
	}

	// Series from raw metadata
	if series, ok := m.RawMetadata["series"].([]string); ok && len(series) > 0 {
		sb.WriteString(fmt.Sprintf("%s Series: %s\n", color.CyanString("📚"), seriesColor(strings.Join(series, ", "))))
	} else if series, ok := m.RawMetadata["series"].(string); ok && series != "" {
		sb.WriteString(fmt.Sprintf("%s Series: %s\n", color.CyanString("📚"), seriesColor(series)))
	}

	// Track from raw metadata
	if track, ok := m.RawMetadata["track"].(int); ok && track > 0 {
		sb.WriteString(fmt.Sprintf("%s Track: %s\n", color.CyanString("🔢"), trackColor(fmt.Sprintf("%d", track))))
	}

	// Format audio-specific fields
	if m.SourceType == "audio" {
		// Album
		if album, ok := m.RawMetadata["album"].(string); ok && album != "" {
			sb.WriteString(fmt.Sprintf("%s Album: %s\n", color.CyanString("💿"), seriesColor(album)))
		}

		// Artist
		if artist, ok := m.RawMetadata["artist"].(string); ok && artist != "" {
			sb.WriteString(fmt.Sprintf("%s Artist: %s\n", color.CyanString("🎤"), authorColor(artist)))
		}

		// Album Artist
		if albumArtist, ok := m.RawMetadata["album_artist"].(string); ok && albumArtist != "" {
			sb.WriteString(fmt.Sprintf("%s Album Artist: %s\n", color.CyanString("🎭"), authorColor(albumArtist)))
		}

		// Composer
		if composer, ok := m.RawMetadata["composer"].(string); ok && composer != "" {
			sb.WriteString(fmt.Sprintf("%s Composer: %s\n", color.CyanString("🎼"), authorColor(composer)))
		}

		// Genre
		if genre, ok := m.RawMetadata["genre"].(string); ok && genre != "" {
			sb.WriteString(fmt.Sprintf("%s Genre: %s\n", color.CyanString("🎭"), color.WhiteString(genre)))
		}

		// Comment
		if comment, ok := m.RawMetadata["comment"].(string); ok && comment != "" {
			sb.WriteString(fmt.Sprintf("%s Comment: %s\n", color.CyanString("💬"), color.WhiteString(comment)))
		}
	}

	return sb.String()
}

// Simple getters - return already-mapped values
func (m *Metadata) GetTitle() string {
	if m.Title != "" {
		return m.Title
	}

	// Fallback to filename
	if m.SourcePath != "" {
		base := filepath.Base(m.SourcePath)
		return strings.TrimSuffix(base, filepath.Ext(base))
	}

	return "Unknown Title"
}

func (m *Metadata) GetAuthors() []string {
	if len(m.Authors) > 0 {
		return m.Authors
	}
	return []string{"Unknown"}
}

func (m *Metadata) GetSeries() []string {
	return m.Series
}

func (m *Metadata) GetFilename() string {
	if m.SourcePath != "" {
		return filepath.Base(m.SourcePath)
	}
	return m.GetTitle()
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
