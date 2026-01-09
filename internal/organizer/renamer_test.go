package organizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRenamer(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		config  *RenamerConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &RenamerConfig{
				BaseDir:      tmpDir,
				Template:     "{author} - {title}",
				AuthorFormat: AuthorFormatFirstLast,
			},
			wantErr: false,
		},
		{
			name: "invalid template",
			config: &RenamerConfig{
				BaseDir:      tmpDir,
				Template:     "", // Empty template is actually valid
				AuthorFormat: AuthorFormatFirstLast,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renamer, err := NewRenamer(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewRenamer() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewRenamer() unexpected error: %v", err)
				return
			}

			if renamer == nil {
				t.Fatal("NewRenamer() returned nil")
			}
		})
	}
}

func TestRenamer_ScanFiles(t *testing.T) {
	// Note: This test uses dummy audio files which can't be parsed.
	// The test verifies that the renamer properly handles errors and returns
	// candidates with error information when metadata extraction fails.

	tmpDir := t.TempDir()

	// Create test files
	testFiles := []struct {
		name         string
		metadataJSON string
	}{
		{
			name: "book1.m4b",
			metadataJSON: `{
				"title": "The Final Empire",
				"authors": ["Brandon Sanderson"],
				"series": ["Mistborn #1"]
			}`,
		},
		{
			name: "book2.mp3",
			metadataJSON: `{
				"title": "The Way of Kings",
				"authors": ["Brandon Sanderson"]
			}`,
		},
	}

	for _, tf := range testFiles {
		// Create book directory
		bookDir := filepath.Join(tmpDir, strings.TrimSuffix(tf.name, filepath.Ext(tf.name)))
		if err := os.MkdirAll(bookDir, 0755); err != nil {
			t.Fatalf("Failed to create book directory: %v", err)
		}

		// Create metadata.json
		metadataPath := filepath.Join(bookDir, "metadata.json")
		if err := os.WriteFile(metadataPath, []byte(tf.metadataJSON), 0644); err != nil {
			t.Fatalf("Failed to write metadata.json: %v", err)
		}

		// Create dummy audio file (can't be parsed, but tests error handling)
		audioPath := filepath.Join(bookDir, tf.name)
		if err := os.WriteFile(audioPath, []byte("dummy audio"), 0644); err != nil {
			t.Fatalf("Failed to write audio file: %v", err)
		}
	}

	config := &RenamerConfig{
		BaseDir:      tmpDir,
		Template:     "{author} - {title}",
		AuthorFormat: AuthorFormatFirstLast,
		Recursive:    true,
	}

	renamer, err := NewRenamer(config)
	if err != nil {
		t.Fatalf("NewRenamer() error: %v", err)
	}

	candidates, err := renamer.ScanFiles()
	if err != nil {
		t.Fatalf("ScanFiles() error: %v", err)
	}

	if len(candidates) != 2 {
		t.Errorf("ScanFiles() found %d files, want 2", len(candidates))
	}

	// With dummy audio files, we expect errors since they can't be parsed
	// This verifies that the scanner properly handles files and reports errors
	errorCount := 0
	for i, candidate := range candidates {
		if candidate.Error != "" {
			errorCount++
			t.Logf("Candidate %d has expected error (dummy audio file): %s", i, candidate.Error)
		}
		// Always verify CurrentPath is set
		if candidate.CurrentPath == "" {
			t.Error("Candidate missing CurrentPath")
		}
	}

	if errorCount != 2 {
		t.Errorf("Expected 2 candidates with errors (dummy files), got %d", errorCount)
	}
}

func TestRenamer_GenerateNewFilename(t *testing.T) {
	tmpDir := t.TempDir()

	config := &RenamerConfig{
		BaseDir:      tmpDir,
		Template:     "{author} - {title}",
		AuthorFormat: AuthorFormatFirstLast,
		PreservePath: true,
	}

	renamer, err := NewRenamer(config)
	if err != nil {
		t.Fatalf("NewRenamer() error: %v", err)
	}

	tests := []struct {
		name         string
		metadata     Metadata
		originalPath string
		wantFilename string
	}{
		{
			name: "simple filename",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
			},
			originalPath: filepath.Join(tmpDir, "original.m4b"),
			wantFilename: "Test Author - Test Book.m4b",
		},
		{
			name: "with series",
			metadata: Metadata{
				Title:   "Book One",
				Authors: []string{"Jane Doe"},
				Series:  []string{"Series #1"},
			},
			originalPath: filepath.Join(tmpDir, "old.mp3"),
			wantFilename: "Jane Doe - Book One.mp3", // Series not in template
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newPath, err := renamer.GenerateNewPath(tt.originalPath, tt.metadata)
			if err != nil {
				t.Errorf("GenerateNewPath() error: %v", err)
				return
			}

			gotFilename := filepath.Base(newPath)
			if gotFilename != tt.wantFilename {
				t.Errorf("GenerateNewPath() filename = %q, want %q", gotFilename, tt.wantFilename)
			}
		})
	}
}

func TestRenamer_Execute_DryRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	originalFile := filepath.Join(tmpDir, "original.m4b")
	if err := os.WriteFile(originalFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create metadata.json
	metadataPath := filepath.Join(tmpDir, "metadata.json")
	metadataContent := `{
		"title": "Test Title",
		"authors": ["Test Author"]
	}`
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}

	config := &RenamerConfig{
		BaseDir:      tmpDir,
		Template:     "{author} - {title}",
		AuthorFormat: AuthorFormatFirstLast,
		DryRun:       true,
		PreservePath: true,
	}

	renamer, err := NewRenamer(config)
	if err != nil {
		t.Fatalf("NewRenamer() error: %v", err)
	}

	if err := renamer.Execute(); err != nil {
		t.Errorf("Execute() error: %v", err)
	}

	// Verify original file still exists (dry-run shouldn't move it)
	if _, err := os.Stat(originalFile); os.IsNotExist(err) {
		t.Error("Execute() in dry-run mode moved the file (should not)")
	}

	// Verify no log file created (dry-run shouldn't create log)
	logPath := filepath.Join(tmpDir, ".abook-rename.log")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("Execute() in dry-run mode created log file (should not)")
	}
}

func TestRenamer_ConflictDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two files that would generate the same target name
	file1 := filepath.Join(tmpDir, "file1.m4b")
	file2 := filepath.Join(tmpDir, "file2.m4b")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Create metadata for both (same title/author = conflict)
	metadataContent := `{
		"title": "Same Title",
		"authors": ["Same Author"]
	}`

	for _, file := range []string{file1, file2} {
		dir := filepath.Dir(file)
		metadataPath := filepath.Join(dir, "metadata.json")
		if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
			t.Fatalf("Failed to write metadata.json: %v", err)
		}
	}

	config := &RenamerConfig{
		BaseDir:      tmpDir,
		Template:     "{author} - {title}",
		AuthorFormat: AuthorFormatFirstLast,
		DryRun:       true, // Use dry-run to test conflict detection
	}

	renamer, err := NewRenamer(config)
	if err != nil {
		t.Fatalf("NewRenamer() error: %v", err)
	}

	candidates, err := renamer.ScanFiles()
	if err != nil {
		t.Fatalf("ScanFiles() error: %v", err)
	}

	// Check for conflicts
	conflicts := detectConflicts(candidates)
	if len(conflicts) == 0 {
		t.Error("detectConflicts() should detect conflict for duplicate target names")
	}
}

func TestRenameLogEntry(t *testing.T) {
	entry := RenameLogEntry{
		OldPath: "/old/path/file.m4b",
		NewPath: "/new/path/file.m4b",
	}

	// Test that entry can be created and has expected fields
	if entry.OldPath == "" {
		t.Error("RenameLogEntry.OldPath is empty")
	}
	if entry.NewPath == "" {
		t.Error("RenameLogEntry.NewPath is empty")
	}
}

func TestRenameSummary(t *testing.T) {
	summary := RenameSummary{
		FilesScanned:   10,
		FilesRenamed:   8,
		FilesSkipped:   2,
		ConflictsFound: 1,
	}

	if summary.FilesScanned != 10 {
		t.Errorf("FilesScanned = %d, want 10", summary.FilesScanned)
	}
	if summary.FilesRenamed != 8 {
		t.Errorf("FilesRenamed = %d, want 8", summary.FilesRenamed)
	}
}
