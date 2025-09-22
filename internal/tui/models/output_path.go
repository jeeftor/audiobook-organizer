package models

import (
	"path/filepath"
	"strings"
)

// generateOutputPathPreview creates a preview of what the output path might be
// based on the book's metadata and a default layout pattern
func generateOutputPathPreview(book AudioBook) string {
	// Default layout pattern: Author/Series/Title
	layout := "{author}/{series}/{title}"

	// Get metadata values
	author := "Unknown Author"
	if len(book.Metadata.Authors) > 0 {
		author = book.Metadata.Authors[0]
	}

	series := ""
	if validSeries := book.Metadata.GetValidSeries(); validSeries != "" {
		series = validSeries
	}

	// For title, prefer the filename over metadata title if it looks more specific
	// This helps when metadata title is generic (like "The Chronicles of Narnia")
	// but the filename is specific (like "The Voyage of the Dawn Treader")
	base := filepath.Base(book.Path)
	fileTitle := strings.TrimSuffix(base, filepath.Ext(base))

	title := book.Metadata.Title
	if title == "" || (fileTitle != "" && len(fileTitle) > len(title)/2) {
		// Use filename if metadata title is empty or filename looks more specific
		title = fileTitle
	}

	// Replace placeholders in layout
	outputPath := layout
	outputPath = strings.Replace(outputPath, "{author}", author, -1)

	if series != "" {
		outputPath = strings.Replace(outputPath, "{series}", series, -1)
	} else {
		// Remove the series part including the slash if no series
		outputPath = strings.Replace(outputPath, "{series}/", "", -1)
		outputPath = strings.Replace(outputPath, "/{series}", "", -1)
		outputPath = strings.Replace(outputPath, "{series}", "", -1)
	}

	outputPath = strings.Replace(outputPath, "{title}", title, -1)

	// Clean up any double slashes that might have been created
	outputPath = strings.Replace(outputPath, "//", "/", -1)

	// Add file extension from original file
	ext := filepath.Ext(book.Path)
	outputPath = outputPath + ext

	return outputPath
}

// GetOutputPath generates the actual output path for a book
// This is a more complete version that would use the actual layout settings
func GetOutputPath(book AudioBook, outputDir string, layout string) string {
	if layout == "" {
		layout = "{author}/{series}/{title}"
	}

	// Process similar to preview but with actual settings
	outputPath := generateOutputPathPreview(book)

	// Combine with output directory
	return filepath.Join(outputDir, outputPath)
}
