//go:build integration

package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	for _, dir := range []string{inputDir, outputDir} {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	for _, file := range files {
		filePath := filepath.Join(inputDir, file.Name)
		err = os.WriteFile(filePath, []byte(file.Content), 0644)
		if err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create test file %s: %v", filePath, err)
		}

		// Set file modification time if specified
		if !file.ModTime.IsZero() {
			err := os.Chtimes(filePath, file.ModTime, file.ModTime)
			if err != nil {
				os.RemoveAll(tempDir)
				t.Fatalf("Failed to set file time for %s: %v", filePath, err)
			}
		}
	}

	return &testEnvironment{
		BaseDir:   tempDir,
		InputDir:  inputDir,
		OutputDir: outputDir,
		Cleanup:   func() { os.RemoveAll(tempDir) },
	}
}

// testFile represents a test file to be created
type testFile struct {
	Name    string
	Content string
	ModTime time.Time
}

func TestOrganizer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name           string
		setup          func(t *testing.T) *testEnvironment
		expectedOutput string
		expectedError  bool
	}{
		{
			name: "simple integration test",
			setup: func(t *testing.T) *testEnvironment {
				files := []testFile{
					{
						Name:    "test.mp3",
						Content: "test content",
					},
				}
				return setupTestEnvironment(t, files)
			},
			expectedOutput: "test.mp3",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := tt.setup(t)
			defer env.Cleanup()

			// Create a test organizer
			config := &organizer.OrganizerConfig{
				BaseDir:      env.InputDir,
				OutputDir:    env.OutputDir,
				Verbose:      true,
				DryRun:       true,
				UseFileTimes: true,
			}

			org := organizer.NewOrganizer(config)

			// Run the organizer
			err := org.Organize()

			// Verify results
			if tt.expectedError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
				// Add more assertions based on expected output
			}
		})
	}
}
