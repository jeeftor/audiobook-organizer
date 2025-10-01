package main

// Build with: go run test_embedded_metadata.go

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	"github.com/jeeftor/audiobook-organizer/internal/tui/models"
)

func main() {
	// Set up test directories
	testDataDir := "./testdata/m4b"

	// Create some test books with metadata
	books := []models.AudioBook{
		{
			Path: filepath.Join(testDataDir, "strange_audiobook_10_Epic_Saga__Adventure__Quest___Glory__John_Smith.m4b"),
			Metadata: organizer.Metadata{
				Title:   "Epic Saga: Adventure, Quest & Glory",
				Authors: []string{"John Smith"},
				Series:  []string{"Epic Saga"},
			},
		},
		{
			Path: filepath.Join(testDataDir, "strange_audiobook_14_Tales_of__ngstr_m___Caf__Chronicles_Mar_a_L_pez_Tr1.m4b"),
			Metadata: organizer.Metadata{
				Title:   "Tales of Ångström & Café Chronicles",
				Authors: []string{"María López"},
				Series:  []string{"Café Chronicles"},
				TrackNumber: 1,
			},
		},
	}

	// Test with embedded metadata enabled
	fmt.Println("=== Testing with embedded metadata ENABLED ===")
	testEmbeddedMetadata(books, true)

	// Test with embedded metadata disabled
	fmt.Println("\n=== Testing with embedded metadata DISABLED ===")
	testEmbeddedMetadata(books, false)
}

func testEmbeddedMetadata(books []models.AudioBook, useEmbeddedMetadata bool) {
	fmt.Printf("Testing with %d books\n", len(books))

	// Display book information with and without embedded metadata
	for i, book := range books {
		// Get filename for display
		base := filepath.Base(book.Path)
		fileTitle := strings.TrimSuffix(base, filepath.Ext(base))

		fmt.Printf("\nBook %d: %s\n", i+1, base)
		fmt.Printf("  Filename: %s\n", fileTitle)
		fmt.Printf("  Metadata Title: %s\n", book.Metadata.Title)
		fmt.Printf("  Authors: %v\n", book.Metadata.Authors)
		fmt.Printf("  Series: %s\n", book.Metadata.GetValidSeries())

		// Generate output paths with different layouts
		layouts := []string{"author-only", "author-title", "author-series-title"}
		for _, layout := range layouts {
			// Use the exported function to generate output path
			outputPath := models.GenerateOutputPathWithLayout(book, layout, useEmbeddedMetadata)
			fmt.Printf("  Output (%s): %s\n", layout, outputPath)
		}
	}
}
