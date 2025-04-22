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
