// internal/organizer/metadata_formatter.go
package organizer

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MetadataFormatter handles all metadata display and formatting
type MetadataFormatter struct {
	metadata     Metadata
	fieldMapping FieldMapping
}

// NewMetadataFormatter creates a new formatter for the given metadata
func NewMetadataFormatter(metadata Metadata, fieldMapping FieldMapping) *MetadataFormatter {
	return &MetadataFormatter{
		metadata:     metadata,
		fieldMapping: fieldMapping,
	}
}

// Source indicator constants and styles
const (
	SourceJSON     = "json"
	SourceEmbedded = "embedded"
	SourceHybrid   = "hybrid"
)

var (
	// Source indicator styles
	jsonSourceStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true) // Gold
	embeddedSourceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CED1")).Bold(true) // Turquoise
	hybridSourceStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF69B4")).Bold(true) // Hot Pink
)

// isHybridMode checks if we're in hybrid metadata mode (metadata.json + embedded)
func (mf *MetadataFormatter) isHybridMode() bool {
	// Check if RawData contains the _embedded_source marker
	if mf.metadata.RawData == nil {
		return false
	}
	_, hasEmbedded := mf.metadata.RawData["_embedded_source"]
	return hasEmbedded && mf.metadata.SourceType == "json"
}

// getFieldSource determines where a specific field comes from
// Returns "json", "embedded", or "hybrid"
func (mf *MetadataFormatter) getFieldSource(fieldName string) string {
	if !mf.isHybridMode() {
		// Not in hybrid mode - just return the source type
		if mf.metadata.SourceType == "json" {
			return SourceJSON
		}
		return SourceEmbedded
	}

	// In hybrid mode, determine based on field type
	// File-level fields come from embedded audio
	switch fieldName {
	case "track", "track_number", "disc", "disc_number":
		return SourceEmbedded
	default:
		// Book-level fields come from JSON
		return SourceJSON
	}
}

// formatSourceIndicator returns a styled source indicator for display
func (mf *MetadataFormatter) formatSourceIndicator(fieldName string) string {
	if !mf.isHybridMode() {
		// Not in hybrid mode - no indicator needed
		return ""
	}

	source := mf.getFieldSource(fieldName)
	switch source {
	case SourceJSON:
		return jsonSourceStyle.Render(" 📁")
	case SourceEmbedded:
		return embeddedSourceStyle.Render(" 🎵")
	case SourceHybrid:
		return hybridSourceStyle.Render(" 🔄")
	default:
		return ""
	}
}

// FormatMetadataWithMapping returns a beautifully formatted string showing the metadata
func (mf *MetadataFormatter) FormatMetadataWithMapping() string {
	var sb strings.Builder

	// Get file type icon and color
	fileIcon, fileType := mf.getFileTypeDisplay()

	// Header with file type
	sb.WriteString(fmt.Sprintf("%s %s Metadata\n", fileIcon, fileType))
	sb.WriteString(strings.Repeat("─", len(fileType)+10) + "\n")

	// Add hybrid mode legend if applicable
	if mf.isHybridMode() {
		sb.WriteString(fmt.Sprintf("%s Hybrid Mode: ", IconColor("ℹ️")))
		sb.WriteString(jsonSourceStyle.Render("📁 metadata.json"))
		sb.WriteString(" | ")
		sb.WriteString(embeddedSourceStyle.Render("🎵 Embedded"))
		sb.WriteString("\n")
		if embeddedSource, ok := mf.metadata.RawData["_embedded_source"].(string); ok {
			sb.WriteString(fmt.Sprintf("  (File-level data from: %s)\n", filepath.Base(embeddedSource)))
		}
		sb.WriteString("\n")
	}

	// Add filename (just the base name, not the full path) - clean display
	if mf.metadata.SourcePath != "" {
		filename := filepath.Base(mf.metadata.SourcePath)
		sb.WriteString(fmt.Sprintf("%s Filename: %s\n", IconColor("📂"), filename))
	}

	// Core fields that every file type should have
	mf.formatCoreFields(&sb)

	// File-type specific fields
	switch mf.metadata.SourceType {
	case "audio":
		mf.formatAudioFields(&sb)
	case "epub":
		mf.formatEPUBFields(&sb)
	}

	// Show combined field mapping and values
	if !mf.fieldMapping.IsEmpty() {
		sb.WriteString(mf.FormatFieldMappingAndValues())
	}

	return sb.String()
}

func (mf *MetadataFormatter) getFileTypeDisplay() (string, string) {
	switch mf.metadata.SourceType {
	case "audio":
		ext := strings.ToLower(filepath.Ext(mf.metadata.SourcePath))
		switch ext {
		case ".mp3":
			return IconColor("🎵"), IconColor("MP3 Audio")
		case ".m4b":
			return IconColor("🎧"), IconColor("M4B Audiobook")
		case ".m4a":
			return IconColor("🔊"), IconColor("M4A Audio")
		case ".flac":
			return IconColor("🎶"), IconColor("FLAC Audio")
		case "":
			return IconColor("❓"), IconColor("UNKNOWN")
		default:
			return IconColor("❓"), IconColor("UNKNOWN")
		}
	case "epub":
		return IconColor("📚"), IconColor("EPUB Book")
	default:
		return IconColor("📄"), IconColor("Metadata")
	}
}

func (mf *MetadataFormatter) formatCoreFields(sb *strings.Builder) {
	// Title - with source indicator
	if mf.metadata.Title != "" {
		sourceIndicator := mf.formatSourceIndicator("title")
		sb.WriteString(fmt.Sprintf("%s Title: %s%s\n", IconColor("📖"), mf.metadata.Title, sourceIndicator))
	}

	// Authors - with source indicator
	if len(mf.metadata.Authors) > 0 {
		sourceIndicator := mf.formatSourceIndicator("authors")
		sb.WriteString(fmt.Sprintf("%s Authors: %s%s\n", IconColor("👥"), strings.Join(mf.metadata.Authors, ", "), sourceIndicator))
	}

	// Series - with source indicator and series index if available
	if len(mf.metadata.Series) > 0 {
		seriesName := mf.metadata.Series[0]
		sourceIndicator := mf.formatSourceIndicator("series")

		// Check if we have a series index
		if seriesIndex, ok := mf.metadata.RawData["series_index"].(float64); ok && seriesIndex > 0 {
			sb.WriteString(fmt.Sprintf("%s Series: %s (#%.1f)%s\n", IconColor("📚"), seriesName, seriesIndex, sourceIndicator))
		} else {
			sb.WriteString(fmt.Sprintf("%s Series: %s%s\n", IconColor("📚"), seriesName, sourceIndicator))
		}
	}

	// Track number - with source indicator (this comes from embedded in hybrid mode)
	if mf.metadata.TrackNumber > 0 {
		sourceIndicator := mf.formatSourceIndicator("track_number")
		sb.WriteString(fmt.Sprintf("%s Track: %d%s\n", IconColor("🔢"), mf.metadata.TrackNumber, sourceIndicator))
	}

	// Disc number - with source indicator (if present)
	if discNum := getDiscNumberFromRaw(mf.metadata.RawData); discNum > 0 {
		sourceIndicator := mf.formatSourceIndicator("disc_number")
		sb.WriteString(fmt.Sprintf("%s Disc: %d%s\n", IconColor("💿"), discNum, sourceIndicator))
	}
}

func (mf *MetadataFormatter) formatAudioFields(sb *strings.Builder) {
	// Album - with source indicator
	if album, ok := mf.metadata.RawData["album"].(string); ok && album != "" {
		sourceIndicator := mf.formatSourceIndicator("album")
		sb.WriteString(fmt.Sprintf("%s Album: %s%s\n", IconColor("💿"), album, sourceIndicator))
	}

	// Artist - with source indicator
	if artist, ok := mf.metadata.RawData["artist"].(string); ok && artist != "" {
		sourceIndicator := mf.formatSourceIndicator("artist")
		sb.WriteString(fmt.Sprintf("%s Artist: %s%s\n", IconColor("🎤"), artist, sourceIndicator))
	}

	// Album Artist - with source indicator
	if albumArtist, ok := mf.metadata.RawData["album_artist"].(string); ok && albumArtist != "" {
		sourceIndicator := mf.formatSourceIndicator("album_artist")
		sb.WriteString(fmt.Sprintf("%s Album Artist: %s%s\n", IconColor("🎭"), albumArtist, sourceIndicator))
	}

	// Composer - with source indicator
	if composer, ok := mf.metadata.RawData["composer"].(string); ok && composer != "" {
		sourceIndicator := mf.formatSourceIndicator("composer")
		sb.WriteString(fmt.Sprintf("%s Composer: %s%s\n", IconColor("🎼"), composer, sourceIndicator))
	}

	// Narrator - with source indicator
	if narrator, ok := mf.metadata.RawData["narrator"].(string); ok && narrator != "" {
		sourceIndicator := mf.formatSourceIndicator("narrator")
		sb.WriteString(fmt.Sprintf("%s Narrator: %s%s\n", IconColor("🗣️"), narrator, sourceIndicator))
	}

	// Genre - with source indicator
	if genre, ok := mf.metadata.RawData["genre"].(string); ok && genre != "" {
		sourceIndicator := mf.formatSourceIndicator("genre")
		sb.WriteString(fmt.Sprintf("%s Genre: %s%s\n", IconColor("🎭"), genre, sourceIndicator))
	}

	// Year - with source indicator
	if year, ok := mf.metadata.RawData["year"].(int); ok && year > 0 {
		sourceIndicator := mf.formatSourceIndicator("year")
		sb.WriteString(fmt.Sprintf("%s Year: %d%s\n", IconColor("📅"), year, sourceIndicator))
	}

	// Track details - with source indicator
	if trackTotal, ok := mf.metadata.RawData["track_total"].(int); ok && trackTotal > 0 {
		sourceIndicator := mf.formatSourceIndicator("track_total")
		sb.WriteString(fmt.Sprintf("%s Track: %d of %d%s\n", IconColor("📊"), mf.metadata.TrackNumber, trackTotal, sourceIndicator))
	}

	// Comment - with source indicator, truncated if too long
	if comment, ok := mf.metadata.RawData["comment"].(string); ok && comment != "" {
		sourceIndicator := mf.formatSourceIndicator("comment")
		if len(comment) > 100 {
			comment = comment[:97] + "..."
		}
		sb.WriteString(fmt.Sprintf("%s Comment: %s%s\n", IconColor("💬"), comment, sourceIndicator))
	}
}

func (mf *MetadataFormatter) formatEPUBFields(sb *strings.Builder) {
	// Publisher - clean display
	if publisher, ok := mf.metadata.RawData["publisher"].(string); ok && publisher != "" {
		sb.WriteString(fmt.Sprintf("%s Publisher: %s\n", IconColor("🏢"), publisher))
	}

	// Language - clean display
	if language, ok := mf.metadata.RawData["language"].(string); ok && language != "" {
		sb.WriteString(fmt.Sprintf("%s Language: %s\n", IconColor("🌍"), language))
	}

	// Identifier - clean display
	if identifier, ok := mf.metadata.RawData["identifier"].(string); ok && identifier != "" {
		sb.WriteString(fmt.Sprintf("%s Identifier: %s\n", IconColor("🆔"), identifier))
	}

	// Subjects/Tags - clean display
	if subjects, ok := mf.metadata.RawData["subjects"].([]string); ok && len(subjects) > 0 {
		if len(subjects) <= 3 {
			sb.WriteString(fmt.Sprintf("%s Subjects: %s\n", IconColor("🏷️"), strings.Join(subjects, ", ")))
		} else {
			sb.WriteString(fmt.Sprintf("%s Subjects: %s, ... (%d total)\n",
				IconColor("🏷️"), strings.Join(subjects[:3], ", "), len(subjects)))
		}
	}
}

func (mf *MetadataFormatter) FormatFieldMappingAndValues() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n%s Field Mapping:\n", IconColor("🔧")))

	// Define consistent colors for each component with background colors
	authorColor := AuthorColor
	seriesColor := SeriesColor
	titleColor := TitleColor
	trackColor := TrackNumberColor

	// Title field and value
	titleField := "title"
	if mf.fieldMapping.TitleField != "" {
		titleField = mf.fieldMapping.TitleField
	}
	sb.WriteString(fmt.Sprintf("   %s %s: %s → %s\n",
		IconColor("📖"),
		IconColor("Title"),
		FieldNameColor(titleField),
		titleColor(mf.metadata.Title)))

	// Series field and value
	seriesField := "series"
	if mf.fieldMapping.SeriesField != "" {
		seriesField = mf.fieldMapping.SeriesField
	}
	if len(mf.metadata.Series) > 0 {
		seriesValue := strings.Join(mf.metadata.Series, ", ")
		sb.WriteString(fmt.Sprintf("   %s %s: %s → %s\n",
			IconColor("📚"),
			IconColor("Series"),
			FieldNameColor(seriesField),
			seriesColor(seriesValue)))
	}

	// Author fields and values
	authorFields := []string{"authors"}
	if len(mf.fieldMapping.AuthorFields) > 0 {
		authorFields = mf.fieldMapping.AuthorFields
	}
	if len(mf.metadata.Authors) > 0 {
		authorsValue := strings.Join(mf.metadata.Authors, ", ")
		sb.WriteString(fmt.Sprintf("   %s %s: %s → %s\n",
			IconColor("👥"),
			IconColor("Authors"),
			FieldNameColor(strings.Join(authorFields, ", ")),
			authorColor(authorsValue)))
	}

	// Track field and value
	trackField := "track"
	if mf.fieldMapping.TrackField != "" {
		trackField = mf.fieldMapping.TrackField
	}
	if mf.metadata.TrackNumber > 0 {
		sb.WriteString(fmt.Sprintf("   %s %s: %s → %s\n",
			IconColor("🔢"),
			IconColor("Track"),
			FieldNameColor(trackField),
			trackColor(fmt.Sprintf("%d", mf.metadata.TrackNumber))))
	}

	return sb.String()
}

// FormatMetadata returns a simple formatted string showing the metadata
func (mf *MetadataFormatter) FormatMetadata() string {
	var sb strings.Builder

	// Get file type display
	icon, typeStr := mf.getFileTypeDisplay()
	sb.WriteString(fmt.Sprintf("%s %s Metadata\n", icon, typeStr))
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	// Filename - clean display
	if mf.metadata.SourcePath != "" {
		sb.WriteString(fmt.Sprintf("%s Filename: %s\n", IconColor("📂"), filepath.Base(mf.metadata.SourcePath)))
	}

	// Core fields
	if mf.metadata.Title != "" {
		sb.WriteString(fmt.Sprintf("%s Title: %s\n", IconColor("📖"), mf.metadata.Title))
	}

	if len(mf.metadata.Authors) > 0 {
		sb.WriteString(fmt.Sprintf("%s Authors: %s\n", IconColor("👥"), strings.Join(mf.metadata.Authors, ", ")))
	}

	if len(mf.metadata.Series) > 0 {
		sb.WriteString(fmt.Sprintf("%s Series: %s\n", IconColor("📚"), strings.Join(mf.metadata.Series, ", ")))
	}

	if mf.metadata.TrackNumber > 0 {
		sb.WriteString(fmt.Sprintf("%s Track: %d\n", IconColor("🔢"), mf.metadata.TrackNumber))
	}

	return sb.String()
}
