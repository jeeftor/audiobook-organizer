package models

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// GenerateOutputPath generates a preview of the output path based on metadata and layout
// This is the universal function used by both settings preview and the actual preview screen
func GenerateOutputPath(book AudioBook, layout string, fieldMapping organizer.FieldMapping, outputDir string) string {
	// Apply field mapping to get updated metadata
	updatedMetadata := book.Metadata
	updatedMetadata.ApplyFieldMapping(fieldMapping)

	// Get filename for fallback
	base := filepath.Base(book.Path)
	fileTitle := strings.TrimSuffix(base, filepath.Ext(base))

	// Get metadata values with fallbacks
	author := "Unknown"
	if len(updatedMetadata.Authors) > 0 {
		// Join all authors with ", " (they're already split by ApplyFieldMapping)
		author = strings.Join(updatedMetadata.Authors, ", ")
	}

	// Get title
	title := fileTitle
	if updatedMetadata.Title != "" {
		title = updatedMetadata.Title
	}

	// Get series
	series := ""
	if validSeries := updatedMetadata.GetValidSeries(); validSeries != "" {
		series = validSeries
	}

	// Get series number if available
	seriesNumber := organizer.GetSeriesNumberFromMetadata(updatedMetadata)

	// Use default output dir if not provided
	if outputDir == "" {
		outputDir = "output"
	}

	// Generate path based on layout
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
				numberedTitle := fmt.Sprintf("#%s - %s", seriesNumber, title)
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
				numberedTitle := fmt.Sprintf("#%s - %s", seriesNumber, title)
				return filepath.Join(outputDir, series, numberedTitle, base)
			}
			return filepath.Join(outputDir, series, title, base)
		}
		return filepath.Join(outputDir, title, base)
	default:
		return filepath.Join(outputDir, base)
	}
}
