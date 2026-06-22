package models

import (
	"path/filepath"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// DefaultCustomLayoutTemplate is the default organize layout template used by the TUI.
const DefaultCustomLayoutTemplate = "{author}/{series|Standalone}/{Vol series_number:02 - }{book_title}{ [narrator]}"

// GenerateOutputPath generates a preview of the output path based on metadata and layout.
// This is the universal function used by both settings preview and the actual preview screen.
func GenerateOutputPath(
	book AudioBook,
	layout string,
	layoutTemplate string,
	fieldMapping organizer.FieldMapping,
	outputDir string,
) string {
	updatedMetadata := book.Metadata
	updatedMetadata.ApplyFieldMapping(fieldMapping)

	if outputDir == "" {
		outputDir = "output"
	}

	if layout == "custom" && strings.TrimSpace(layoutTemplate) != "" {
		config := &organizer.OrganizerConfig{
			BaseDir:        outputDir,
			OutputDir:      outputDir,
			LayoutTemplate: layoutTemplate,
		}
		lc := organizer.NewLayoutCalculator(config, previewPathSanitizer)
		targetDir, err := lc.CalculateTargetPathInBaseE(updatedMetadata, outputDir)
		if err == nil {
			return filepath.Join(targetDir, filepath.Base(book.Path))
		}
	}

	// Get filename for fallback
	base := filepath.Base(book.Path)
	fileTitle := strings.TrimSuffix(base, filepath.Ext(base))

	// Get metadata values with fallbacks
	author := "Unknown"
	if len(updatedMetadata.Authors) > 0 {
		author = strings.Join(updatedMetadata.Authors, ", ")
	}

	title := fileTitle
	if updatedMetadata.Title != "" {
		title = updatedMetadata.Title
	}

	series := ""
	if validSeries := updatedMetadata.GetValidSeries(); validSeries != "" {
		series = validSeries
	}

	seriesNumber := organizer.GetSeriesNumberFromMetadata(updatedMetadata)

	switch layout {
	case "author-only":
		return filepath.Join(outputDir, author, base)
	case "author-title":
		return filepath.Join(outputDir, author, title, base)
	case "author-series-title":
		if series != "" {
			return filepath.Join(outputDir, author, series, title, base)
		}
		return filepath.Join(outputDir, author, title, base)
	case "author-series-title-number":
		if series != "" {
			if seriesNumber != "" {
				numberedTitle := "#" + seriesNumber + " - " + title
				return filepath.Join(outputDir, author, series, numberedTitle, base)
			}
			return filepath.Join(outputDir, author, series, title, base)
		}
		return filepath.Join(outputDir, author, title, base)
	case "series-title":
		if series != "" {
			return filepath.Join(outputDir, series, title, base)
		}
		return filepath.Join(outputDir, title, base)
	case "series-title-number":
		if series != "" {
			if seriesNumber != "" {
				numberedTitle := "#" + seriesNumber + " - " + title
				return filepath.Join(outputDir, series, numberedTitle, base)
			}
			return filepath.Join(outputDir, series, title, base)
		}
		return filepath.Join(outputDir, title, base)
	default:
		return filepath.Join(outputDir, base)
	}
}

func previewPathSanitizer(value string) string {
	for _, char := range []string{"/", "<", ">", ":", "|", "?", "*", "`", "\""} {
		value = strings.ReplaceAll(value, char, "_")
	}
	return strings.Trim(value, " ._")
}

func truncateLayoutTemplate(template string) string {
	template = strings.TrimSpace(template)
	if len(template) <= 42 {
		return template
	}
	return template[:39] + "..."
}
