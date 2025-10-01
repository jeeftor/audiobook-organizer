package tests

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	"github.com/jeeftor/audiobook-organizer/internal/tui/models"
)

// TestEmbeddedMetadataFlag tests the embedded metadata flag functionality
func TestEmbeddedMetadataFlag(t *testing.T) {
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
	t.Log("=== Testing with embedded metadata ENABLED ===")
	testEmbeddedMetadata(t, books, true)

	// Test with embedded metadata disabled
	t.Log("\n=== Testing with embedded metadata DISABLED ===")
	testEmbeddedMetadata(t, books, false)
}

func testEmbeddedMetadata(t *testing.T, books []models.AudioBook, useEmbeddedMetadata bool) {
	t.Logf("Testing with %d books\n", len(books))

	// Display book information with and without embedded metadata
	for i, book := range books {
		// Get filename for display
		base := filepath.Base(book.Path)
		fileTitle := strings.TrimSuffix(base, filepath.Ext(base))

		t.Logf("\nBook %d: %s", i+1, base)
		t.Logf("  Filename: %s", fileTitle)
		t.Logf("  Metadata Title: %s", book.Metadata.Title)
		t.Logf("  Authors: %v", book.Metadata.Authors)
		t.Logf("  Series: %s", book.Metadata.GetValidSeries())

		// Generate output paths with different layouts
		layouts := []string{"author-only", "author-title", "author-series-title"}
		for _, layout := range layouts {
			// Use the exported function to generate output path
			outputPath := models.GenerateOutputPathWithLayout(book, layout, useEmbeddedMetadata)
			t.Logf("  Output (%s): %s", layout, outputPath)

			// Verify that the output path is correct based on the embedded metadata flag
			if useEmbeddedMetadata {
				// When embedded metadata is enabled, the path should use the metadata title
				if !strings.Contains(outputPath, book.Metadata.Title) && !strings.Contains(outputPath, fileTitle) {
					t.Errorf("Expected output path to contain metadata title '%s' or filename '%s', but got '%s'",
						book.Metadata.Title, fileTitle, outputPath)
				}
			} else {
				// When embedded metadata is disabled, the path should use the filename
				if !strings.Contains(outputPath, fileTitle) {
					t.Errorf("Expected output path to contain filename '%s', but got '%s'",
						fileTitle, outputPath)
				}
			}
		}
	}
}
