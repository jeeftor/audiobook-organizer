package main

import (
	"fmt"
	"os"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	"github.com/jeeftor/audiobook-organizer/internal/tui/models"
)

// This is a simple test program to verify the field mapping integration
func main() {
	// Create test books
	books := []models.AudioBook{
		{
			Path: "./testdata/m4b/strange_audiobook_10_Epic_Saga__Adventure__Quest___Glory__John_Smith.m4b",
			Metadata: organizer.Metadata{
				Title:   "Epic Saga: Adventure, Quest & Glory",
				Authors: []string{"John Smith"},
				RawData: map[string]interface{}{
					"title":        "Epic Saga: Adventure, Quest & Glory",
					"authors":      []string{"John Smith"},
					"album":        "Epic Series",
					"artist":       "John Smith",
					"album_artist": "John Smith",
					"series":       "Epic Series",
					"track":        1,
				},
			},
		},
	}

	// Create default field mapping
	defaultMapping := organizer.FieldMapping{
		TitleField:   "title",
		SeriesField:  "series",
		AuthorFields: []string{"authors", "artist", "album_artist"},
		TrackField:   "track",
	}

	// Create custom field mapping (using album as series)
	customMapping := organizer.FieldMapping{
		TitleField:   "title",
		SeriesField:  "album",
		AuthorFields: []string{"artist"},
		TrackField:   "track",
	}

	// Create config
	config := map[string]string{
		"Layout":              "author-series-title",
		"Output Directory":    "./test_output",
		"Use Embedded Metadata": "Yes",
		"Flat Mode":           "No",
		"Dry Run":             "Yes",
		"Verbose":             "Yes",
	}

	// Test with default mapping
	fmt.Println("Testing with default field mapping:")
	previewModel := models.NewPreviewModel(books, config, defaultMapping)
	for _, move := range previewModel.GetMoves() {
		fmt.Printf("Source: %s\nTarget: %s\n\n", move.SourcePath, move.TargetPath)
	}

	// Test with custom mapping
	fmt.Println("Testing with custom field mapping (using album as series):")
	previewModel = models.NewPreviewModel(books, config, customMapping)
	for _, move := range previewModel.GetMoves() {
		fmt.Printf("Source: %s\nTarget: %s\n\n", move.SourcePath, move.TargetPath)
	}

	// Create output directory if it doesn't exist
	outputDir := "./test_output"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}

	// Create process model with default mapping
	fmt.Println("Creating process model with default mapping:")
	_ = models.NewProcessModel(books, config, previewModel.GetMoves(), defaultMapping)
	fmt.Printf("Process model created with %d items\n", len(previewModel.GetMoves()))
	fmt.Printf("Field mapping: title=%s, series=%s, authors=%v\n\n",
		defaultMapping.TitleField, defaultMapping.SeriesField, defaultMapping.AuthorFields)

	// Create process model with custom mapping
	fmt.Println("Creating process model with custom mapping:")
	_ = models.NewProcessModel(books, config, previewModel.GetMoves(), customMapping)
	fmt.Printf("Process model created with %d items\n", len(previewModel.GetMoves()))
	fmt.Printf("Field mapping: title=%s, series=%s, authors=%v\n",
		customMapping.TitleField, customMapping.SeriesField, customMapping.AuthorFields)
}
