//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// testFile represents a test file with its expected metadata
type testFile struct {
	Path     string
	Metadata *organizer.Metadata
}

// testEnvironment holds the test environment setup
type testEnvironment struct {
	BaseDir   string
	InputDir  string
	OutputDir string
	Cleanup   func()
}

// setupTestEnvironment creates a test environment with the given configuration
func setupTestEnvironment(t *testing.T, files []testFile) *testEnvironment {
	t.Helper()

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "audiobook-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	// Create input and output directories
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create test files
	for _, tf := range files {
		// Create any necessary subdirectories
		fullPath := filepath.Join(inputDir, tf.Path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", fullPath, err)
		}

		// Create an empty file
		file, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", fullPath, err)
		}
		file.Close()

		// TODO: Add metadata to the file if needed
	}

	return &testEnvironment{
		BaseDir:   tempDir,
		InputDir:  inputDir,
		OutputDir: outputDir,
		Cleanup: func() {
			os.RemoveAll(tempDir)
		},
	}
}

// assertFileExists checks if a file exists and fails the test if it doesn't
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist, but it doesn't", path)
	}
}

// assertFileNotExists checks if a file doesn't exist and fails the test if it does
func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected file %s to not exist, but it does", path)
	}
}
