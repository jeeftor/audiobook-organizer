package organizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type renameTestMetadataResolver struct {
	metadata Metadata
}

func (r renameTestMetadataResolver) MetadataForPath(string) (Metadata, error) {
	return r.metadata, nil
}

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

func TestRenamerScanFilesUsesConfiguredMetadataResolver(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "source.mp3")
	if err := os.WriteFile(filePath, []byte("audio"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	renamer, err := NewRenamer(&RenamerConfig{
		BaseDir:      tmpDir,
		Template:     "{track} - {title}",
		AuthorFormat: AuthorFormatFirstLast,
		Recursive:    true,
		MetadataResolver: renameTestMetadataResolver{metadata: Metadata{
			Title:       "ABS Title",
			Authors:     []string{"ABS Author"},
			TrackNumber: 7,
			RawData:     map[string]interface{}{},
		}},
	})
	if err != nil {
		t.Fatalf("NewRenamer() error = %v", err)
	}

	candidates, err := renamer.ScanFiles()
	if err != nil {
		t.Fatalf("ScanFiles() error = %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("ScanFiles() candidates = %d, want 1", len(candidates))
	}
	if got, want := filepath.Base(candidates[0].ProposedPath), "07 - ABS Title.mp3"; got != want {
		t.Fatalf("proposed filename = %q, want %q", got, want)
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
		if err := os.MkdirAll(bookDir, 0o755); err != nil {
			t.Fatalf("Failed to create book directory: %v", err)
		}

		// Create metadata.json
		metadataPath := filepath.Join(bookDir, "metadata.json")
		if err := os.WriteFile(metadataPath, []byte(tf.metadataJSON), 0o644); err != nil {
			t.Fatalf("Failed to write metadata.json: %v", err)
		}

		// Create dummy audio file (can't be parsed, but tests error handling)
		audioPath := filepath.Join(bookDir, tf.name)
		if err := os.WriteFile(audioPath, []byte("dummy audio"), 0o644); err != nil {
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

func TestRenamerScanFilesHonorsAllowedCurrentPaths(t *testing.T) {
	tmpDir := t.TempDir()
	selectedPath := createDummyRenameBook(t, tmpDir, "selected", "selected.mp3")
	createDummyRenameBook(t, tmpDir, "ignored", "ignored.mp3")

	renamer, err := NewRenamer(&RenamerConfig{
		BaseDir:             tmpDir,
		Template:            "{author} - {title}",
		AuthorFormat:        AuthorFormatFirstLast,
		Recursive:           true,
		AllowedCurrentPaths: []string{selectedPath},
	})
	if err != nil {
		t.Fatalf("NewRenamer() error: %v", err)
	}

	candidates, err := renamer.ScanFiles()
	if err != nil {
		t.Fatalf("ScanFiles() error: %v", err)
	}

	if got := len(candidates); got != 1 {
		t.Fatalf("candidate count = %d, want 1", got)
	}
	if candidates[0].CurrentPath != selectedPath {
		t.Fatalf("candidate path = %q, want %q", candidates[0].CurrentPath, selectedPath)
	}
	if summary := renamer.GetSummary(); summary.FilesScanned != 1 {
		t.Fatalf("FilesScanned = %d, want 1", summary.FilesScanned)
	}
}

func TestRenamerScanFilesRejectsInvalidAllowedCurrentPath(t *testing.T) {
	tmpDir := t.TempDir()
	createDummyRenameBook(t, tmpDir, "selected", "selected.mp3")
	missingPath := filepath.Join(tmpDir, "missing", "missing.mp3")

	renamer, err := NewRenamer(&RenamerConfig{
		BaseDir:             tmpDir,
		Template:            "{author} - {title}",
		AuthorFormat:        AuthorFormatFirstLast,
		Recursive:           true,
		AllowedCurrentPaths: []string{missingPath},
	})
	if err != nil {
		t.Fatalf("NewRenamer() error: %v", err)
	}

	_, err = renamer.ScanFiles()
	if err == nil {
		t.Fatal("ScanFiles() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "resolve allowed rename path") {
		t.Fatalf("ScanFiles() error = %q, want allowed path resolution context", err)
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

func createDummyRenameBook(t *testing.T, root, dirName, audioName string) string {
	t.Helper()
	bookDir := filepath.Join(root, dirName)
	if err := os.MkdirAll(bookDir, 0o755); err != nil {
		t.Fatalf("Failed to create book directory: %v", err)
	}
	metadataPath := filepath.Join(bookDir, "metadata.json")
	if err := os.WriteFile(
		metadataPath,
		[]byte(`{"title":"Allowed Book","authors":["Allowed Author"]}`),
		0o644,
	); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}
	audioPath := filepath.Join(bookDir, audioName)
	if err := os.WriteFile(audioPath, []byte("dummy audio"), 0o644); err != nil {
		t.Fatalf("Failed to write audio file: %v", err)
	}
	return audioPath
}

func TestRenamer_GenerateNewPathTemplateCompatibility(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		template     string
		metadata     Metadata
		originalPath string
		replaceSpace string
		wantFilename string
	}{
		{
			name:     "dollar brace syntax with series count and narrators",
			template: "${author} - ${series-count} - ${title} (${narrators})",
			metadata: Metadata{
				Title:   "Template Book",
				Authors: []string{"Template Author"},
				Series:  []string{"Template Series #7"},
				RawData: map[string]interface{}{
					"narrators": []interface{}{"Narrator One", "Narrator Two"},
				},
			},
			originalPath: filepath.Join(tmpDir, "original.mp3"),
			wantFilename: "Template Author - 7 - Template Book (Narrator One, Narrator Two).mp3",
		},
		{
			name:     "missing optional fields use fallback values",
			template: "{author} - {series|Standalone} - {title} ({narrator|Unknown Narrator})",
			metadata: Metadata{
				Title:   "Standalone Book",
				Authors: []string{"Standalone Author"},
				RawData: map[string]interface{}{},
			},
			originalPath: filepath.Join(tmpDir, "original.m4b"),
			wantFilename: "Standalone Author - Standalone - Standalone Book (Unknown Narrator).m4b",
		},
		{
			name:     "path separators render as safe filename characters",
			template: "{author}/{title}",
			metadata: Metadata{
				Title:   "Nested Title",
				Authors: []string{"Nested Author"},
			},
			originalPath: filepath.Join(tmpDir, "original.mp3"),
			wantFilename: "Nested Author_Nested Title.mp3",
		},
		{
			name:     "raw field aliases and space replacement are applied",
			template: "{publisher-name} - {title}",
			metadata: Metadata{
				Title:   "Raw Alias Book",
				Authors: []string{"Raw Author"},
				RawData: map[string]interface{}{
					"publisher_name": "Raw Publisher",
				},
			},
			originalPath: filepath.Join(tmpDir, "original.mp3"),
			replaceSpace: "_",
			wantFilename: "Raw_Publisher_-_Raw_Alias_Book.mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renamer, err := NewRenamer(&RenamerConfig{
				BaseDir:      tmpDir,
				Template:     tt.template,
				AuthorFormat: AuthorFormatFirstLast,
				PreservePath: true,
				ReplaceSpace: tt.replaceSpace,
			})
			if err != nil {
				t.Fatalf("NewRenamer() error: %v", err)
			}

			newPath, err := renamer.GenerateNewPath(tt.originalPath, tt.metadata)
			if err != nil {
				t.Fatalf("GenerateNewPath() error: %v", err)
			}

			if got := filepath.Base(newPath); got != tt.wantFilename {
				t.Fatalf("GenerateNewPath() filename = %q, want %q", got, tt.wantFilename)
			}
			if gotDir := filepath.Dir(newPath); gotDir != tmpDir {
				t.Fatalf("GenerateNewPath() dir = %q, want %q", gotDir, tmpDir)
			}
		})
	}
}

func TestRenamer_Execute_DryRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	originalFile := filepath.Join(tmpDir, "original.m4b")
	if err := os.WriteFile(originalFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create metadata.json
	metadataPath := filepath.Join(tmpDir, "metadata.json")
	metadataContent := `{
		"title": "Test Title",
		"authors": ["Test Author"]
	}`
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0o644); err != nil {
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

	candidates := []RenameCandidate{
		{
			CurrentPath:  filepath.Join(tmpDir, "file1.m4b"),
			ProposedPath: filepath.Join(tmpDir, "Same Author - Same Title.m4b"),
		},
		{
			CurrentPath:  filepath.Join(tmpDir, "file2.m4b"),
			ProposedPath: filepath.Join(tmpDir, "Same Author - Same Title.m4b"),
		},
	}
	renamer.finalizePreviewSummary(candidates)

	if !candidates[1].IsConflict {
		t.Fatal("second duplicate candidate should be marked as a conflict")
	}
	if got := filepath.Base(candidates[1].ProposedPath); got != "Same Author - Same Title (2).m4b" {
		t.Fatalf(
			"resolved conflict filename = %q, want %q",
			got,
			"Same Author - Same Title (2).m4b",
		)
	}
	if got := renamer.GetSummary().ConflictsFound; got != 1 {
		t.Fatalf("ConflictsFound = %d, want 1", got)
	}
}

func TestRenamer_PreviewSummaryCountsSkippedErrorsAndConflicts(t *testing.T) {
	tmpDir := t.TempDir()
	renamer, err := NewRenamer(&RenamerConfig{
		BaseDir:      tmpDir,
		Template:     "{author} - {title}",
		AuthorFormat: AuthorFormatFirstLast,
	})
	if err != nil {
		t.Fatalf("NewRenamer() error: %v", err)
	}

	candidates := []RenameCandidate{
		{
			CurrentPath:  filepath.Join(tmpDir, "source-one.mp3"),
			ProposedPath: filepath.Join(tmpDir, "Preview Author - Preview Book.mp3"),
		},
		{
			CurrentPath:  filepath.Join(tmpDir, "source-two.mp3"),
			ProposedPath: filepath.Join(tmpDir, "Preview Author - Preview Book.mp3"),
		},
		{
			CurrentPath:  filepath.Join(tmpDir, "Noop Author - Noop Book.mp3"),
			ProposedPath: filepath.Join(tmpDir, "Noop Author - Noop Book.mp3"),
			IsNoOp:       true,
		},
		{
			CurrentPath: filepath.Join(tmpDir, "broken.mp3"),
			Error:       "Failed to extract metadata: test error",
		},
	}
	renamer.finalizePreviewSummary(candidates)

	summary := renamer.GetSummary()
	if summary.FilesScanned != 4 {
		t.Fatalf("FilesScanned = %d, want 4", summary.FilesScanned)
	}
	if summary.FilesSkipped != 2 {
		t.Fatalf("FilesSkipped = %d, want 2", summary.FilesSkipped)
	}
	if summary.ConflictsFound != 1 {
		t.Fatalf("ConflictsFound = %d, want 1", summary.ConflictsFound)
	}
	if len(summary.Errors) != 1 {
		t.Fatalf("Errors length = %d, want 1", len(summary.Errors))
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
