package organizer

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"os"
)

func TestEPUBMetadataPathDetermination(t *testing.T) {
	// Skip this test if books directory doesn't exist
	testDataDir := filepath.Join("..", "..", "books")

	// Create a test organizer with dry-run mode
	config := &OrganizerConfig{
		BaseDir:             testDataDir,
		OutputDir:           "",
		ReplaceSpace:        "",
		Verbose:             true,
		DryRun:              true,
		UseEmbeddedMetadata: true,
		Flat:                true,
	}
	org := NewOrganizer(config)

	// Test all EPUB files in the books directory
	tests := []struct {
		filename      string
		expectedPath  string
		expectedTitle string
		expectedSeries string
		hasSeries     bool
	}{
		{
			filename:      "title-author-series1.epub",
			expectedPath:  "Jeef of Github,Some random guy/Test Books",
			expectedTitle: "First book of testing knowledge",
			expectedSeries: "Test Books",
			hasSeries:     true,
		},
		{
			filename:      "title-author-series2.epub",
			expectedPath:  "Jeef of Github,Some random guy/Test Books",
			expectedTitle: "Testing is dumb",
			expectedSeries: "Test Books",
			hasSeries:     true,
		},
		{
			filename:      "title-author-series3.epub",
			expectedPath:  "Jeef of Github,Some random guy/Test Books",
			expectedTitle: "Why is everything broken",
			expectedSeries: "Test Books",
			hasSeries:     true,
		},
		{
			filename:      "title-author.epub",
			expectedPath:  "Jeef of Github,Some random guy",
			expectedTitle: "The book of cool stuff",
			expectedSeries: "",
			hasSeries:     false,
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
			if tt.hasSeries {
				if len(metadata.Series) == 0 {
					t.Errorf("Expected series %q, but no series found", tt.expectedSeries)
				} else if metadata.Series[0] != tt.expectedSeries {
					t.Errorf("Expected series %q, got %q", tt.expectedSeries, metadata.Series[0])
				}
			} else if len(metadata.Series) > 0 {
				t.Errorf("Expected no series, but found %q", metadata.Series[0])
			}

			// Determine the target directory using the organizer's logic
			targetDir, err := org.calculateTargetPath(metadata)
			if err != nil {
				t.Fatalf("Failed to calculate target path: %v", err)
			}

			// Check if the target directory contains the expected path
			if !strings.Contains(targetDir, tt.expectedPath) {
				t.Errorf("Expected path to contain %q, got %q", tt.expectedPath, targetDir)
			}

			// For files with series, verify that the series is in the path
			if tt.hasSeries && !strings.Contains(targetDir, tt.expectedSeries) {
				t.Errorf("Expected path to contain series %q, got %q", tt.expectedSeries, targetDir)
			}

			// Print the full target path for debugging
			t.Logf("File: %s\nMetadata: %+v\nTarget path: %s", tt.filename, metadata, targetDir)
		})
	}
}

func TestExtractCalibreSeriesFromOPF(t *testing.T) {
	// Test the direct OPF parsing function
	testDataDir := filepath.Join("..", "..", "books")

	tests := []struct {
		filename      string
		expectedSeries string
		expectFound   bool
	}{
		{
			filename:      "title-author-series1.epub",
			expectedSeries: "Test Books",
			expectFound:   true,
		},
		{
			filename:      "title-author-series2.epub",
			expectedSeries: "Test Books",
			expectFound:   true,
		},
		{
			filename:      "title-author-series3.epub",
			expectedSeries: "Test Books",
			expectFound:   true,
		},
		{
			filename:      "title-author.epub",
			expectedSeries: "",
			expectFound:   false, // No series
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			epubPath := filepath.Join(testDataDir, tt.filename)
			series, found := extractCalibreSeriesFromOPF(epubPath)

			if found != tt.expectFound {
				t.Errorf("Expected found=%v, got %v for %s", tt.expectFound, found, tt.filename)
			}

			if tt.expectFound && series != tt.expectedSeries {
				t.Errorf("Expected series %q, got %q for %s", tt.expectedSeries, series, tt.filename)
			}
		})
	}
}

func TestEPUBMetadataWithProblematicFiles(t *testing.T) {
	// This test processes EPUB files with problematic metadata
	// and verifies that the paths are correctly sanitized

	// Skip this test if books directory doesn't exist
	testDataDir := filepath.Join("..", "..", "books")

	// Create a test organizer with various configurations to test sanitization
	configs := []struct {
		name         string
		replaceSpace string
	}{
		{"default", ""},
		{"underscore", "_"},
		{"dot", "."},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			config := &OrganizerConfig{
				BaseDir:             testDataDir,
				OutputDir:           "",
				ReplaceSpace:        cfg.replaceSpace,
				Verbose:             false,
				DryRun:              true,
				UseEmbeddedMetadata: true,
				Flat:                true,
			}
			org := NewOrganizer(config)

			// Find all EPUB files that match the pattern for problematic files
			files, err := filepath.Glob(filepath.Join(testDataDir, "strange_book_*.epub"))
			if err != nil {
				t.Fatalf("Failed to find test files: %v", err)
			}

			// Skip this test if no files are found
			if len(files) == 0 {
				t.Skip("No strange_book_*.epub files found in books directory")
			}

			t.Logf("Found %d problematic EPUB files to test", len(files))

			// Process each file
			for _, file := range files {
				filename := filepath.Base(file)

				t.Run(filename, func(t *testing.T) {
					// Create the EPUB metadata provider
					provider := NewEPUBMetadataProvider(file)

					// Get the metadata
					metadata, err := provider.GetMetadata()
					if err != nil {
						t.Fatalf("Failed to get metadata for %s: %v", filename, err)
					}

					// Determine the target directory using the organizer's logic
					targetDir, err := org.calculateTargetPath(metadata)
					if err != nil {
						t.Fatalf("Failed to calculate target path: %v", err)
					}

					// Verify that the path doesn't contain invalid characters
					verifyPathSanitization(t, targetDir, cfg.replaceSpace)

					// Log the metadata and target path for debugging
					t.Logf("File: %s\nMetadata: %+v\nTarget path: %s", filename, metadata, targetDir)
				})
			}
		})
	}
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
	// Extract the sanitized part of the path (after books/)
	parts := strings.Split(path, "books/")
	if len(parts) < 2 {
		t.Errorf("Path does not contain 'books/': %s", path)
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
