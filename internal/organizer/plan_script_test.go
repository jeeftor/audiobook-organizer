// internal/organizer/plan_script_test.go
package organizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanScriptWriter(t *testing.T) {
	// Create a temp directory for the test
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "test_plan.sh")

	// Create a plan writer
	writer := NewPlanScriptWriter(scriptPath, "/source/dir", "/output/dir")

	// Add some test moves
	metadata1 := &Metadata{
		Title:   "Test Book 1",
		Authors: []string{"Test Author"},
		Series:  []string{"Test Series"},
	}
	writer.AddMove("/source/dir/file1.mp3", "/output/dir/Author/Series/file1.mp3", metadata1)
	writer.AddMove("/source/dir/file2.mp3", "/output/dir/Author/Series/file2.mp3", metadata1)

	metadata2 := &Metadata{
		Title:   "Test Book 2",
		Authors: []string{"Another Author"},
	}
	writer.AddMove("/source/dir/file3.mp3", "/output/dir/Another/file3.mp3", metadata2)

	// Write the script
	if err := writer.WriteScript(); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	// Verify the script was created
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Fatal("Script file was not created")
	}

	// Read and verify the script content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read script: %v", err)
	}

	scriptContent := string(content)

	// Check for expected content
	expectedStrings := []string{
		"#!/bin/bash",
		"Audiobook Organizer - Move Plan Script",
		"Source Directory: /source/dir",
		"Output Directory: /output/dir",
		"Total Moves: 3",
		"move_file()",
		"DRY_RUN=${DRY_RUN:-0}",
		"# Book: Test Book 1",
		"# Author: Test Author",
		"# Book: Test Book 2",
		"# Author: Another Author",
		"Plan complete: 3 files processed",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(scriptContent, expected) {
			t.Errorf("Script missing expected content: %q", expected)
		}
	}

	// Verify the script is executable
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("Failed to stat script: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("Script is not executable")
	}
}

func TestPlanScriptWriterEmptyPath(t *testing.T) {
	// Test that empty path doesn't cause errors
	writer := NewPlanScriptWriter("", "/source", "/output")
	writer.AddMove("/source/file.mp3", "/output/file.mp3", nil)

	// WriteScript should return nil without doing anything
	if err := writer.WriteScript(); err != nil {
		t.Errorf("WriteScript with empty path should not error: %v", err)
	}
}

func TestPlanScriptWriterMoveCount(t *testing.T) {
	writer := NewPlanScriptWriter("", "/source", "/output")

	if writer.MoveCount() != 0 {
		t.Error("New writer should have 0 moves")
	}

	writer.AddMove("/source/file1.mp3", "/output/file1.mp3", nil)
	writer.AddMove("/source/file2.mp3", "/output/file2.mp3", nil)

	if writer.MoveCount() != 2 {
		t.Errorf("Expected 2 moves, got %d", writer.MoveCount())
	}
}

func TestPlanFileWriter(t *testing.T) {
	// Create a temp directory for the test
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_plan.txt")

	// Create a plan file writer
	writer := NewPlanFileWriter(filePath, "/source/dir", "/output/dir")

	// Add some test moves
	metadata1 := &Metadata{
		Title:   "Test Book 1",
		Authors: []string{"Test Author"},
		Series:  []string{"Test Series"},
	}
	writer.AddMove("/source/dir/file1.mp3", "/output/dir/Author/Series/file1.mp3", metadata1)
	writer.AddMove("/source/dir/file2.mp3", "/output/dir/Author/Series/file2.mp3", metadata1)

	// Write the file
	if err := writer.WriteFile(); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Plan file was not created")
	}

	// Read and verify the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	fileContent := string(content)

	// Check for expected content
	expectedStrings := []string{
		"AUDIOBOOK ORGANIZER - MOVE PLAN",
		"Source Directory: /source/dir",
		"Output Directory: /output/dir",
		"Total Files: 2",
		"BOOK: Test Book 1",
		"AUTHOR: Test Author",
		"FROM: /source/dir/file1.mp3",
		"TO: /output/dir/Author/Series/file1.mp3",
		"This is a DRY-RUN plan",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(fileContent, expected) {
			t.Errorf("File missing expected content: %q", expected)
		}
	}
}

func TestPlanFileWriterEmptyPath(t *testing.T) {
	// Test that empty path doesn't cause errors
	writer := NewPlanFileWriter("", "/source", "/output")
	writer.AddMove("/source/file.mp3", "/output/file.mp3", nil)

	// WriteFile should return nil without doing anything
	if err := writer.WriteFile(); err != nil {
		t.Errorf("WriteFile with empty path should not error: %v", err)
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "\"simple\""},
		{"with space", "'with space'"},
		{"with'quote", "'with'\"'\"'quote'"},
		{"path/to/file", "\"path/to/file\""},
		{"file with spaces.mp3", "'file with spaces.mp3'"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := shellQuote(tt.input)
			if result != tt.expected {
				t.Errorf("shellQuote(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
