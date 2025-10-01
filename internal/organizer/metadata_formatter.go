// internal/organizer/metadata_formatter.go
package organizer

import (
	"fmt"
	"path/filepath"
	"strings"
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

// FormatMetadataWithMapping returns a beautifully formatted string showing the metadata
func (mf *MetadataFormatter) FormatMetadataWithMapping() string {
	var sb strings.Builder

	// Get file type icon and color
	fileIcon, fileType := mf.getFileTypeDisplay()

	// Header with file type
	sb.WriteString(fmt.Sprintf("%s %s Metadata\n", fileIcon, fileType))
	sb.WriteString(strings.Repeat("─", len(fileType)+10) + "\n")

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
	// Title - clean display with just icon colored
	if mf.metadata.Title != "" {
		sb.WriteString(fmt.Sprintf("%s Title: %s\n", IconColor("📖"), mf.metadata.Title))
	}

	// Authors - clean display
	if len(mf.metadata.Authors) > 0 {
		sb.WriteString(fmt.Sprintf("%s Authors: %s\n", IconColor("👥"), strings.Join(mf.metadata.Authors, ", ")))
	}

	// Series - clean display with series index if available
	if len(mf.metadata.Series) > 0 {
		seriesName := mf.metadata.Series[0]

		// Check if we have a series index
		if seriesIndex, ok := mf.metadata.RawData["series_index"].(float64); ok && seriesIndex > 0 {
			sb.WriteString(fmt.Sprintf("%s Series: %s (#%.1f)\n", IconColor("📚"), seriesName, seriesIndex))
		} else {
			sb.WriteString(fmt.Sprintf("%s Series: %s\n", IconColor("📚"), seriesName))
		}
	}

	// Track number - clean display
	if mf.metadata.TrackNumber > 0 {
		sb.WriteString(fmt.Sprintf("%s Track: %d\n", IconColor("🔢"), mf.metadata.TrackNumber))
	}
}

func (mf *MetadataFormatter) formatAudioFields(sb *strings.Builder) {
	// Album - clean display
	if album, ok := mf.metadata.RawData["album"].(string); ok && album != "" {
		sb.WriteString(fmt.Sprintf("%s Album: %s\n", IconColor("💿"), album))
	}

	// Artist - clean display
	if artist, ok := mf.metadata.RawData["artist"].(string); ok && artist != "" {
		sb.WriteString(fmt.Sprintf("%s Artist: %s\n", IconColor("🎤"), artist))
	}

	// Album Artist - clean display
	if albumArtist, ok := mf.metadata.RawData["album_artist"].(string); ok && albumArtist != "" {
		sb.WriteString(fmt.Sprintf("%s Album Artist: %s\n", IconColor("🎭"), albumArtist))
	}

	// Composer - clean display
	if composer, ok := mf.metadata.RawData["composer"].(string); ok && composer != "" {
		sb.WriteString(fmt.Sprintf("%s Composer: %s\n", IconColor("🎼"), composer))
	}

	// Narrator - clean display
	if narrator, ok := mf.metadata.RawData["narrator"].(string); ok && narrator != "" {
		sb.WriteString(fmt.Sprintf("%s Narrator: %s\n", IconColor("🗣️"), narrator))
	}

	// Genre - clean display
	if genre, ok := mf.metadata.RawData["genre"].(string); ok && genre != "" {
		sb.WriteString(fmt.Sprintf("%s Genre: %s\n", IconColor("🎭"), genre))
	}

	// Year - clean display
	if year, ok := mf.metadata.RawData["year"].(int); ok && year > 0 {
		sb.WriteString(fmt.Sprintf("%s Year: %d\n", IconColor("📅"), year))
	}

	// Track details - clean display
	if trackTotal, ok := mf.metadata.RawData["track_total"].(int); ok && trackTotal > 0 {
		sb.WriteString(fmt.Sprintf("%s Track: %d of %d\n", IconColor("📊"), mf.metadata.TrackNumber, trackTotal))
	}

	// Comment - clean display, truncated if too long
	if comment, ok := mf.metadata.RawData["comment"].(string); ok && comment != "" {
		if len(comment) > 100 {
			comment = comment[:97] + "..."
		}
		sb.WriteString(fmt.Sprintf("%s Comment: %s\n", IconColor("💬"), comment))
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
