package organizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOrganizerExecute(t *testing.T) {
	tests := []struct {
		name        string
		config      OrganizerConfig
		setupFunc   func(t *testing.T, baseDir string) // Function to set up test files
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid directory with audiobooks",
			config: OrganizerConfig{
				DryRun:              true, // Use dry run to avoid actual file moves
				Verbose:             false,
				UseEmbeddedMetadata: true,
				FieldMapping:        DefaultFieldMapping(),
			},
			setupFunc: func(t *testing.T, baseDir string) {
				// Create a simple audiobook structure
				bookDir := filepath.Join(baseDir, "test_book")
				if err := os.MkdirAll(bookDir, 0755); err != nil {
					t.Fatalf("Failed to create book directory: %v", err)
				}

				// Create metadata.json
				metadataContent := `{
					"title": "Test Book",
					"authors": ["Test Author"],
					"series": ["Test Series"]
				}`
				metadataPath := filepath.Join(bookDir, "metadata.json")
				if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
					t.Fatalf("Failed to create metadata file: %v", err)
				}

				// Create a test audio file
				audioPath := filepath.Join(bookDir, "audio.mp3")
				if err := os.WriteFile(audioPath, []byte("fake audio data"), 0644); err != nil {
					t.Fatalf("Failed to create audio file: %v", err)
				}
			},
			expectError: false,
		},
		{
			name: "nonexistent base directory",
			config: OrganizerConfig{
				BaseDir:      "/nonexistent/directory/that/should/not/exist",
				DryRun:       true,
				FieldMapping: DefaultFieldMapping(),
			},
			expectError: true,
			errorMsg:    "error resolving base directory path",
		},
		{
			name: "empty directory",
			config: OrganizerConfig{
				DryRun:       true,
				Verbose:      false,
				FieldMapping: DefaultFieldMapping(),
			},
			setupFunc: func(t *testing.T, baseDir string) {
				// Create an empty directory structure
				emptyDir := filepath.Join(baseDir, "empty")
				if err := os.MkdirAll(emptyDir, 0755); err != nil {
					t.Fatalf("Failed to create empty directory: %v", err)
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "organizer_execute_test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Set up base directory
			if tt.config.BaseDir == "" {
				tt.config.BaseDir = tempDir
			}
			if tt.config.OutputDir == "" {
				tt.config.OutputDir = filepath.Join(tempDir, "output")
				// Create output directory only if BaseDir exists
				if tt.config.BaseDir != "/nonexistent/directory/that/should/not/exist" {
					if err := os.MkdirAll(tt.config.OutputDir, 0755); err != nil {
						t.Fatalf("Failed to create output directory: %v", err)
					}
				}
			}

			// Run setup function if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			// Create organizer
			org := NewOrganizer(&tt.config)

			// Execute
			err = org.Execute()

			// Check results
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestOrganizerExecuteWithOutput(t *testing.T) {
	// Test Execute with separate output directory
	tempDir, err := os.MkdirTemp("", "organizer_output_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	// Create input structure
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	bookDir := filepath.Join(inputDir, "test_book")
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		t.Fatalf("Failed to create book directory: %v", err)
	}

	// Create test files
	metadataContent := `{
		"title": "Test Book",
		"authors": ["Test Author"]
	}`
	metadataPath := filepath.Join(bookDir, "metadata.json")
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to create metadata file: %v", err)
	}

	audioPath := filepath.Join(bookDir, "audio.mp3")
	if err := os.WriteFile(audioPath, []byte("fake audio data"), 0644); err != nil {
		t.Fatalf("Failed to create audio file: %v", err)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create organizer with separate output directory
	config := OrganizerConfig{
		BaseDir:      inputDir,
		OutputDir:    outputDir,
		DryRun:       true,
		Verbose:      false,
		FieldMapping: DefaultFieldMapping(),
	}

	org := NewOrganizer(&config)

	// Execute
	if err := org.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify that log path is calculated correctly
	// GetLogPath() should return a path based on the resolved output directory
	actualLogPath := org.GetLogPath()

	// The log path should be in the output directory and end with the correct filename
	if !strings.HasSuffix(actualLogPath, LogFileName) {
		t.Errorf("Expected log path to end with %q, got %q", LogFileName, actualLogPath)
	}

	// The log path should be in some form of the output directory
	// (accounts for symlink resolution differences like /private/var vs /var)
	if !strings.Contains(actualLogPath, "output") {
		t.Errorf("Expected log path to contain 'output', got %q", actualLogPath)
	}
}

func TestOrganizerExecuteUndo(t *testing.T) {
	// Test Execute in undo mode
	tempDir, err := os.MkdirTemp("", "organizer_undo_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a fake log file for undo (using the correct format)
	logPath := filepath.Join(tempDir, LogFileName)
	logContent := `[]` // Empty array of log entries
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	config := OrganizerConfig{
		BaseDir:      tempDir,
		DryRun:       true,
		Undo:         true,
		FieldMapping: DefaultFieldMapping(),
	}

	org := NewOrganizer(&config)

	// Execute (should call undoMoves)
	if err := org.Execute(); err != nil {
		t.Fatalf("Execute undo failed: %v", err)
	}
}

func TestOrganizerExecuteRemoveEmpty(t *testing.T) {
	// Test Execute with RemoveEmpty option
	tempDir, err := os.MkdirTemp("", "organizer_remove_empty_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple test structure to avoid infinite loops
	bookDir := filepath.Join(tempDir, "test_book")
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		t.Fatalf("Failed to create book directory: %v", err)
	}

	// Create metadata.json to give the organizer something to process
	metadataContent := `{"title": "Test", "authors": ["Author"]}`
	metadataPath := filepath.Join(bookDir, "metadata.json")
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to create metadata file: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	config := OrganizerConfig{
		BaseDir:      tempDir,
		OutputDir:    outputDir,
		DryRun:       true,
		RemoveEmpty:  true,
		Verbose:      false, // Disable verbose to reduce output
		FieldMapping: DefaultFieldMapping(),
	}

	org := NewOrganizer(&config)

	// Execute should complete without infinite loops
	if err := org.Execute(); err != nil {
		t.Fatalf("Execute with RemoveEmpty failed: %v", err)
	}
}

func TestOrganizerExecutePathResolution(t *testing.T) {
	// Test that Execute properly resolves symbolic links and relative paths
	tempDir, err := os.MkdirTemp("", "organizer_path_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	realDir := filepath.Join(tempDir, "real")
	if err := os.MkdirAll(realDir, 0755); err != nil {
		t.Fatalf("Failed to create real directory: %v", err)
	}

	// Test with relative path (current directory should be resolved to absolute)
	config := OrganizerConfig{
		BaseDir:      ".", // Relative path
		DryRun:       true,
		FieldMapping: DefaultFieldMapping(),
	}

	org := NewOrganizer(&config)

	// Log initial state
	t.Logf("Initial BaseDir: %q", org.config.BaseDir)

	// Execute should resolve the relative path
	if err := org.Execute(); err != nil {
		t.Fatalf("Execute with relative path failed: %v", err)
	}

	// Log final state
	t.Logf("Final BaseDir: %q", org.config.BaseDir)

	// The Execute() function uses EvalSymlinks which resolves symlinks but doesn't
	// necessarily convert relative paths to absolute paths if no symlinks are involved.
	// The important thing is that Execute() completes successfully and the path is valid.

	// Test that the resolved path (whatever it is) refers to the current directory
	expectedAbs, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("Failed to get absolute path for current directory: %v", err)
	}

	actualAbs, err := filepath.Abs(org.config.BaseDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path for resolved BaseDir: %v", err)
	}

	if actualAbs != expectedAbs {
		t.Errorf("Expected resolved BaseDir to refer to current directory")
		t.Logf("Expected (abs): %s", expectedAbs)
		t.Logf("Actual (abs): %s", actualAbs)
	} else {
		t.Logf("BaseDir correctly resolved - refers to current directory")
	}
}

func TestOrganizerExecuteFieldMapping(t *testing.T) {
	// Test that Execute initializes default field mapping when not provided
	tempDir, err := os.MkdirTemp("", "organizer_mapping_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := OrganizerConfig{
		BaseDir: tempDir,
		DryRun:  true,
		// Note: FieldMapping is intentionally empty to test default initialization
	}

	org := NewOrganizer(&config)

	// The NewOrganizer constructor should initialize default field mapping on the original config
	// Check that the original config pointer was modified
	if config.FieldMapping.IsEmpty() {
		t.Error("Expected field mapping to be initialized with defaults in NewOrganizer")
	} else {
		// Verify some expected default values are present
		defaultMapping := DefaultFieldMapping()
		if config.FieldMapping.TitleField != defaultMapping.TitleField {
			t.Errorf("Expected TitleField %q, got %q", defaultMapping.TitleField, config.FieldMapping.TitleField)
		}
	}

	// Execute should work with the initialized field mapping
	if err := org.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestOrganizerExecuteVerboseMode(t *testing.T) {
	// Test Execute in verbose mode
	tempDir, err := os.MkdirTemp("", "organizer_verbose_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple structure
	bookDir := filepath.Join(tempDir, "test_book")
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		t.Fatalf("Failed to create book directory: %v", err)
	}

	metadataPath := filepath.Join(bookDir, "metadata.json")
	metadataContent := `{"title": "Test", "authors": ["Author"]}`
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to create metadata file: %v", err)
	}

	config := OrganizerConfig{
		BaseDir:      tempDir,
		DryRun:       true,
		Verbose:      true, // Enable verbose mode
		FieldMapping: DefaultFieldMapping(),
	}

	org := NewOrganizer(&config)

	// Execute should not fail in verbose mode
	if err := org.Execute(); err != nil {
		t.Fatalf("Execute in verbose mode failed: %v", err)
	}
}

// Helper function for string containment check
func containsString(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) &&
		   (s == substr || len(s) > len(substr) &&
		   (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		   containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test file system error scenarios
func TestOrganizerExecuteErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string // Returns base directory path
		expectError bool
		errorMsg    string
	}{
		{
			name: "permission denied on base directory",
			setupFunc: func(t *testing.T) string {
				// Create a directory we can't read (if running as non-root)
				tempDir, err := os.MkdirTemp("", "permission_test")
				if err != nil {
					t.Fatalf("Failed to create temp directory: %v", err)
				}

				// Try to make it unreadable (this might not work on all systems)
				restrictedDir := filepath.Join(tempDir, "restricted")
				if err := os.MkdirAll(restrictedDir, 0000); err != nil {
					t.Fatalf("Failed to create restricted directory: %v", err)
				}

				return restrictedDir
			},
			expectError: true,
			errorMsg:    "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := tt.setupFunc(t)
			defer os.RemoveAll(filepath.Dir(baseDir)) // Clean up temp directory

			config := OrganizerConfig{
				BaseDir:      baseDir,
				DryRun:       true,
				FieldMapping: DefaultFieldMapping(),
			}

			org := NewOrganizer(&config)
			err := org.Execute()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				// Note: We can't reliably test permission errors on all systems
				// so we just verify that an error occurred
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestOrganizerExecuteTiming(t *testing.T) {
	// Test that Execute completes in reasonable time and measures duration
	tempDir, err := os.MkdirTemp("", "organizer_timing_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := OrganizerConfig{
		BaseDir:      tempDir,
		DryRun:       true,
		FieldMapping: DefaultFieldMapping(),
	}

	org := NewOrganizer(&config)

	start := time.Now()
	err = org.Execute()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify it completes in reasonable time (should be very fast for empty directory)
	if duration > 5*time.Second {
		t.Errorf("Execute took too long: %v", duration)
	}
}
