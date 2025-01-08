package main_test

import (
	"audiobook-organizer/internal/organizer"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "audiobook-organizer/internal/organizer" // Add this import
)

func TestOrganizer(t *testing.T) {
	tests := []struct {
		name         string
		metadata     organizer.Metadata
		replaceSpace string
		wantDir      string
	}{
		{
			name: "single_author",
			metadata: organizer.Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
			},
			replaceSpace: "",
			wantDir:      "John Smith/Test Book",
		},
		{
			name: "multiple_authors",
			metadata: organizer.Metadata{
				Authors: []string{"John Smith", "Jane Doe"},
				Title:   "Test Book",
			},
			replaceSpace: "",
			wantDir:      "John Smith,Jane Doe/Test Book",
		},
		{
			name: "with_series",
			metadata: organizer.Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series #12"},
			},
			replaceSpace: "",
			wantDir:      "John Smith/Test Series/Test Book",
		},
		{
			name: "with_series_and_space_replacement",
			metadata: organizer.Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series #1"},
			},
			replaceSpace: ".",
			wantDir:      "John.Smith/Test.Series/Test.Book",
		},
		{
			name: "directory_with_spaces",
			metadata: organizer.Metadata{
				Authors: []string{"John Smith Jr"},
				Title:   "My Book Title",
				Series:  []string{"My Series Name #3"},
			},
			replaceSpace: "",
			wantDir:      "John Smith Jr/My Series Name/My Book Title",
		},
		{
			name: "multiple_hash_in_series",
			metadata: organizer.Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test #Series Part 1 #12"},
			},
			replaceSpace: "",
			wantDir:      "John Smith/Test #Series Part 1/Test Book",
		},
		{
			name: "special_characters",
			metadata: organizer.Metadata{
				Authors: []string{"John Smith-Jones", "O'Brien, Pat"},
				Title:   "Test & Book!",
				Series:  []string{"Test's Series #1"},
			},
			replaceSpace: "",
			wantDir:      "John Smith-Jones,O'Brien, Pat/Test's Series/Test & Book!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "audiobook-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

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

			org := organizer.New(
				tempDir,
				"", // no output dir
				tt.replaceSpace,
				false, // verbose
				false, // dryRun
				false, // undo
				false, // prompt
			)

			if err := org.OrganizeAudiobook(sourceDir, filepath.Join(sourceDir, "metadata.json")); err != nil {
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

func TestOutputDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audiobook-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	outputDir := filepath.Join(tempDir, "output")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	metadata := organizer.Metadata{
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

	org := organizer.New(
		sourceDir,
		outputDir,
		"",    // replaceSpace
		false, // verbose
		false, // dryRun
		false, // undo
		false, // prompt
	)

	if err := org.OrganizeAudiobook(sourceDir, filepath.Join(sourceDir, "metadata.json")); err != nil {
		t.Fatal(err)
	}

	wantPath := filepath.Join(outputDir, "John Smith/Test Book")
	if _, err := os.Stat(wantPath); os.IsNotExist(err) {
		t.Errorf("directory %s was not created in output directory", wantPath)
	}
}

func TestMissingMetadata(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audiobook-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test missing authors
	missingAuthors := organizer.Metadata{
		Title: "Test Book",
	}
	if err := testInvalidMetadata(t, tempDir, missingAuthors); err == nil {
		t.Error("expected error for missing authors")
	}

	// Test missing title
	missingTitle := organizer.Metadata{
		Authors: []string{"John Smith"},
	}
	if err := testInvalidMetadata(t, tempDir, missingTitle); err == nil {
		t.Error("expected error for missing title")
	}
}

func TestInvalidMetadataJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audiobook-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write invalid JSON
	invalidJSON := []byte(`{"authors": ["Test Author"], "title": "Test Book", invalid json}`)
	metadataPath := filepath.Join(sourceDir, "metadata.json")
	if err := os.WriteFile(metadataPath, invalidJSON, 0644); err != nil {
		t.Fatal(err)
	}

	org := organizer.New(
		tempDir,
		"",    // outputDir
		"",    // replaceSpace
		false, // verbose
		false, // dryRun
		false, // undo
		false, // prompt
	)

	if err := org.OrganizeAudiobook(sourceDir, metadataPath); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLogFileCreation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audiobook-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	metadata := organizer.Metadata{
		Authors: []string{"Test Author"},
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

	org := organizer.New(
		tempDir,
		"",    // outputDir
		"",    // replaceSpace
		false, // verbose
		false, // dryRun
		false, // undo
		false, // prompt
	)

	if err := org.OrganizeAudiobook(sourceDir, filepath.Join(sourceDir, "metadata.json")); err != nil {
		t.Fatal(err)
	}

	// Check log file exists
	logPath := filepath.Join(tempDir, organizer.LogFileName)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("log file was not created")
	}

	// Check log content
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	var logEntries []organizer.LogEntry
	if err := json.Unmarshal(logData, &logEntries); err != nil {
		t.Error("invalid log file format")
	}

	if len(logEntries) == 0 {
		t.Error("log file is empty")
	}

	if !strings.Contains(logEntries[0].TargetPath, "Test Author/Test Book") {
		t.Error("incorrect target path in log")
	}
}

func testInvalidMetadata(t *testing.T, tempDir string, metadata organizer.Metadata) error {
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

	org := organizer.New(
		tempDir,
		"",    // outputDir
		"",    // replaceSpace
		false, // verbose
		false, // dryRun
		false, // undo
		false, // prompt
	)

	return org.OrganizeAudiobook(sourceDir, metadataPath)
}
