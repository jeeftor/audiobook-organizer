package main

import (
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
				Title:  "Test Book",
			},
			replaceSpace: ".",
			wantDir:     "John.Smith/Test.Book",
		},
		{
			name: "multiple_authors",
			metadata: Metadata{
				Authors: []string{"John Smith", "Jane Doe"},
				Title:  "Test Book",
			},
			replaceSpace: ".",
			wantDir:     "John.Smith,Jane.Doe/Test.Book",
		},
		{
			name: "custom_space_replacement",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:  "Test Book",
			},
			replaceSpace: "_",
			wantDir:     "John_Smith/Test_Book",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp test directory
			tempDir, err := os.MkdirTemp("", "audiobook-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			// Create source directory with metadata
			sourceDir := filepath.Join(tempDir, "source")
			if err := os.MkdirAll(sourceDir, 0755); err != nil {
				t.Fatal(err)
			}

			// Create metadata file
			metadataBytes, err := json.Marshal(tt.metadata)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(sourceDir, "metadata.json"), metadataBytes, 0644); err != nil {
				t.Fatal(err)
			}

			// Create test file
			testFile := filepath.Join(sourceDir, "test.mp3")
			if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
				t.Fatal(err)
			}

			// Run organizer
			baseDir = tempDir
			replaceSpace = tt.replaceSpace
			dryRun = false
			verbose = false

			if err := organizeAudiobook(sourceDir, filepath.Join(sourceDir, "metadata.json")); err != nil {
				t.Fatal(err)
			}

			// Check if files were moved correctly
			wantPath := filepath.Join(tempDir, tt.wantDir)
			if _, err := os.Stat(wantPath); os.IsNotExist(err) {
				t.Errorf("directory %s was not created", wantPath)
			}

			wantFile := filepath.Join(wantPath, "test.mp3")
			if _, err := os.Stat(wantFile); os.IsNotExist(err) {
				t.Errorf("file was not moved to %s", wantFile)
			}
		})
	}
}

func TestMissingMetadata(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audiobook-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test missing authors
	missingAuthors := Metadata{
		Title: "Test Book",
	}
	if err := testInvalidMetadata(t, tempDir, missingAuthors); err == nil {
		t.Error("expected error for missing authors")
	}

	// Test missing title
	missingTitle := Metadata{
		Authors: []string{"John Smith"},
	}
	if err := testInvalidMetadata(t, tempDir, missingTitle); err == nil {
		t.Error("expected error for missing title")
	}
}

func testInvalidMetadata(t *testing.T, tempDir string, metadata Metadata) error {
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	
	metadataPath := filepath.Join(sourceDir, "metadata.json")
	if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
		t.Fatal(err)
	}

	baseDir = tempDir
	replaceSpace = "."
	return organizeAudiobook(sourceDir, metadataPath)
}