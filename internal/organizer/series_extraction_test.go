//go:build !integration

package organizer

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// TestExtractCalibreSeriesFromOPF tests the extraction of series information from both
// EPUB3 standard and Calibre-specific metadata formats
func TestExtractCalibreSeriesFromOPF(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test-series-extraction-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases for different metadata formats
	testCases := []struct {
		name     string
		content  string
		expected string
		index    float64
		found    bool
	}{
		{
			name: "EPUB3 Standard Format",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <meta property="belongs-to-collection" id="series-id">Test Series EPUB3</meta>
    <meta refines="#series-id" property="group-position">2.5</meta>
  </metadata>
</package>`,
			expected: "Test Series EPUB3",
			index:    2.5,
			found:    true,
		},
		{
			name: "Calibre Format",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <meta name="calibre:series" content="Test Series Calibre"/>
    <meta name="calibre:series_index" content="3.0"/>
  </metadata>
</package>`,
			expected: "Test Series Calibre",
			index:    3.0,
			found:    true,
		},
		{
			name: "No Series Metadata",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Book Without Series</dc:title>
  </metadata>
</package>`,
			expected: "",
			index:    0,
			found:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary EPUB file
			epubPath := filepath.Join(tempDir, tc.name+".epub")

			// Create a zip file
			zipFile, err := os.Create(epubPath)
			if err != nil {
				t.Fatalf("Failed to create zip file: %v", err)
			}

			// Create a zip writer
			zipWriter := zip.NewWriter(zipFile)

			// Add content.opf file to the zip
			opfWriter, err := zipWriter.Create("content.opf")
			if err != nil {
				zipWriter.Close()
				zipFile.Close()
				t.Fatalf("Failed to create opf entry: %v", err)
			}

			// Write the OPF content
			_, err = opfWriter.Write([]byte(tc.content))
			if err != nil {
				zipWriter.Close()
				zipFile.Close()
				t.Fatalf("Failed to write opf content: %v", err)
			}

			// Close the zip writer and file
			zipWriter.Close()
			zipFile.Close()

			// Call the function with the path to our test EPUB
			series, seriesIndex, found := ExtractCalibreSeriesFromOPF(epubPath)

			// Verify the results
			if found != tc.found {
				t.Errorf("Expected found=%v, got %v", tc.found, found)
			}

			if found {
				if series != tc.expected {
					t.Errorf("Expected series name %q, got %q", tc.expected, series)
				}

				if seriesIndex != tc.index {
					t.Errorf("Expected series index %.1f, got %.1f", tc.index, seriesIndex)
				}
			}
		})
	}
}
