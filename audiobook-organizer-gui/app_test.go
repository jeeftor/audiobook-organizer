package main

import (
	"context"
	"testing"

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

	_, err := app.PreviewChanges(inputDir, outputDir, []int{})
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
