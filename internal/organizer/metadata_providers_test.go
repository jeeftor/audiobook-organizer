//go:build !integration

package organizer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

//)
//
//func TestExtractCalibreSeriesFromOPF(t *testing.T) {
//	// Create a temporary directory for test files
//	tempDir, err := os.MkdirTemp("", "test-calibre-series-*")
//	require.NoError(t, err, "Failed to create temp directory")
//	defer os.RemoveAll(tempDir)
//
//	tests := []struct {
//		name        string
//		content     string
//		expected    string
//		seriesIndex float64
//		hasSeries   bool
//	}{
//		{
//			name: "valid calibre series",
//			content: `<?xml version="1.0" encoding="UTF-8"?>
//<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
//  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
//    <meta name="calibre:series" content="Test Series"/>
//    <meta name="calibre:series_index" content="1.0"/>
//  </metadata>
//</package>`,
//			expected:    "Test Series",
//			seriesIndex: 1.0,
//			hasSeries:   true,
//		},
//		{
//			name:     "no series info",
//			content:  `<?xml version="1.0" encoding="UTF-8"?><package></package>`,
//			expected: "",
//			hasSeries: false,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// Create a temporary EPUB file with the test content
//			epubPath := filepath.Join(tempDir, tt.name+".epub")
//			err := os.MkdirAll(epubPath, 0755)
//			require.NoError(t, err, "Failed to create test EPUB directory")
//
//			// Create a minimal OPF file
//			opfPath := filepath.Join(epubPath, "content.opf")
//			err = os.WriteFile(opfPath, []byte(tt.content), 0644)
//			require.NoError(t, err, "Failed to write test OPF file")
//
//			// Call the function with the path to our test EPUB
//			series, seriesIndex, found := extractCalibreSeriesFromOPF(epubPath)
//
//			if tt.hasSeries {
//				assert.True(t, found, "Expected series to be found for test case: %s", tt.name)
//				assert.Equal(t, tt.expected, series, "Series name mismatch for test case: %s", tt.name)
//				assert.Equal(t, tt.seriesIndex, seriesIndex, "Series index mismatch for test case: %s", tt.name)
//			} else {
//				assert.False(t, found, "Expected no series for test case: %s", tt.name)
//			}
//		})
//	}
//}

func TestEPUBMetadataExtraction(t *testing.T) {
	// Skip this test if books directory doesn't exist
	testDataDir := filepath.Join("..", "..", "testdata", "epub")
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skipf("Skipping test: test data directory %s does not exist", testDataDir)
	}

	tests := []struct {
		name           string
		filename       string
		expectedTitle  string
		expectedSeries string
		hasSeries      bool
		shouldSkip     bool
	}{
		//{
		//	name:          "valid book with series",
		//	filename:      "valid-book.epub",
		//	expectedTitle:  "Test Book",
		//	expectedSeries: "Test Series",
		//	hasSeries:      true,
		//	shouldSkip:     false,
		//},
		{
			name:           "book without series",
			filename:       "title-author.epub",
			expectedTitle:  "The book of cool stuff",
			expectedSeries: "",
			hasSeries:      false,
			shouldSkip:     false,
		},
		{
			name:           "book with series 1",
			filename:       "title-author-series1.epub",
			expectedTitle:  "First book of testing knowledge",
			expectedSeries: "Test Books",
			hasSeries:      true,
			shouldSkip:     false,
		},
		{
			name:           "book with series 2",
			filename:       "title-author-series2.epub",
			expectedTitle:  "Testing is dumb",
			expectedSeries: "Test Books",
			hasSeries:      true,
			shouldSkip:     false,
		},
		{
			name:           "book with series 3",
			filename:       "title-author-series3.epub",
			expectedTitle:  "Why is everything broken",
			expectedSeries: "Test Books",
			hasSeries:      true,
			shouldSkip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldSkip {
				t.Skip("Test case marked to be skipped")
			}

			// Create the EPUB metadata provider
			epubPath := filepath.Join(testDataDir, tt.filename)
			provider := NewEPUBMetadataProvider(epubPath)

			// Get the metadata
			metadata, err := provider.GetMetadata()
			if err != nil {
				t.Fatalf("Failed to get metadata for %s: %v", tt.filename, err)
			}

			// Verify the title
			if metadata.Title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, metadata.Title)
			}

			// Verify series metadata
			if tt.hasSeries {
				if len(metadata.Series) == 0 {
					t.Errorf("Expected series %q, but no series found", tt.expectedSeries)
				} else if metadata.Series[0] != tt.expectedSeries {
					t.Errorf("Expected series %q, got %q", tt.expectedSeries, metadata.Series[0])
				}
			} else if len(metadata.Series) > 0 {
				t.Errorf("Expected no series, but found %q", metadata.Series[0])
			}
		})
	}
}

func TestEPUBMetadataWithProblematicFiles(t *testing.T) {
	t.Skip("Skipping test that requires specific test data files")
}

func TestMP3MetadataWithProblematicFiles(t *testing.T) {
	testDataDir := "../../testdata/mp3"

	cwd, _ := os.Getwd()
	t.Logf("Current working directory: %s", cwd)

	dirEntries, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", testDataDir, err)
	}

	var found bool
	for _, entry := range dirEntries {
		if entry.Type().IsRegular() && filepath.Ext(entry.Name()) == ".mp3" {
			found = true
			filename := entry.Name()
			filePath := filepath.Join(testDataDir, filename)
			t.Run(filename, func(t *testing.T) {
				provider := NewFileMetadataProvider(filePath)
				metadata, err := provider.GetMetadata()
				if err != nil {
					t.Fatalf("Failed to get metadata for %s: %v", filename, err)
				}
				t.Logf("File: %s\nMetadata: %+v", filename, metadata)
			})
		}
	}
	if !found {
		t.Fatalf("No mp3 files found in %s (cwd: %s)", testDataDir, cwd)
	}
}

func TestM4BMetadataWithProblematicFiles(t *testing.T) {
	testDataDir := "../../testdata/m4b"

	cwd, _ := os.Getwd()
	t.Logf("Current working directory: %s", cwd)

	dirEntries, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", testDataDir, err)
	}

	var found bool
	for _, entry := range dirEntries {
		if entry.Type().IsRegular() && filepath.Ext(entry.Name()) == ".m4b" {
			found = true
			filename := entry.Name()
			filePath := filepath.Join(testDataDir, filename)
			t.Run(filename, func(t *testing.T) {
				provider := NewFileMetadataProvider(filePath)
				metadata, err := provider.GetMetadata()
				if err != nil {
					t.Fatalf("Failed to get metadata for %s: %v", filename, err)
				}
				t.Logf("File: %s\nMetadata: %+v", filename, metadata)
			})
		}
	}
	if !found {
		t.Fatalf("No m4b files found in %s (cwd: %s)", testDataDir, cwd)
	}
}

// verifyPathSanitization checks that a path doesn't contain invalid characters
func verifyPathSanitization(t *testing.T, path string, replaceSpace string) {
	// Extract the sanitized part of the path (after epub/)
	parts := strings.Split(path, "epub/")
	if len(parts) < 2 {
		t.Errorf("Path does not contain 'epub/': %s", path)
		return
	}
	sanitizedPath := parts[1]

	// Split the path into components (author/series/title)
	pathComponents := strings.Split(sanitizedPath, "/")

	// Check each component for invalid characters
	for _, component := range pathComponents {
		// Check for invalid characters based on OS
		var invalidChars []string
		if runtime.GOOS == "windows" {
			invalidChars = []string{"<", ">", ":", "\"", "\\", "|", "?", "*"}
		} else {
			invalidChars = []string{"/"}
		}

		// Check for each invalid character
		for _, char := range invalidChars {
			if strings.Contains(component, char) {
				t.Errorf("Path component %q contains invalid character %q", component, char)
			}
		}

		// Check for space replacement if configured
		if replaceSpace != "" && strings.Contains(component, " ") {
			t.Errorf("Path component %q contains spaces when replace_space=%q", component, replaceSpace)
		}
	}
}
