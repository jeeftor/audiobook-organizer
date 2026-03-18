package organizer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// createBookDir creates a book subdirectory with a metadata.json and a fake .mp3 file.
func createBookDir(t *testing.T, baseDir, name, title, author string) string {
	t.Helper()
	bookDir := filepath.Join(baseDir, name)
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		t.Fatalf("failed to create book directory %s: %v", bookDir, err)
	}
	meta := map[string]interface{}{
		"title":   title,
		"authors": []string{author},
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("failed to marshal metadata: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bookDir, "metadata.json"), metaBytes, 0644); err != nil {
		t.Fatalf("failed to write metadata.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bookDir, "audio.mp3"), []byte("fake audio data"), 0644); err != nil {
		t.Fatalf("failed to write audio.mp3: %v", err)
	}
	return bookDir
}

// --- Change 1: AllowedSourcePaths ---

func TestOrganizerAllowedSourcePaths_HierarchicalMode(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := t.TempDir()

	createBookDir(t, baseDir, "BookA", "Book A", "Author A")
	bookBDir := createBookDir(t, baseDir, "BookB", "Book B", "Author B")
	createBookDir(t, baseDir, "BookC", "Book C", "Author C")

	config := OrganizerConfig{
		BaseDir:            baseDir,
		OutputDir:          outputDir,
		DryRun:             false,
		FieldMapping:       DefaultFieldMapping(),
		AllowedSourcePaths: []string{bookBDir},
	}

	org, err := NewOrganizer(&config)
	if err != nil {
		t.Fatalf("NewOrganizer() error = %v", err)
	}

	if err := org.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// BookB should have been moved into the output directory.
	// The target path is OutputDir/Author B/Book B/.
	bookBTarget := filepath.Join(outputDir, "Author B", "Book B")
	if _, err := os.Stat(bookBTarget); os.IsNotExist(err) {
		t.Errorf("expected BookB to be moved to %s, but it does not exist", bookBTarget)
	}

	// BookA and BookC should NOT appear in the output directory.
	bookATarget := filepath.Join(outputDir, "Author A", "Book A")
	if _, err := os.Stat(bookATarget); err == nil {
		t.Errorf("expected BookA NOT to be moved, but %s exists", bookATarget)
	}

	bookCTarget := filepath.Join(outputDir, "Author C", "Book C")
	if _, err := os.Stat(bookCTarget); err == nil {
		t.Errorf("expected BookC NOT to be moved, but %s exists", bookCTarget)
	}
}

func TestOrganizerAllowedSourcePaths_Empty_ProcessesAll(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := t.TempDir()

	createBookDir(t, baseDir, "BookA", "Book A", "Author A")
	createBookDir(t, baseDir, "BookB", "Book B", "Author B")
	createBookDir(t, baseDir, "BookC", "Book C", "Author C")

	config := OrganizerConfig{
		BaseDir:            baseDir,
		OutputDir:          outputDir,
		DryRun:             false,
		FieldMapping:       DefaultFieldMapping(),
		AllowedSourcePaths: []string{}, // empty = process all
	}

	org, err := NewOrganizer(&config)
	if err != nil {
		t.Fatalf("NewOrganizer() error = %v", err)
	}

	if err := org.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// All three books should appear in the output directory.
	for _, tc := range []struct{ dir, author, title string }{
		{"BookA", "Author A", "Book A"},
		{"BookB", "Author B", "Book B"},
		{"BookC", "Author C", "Book C"},
	} {
		target := filepath.Join(outputDir, tc.author, tc.title)
		if _, err := os.Stat(target); os.IsNotExist(err) {
			t.Errorf("expected %s to be moved to %s, but it does not exist", tc.dir, target)
		}
	}
}

// --- Change 2: Absolute paths in log ---

func TestOrganizerExecute_AbsolutePathsInLog(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	createBookDir(t, inputDir, "BookA", "Book A", "Author A")

	config := OrganizerConfig{
		BaseDir:      inputDir,
		OutputDir:    outputDir,
		DryRun:       false,
		FieldMapping: DefaultFieldMapping(),
	}

	org, err := NewOrganizer(&config)
	if err != nil {
		t.Fatalf("NewOrganizer() error = %v", err)
	}

	if err := org.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Read the log file from the output directory.
	logPath := filepath.Join(outputDir, LogFileName)
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file at %s: %v", logPath, err)
	}

	var entries []LogEntry
	if err := json.Unmarshal(logData, &entries); err != nil {
		t.Fatalf("failed to parse log file: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("log file contains no entries")
	}

	for i, entry := range entries {
		if !filepath.IsAbs(entry.SourcePath) {
			t.Errorf("log entry[%d].SourcePath %q is not an absolute path", i, entry.SourcePath)
		}
		if !filepath.IsAbs(entry.TargetPath) {
			t.Errorf("log entry[%d].TargetPath %q is not an absolute path", i, entry.TargetPath)
		}
	}
}

// --- Change 3: FilePair — store original + target filename in log ---

func TestOrganizerFilePairLog_StoresBothNames(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := t.TempDir()

	// Create a book directory manually so we control the filename.
	bookDir := filepath.Join(baseDir, "MyBook")
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		t.Fatalf("failed to create book dir: %v", err)
	}

	// Metadata with a TrackNumber so the organizer adds a track prefix.
	meta := map[string]interface{}{
		"title":        "My Book",
		"authors":      []string{"My Author"},
		"track_number": 1,
	}
	metaBytes, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(bookDir, "metadata.json"), metaBytes, 0644); err != nil {
		t.Fatalf("failed to write metadata.json: %v", err)
	}

	originalName := "original_name.mp3"
	if err := os.WriteFile(filepath.Join(bookDir, originalName), []byte("fake audio"), 0644); err != nil {
		t.Fatalf("failed to write audio file: %v", err)
	}

	config := OrganizerConfig{
		BaseDir:      baseDir,
		OutputDir:    outputDir,
		DryRun:       false,
		FieldMapping: DefaultFieldMapping(),
	}

	org, err := NewOrganizer(&config)
	if err != nil {
		t.Fatalf("NewOrganizer() error = %v", err)
	}

	if err := org.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	logPath := filepath.Join(outputDir, LogFileName)
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var entries []LogEntry
	if err := json.Unmarshal(logData, &entries); err != nil {
		t.Fatalf("failed to parse log file: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("log file contains no entries")
	}

	// Find the FilePair for the audio file.
	var found bool
	for _, entry := range entries {
		for _, fp := range entry.Files {
			if fp.From == originalName {
				found = true
				// With track_number=1, the file should get a "01 - " prefix.
				expectedTo := "01 - " + originalName
				if fp.To != expectedTo {
					t.Errorf("FilePair.To = %q, want %q", fp.To, expectedTo)
				}
			}
		}
	}
	if !found {
		t.Errorf("no FilePair with From == %q found in log entries", originalName)
	}
}

func TestOrganizerUndoRestoresOriginalFilename(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := t.TempDir()

	bookDir := filepath.Join(baseDir, "MyBook")
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		t.Fatalf("failed to create book dir: %v", err)
	}

	meta := map[string]interface{}{
		"title":        "My Book",
		"authors":      []string{"My Author"},
		"track_number": 1,
	}
	metaBytes, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(bookDir, "metadata.json"), metaBytes, 0644); err != nil {
		t.Fatalf("failed to write metadata.json: %v", err)
	}

	originalName := "original_name.mp3"
	originalContent := []byte("fake audio content for undo test")
	if err := os.WriteFile(filepath.Join(bookDir, originalName), originalContent, 0644); err != nil {
		t.Fatalf("failed to write audio file: %v", err)
	}

	// Step 1: organize (rename + move).
	config := OrganizerConfig{
		BaseDir:      baseDir,
		OutputDir:    outputDir,
		DryRun:       false,
		FieldMapping: DefaultFieldMapping(),
	}

	org, err := NewOrganizer(&config)
	if err != nil {
		t.Fatalf("NewOrganizer() error = %v", err)
	}

	if err := org.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify the renamed file exists in output.
	renamedName := "01 - " + originalName
	targetBookDir := filepath.Join(outputDir, "My Author", "My Book")
	renamedPath := filepath.Join(targetBookDir, renamedName)
	if _, err := os.Stat(renamedPath); os.IsNotExist(err) {
		t.Fatalf("expected renamed file at %s, but not found", renamedPath)
	}

	// Step 2: undo — should restore to bookDir with the ORIGINAL filename.
	undoConfig := OrganizerConfig{
		BaseDir:      baseDir,
		OutputDir:    outputDir,
		DryRun:       false,
		Undo:         true,
		FieldMapping: DefaultFieldMapping(),
	}

	undoOrg, err := NewOrganizer(&undoConfig)
	if err != nil {
		t.Fatalf("NewOrganizer() undo error = %v", err)
	}

	if err := undoOrg.Execute(); err != nil {
		t.Fatalf("Execute() undo error = %v", err)
	}

	// The file should be back in bookDir with the ORIGINAL name, not the renamed name.
	restoredPath := filepath.Join(bookDir, originalName)
	if _, err := os.Stat(restoredPath); os.IsNotExist(err) {
		t.Errorf("expected file restored to %s, but not found", restoredPath)
	}

	// The renamed file should no longer exist in the target.
	if _, err := os.Stat(renamedPath); err == nil {
		t.Errorf("renamed file %s should not exist after undo", renamedPath)
	}

	// File content should be preserved.
	restoredContent, err := os.ReadFile(restoredPath)
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if string(restoredContent) != string(originalContent) {
		t.Errorf("restored file content mismatch: got %q, want %q", string(restoredContent), string(originalContent))
	}
}
