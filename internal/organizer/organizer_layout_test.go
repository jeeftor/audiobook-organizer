package organizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testFile and testEnvironment types are defined in organizer_test.go

// TestLayoutDefaultLayoutWithSeries verifies the default layout behavior with series information
// It tests the case where --layout=author-series-title and --use-series-as-title=false
func TestLayoutDefaultLayoutWithSeries(t *testing.T) {
	testCases := []struct {
		name      string
		hasSeries bool
		series    string
		expected  string
	}{
		{
			name:      "with_series",
			hasSeries: true,
			series:    "Test Series",
			expected:  "John Smith/Test Series/Test Book/audio.mp3",
		},
		{
			name:      "without_series",
			hasSeries: false,
			series:    "",
			expected:  "John Smith/Test Book/audio.mp3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			tf := testFile{
				Path: "test_book/audio.mp3",
				Metadata: &Metadata{
					Title:   "Test Book",
					Authors: []string{"John Smith"},
				},
			}

			if tc.hasSeries {
				tf.Metadata.Series = []string{tc.series}
			}

			env := setupTestEnvironment(t, []testFile{tf})
			defer env.Cleanup()

			// Create organizer with default settings
			config := &OrganizerConfig{
				BaseDir:             env.InputDir,
				OutputDir:           env.OutputDir,
				UseEmbeddedMetadata: false,
				Flat:                false,
				Layout:              "author-series-title",
				Verbose:             testing.Verbose(),
			}

			org := NewOrganizer(config)
			err := org.Execute()
			if err != nil {
				t.Fatalf("Execute() returned error: %v", err)
			}

			// Verify the output directory structure
			expectedPath := filepath.Join(env.OutputDir, filepath.FromSlash(tc.expected))

			t.Logf("\n=== Test Case: %s ===", tc.name)
			t.Logf("Input Dir: %s", env.InputDir)
			t.Logf("Output Dir: %s", env.OutputDir)
			t.Logf("Expected path: %s", expectedPath)

			// Print the entire directory structure
			t.Log("\nDirectory structure:")
			err = filepath.Walk(env.OutputDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				shortPath := strings.TrimPrefix(path, env.OutputDir)
				if info.IsDir() {
					t.Logf("DIR:  %s/", shortPath)
				} else {
					t.Logf("FILE: %s", shortPath)
				}
				return nil
			})
			if err != nil {
				t.Logf("Error walking directory: %v", err)
			}

			// Check if expected path exists
			_, err = os.Stat(expectedPath)
			if os.IsNotExist(err) {
				t.Errorf("❌ Expected path does not exist: %s", expectedPath)
				t.Logf("Error: %v", err)
			} else if err != nil {
				t.Errorf("❌ Error checking path %s: %v", expectedPath, err)
			} else {
				t.Logf("✅ Found expected path: %s", expectedPath)
			}
		})
	}
}

// TestSeriesAsTitleFlag verifies the behavior of the --use-series-as-title flag
// It tests both flat and non-flat modes to ensure consistent behavior
func TestLayoutSeriesAsTitleFlag(t *testing.T) {
	testCases := []struct {
		name              string
		flat              bool
		useSeriesAsTitle  bool
		hasSeries         bool
		series            string
		expectedStructure string
	}{
		{
			name:              "non_flat_with_series",
			flat:              false,
			useSeriesAsTitle:  true,
			hasSeries:         true,
			series:            "Test Series",
			expectedStructure: "John Smith/Test Series",
		},
		{
			name:              "flat_with_series",
			flat:              true,
			useSeriesAsTitle:  true,
			hasSeries:         true,
			series:            "Test Series",
			expectedStructure: "John Smith/Test Series/audio.mp3", // Flat mode with series as title
		},
		{
			name:              "no_series_should_ignore_flag",
			flat:              false,
			useSeriesAsTitle:  true,
			hasSeries:         false,
			series:            "",
			expectedStructure: "John Smith/Test Book",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			tf := testFile{
				Path: "test_book/audio.mp3",
				Metadata: &Metadata{
					Title:   "Test Book",
					Authors: []string{"John Smith"},
				},
			}

			if tc.hasSeries {
				tf.Metadata.Series = []string{tc.series}
			}

			env := setupTestEnvironment(t, []testFile{tf})
			defer env.Cleanup()

			// Create organizer with test configuration
			config := &OrganizerConfig{
				BaseDir:             env.InputDir,
				OutputDir:           env.OutputDir,
				UseEmbeddedMetadata: false,
				Flat:                tc.flat,
				Layout:              "author-series-title",
				Verbose:             testing.Verbose(),
			}

			org := NewOrganizer(config)
			err := org.Execute()
			if err != nil {
				t.Fatalf("Execute() returned error: %v", err)
			}

			// Verify the output directory structure
			expectedPath := filepath.Join(env.OutputDir, filepath.FromSlash(tc.expectedStructure))
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("Expected path does not exist: %s", expectedPath)
				// Print actual directory structure for debugging
				filepath.Walk(env.OutputDir, func(path string, info os.FileInfo, err error) error {
					if !info.IsDir() {
						t.Logf("Found file: %s", path)
					}
					return nil
				})
			}
		})
	}
}
