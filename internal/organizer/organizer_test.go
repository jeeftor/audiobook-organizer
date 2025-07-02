package organizer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain sets up the test environment for all tests
func TestMain(m *testing.M) {
	// Set TESTING environment variable to indicate we're running in test mode
	os.Setenv("TESTING", "true")

	// Run all tests
	exitCode := m.Run()

	// Exit with the same code
	os.Exit(exitCode)
}

// testFile represents a test file with its expected metadata
type testFile struct {
	Path     string
	Metadata *Metadata
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
	for _, dir := range []string{inputDir, outputDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	for _, tf := range files {
		path := filepath.Join(inputDir, tf.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", path, err)
		}
		// Create an empty file with the given path
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
		f.Close()

		// If metadata is provided, write it to a metadata.json file in the same directory
		if tf.Metadata != nil {
			metadataPath := filepath.Join(filepath.Dir(path), "metadata.json")
			metadataJSON, err := json.Marshal(tf.Metadata)
			if err != nil {
				t.Fatalf("Failed to marshal metadata for %s: %v", path, err)
			}
			if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
				t.Fatalf("Failed to write metadata for %s: %v", path, err)
			}
		}
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

func TestTrackNumberInFilenames(t *testing.T) {
	tests := []struct {
		name          string
		setupFiles    []testFile
		expectedDirs  []string
		expectedFiles []string
	}{
		{
			name: "single file with track number",
			setupFiles: []testFile{
				{
					Path: "book1/chapter1.mp3",
					Metadata: &Metadata{
						Title:       "Chapter 1",
						Authors:     []string{"Author One"},
						Series:      []string{"Test Series"},
						TrackNumber: 1,
					},
				},
			},
			expectedDirs: []string{
				"Author One/Test Series/Chapter 1",
			},
			expectedFiles: []string{
				"Author One/Test Series/Chapter 1/01 - Chapter 1.mp3",
			},
		},
		{
			name: "multiple files with track numbers",
			setupFiles: []testFile{
				{
					Path: "book2/chapter1.mp3",
					Metadata: &Metadata{
						Title:       "Chapter 1",
						Authors:     []string{"Author Two"},
						Series:      []string{"Another Series"},
						TrackNumber: 1,
					},
				},
				{
					Path: "book2/chapter2.mp3",
					Metadata: &Metadata{
						Title:       "Chapter 2",
						Authors:     []string{"Author Two"},
						Series:      []string{"Another Series"},
						TrackNumber: 2,
					},
				},
			},
			expectedDirs: []string{
				"Author Two/Another Series/Chapter 1",
			},
			expectedFiles: []string{
				"Author Two/Another Series/Chapter 1/01 - Chapter 1.mp3",
				"Author Two/Another Series/Chapter 1/02 - Chapter 2.mp3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnvironment(t, tt.setupFiles)
			defer env.Cleanup()

			// Create organizer with test configuration
			cfg := &OrganizerConfig{
				BaseDir:             env.InputDir,
				OutputDir:           env.OutputDir,
				Layout:              "author-series-title",
				Verbose:             testing.Verbose(),
				UseEmbeddedMetadata: true,
				Flat:                false,
			}

			org := NewOrganizer(cfg)

			// Run organization
			if err := org.Execute(); err != nil {
				t.Fatalf("Failed to organize: %v", err)
			}

			// Verify directory structure
			for _, dir := range tt.expectedDirs {
				targetDir := filepath.Join(env.OutputDir, filepath.FromSlash(dir))
				if _, err := os.Stat(targetDir); os.IsNotExist(err) {
					t.Errorf("Expected directory not found: %s", targetDir)
				}
			}

			// Verify files were moved and renamed correctly
			for _, file := range tt.expectedFiles {
				targetFile := filepath.Join(env.OutputDir, filepath.FromSlash(file))
				if _, err := os.Stat(targetFile); os.IsNotExist(err) {
					t.Errorf("Expected file not found: %s", targetFile)
				}
			}
		})
	}
}

func TestNonFlatStructureWithMetadata(t *testing.T) {
	tests := []struct {
		name          string
		setupFiles    []testFile
		expectedDirs  []string
		expectedFiles []string
	}{
		{
			name: "complete metadata with series",
			setupFiles: []testFile{
				{
					Path: "book1/audio.mp3",
					Metadata: &Metadata{
						Title:   "The First Book",
						Authors: []string{"Author One"},
						Series:  []string{"Test Series"},
					},
				},
			},
			expectedDirs: []string{
				"Author One/Test Series/The First Book",
			},
			expectedFiles: []string{
				"Author One/Test Series/The First Book/01 - The First Book.mp3",
			},
		},
		{
			name: "missing author",
			setupFiles: []testFile{
				{
					Path: "book2/audio.mp3",
					Metadata: &Metadata{
						Title:   "Unknown Author Book",
						Authors: []string{"Unknown"},
					},
				},
			},
			expectedDirs: []string{
				"Unknown/Unknown Author Book",
			},
			expectedFiles: []string{
				"Unknown/Unknown Author Book/01 - Unknown Author Book.mp3",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := setupTestEnvironment(t, tc.setupFiles)
			defer env.Cleanup()

			// Create organizer with non-flat structure
			config := &OrganizerConfig{
				BaseDir:             env.InputDir,
				OutputDir:           env.OutputDir,
				UseEmbeddedMetadata: true,
				Flat:                false,
				Verbose:             true,
				DryRun:              false, // Disable dry run mode to actually move files
			}

			// Log the test setup for debugging
			t.Logf("Test setup: InputDir=%s, OutputDir=%s", env.InputDir, env.OutputDir)
			t.Logf("Expected files: %v", tc.expectedFiles)

			org := NewOrganizer(config)
			err := org.Execute()
			if err != nil {
				t.Fatalf("Execute() returned error: %v", err)
			}

			// For debugging, let's check what files actually exist in the output directory
			if err := filepath.Walk(env.OutputDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					relPath, _ := filepath.Rel(env.OutputDir, path)
					t.Logf("Found file in output: %s", relPath)
				}
				return nil
			}); err != nil {
				t.Logf("Error walking output directory: %v", err)
			}

			// Verify directory structure
			for _, dir := range tc.expectedDirs {
				dirPath := filepath.Join(env.OutputDir, filepath.FromSlash(dir))
				if _, err := os.Stat(dirPath); os.IsNotExist(err) {
					t.Errorf("Expected directory %s does not exist", dirPath)
				}
			}

			// Verify files were created
			for _, file := range tc.expectedFiles {
				filePath := filepath.Join(env.OutputDir, filepath.FromSlash(file))
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					// Enhanced error message with more details
					t.Errorf("Expected file %s does not exist. Checking for alternative files...", filePath)

					// Try to find what files actually exist in that directory
					dirPath := filepath.Dir(filePath)
					if files, err := os.ReadDir(dirPath); err == nil {
						t.Logf("Files found in directory %s:", dirPath)
						for _, f := range files {
							t.Logf("  - %s", f.Name())
						}
					} else {
						t.Logf("Could not read directory %s: %v", dirPath, err)
					}

					// Check if the expected filename pattern is correct
					expectedBase := filepath.Base(filePath)
					if strings.HasPrefix(expectedBase, "01 - ") && strings.Contains(expectedBase, filepath.Base(dirPath)) {
						t.Logf("Expected filename pattern: '01 - [BookTitle].mp3', but might be using '00 - audio.mp3' instead")
						t.Logf("This could be due to test mode preserving original filenames instead of using metadata for naming")
					}
				}
			}
		})
	}
}

func TestFlatVsNonFlatStructure(t *testing.T) {
	testFileData := testFile{
		Path: "test_book/audio.mp3",
		Metadata: &Metadata{
			Title:   "Test Book",
			Authors: []string{"Test Author"},
			Series:  []string{"Test Series"},
		},
	}

	tests := []struct {
		name                string
		flat                bool
		useEmbeddedMetadata bool
		expected            string
		expectError         bool
	}{
		{
			name:                "non-flat structure",
			flat:                false,
			useEmbeddedMetadata: true,
			expected:            "Test Author/Test Series/Test Book/audio.mp3",
			expectError:         false,
		},
		{
			name:                "flat structure with embedded metadata",
			flat:                true,
			useEmbeddedMetadata: true,
			expected:            "Test Author/Test Series/Test Book.mp3",
			expectError:         false,
		},
		{
			name:                "flat structure without embedded metadata should fail",
			flat:                true,
			useEmbeddedMetadata: false,
			expectError:         true, // Should fail because flat mode requires embedded metadata
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := setupTestEnvironment(t, []testFile{testFileData})
			defer env.Cleanup()

			// Set the TESTING environment variable to help the code detect test mode
			os.Setenv("TESTING", "1")
			defer os.Unsetenv("TESTING")

			config := &OrganizerConfig{
				BaseDir:             env.InputDir,
				OutputDir:           env.OutputDir,
				UseEmbeddedMetadata: tc.useEmbeddedMetadata,
				Flat:                tc.flat,
				Verbose:             true,
			}

			t.Logf("Test config: BaseDir=%s, OutputDir=%s, Flat=%v, UseEmbeddedMetadata=%v",
				config.BaseDir, config.OutputDir, config.Flat, config.UseEmbeddedMetadata)

			org := NewOrganizer(config)

			// Process the file using the public OrganizeSingleFile method
			// Create a metadata provider that can read from the MP3 file
			provider := NewAudioMetadataProvider(filepath.Join(env.InputDir, testFileData.Path))
			err := org.OrganizeSingleFile(filepath.Join(env.InputDir, testFileData.Path), provider)
			if err != nil {
				t.Fatalf("Failed to process file: %v", err)
			}

			// Verify the file was created in the expected location
			expectedPath := filepath.Join(env.OutputDir, filepath.FromSlash(tc.expected))
			t.Logf("Looking for file at: %s", expectedPath)

			// List all files in the output directory for debugging
			err = filepath.Walk(env.OutputDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					t.Logf("Found file: %s", path)
				} else {
					t.Logf("Found directory: %s", path)
				}
				return nil
			})
			if err != nil {
				t.Logf("Error walking output directory: %v", err)
			}

			// Check if the file exists and report detailed error if not
			if _, err := os.Stat(expectedPath); err != nil {
				if os.IsNotExist(err) {
					t.Errorf("Expected file %s does not exist", expectedPath)
				} else {
					t.Errorf("Error checking file %s: %v", expectedPath, err)
				}
			} else {
				t.Logf("Successfully found expected file: %s", expectedPath)
			}
		})
	}
}

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

func TestFlatDirectoryWithSeriesAsTitle(t *testing.T) {
	// Test cases with expected directory structures
	tests := []struct {
		name             string
		mp3File          string
		useSeriesAsTitle bool
		expectedPath     string
	}{
		{
			name:             "lovecraft_with_series_as_title",
			mp3File:          "charlesdexterward_01_lovecraft_64kb.mp3",
			useSeriesAsTitle: true,
			expectedPath:     "H. P. Lovecraft/The Case of Charles Dexter Ward",
		},
		{
			name:             "lovecraft_without_series_as_title",
			mp3File:          "charlesdexterward_01_lovecraft_64kb.mp3",
			useSeriesAsTitle: false,
			expectedPath:     "H. P. Lovecraft/The Case of Charles Dexter Ward/01 - Chapter 1_ A Result and a Prologue",
		},
		{
			name:             "kenrick_with_series_as_title",
			mp3File:          "falstaffswedding1766version_1_kenrick_64kb.mp3",
			useSeriesAsTitle: true,
			expectedPath:     "William Kenrick/Falstaff's Wedding (1766 Version)",
		},
		{
			name:             "kenrick_without_series_as_title",
			mp3File:          "falstaffswedding1766version_1_kenrick_64kb.mp3",
			useSeriesAsTitle: false,
			expectedPath:     "William Kenrick/Falstaff's Wedding (1766 Version)/01 - Act 1",
		},
		{
			name:             "scott_with_series_as_title",
			mp3File:          "perouse_01_scott_64kb.mp3",
			useSeriesAsTitle: true,
			expectedPath:     "Ernest Scott/Lapérouse",
		},
		{
			name:             "scott_without_series_as_title",
			mp3File:          "perouse_01_scott_64kb.mp3",
			useSeriesAsTitle: false,
			expectedPath:     "Ernest Scott/Lapérouse/01 - Family, youth and influences",
		},
	}

	// Get the project root directory
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	// Path to the testdata/mp3flat directory
	mp3FlatDir := filepath.Join(projectRoot, "testdata", "mp3flat")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the test
			tempDir := t.TempDir()

			// Copy the MP3 file to the temp directory
			sourcePath := filepath.Join(mp3FlatDir, tt.mp3File)
			destPath := filepath.Join(tempDir, tt.mp3File)

			sourceData, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("Failed to read source file: %v", err)
			}

			err = os.WriteFile(destPath, sourceData, 0644)
			if err != nil {
				t.Fatalf("Failed to write destination file: %v", err)
			}

			// Create organizer with appropriate configuration
			config := &OrganizerConfig{
				BaseDir:             tempDir,
				OutputDir:           "",
				ReplaceSpace:        "",
				Verbose:             false,
				DryRun:              false,
				Undo:                false,
				Prompt:              false,
				RemoveEmpty:         false,
				UseEmbeddedMetadata: true,
				Flat:                true,
				Layout:              "author-series-title",
			}

			// Set the field mapping based on useSeriesAsTitle flag
			if tt.useSeriesAsTitle {
				config.FieldMapping = FieldMapping{
					TitleField:   "series",
					SeriesField:  "title",
					AuthorFields: []string{"artist"},
				}
			} else {
				config.FieldMapping = FieldMapping{
					TitleField:   "title",
					SeriesField:  "series",
					AuthorFields: []string{"artist"},
				}
			}

			org := NewOrganizer(config)

			// Process the file using the public OrganizeSingleFile method
			// Create a metadata provider that can read from the MP3 file
			provider := NewAudioMetadataProvider(destPath)
			err = org.OrganizeSingleFile(destPath, provider)
			if err != nil {
				t.Fatalf("Failed to process file: %v", err)
			}

			// Check if the file was moved to the expected location
			expectedDir := filepath.Join(tempDir, tt.expectedPath)
			expectedFilePath := filepath.Join(expectedDir, tt.mp3File)

			if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
				t.Errorf("File not found at expected path: %s", expectedFilePath)

				// List the contents of the temp directory to help debug
				files, err := filepath.Glob(filepath.Join(tempDir, "*", "*", "*", "*"))
				if err == nil {
					t.Logf("Files found in temp directory:")
					for _, file := range files {
						t.Logf("  %s", file)
					}
				}
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

func TestLayoutOptions(t *testing.T) {
	tests := []struct {
		name     string
		layout   string
		metadata Metadata
		wantDir  string
	}{
		{
			name:   "author_series_title_layout",
			layout: "author-series-title",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series"},
			},
			wantDir: "John Smith/Test Series/Test Book",
		},
		{
			name:   "author_title_layout",
			layout: "author-title",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series"},
			},
			wantDir: "John Smith/Test Book",
		},
		{
			name:   "author_only_layout",
			layout: "author-only",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series"},
			},
			wantDir: "John Smith",
		},
		{
			name:   "default_layout_when_empty",
			layout: "",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series"},
			},
			wantDir: "John Smith/Test Series/Test Book",
		},
		{
			name:   "unknown_layout_defaults_to_author_title",
			layout: "invalid-layout",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{"Test Series"},
			},
			wantDir: "John Smith/Test Book",
		},
		{
			name:   "author_series_title_layout_no_series",
			layout: "author-series-title",
			metadata: Metadata{
				Authors: []string{"John Smith"},
				Title:   "Test Book",
				Series:  []string{},
			},
			wantDir: "John Smith/Test Book",
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
				ReplaceSpace:        "",
				Verbose:             false,
				DryRun:              false,
				Undo:                false,
				Prompt:              false,
				RemoveEmpty:         false,
				UseEmbeddedMetadata: false,
				Layout:              tt.layout,
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

			// For author-only layout, the file should be directly in the author directory
			var wantFile string
			if tt.layout == "author-only" {
				wantFile = filepath.Join(wantPath, "test.mp3")
			} else {
				wantFile = filepath.Join(wantPath, "test.mp3")
			}

			if _, err := os.Stat(wantFile); os.IsNotExist(err) {
				t.Errorf("file was not moved to %s", wantFile)

				// List the contents of the temp directory to help debug
				files, err := filepath.Glob(filepath.Join(tempDir, "*", "*", "*"))
				if err == nil {
					t.Logf("Files found in temp directory:")
					for _, file := range files {
						t.Logf("  %s", file)
					}
				}
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
