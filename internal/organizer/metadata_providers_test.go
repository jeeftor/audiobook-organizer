//go:build !integration

package organizer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Helper function to check if an author is in the authors list
func containsAuthor(authors []string, author string) bool {
	for _, a := range authors {
		if a == author {
			return true
		}
	}
	return false
}

func TestEPUBMetadataExtraction(t *testing.T) {
	// Skip this test if books directory doesn't exist
	testDataDir := filepath.Join("..", "..", "testdata", "epub")
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skipf("Skipping test: test data directory %s does not exist", testDataDir)
	}

	tests := []struct {
		filename       string
		expectedTitle  string
		expectedSeries string
		expectedAuthor string
	}{
		{
			filename:       "title-author.epub",
			expectedTitle:  "The book of cool stuff",
			expectedSeries: "",
			expectedAuthor: "Jeef of Github",
		},
		{
			filename:       "title-author-series1.epub",
			expectedTitle:  "First book of testing knowledge",
			expectedSeries: "Test Books",
			expectedAuthor: "Jeef of Github",
		},
		{
			filename:       "title-author-series2.epub",
			expectedTitle:  "Testing is dumb",
			expectedSeries: "Test Books",
			expectedAuthor: "Jeef of Github",
		},
		{
			filename:       "title-author-series3.epub",
			expectedTitle:  "Why is everything broken",
			expectedSeries: "Test Books",
			expectedAuthor: "Jeef of Github",
		},
		{
			filename:       "strange_book_1_The_Book_With_Colons_.epub",
			expectedTitle:  "The Book: With Colons",
			expectedSeries: "Series/With/Slashes",
			expectedAuthor: "Author*With|Invalid",
		},
		{
			filename:       "strange_book_2_Book_&_Symbols_%_$_#_@_!_.epub",
			expectedTitle:  "Book & Symbols % $ # @ !",
			expectedSeries: "Series™ with ® symbols",
			expectedAuthor: "Author+Plus-Minus±§",
		},
		{
			filename:       "strange_book_3_Café_au_lait_.epub",
			expectedTitle:  "Café au lait",
			expectedSeries: "Résumé Series",
			expectedAuthor: "José",
		},
		{
			filename:       "strange_book_4_This_is_an_extremely_long_title_that_goes_on_and_on_.epub",
			expectedTitle:  "This is an extremely long title that goes on and on",
			expectedSeries: "The Long Series",
			expectedAuthor: "Hubert",
		},
		{
			filename:       "strange_book_5_Book_With_Control_Characters_.epub",
			expectedTitle:  "The book of cool stuff",
			expectedSeries: "",
			expectedAuthor: "Jeef of Github",
		},
		{
			filename:       "strange_book_6__Book_With_Many_Spaces_.epub",
			expectedTitle:  " Book With Many Spaces ",
			expectedSeries: "Series With Spaces",
			expectedAuthor: "Author",
		},
		{
			filename:       "strange_book_7_Book_Quoted_Title_.epub",
			expectedTitle:  "Book \"Quoted\" Title",
			expectedSeries: "Series\\With\\Backslashes",
			expectedAuthor: "Author",
		},
		{
			filename:       "strange_book_8__.epub",
			expectedTitle:  "The book of cool stuff",
			expectedSeries: "",
			expectedAuthor: "Jeef of Github",
		},
		{
			filename:       "strange_book_9_Book_Café_&_Symbols!_.epub",
			expectedTitle:  "Book: Café & Symbols!",
			expectedSeries: "Ångström's Collection",
			expectedAuthor: "José Martínez",
		},
		{
			filename:       "strange_book_10__.epub",
			expectedTitle:  "The book of cool stuff",
			expectedSeries: "Series™ with ® symbols",
			expectedAuthor: "Author",
		},
		{
			filename:       "strange_book_11_Long_Title_With_Colons_(Part_1)_.epub",
			expectedTitle:  "Long Title: With Colons (Part 1)",
			expectedSeries: "",
			expectedAuthor: "Hubert Blaine",
		},
		{
			filename:       "strange_book_12_Book.With.Dots_.epub",
			expectedTitle:  "Book.With.Dots",
			expectedSeries: "Series.With.Dots",
			expectedAuthor: "Author.With.Dots",
		},
		{
			filename:       "strange_book_13__Book_With_Leading_Spaces_.epub",
			expectedTitle:  " Book With Leading Spaces",
			expectedSeries: "Series With Leading Spaces",
			expectedAuthor: "Author",
		},
		{
			filename:       "strange_book_14_Book_With_Trailing_Spaces_.epub",
			expectedTitle:  "Book With Trailing Spaces ",
			expectedSeries: "Series With Trailing Spaces",
			expectedAuthor: "Author",
		},
		{
			filename:       "strange_book_15_Book_With_Multiple_Spaces_.epub",
			expectedTitle:  "Book  With  Multiple  Spaces",
			expectedSeries: "Series  With  Multiple  Spaces",
			expectedAuthor: "Author",
		},
		{
			filename:       "strange_book_16_Book_With_Emoji_🔍_.epub",
			expectedTitle:  "Book With Emoji 🔍",
			expectedSeries: "Series With Emoji 🔍",
			expectedAuthor: "Author",
		},
		{
			filename:       "strange_book_17_Book_With_HTML_bTagsb_.epub",
			expectedTitle:  "Book With HTML <b>Tags</b>",
			expectedSeries: "Series With <i>HTML</i> Tags",
			expectedAuthor: "Author",
		},
		{
			filename:       "strange_book_18_Multi-Author_Book_.epub",
			expectedTitle:  "Multi-Author Book",
			expectedSeries: "Collaboration Series",
			expectedAuthor: "John Doe",
		},
		{
			filename:       "strange_book_19_Three_Author_Book_.epub",
			expectedTitle:  "Three Author Book",
			expectedSeries: "Team Series",
			expectedAuthor: "Alice Johnson",
		},
		{
			filename:       "strange_book_20_Complex_Authors_.epub",
			expectedTitle:  "Complex Authors",
			expectedSeries: "Mixed Series",
			expectedAuthor: "José Martínez",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
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
			if tt.expectedSeries != "" {
				if len(metadata.Series) == 0 {
					t.Errorf("Expected series %q, but no series found", tt.expectedSeries)
				} else if metadata.Series[0] != tt.expectedSeries {
					t.Errorf("Expected series %q, got %q", tt.expectedSeries, metadata.Series[0])
				}
			} else if len(metadata.Series) > 0 {
				t.Errorf("Expected no series, but found %q", metadata.Series[0])
			}

			// Verify author metadata - just check first author since we're seeing only first author in results
			if tt.expectedAuthor != "" {
				if len(metadata.Authors) == 0 {
					t.Errorf("Expected author %q, but no author found", tt.expectedAuthor)
				} else if !containsAuthor(metadata.Authors, tt.expectedAuthor) {
					t.Errorf("Expected author %q, got %v", tt.expectedAuthor, metadata.Authors)
				}
			}
		})
	}
}

func TestMP3MetadataWithProblematicFiles(t *testing.T) {
	// Test both mp3track and mp3flat directories
	testDataDirs := []string{
		"../../testdata/mp3track",
		"../../testdata/mp3flat",
		"../../testdata/mp3-badmetadata",
	}

	cwd, _ := os.Getwd()
	t.Logf("Current working directory: %s", cwd)

	var found bool
	for _, testDataDir := range testDataDirs {
		dirEntries, err := os.ReadDir(testDataDir)
		if err != nil {
			t.Logf("Failed to read directory %s: %v", testDataDir, err)
			continue
		}

		for _, entry := range dirEntries {
			if entry.Type().IsRegular() && filepath.Ext(entry.Name()) == ".mp3" {
				found = true
				filename := entry.Name()
				filePath := filepath.Join(testDataDir, filename)
				t.Run(filename, func(t *testing.T) {
					provider := NewFileMetadataProvider(filePath)
					metadata, err := provider.GetMetadata()
					if err != nil {
						t.Logf("Note: Failed to get metadata for %s: %v", filename, err)
						return // Skip this file but don't fail the test
					}
					t.Logf("File: %s\nMetadata: %+v", filename, metadata)
				})
			}
		}
	}
	if !found {
		t.Fatalf("No mp3 files found in any of the test directories (cwd: %s)", cwd)
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
