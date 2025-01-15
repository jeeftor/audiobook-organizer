package organizer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogFileCreation(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	metadata := Metadata{
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

	org := New(
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
	logPath := filepath.Join(tempDir, LogFileName)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("log file was not created")
	}

	// Check log content
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	var logEntries []LogEntry
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

func TestUndoMoves(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test file and metadata
	metadata := Metadata{
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
	testData := []byte("test data")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	// First, organize the files
	org := New(
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

	// Verify files were moved
	targetPath := filepath.Join(tempDir, "Test Author/Test Book")
	movedFile := filepath.Join(targetPath, "test.mp3")
	if _, err := os.Stat(movedFile); os.IsNotExist(err) {
		t.Fatal("file was not moved to target location")
	}

	// Now undo the moves
	undoOrg := New(
		tempDir,
		"",    // outputDir
		"",    // replaceSpace
		false, // verbose
		false, // dryRun
		true,  // undo
		false, // prompt
	)

	if err := undoOrg.Execute(); err != nil {
		t.Fatal(err)
	}

	// Verify files were moved back
	restoredFile := filepath.Join(sourceDir, "test.mp3")
	if _, err := os.Stat(restoredFile); os.IsNotExist(err) {
		t.Error("file was not restored to original location")
	}

	// Verify target directory is empty or removed
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		entries, _ := os.ReadDir(targetPath)
		if len(entries) > 0 {
			t.Error("target directory still contains files after undo")
		}
	}

	// Verify file contents are preserved
	restoredData, err := os.ReadFile(restoredFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(restoredData) != string(testData) {
		t.Error("restored file contents do not match original")
	}

	// Verify log file was removed
	logPath := filepath.Join(tempDir, LogFileName)
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("log file was not removed after undo")
	}
}

func TestLogFileInOutputDirectory(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	outputDir := filepath.Join(tempDir, "output")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}

	metadata := Metadata{
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

	org := New(
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

	// Check log file is in output directory
	logPath := filepath.Join(outputDir, LogFileName)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("log file was not created in output directory")
	}

	// Verify log contents
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	var logEntries []LogEntry
	if err := json.Unmarshal(logData, &logEntries); err != nil {
		t.Error("invalid log file format")
	}

	if len(logEntries) == 0 {
		t.Error("log file is empty")
	}
}
