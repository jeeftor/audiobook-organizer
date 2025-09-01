package organizer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestOrganizer(t *testing.T) {
	tests := []struct {
		name         string
		metadata     Metadata
		replaceSpace string
		wantDir      string
	}{
		{
			name: "single_author",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
			},
			replaceSpace: "",
			wantDir:      "John Smith/Test Book",
		},
		{
			name: "multiple_authors",
			metadata: Metadata{
				Authors: []string{"John Smith", "Jane Doe"},
				Title:   "Test Book",
			},
			replaceSpace: "",
			wantDir:      "John Smith,Jane Doe/Test Book",
		},
		{
			name: "with_series",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series #12"},
			},
			replaceSpace: "",
			wantDir:      "John Smith/Test Series/Test Book",
		},
		{
			name: "with_series_and_space_replacement",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series #1"},
			},
			replaceSpace: ".",
			wantDir:      "John.Smith/Test.Series/Test.Book",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir() // Using t.TempDir() for automatic cleanup

			sourceDir := filepath.Join(tempDir, "source")
			if err := os.MkdirAll(sourceDir, 0755); err != nil {
				t.Fatal(err)
			}

			metadataBytes, err := json.Marshal(tt.metadata)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(sourceDir, "metadata.json"), metadataBytes, 0644); err != nil {
				t.Fatal(err)
			}

			testData := []byte("test data")
			testFile := filepath.Join(sourceDir, "test.mp3")
			if err := os.WriteFile(testFile, testData, 0644); err != nil {
				t.Fatal(err)
			}

			config := &OrganizerConfig{
				BaseDir:             tempDir,
				OutputDir:           "",
				ReplaceSpace:        tt.replaceSpace,
				Verbose:             false,
				DryRun:              false,
				Undo:                false,
				Prompt:              false,
				RemoveEmpty:         false,
				UseEmbeddedMetadata: false,
			}
			org := NewOrganizer(config)

			provider := NewJSONMetadataProvider(filepath.Join(sourceDir, "metadata.json"))
			if err := org.OrganizeAudiobook(sourceDir, provider); err != nil {
				t.Fatal(err)
			}

			wantPath := filepath.Join(tempDir, tt.wantDir)
			if _, err := os.Stat(wantPath); os.IsNotExist(err) {
				t.Errorf("directory %s was not created", wantPath)
			}

			wantFile := filepath.Join(wantPath, "test.mp3")
			if _, err := os.Stat(wantFile); os.IsNotExist(err) {
				t.Errorf("file was not moved to %s", wantFile)
			}

			// Verify file contents
			movedData, err := os.ReadFile(wantFile)
			if err != nil {
				t.Errorf("error reading moved file: %v", err)
			}
			if !bytes.Equal(movedData, testData) {
				t.Error("moved file contents do not match original")
			}
		})
	}
}

func TestFlatDirectoryWithSeriesAsTitle(t *testing.T) {
	// Test cases with expected directory structures
	tests := []struct {
		name            string
		mp3File         string
		useSeriesAsTitle bool
		expectedPath    string
	}{
		{
			name:            "lovecraft_with_series_as_title",
			mp3File:         "charlesdexterward_01_lovecraft_64kb.mp3",
			useSeriesAsTitle: true,
			expectedPath:    "H. P. Lovecraft/The Case of Charles Dexter Ward",
		},
		{
			name:            "lovecraft_without_series_as_title",
			mp3File:         "charlesdexterward_01_lovecraft_64kb.mp3",
			useSeriesAsTitle: false,
			expectedPath:    "H. P. Lovecraft/The Case of Charles Dexter Ward/01 - Chapter 1_ A Result and a Prologue",
		},
		{
			name:            "kenrick_with_series_as_title",
			mp3File:         "falstaffswedding1766version_1_kenrick_64kb.mp3",
			useSeriesAsTitle: true,
			expectedPath:    "William Kenrick/Falstaff's Wedding (1766 Version)",
		},
		{
			name:            "kenrick_without_series_as_title",
			mp3File:         "falstaffswedding1766version_1_kenrick_64kb.mp3",
			useSeriesAsTitle: false,
			expectedPath:    "William Kenrick/Falstaff's Wedding (1766 Version)/01 - Act 1",
		},
		{
			name:            "scott_with_series_as_title",
			mp3File:         "perouse_01_scott_64kb.mp3",
			useSeriesAsTitle: true,
			expectedPath:    "Ernest Scott/Lapérouse",
		},
		{
			name:            "scott_without_series_as_title",
			mp3File:         "perouse_01_scott_64kb.mp3",
			useSeriesAsTitle: false,
			expectedPath:    "Ernest Scott/Lapérouse/01 - Family, youth and influences",
		},
	}

	// Get the project root directory
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	// Path to the testdata/mp3flat directory
	mp3FlatDir := filepath.Join(projectRoot, "testdata", "mp3flat")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the test
			tempDir := t.TempDir()

			// Copy the MP3 file to the temp directory
			sourcePath := filepath.Join(mp3FlatDir, tt.mp3File)
			destPath := filepath.Join(tempDir, tt.mp3File)

			sourceData, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("Failed to read source file: %v", err)
			}

			err = os.WriteFile(destPath, sourceData, 0644)
			if err != nil {
				t.Fatalf("Failed to write destination file: %v", err)
			}

			// Create organizer with appropriate configuration
			config := &OrganizerConfig{
				BaseDir:             tempDir,
				OutputDir:           "",
				ReplaceSpace:        "",
				Verbose:             false,
				DryRun:              false,
				Undo:                false,
				Prompt:              false,
				RemoveEmpty:         false,
				UseEmbeddedMetadata: true,
				Flat:                true,
				Layout:              "author-series-title",
				UseSeriesAsTitle:    tt.useSeriesAsTitle,
			}

			org := NewOrganizer(config)

			// Process the file using the public OrganizeSingleFile method
			// Create a metadata provider that can read from the MP3 file
			provider := NewAudioMetadataProvider(destPath)
			err = org.OrganizeSingleFile(destPath, provider)
			if err != nil {
				t.Fatalf("Failed to process file: %v", err)
			}

			// Check if the file was moved to the expected location
			expectedDir := filepath.Join(tempDir, tt.expectedPath)
			expectedFilePath := filepath.Join(expectedDir, tt.mp3File)

			if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
				t.Errorf("File not found at expected path: %s", expectedFilePath)

				// List the contents of the temp directory to help debug
				files, err := filepath.Glob(filepath.Join(tempDir, "*", "*", "*", "*"))
				if err == nil {
					t.Logf("Files found in temp directory:")
					for _, file := range files {
						t.Logf("  %s", file)
					}
				}
			}
		})
	}
}

func TestOutputDirectory(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	outputDir := filepath.Join(tempDir, "output")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	metadata := Metadata{
		Authors: []string{"John Smith"},
		Title:   "Test Book",
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "metadata.json"), metadataBytes, 0644); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(sourceDir, "test.mp3")
	if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
		t.Fatal(err)
	}

	config := &OrganizerConfig{
		BaseDir:      sourceDir,
		OutputDir:    outputDir,
		ReplaceSpace: "",
		Verbose:      false,
		DryRun:       false,
		Undo:         false,
		Prompt:       false,
		RemoveEmpty:  false,
	}
	org := NewOrganizer(config)

	provider := NewJSONMetadataProvider(filepath.Join(sourceDir, "metadata.json"))
	if err := org.OrganizeAudiobook(sourceDir, provider); err != nil {
		t.Fatal(err)
	}

	wantPath := filepath.Join(outputDir, "John Smith/Test Book")
	if _, err := os.Stat(wantPath); os.IsNotExist(err) {
		t.Errorf("directory %s was not created in output directory", wantPath)
	}
}

func TestLayoutOptions(t *testing.T) {
	tests := []struct {
		name       string
		layout     string
		metadata   Metadata
		wantDir    string
	}{
		{
			name:   "author_series_title_layout",
			layout: "author-series-title",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series"},
			},
			wantDir: "John Smith/Test Series/Test Book",
		},
		{
			name:   "author_title_layout",
			layout: "author-title",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series"},
			},
			wantDir: "John Smith/Test Book",
		},
		{
			name:   "author_only_layout",
			layout: "author-only",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series"},
			},
			wantDir: "John Smith",
		},
		{
			name:   "default_layout_when_empty",
			layout: "",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series"},
			},
			wantDir: "John Smith/Test Series/Test Book",
		},
		{
			name:   "unknown_layout_defaults_to_author_title",
			layout: "invalid-layout",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series"},
			},
			wantDir: "John Smith/Test Book",
		},
		{
			name:   "author_series_title_layout_no_series",
			layout: "author-series-title",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{},
			},
			wantDir: "John Smith/Test Book",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir() // Using t.TempDir() for automatic cleanup

			sourceDir := filepath.Join(tempDir, "source")
			if err := os.MkdirAll(sourceDir, 0755); err != nil {
				t.Fatal(err)
			}

			metadataBytes, err := json.Marshal(tt.metadata)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(sourceDir, "metadata.json"), metadataBytes, 0644); err != nil {
				t.Fatal(err)
			}

			testData := []byte("test data")
			testFile := filepath.Join(sourceDir, "test.mp3")
			if err := os.WriteFile(testFile, testData, 0644); err != nil {
				t.Fatal(err)
			}

			config := &OrganizerConfig{
				BaseDir:             tempDir,
				OutputDir:           "",
				ReplaceSpace:        "",
				Verbose:             false,
				DryRun:              false,
				Undo:                false,
				Prompt:              false,
				RemoveEmpty:         false,
				UseEmbeddedMetadata: false,
				Layout:              tt.layout,
			}
			org := NewOrganizer(config)

			provider := NewJSONMetadataProvider(filepath.Join(sourceDir, "metadata.json"))
			if err := org.OrganizeAudiobook(sourceDir, provider); err != nil {
				t.Fatal(err)
			}

			wantPath := filepath.Join(tempDir, tt.wantDir)
			if _, err := os.Stat(wantPath); os.IsNotExist(err) {
				t.Errorf("directory %s was not created", wantPath)
			}

			// For author-only layout, the file should be directly in the author directory
			var wantFile string
			if tt.layout == "author-only" {
				wantFile = filepath.Join(wantPath, "test.mp3")
			} else {
				wantFile = filepath.Join(wantPath, "test.mp3")
			}

			if _, err := os.Stat(wantFile); os.IsNotExist(err) {
				t.Errorf("file was not moved to %s", wantFile)

				// List the contents of the temp directory to help debug
				files, err := filepath.Glob(filepath.Join(tempDir, "*", "*", "*"))
				if err == nil {
					t.Logf("Files found in temp directory:")
					for _, file := range files {
						t.Logf("  %s", file)
					}
				}
			}

			// Verify file contents
			movedData, err := os.ReadFile(wantFile)
			if err != nil {
				t.Errorf("error reading moved file: %v", err)
			}
			if !bytes.Equal(movedData, testData) {
				t.Error("moved file contents do not match original")
			}
		})
	}
}
