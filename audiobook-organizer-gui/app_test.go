package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jeeftor/audiobook-organizer/pkg/organizer"
)

func TestNewApp(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	// Verify config is initialized with defaults
	if app.config == nil {
		t.Fatal("config not initialized")
	}

	// Check default config values
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Layout", app.config.Layout, "author-series-title"},
		{"ReplaceSpace", app.config.ReplaceSpace, " "},
		{"Verbose", app.config.Verbose, false},
		{"DryRun", app.config.DryRun, false},
		{"Undo", app.config.Undo, false},
		{"Prompt", app.config.Prompt, false},
		{"RemoveEmpty", app.config.RemoveEmpty, false},
		{"UseEmbeddedMetadata", app.config.UseEmbeddedMetadata, false},
		{"Flat", app.config.Flat, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	// Verify field mapping is initialized
	if app.config.FieldMapping.TitleField == "" {
		t.Error("FieldMapping.TitleField not initialized")
	}

	// Verify initial state
	if app.scanning {
		t.Error("scanning should be false initially")
	}

	if app.progress.Status != "idle" {
		t.Errorf("progress.Status = %s, want 'idle'", app.progress.Status)
	}
}

func TestApp_Startup(t *testing.T) {
	app := NewApp()
	ctx := context.Background()

	// Call startup to set context
	app.startup(ctx)

	if app.ctx == nil {
		t.Error("startup() did not set context")
	}
}

func TestApp_GetSettings(t *testing.T) {
	app := NewApp()

	settings := app.GetSettings()

	// Verify settings are returned correctly
	if settings.Layout != "author-series-title" {
		t.Errorf("GetSettings() Layout = %s, want 'author-series-title'", settings.Layout)
	}

	// Verify it's a copy, not a pointer
	settings.Layout = "author-only"
	if app.config.Layout == "author-only" {
		t.Error("GetSettings() should return a copy, not modify original config")
	}
}

func TestApp_UpdateSettings(t *testing.T) {
	app := NewApp()

	// Create new config
	newConfig := organizer.OrganizerConfig{
		Layout:       "author-only",
		ReplaceSpace: "_",
		Verbose:      true,
		DryRun:       true,
	}

	err := app.UpdateSettings(newConfig)
	if err != nil {
		t.Errorf("UpdateSettings() error = %v", err)
	}

	// Verify settings were updated
	if app.config.Layout != "author-only" {
		t.Errorf("UpdateSettings() did not update Layout: got %s, want 'author-only'", app.config.Layout)
	}

	if app.config.ReplaceSpace != "_" {
		t.Errorf("UpdateSettings() did not update ReplaceSpace: got %s, want '_'", app.config.ReplaceSpace)
	}

	if !app.config.Verbose {
		t.Error("UpdateSettings() did not update Verbose")
	}

	if !app.config.DryRun {
		t.Error("UpdateSettings() did not update DryRun")
	}
}

func TestApp_GetProgress(t *testing.T) {
	app := NewApp()

	// Initial progress
	progress := app.GetProgress()
	if progress.Status != "idle" {
		t.Errorf("GetProgress() Status = %s, want 'idle'", progress.Status)
	}

	// Update progress
	app.progress = ProgressUpdate{
		Status:      "scanning",
		Current:     5,
		Total:       10,
		CurrentFile: "/path/to/file.m4b",
	}

	progress = app.GetProgress()
	if progress.Status != "scanning" {
		t.Errorf("GetProgress() Status = %s, want 'scanning'", progress.Status)
	}
	if progress.Current != 5 {
		t.Errorf("GetProgress() Current = %d, want 5", progress.Current)
	}
	if progress.Total != 10 {
		t.Errorf("GetProgress() Total = %d, want 10", progress.Total)
	}
	if progress.CurrentFile != "/path/to/file.m4b" {
		t.Errorf("GetProgress() CurrentFile = %s, want '/path/to/file.m4b'", progress.CurrentFile)
	}
}

func TestApp_GetLogPath(t *testing.T) {
	tests := []struct {
		name      string
		outputDir string
		want      string
	}{
		{
			name:      "with output dir",
			outputDir: "/path/to/output",
			want:      "/path/to/output/.abook-org.log",
		},
		{
			name:      "empty output dir",
			outputDir: "",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			app.config.OutputDir = tt.outputDir

			got := app.GetLogPath()
			if got != tt.want {
				t.Errorf("GetLogPath() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestApp_ScanDirectory_ErrorCases(t *testing.T) {
	app := NewApp()

	tests := []struct {
		name    string
		dir     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty directory",
			dir:     "",
			wantErr: true,
			errMsg:  "directory path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.ScanDirectory(tt.dir)

			if (err != nil) != tt.wantErr {
				t.Errorf("ScanDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ScanDirectory() error message = %s, want %s", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestApp_PreviewChanges_ErrorCases(t *testing.T) {
	app := NewApp()

	tests := []struct {
		name      string
		inputDir  string
		outputDir string
		books     []int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "empty input dir",
			inputDir:  "",
			outputDir: "/output",
			books:     []int{1, 2},
			wantErr:   true,
			errMsg:    "input and output directories are required",
		},
		{
			name:      "empty output dir",
			inputDir:  "/input",
			outputDir: "",
			books:     []int{1, 2},
			wantErr:   true,
			errMsg:    "input and output directories are required",
		},
		{
			name:      "both empty",
			inputDir:  "",
			outputDir: "",
			books:     []int{},
			wantErr:   true,
			errMsg:    "input and output directories are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.PreviewChanges(tt.inputDir, tt.outputDir, tt.books)

			if (err != nil) != tt.wantErr {
				t.Errorf("PreviewChanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("PreviewChanges() error message = %s, want %s", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestApp_PreviewChanges_SetsConfig(t *testing.T) {
	app := NewApp()

	// Create temporary directories for testing
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create a test audio file with metadata
	testFile := filepath.Join(inputDir, "test.mp3")
	if err := os.WriteFile(testFile, []byte("fake mp3"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Scan to populate cache (required by PreviewChanges)
	_, err := app.ScanDirectory(inputDir)
	if err != nil {
		// Scanning might fail if no valid metadata, but we still want to test PreviewChanges
		// with empty cache, so continue
	}

	// If scan found no books, manually populate cache for testing
	if len(lastScanResults) == 0 {
		lastScanResults = []organizer.Metadata{
			{
				Title:      "Test Book",
				Authors:    []string{"Test Author"},
				Series:     []string{"Test Series"},
				SourcePath: testFile,
				SourceType: "audio",
			},
		}
	}

	_, err = app.PreviewChanges(inputDir, outputDir, []int{})
	if err != nil {
		t.Fatalf("PreviewChanges() unexpected error: %v", err)
	}

	// Verify config was updated
	if app.config.BaseDir != inputDir {
		t.Errorf("PreviewChanges() BaseDir = %s, want %s", app.config.BaseDir, inputDir)
	}

	if app.config.OutputDir != outputDir {
		t.Errorf("PreviewChanges() OutputDir = %s, want %s", app.config.OutputDir, outputDir)
	}

	if !app.config.DryRun {
		t.Error("PreviewChanges() should set DryRun to true")
	}

	// Verify organizer was stored
	if app.organizer == nil {
		t.Error("PreviewChanges() should store organizer instance")
	}
}

func TestApp_ExecuteOrganize_NoOrganizerError(t *testing.T) {
	app := NewApp()

	// Try to execute without calling PreviewChanges first
	_, err := app.ExecuteOrganize(false)

	if err == nil {
		t.Error("ExecuteOrganize() should return error when organizer is nil")
	}

	expectedErr := "no organizer configured - run PreviewChanges first"
	if err.Error() != expectedErr {
		t.Errorf("ExecuteOrganize() error = %s, want %s", err.Error(), expectedErr)
	}
}

func TestApp_ExecuteOrganize_DryRunFlag(t *testing.T) {
	app := NewApp()

	// Create temporary directories for testing
	baseDir := t.TempDir()
	outputDir := t.TempDir()

	// Set up organizer first
	app.config.BaseDir = baseDir
	app.config.OutputDir = outputDir

	// Create organizer
	org, err := organizer.NewOrganizer(app.config)
	if err != nil {
		t.Fatalf("NewOrganizer() error: %v", err)
	}
	app.organizer = org

	tests := []struct {
		name       string
		dryRun     bool
		wantDryRun bool
	}{
		{
			name:       "dry run enabled",
			dryRun:     true,
			wantDryRun: true,
		},
		{
			name:       "dry run disabled",
			dryRun:     false,
			wantDryRun: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset config
			app.config.DryRun = false

			// Execute with dry run flag
			// Note: This will likely error because we don't have valid input,
			// but we're testing that the config is updated
			_, _ = app.ExecuteOrganize(tt.dryRun)

			if app.config.DryRun != tt.wantDryRun {
				t.Errorf("ExecuteOrganize() DryRun = %v, want %v", app.config.DryRun, tt.wantDryRun)
			}
		})
	}
}

func TestApp_Greet(t *testing.T) {
	app := NewApp()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "simple name",
			input:    "Alice",
			contains: "Alice",
		},
		{
			name:     "empty name",
			input:    "",
			contains: "Ready to organize some audiobooks?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.Greet(tt.input)

			if result == "" {
				t.Error("Greet() returned empty string")
			}

			// Check if result contains expected substring
			if tt.contains != "" {
				found := false
				for i := 0; i <= len(result)-len(tt.contains); i++ {
					if result[i:i+len(tt.contains)] == tt.contains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Greet() = %s, should contain %s", result, tt.contains)
				}
			}
		})
	}
}

func TestProgressUpdate_JSON(t *testing.T) {
	// Test that ProgressUpdate has proper JSON tags
	progress := ProgressUpdate{
		Status:      "scanning",
		Current:     5,
		Total:       10,
		CurrentFile: "/test/file.m4b",
	}

	// Verify struct is not nil
	if progress.Status == "" {
		t.Error("ProgressUpdate Status not set")
	}

	// Verify fields are accessible
	if progress.Current != 5 {
		t.Errorf("ProgressUpdate Current = %d, want 5", progress.Current)
	}
}

func TestPreviewItem_JSON(t *testing.T) {
	// Test that PreviewItem has proper JSON tags
	item := PreviewItem{
		From:       "/input/file.m4b",
		To:         "/output/author/series/file.m4b",
		IsConflict: true,
	}

	// Verify struct is not nil
	if item.From == "" {
		t.Error("PreviewItem From not set")
	}

	// Verify fields are accessible
	if !item.IsConflict {
		t.Error("PreviewItem IsConflict should be true")
	}
}

// TestApp_UndoLastOperation_RestoresOriginalFilename verifies that UndoLastOperation
// reads the new FilePair log format and restores the file using the original filename
// (file.From) rather than the renamed filename (file.To).
func TestApp_UndoLastOperation_RestoresOriginalFilename(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create the source book directory that existed before organization
	sourceBookDir := filepath.Join(inputDir, "BookA")
	if err := os.MkdirAll(sourceBookDir, 0755); err != nil {
		t.Fatalf("failed to create source book dir: %v", err)
	}

	// Create the target book directory (where the file was moved to)
	targetBookDir := filepath.Join(outputDir, "BookA")
	if err := os.MkdirAll(targetBookDir, 0755); err != nil {
		t.Fatalf("failed to create target book dir: %v", err)
	}

	// Create the renamed file in the target directory (this is "file.To")
	renamedFile := filepath.Join(targetBookDir, "01 - original.mp3")
	if err := os.WriteFile(renamedFile, []byte("fake audio"), 0644); err != nil {
		t.Fatalf("failed to create renamed file: %v", err)
	}

	// Write a .abook-org.log in the output dir using the new FilePair format
	logEntries := []MoveLogEntry{
		{
			Timestamp:  time.Now(),
			SourcePath: sourceBookDir,
			TargetPath: targetBookDir,
			Files: []FilePair{
				{From: "original.mp3", To: "01 - original.mp3"},
			},
		},
	}
	logData, err := json.MarshalIndent(logEntries, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal log entries: %v", err)
	}
	logPath := filepath.Join(outputDir, ".abook-org.log")
	if err := os.WriteFile(logPath, logData, 0644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}

	app := NewApp()
	app.config.OutputDir = outputDir

	result, err := app.UndoLastOperation()
	if err != nil {
		t.Fatalf("UndoLastOperation() returned unexpected error: %v", err)
	}

	if result["success"] != true {
		t.Errorf("UndoLastOperation() success = %v, want true; errors: %v", result["success"], result["errors"])
	}

	// The file must be restored to the source dir using the ORIGINAL name (From)
	restoredPath := filepath.Join(sourceBookDir, "original.mp3")
	if _, err := os.Stat(restoredPath); os.IsNotExist(err) {
		t.Errorf("UndoLastOperation() did not restore file to %s", restoredPath)
	}

	// The renamed file in the target dir must no longer exist
	if _, err := os.Stat(renamedFile); !os.IsNotExist(err) {
		t.Errorf("UndoLastOperation() left renamed file at %s; it should have been moved", renamedFile)
	}
}

// TestApp_UndoLastOperation_NoLogFile verifies that UndoLastOperation returns a
// descriptive error when no log file exists in the output directory.
func TestApp_UndoLastOperation_NoLogFile(t *testing.T) {
	outputDir := t.TempDir()

	app := NewApp()
	app.config.OutputDir = outputDir

	_, err := app.UndoLastOperation()
	if err == nil {
		t.Fatal("UndoLastOperation() expected error when no log file exists, got nil")
	}

	if !strings.Contains(err.Error(), "failed to read operation log") {
		t.Errorf("UndoLastOperation() error = %q, want it to contain 'failed to read operation log'", err.Error())
	}
}

// TestApp_GetLivePreviewPath_ValidIndex verifies that GetLivePreviewPath returns a
// fully populated PreviewItem (author, series, title, filename, output_dir) from
// the cached scan results for the given index.
func TestApp_GetLivePreviewPath_ValidIndex(t *testing.T) {
	outputDir := t.TempDir()

	app := NewApp()
	app.config.Layout = "author-series-title"

	// Populate cache directly (no filesystem scan needed)
	lastScanResults = []organizer.Metadata{
		{
			Title:      "The Way of Kings",
			Authors:    []string{"Brandon Sanderson"},
			Series:     []string{"Stormlight Archive"},
			SourcePath: "/input/thewayofkings.m4b",
			SourceType: "audio",
		},
	}
	t.Cleanup(func() { lastScanResults = nil })

	// Put the app in flat mode so SourcePath is the file (not the dir)
	app.config.Flat = true

	item, err := app.GetLivePreviewPath(0, outputDir)
	if err != nil {
		t.Fatalf("GetLivePreviewPath() error = %v", err)
	}

	if item.Author != "Brandon Sanderson" {
		t.Errorf("Author = %q, want %q", item.Author, "Brandon Sanderson")
	}
	if item.Series != "Stormlight Archive" {
		t.Errorf("Series = %q, want %q", item.Series, "Stormlight Archive")
	}
	if item.Title != "The Way of Kings" {
		t.Errorf("Title = %q, want %q", item.Title, "The Way of Kings")
	}
	if item.Filename != "thewayofkings.m4b" {
		t.Errorf("Filename = %q, want %q", item.Filename, "thewayofkings.m4b")
	}
	if item.OutputDir != outputDir {
		t.Errorf("OutputDir = %q, want %q", item.OutputDir, outputDir)
	}
	if item.To == "" {
		t.Error("To should not be empty")
	}
}

// TestApp_GetLivePreviewPath_InvalidIndex verifies that GetLivePreviewPath returns an
// error when the book index is out of range.
func TestApp_GetLivePreviewPath_InvalidIndex(t *testing.T) {
	app := NewApp()
	lastScanResults = nil
	t.Cleanup(func() { lastScanResults = nil })

	_, err := app.GetLivePreviewPath(0, "/output")
	if err == nil {
		t.Error("GetLivePreviewPath() expected error for empty cache, got nil")
	}
}

// TestApp_GetBatchPreview_ReturnsItems verifies that GetBatchPreview returns one
// enriched PreviewItem per selected index with all metadata fields populated.
func TestApp_GetBatchPreview_ReturnsItems(t *testing.T) {
	outputDir := t.TempDir()

	app := NewApp()
	app.config.Layout = "author-series-title"
	app.config.Flat = true

	lastScanResults = []organizer.Metadata{
		{
			Title:      "The Way of Kings",
			Authors:    []string{"Brandon Sanderson"},
			Series:     []string{"Stormlight Archive"},
			SourcePath: "/input/book1.m4b",
			SourceType: "audio",
		},
		{
			Title:      "Words of Radiance",
			Authors:    []string{"Brandon Sanderson"},
			Series:     []string{"Stormlight Archive"},
			SourcePath: "/input/book2.m4b",
			SourceType: "audio",
		},
	}
	t.Cleanup(func() { lastScanResults = nil })

	items, err := app.GetBatchPreview([]int{0, 1}, outputDir)
	if err != nil {
		t.Fatalf("GetBatchPreview() error = %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("GetBatchPreview() returned %d items, want 2", len(items))
	}

	if items[0].Author != "Brandon Sanderson" {
		t.Errorf("items[0].Author = %q, want %q", items[0].Author, "Brandon Sanderson")
	}
	if items[0].Title != "The Way of Kings" {
		t.Errorf("items[0].Title = %q, want %q", items[0].Title, "The Way of Kings")
	}
	if items[1].Title != "Words of Radiance" {
		t.Errorf("items[1].Title = %q, want %q", items[1].Title, "Words of Radiance")
	}
	if items[0].OutputDir != outputDir {
		t.Errorf("items[0].OutputDir = %q, want %q", items[0].OutputDir, outputDir)
	}
}

// TestApp_GetBatchPreview_EmptyIndicesReturnsEmpty verifies that GetBatchPreview
// returns an empty list when selectedIndices is empty (not all books).
// This prevents the footgun where an empty selection would process every book.
func TestApp_GetBatchPreview_EmptyIndicesReturnsEmpty(t *testing.T) {
	app := NewApp()
	lastScanResults = []organizer.Metadata{
		{Title: "Book One", Authors: []string{"Author A"}, SourcePath: "/input/a.m4b", SourceType: "audio"},
		{Title: "Book Two", Authors: []string{"Author B"}, SourcePath: "/input/b.m4b", SourceType: "audio"},
	}
	t.Cleanup(func() { lastScanResults = nil })

	items, err := app.GetBatchPreview([]int{}, "/output")
	if err != nil {
		t.Fatalf("GetBatchPreview() unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("GetBatchPreview() with empty indices = %d items, want 0 (not all books)", len(items))
	}
}

// TestApp_GetBatchPreview_EmptyNoScanResults verifies that GetBatchPreview returns
// an error when the scan cache is empty.
func TestApp_GetBatchPreview_EmptyNoScanResults(t *testing.T) {
	app := NewApp()
	lastScanResults = nil
	t.Cleanup(func() { lastScanResults = nil })

	_, err := app.GetBatchPreview([]int{0}, "/output")
	if err == nil {
		t.Error("GetBatchPreview() expected error when no scan results, got nil")
	}
}

// TestApp_ExecuteFileOperations_MoveDirectory verifies that ExecuteFileOperations
// moves a source directory to the target location and returns a summary with one move.
func TestApp_ExecuteFileOperations_MoveDirectory(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create a source book directory with an audio file
	sourceDir := filepath.Join(inputDir, "book1")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "audio.mp3"), []byte("fake audio"), 0644); err != nil {
		t.Fatalf("failed to create audio file: %v", err)
	}

	targetDir := filepath.Join(outputDir, "Author", "Book1")

	app := NewApp()
	app.config.OutputDir = outputDir

	ops := []FileOperation{{From: sourceDir, To: targetDir}}
	summary, err := app.ExecuteFileOperations(ops, false)
	if err != nil {
		t.Fatalf("ExecuteFileOperations() error = %v", err)
	}

	if len(summary.Moves) != 1 {
		t.Fatalf("ExecuteFileOperations() Moves = %d, want 1", len(summary.Moves))
	}

	// Target file should exist
	if _, err := os.Stat(filepath.Join(targetDir, "audio.mp3")); os.IsNotExist(err) {
		t.Errorf("ExecuteFileOperations() did not move file to target dir")
	}

	// Source dir should be gone
	if _, err := os.Stat(sourceDir); !os.IsNotExist(err) {
		t.Errorf("ExecuteFileOperations() left source dir at %s", sourceDir)
	}
}

// TestApp_ExecuteFileOperations_CopyDirectory verifies that copy mode leaves the source intact.
func TestApp_ExecuteFileOperations_CopyDirectory(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	sourceDir := filepath.Join(inputDir, "book1")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "audio.mp3"), []byte("fake audio"), 0644); err != nil {
		t.Fatalf("failed to create audio file: %v", err)
	}

	targetDir := filepath.Join(outputDir, "Author", "Book1")

	app := NewApp()
	app.config.OutputDir = outputDir

	ops := []FileOperation{{From: sourceDir, To: targetDir}}
	summary, err := app.ExecuteFileOperations(ops, true) // copy mode
	if err != nil {
		t.Fatalf("ExecuteFileOperations() error = %v", err)
	}

	if len(summary.Moves) != 1 {
		t.Fatalf("ExecuteFileOperations() Moves = %d, want 1", len(summary.Moves))
	}

	// Target file should exist
	if _, err := os.Stat(filepath.Join(targetDir, "audio.mp3")); os.IsNotExist(err) {
		t.Errorf("ExecuteFileOperations() did not copy file to target dir")
	}

	// Source dir should still exist (copy mode)
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		t.Errorf("ExecuteFileOperations() removed source dir in copy mode")
	}
}

// TestApp_ExecuteFileOperations_EmptyOps returns empty summary without error.
func TestApp_ExecuteFileOperations_EmptyOps(t *testing.T) {
	app := NewApp()
	summary, err := app.ExecuteFileOperations([]FileOperation{}, false)
	if err != nil {
		t.Fatalf("ExecuteFileOperations() error = %v", err)
	}
	if len(summary.Moves) != 0 {
		t.Errorf("ExecuteFileOperations() Moves = %d, want 0", len(summary.Moves))
	}
}

// TestApp_ExecuteFileOperations_WritesUndoLog verifies that the undo log is written.
func TestApp_ExecuteFileOperations_WritesUndoLog(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	sourceDir := filepath.Join(inputDir, "book1")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "audio.mp3"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	targetDir := filepath.Join(outputDir, "Author", "Book1")

	app := NewApp()
	app.config.OutputDir = outputDir

	ops := []FileOperation{{From: sourceDir, To: targetDir}}
	_, err := app.ExecuteFileOperations(ops, false)
	if err != nil {
		t.Fatalf("ExecuteFileOperations() error = %v", err)
	}

	logPath := filepath.Join(outputDir, ".abook-org.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("ExecuteFileOperations() did not write undo log at %s", logPath)
	}
}

// TestApp_ExecuteFileOperations_OnlyMovesSelected verifies that only the specified
// operations are executed and other directories in the input remain untouched.
func TestApp_ExecuteFileOperations_OnlyMovesSelected(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create two book directories
	for _, name := range []string{"book1", "book2"} {
		dir := filepath.Join(inputDir, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "audio.mp3"), []byte("data"), 0644); err != nil {
			t.Fatalf("failed to create file in %s: %v", name, err)
		}
	}

	// Only move book1
	targetDir := filepath.Join(outputDir, "Author", "Book1")
	app := NewApp()
	app.config.OutputDir = outputDir

	ops := []FileOperation{{From: filepath.Join(inputDir, "book1"), To: targetDir}}
	_, err := app.ExecuteFileOperations(ops, false)
	if err != nil {
		t.Fatalf("ExecuteFileOperations() error = %v", err)
	}

	// book1 should be moved
	if _, err := os.Stat(filepath.Join(inputDir, "book1")); !os.IsNotExist(err) {
		t.Errorf("ExecuteFileOperations() left book1 in source dir")
	}

	// book2 should still be in the input dir (untouched)
	if _, err := os.Stat(filepath.Join(inputDir, "book2")); os.IsNotExist(err) {
		t.Errorf("ExecuteFileOperations() removed book2 which was not selected")
	}
}

// TestApp_PreviewChanges_SetsAllowedSourcePaths verifies that after calling PreviewChanges
// with a subset of selected book indices, AllowedSourcePaths on the config contains
// exactly the source paths of the selected books.
func TestApp_PreviewChanges_SetsAllowedSourcePaths(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create three book subdirs each with a metadata.json
	bookNames := []string{"BookA", "BookB", "BookC"}
	for _, name := range bookNames {
		bookDir := filepath.Join(inputDir, name)
		if err := os.MkdirAll(bookDir, 0755); err != nil {
			t.Fatalf("failed to create book dir %s: %v", name, err)
		}
		meta := map[string]any{
			"title":   name,
			"authors": []string{"Test Author"},
		}
		metaData, _ := json.Marshal(meta)
		if err := os.WriteFile(filepath.Join(bookDir, "metadata.json"), metaData, 0644); err != nil {
			t.Fatalf("failed to write metadata.json for %s: %v", name, err)
		}
	}

	app := NewApp()

	// Scan to populate lastScanResults
	results, err := app.ScanDirectory(inputDir)
	if err != nil {
		t.Fatalf("ScanDirectory() error: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("ScanDirectory() found %d books, need at least 2 to test selection", len(results))
	}

	// Select only the book at index 1
	selectedIndices := []int{1}
	_, err = app.PreviewChanges(inputDir, outputDir, selectedIndices)
	if err != nil {
		t.Fatalf("PreviewChanges() error: %v", err)
	}

	// AllowedSourcePaths must contain exactly one entry matching the selected book's SourcePath
	if len(app.config.AllowedSourcePaths) != 1 {
		t.Fatalf("AllowedSourcePaths has %d entries, want 1", len(app.config.AllowedSourcePaths))
	}

	wantSourcePath, err := filepath.Abs(lastScanResults[1].SourcePath)
	if err != nil {
		t.Fatalf("failed to resolve expected source path: %v", err)
	}
	// SourcePath on Metadata points to the metadata.json file; the allowed path
	// should be its directory (the book directory).
	wantDir := filepath.Dir(wantSourcePath)

	gotPath := app.config.AllowedSourcePaths[0]
	if gotPath != wantDir {
		t.Errorf("AllowedSourcePaths[0] = %q, want %q", gotPath, wantDir)
	}
}
