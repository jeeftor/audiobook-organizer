package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jeeftor/audiobook-organizer/pkg/organizer"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	config       *organizer.OrganizerConfig
	organizer    *organizer.Organizer
	scanning     bool
	progress     ProgressUpdate
	initialDirs  InitialDirectories
	logLevel     string
	verbose      bool
	renameConfig RenameConfig
}

// RenameConfig holds file rename template configuration
type RenameConfig struct {
	Enabled       bool   `json:"enabled"`
	Template      string `json:"template"`
	Preset        string `json:"preset"`
	Separator     string `json:"separator"`
	AuthorFormat  string `json:"author_format"`
	ReplaceSpaces bool   `json:"replace_spaces"`
	SpaceChar     string `json:"space_char"`
}

// InitialDirectories holds pre-set directories from CLI args
type InitialDirectories struct {
	InputDir  string `json:"input_dir"`
	OutputDir string `json:"output_dir"`
}

// ScanMode represents the different scanning modes available
type ScanMode struct {
	Name                string `json:"name"`
	UseEmbeddedMetadata bool   `json:"use_embedded_metadata"`
	Flat                bool   `json:"flat"`
	Description         string `json:"description"`
}

// GetAvailableScanModes returns all available scanning modes
func (a *App) GetAvailableScanModes() []ScanMode {
	return []ScanMode{
		{
			Name:                "metadata.json",
			UseEmbeddedMetadata: false,
			Flat:                false,
			Description:         "metadata.json files only",
		},
		{
			Name:                "embedded (directory)",
			UseEmbeddedMetadata: true,
			Flat:                false,
			Description:         "Embedded metadata - directories as albums",
		},
		{
			Name:                "embedded (file)",
			UseEmbeddedMetadata: true,
			Flat:                true,
			Description:         "Embedded metadata - each file separate",
		},
	}
}

// UpdateScanMode updates the scanning mode configuration
func (a *App) UpdateScanMode(modeName string) error {
	modes := a.GetAvailableScanModes()
	for _, mode := range modes {
		if mode.Name == modeName {
			a.config.UseEmbeddedMetadata = mode.UseEmbeddedMetadata
			a.config.Flat = mode.Flat

			// Update field mapping based on mode
			if mode.UseEmbeddedMetadata {
				a.config.FieldMapping = organizer.AudioFieldMapping()
			} else {
				a.config.FieldMapping = organizer.DefaultFieldMapping()
			}

			a.log("Updated scan mode to %s (UseEmbedded=%v, Flat=%v)",
				modeName, mode.UseEmbeddedMetadata, mode.Flat)
			return nil
		}
	}
	return fmt.Errorf("unknown scan mode: %s", modeName)
}

// GetCurrentScanMode returns the current scanning mode
func (a *App) GetCurrentScanMode() string {
	if a.config.UseEmbeddedMetadata && a.config.Flat {
		return "embedded (by file)"
	} else if a.config.UseEmbeddedMetadata && !a.config.Flat {
		return "embedded (by directory)"
	} else {
		return "metadata.json"
	}
}

// LayoutOption represents a layout/organization option
type LayoutOption struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetAvailableLayouts returns all available layout options
func (a *App) GetAvailableLayouts() []LayoutOption {
	return []LayoutOption{
		{
			Name:        "author-series-title",
			Description: "Author/Series/Title/ (default)",
		},
		{
			Name:        "author-series-title-number",
			Description: "Author/Series/#1 - Title/ (include series number)",
		},
		{
			Name:        "author-series",
			Description: "Author/Series/ (for multi-file books, files go directly in series folder)",
		},
		{
			Name:        "author-title",
			Description: "Author/Title/ (ignore series)",
		},
		{
			Name:        "author-only",
			Description: "Author/ (flatten all books)",
		},
	}
}

// GetCurrentLayout returns the current layout setting
func (a *App) GetCurrentLayout() string {
	if a.config.Layout == "" {
		return "author-series-title"
	}
	return a.config.Layout
}

// UpdateLayout updates the layout configuration
func (a *App) UpdateLayout(layout string) error {
	a.config.Layout = layout
	a.log("Updated layout to %s", layout)
	return nil
}

// GetCurrentAuthorFormat returns the current author format setting
func (a *App) GetCurrentAuthorFormat() string {
	if a.config.AuthorFormat == "" {
		return "preserve"
	}
	return a.config.AuthorFormat
}

// UpdateAuthorFormat updates the author format configuration
func (a *App) UpdateAuthorFormat(format string) error {
	a.config.AuthorFormat = format
	a.log("Updated author format to %s", format)
	return nil
}

// FieldMappingPreset represents a preset field mapping configuration
type FieldMappingPreset struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Mapping     organizer.FieldMapping `json:"mapping"`
}

// GetFieldMappingPresets returns available field mapping presets
func (a *App) GetFieldMappingPresets() []FieldMappingPreset {
	return []FieldMappingPreset{
		{
			Name:        "Audio",
			Description: "For embedded audio tags (MP3, M4B, etc.)",
			Mapping:     organizer.AudioFieldMapping(),
		},
		{
			Name:        "Default",
			Description: "For metadata.json files",
			Mapping:     organizer.DefaultFieldMapping(),
		},
		{
			Name:        "EPUB",
			Description: "For EPUB files",
			Mapping:     organizer.EpubFieldMapping(),
		},
	}
}

// UpdateFieldMapping updates the field mapping configuration
func (a *App) UpdateFieldMapping(mapping organizer.FieldMapping) error {
	a.config.FieldMapping = mapping
	a.log("Updated field mapping: %+v", mapping)
	return nil
}

// GetCurrentFieldMapping returns the current field mapping
func (a *App) GetCurrentFieldMapping() organizer.FieldMapping {
	return a.config.FieldMapping
}

// FieldMappingOption represents a field mapping configuration option
type FieldMappingOption struct {
	Field       string   `json:"field"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Options     []string `json:"options"`
	Current     string   `json:"current"`
}

// GetFieldMappingOptions returns configurable field mapping options
// Uses shared field option lists from organizer package
func (a *App) GetFieldMappingOptions() []FieldMappingOption {
	return []FieldMappingOption{
		{
			Field:       organizer.TitleFieldKey,
			Label:       "Title Field",
			Description: "Select field for title (title, album, series, or any available field)",
			Options:     organizer.TextFieldOptions(),
			Current:     a.config.FieldMapping.TitleField,
		},
		{
			Field:       organizer.SeriesFieldKey,
			Label:       "Series Field",
			Description: "Select field for series (series, album, title, or any available field)",
			Options:     organizer.TextFieldOptions(),
			Current:     a.config.FieldMapping.SeriesField,
		},
		{
			Field:       organizer.AuthorsFieldKey,
			Label:       "Author Fields (Priority Order)",
			Description: "Select fields in priority order (comma-separated: authors,artist,album_artist)",
			Options:     organizer.AuthorFieldOptions(),
			Current:     strings.Join(a.config.FieldMapping.AuthorFields, ","),
		},
		{
			Field:       organizer.TrackFieldKey,
			Label:       "Track Field",
			Description: "Field to use for track number",
			Options:     organizer.TrackFieldOptions(),
			Current:     a.config.FieldMapping.TrackField,
		},
		{
			Field:       organizer.DiscFieldKey,
			Label:       "Disc Field",
			Description: "Field to use for disc number",
			Options:     organizer.DiscFieldOptions(),
			Current:     a.config.FieldMapping.DiscField,
		},
	}
}

// prioritizeFields returns all fields with common ones first
func prioritizeFields(allFields []string, prioritized []string) []string {
	result := []string{}
	seen := make(map[string]bool)

	// Add prioritized fields first if they exist
	for _, field := range prioritized {
		for _, available := range allFields {
			if field == available {
				result = append(result, field)
				seen[field] = true
				break
			}
		}
	}

	// Add remaining fields
	for _, field := range allFields {
		if !seen[field] {
			result = append(result, field)
		}
	}

	return result
}

// getAvailableMetadataFields scans sample audiobook to find available fields
func (a *App) getAvailableMetadataFields() map[string]bool {
	fields := make(map[string]bool)

	// Try to scan and get first audiobook
	if a.config.BaseDir == "" {
		return fields
	}

	audiobooks, err := organizer.ScanForAudiobooks(a.config.BaseDir, a.config)
	if err != nil || len(audiobooks) == 0 {
		return fields
	}

	// Collect fields from first audiobook
	sample := audiobooks[0]
	for key := range sample.RawData {
		if key != "" {
			fields[key] = true
		}
	}

	return fields
}

// filterFields returns fields that match any of the patterns
func filterFields(available map[string]bool, patterns []string) []string {
	var result []string

	// Add patterns that exist in available fields
	for _, pattern := range patterns {
		if available[pattern] {
			result = append(result, pattern)
		}
	}

	// If no matches, return the patterns anyway as fallback
	if len(result) == 0 {
		result = patterns
	}

	return result
}

// UpdateFieldMappingField updates a single field in the field mapping
func (a *App) UpdateFieldMappingField(field string, value string) error {
	switch field {
	case organizer.TitleFieldKey:
		a.config.FieldMapping.TitleField = value
	case organizer.SeriesFieldKey:
		a.config.FieldMapping.SeriesField = value
	case organizer.AuthorsFieldKey:
		// Parse comma-separated list
		a.config.FieldMapping.AuthorFields = strings.Split(value, ",")
	case organizer.TrackFieldKey:
		a.config.FieldMapping.TrackField = value
	case organizer.DiscFieldKey:
		a.config.FieldMapping.DiscField = value
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	a.log("Updated field mapping field %s=%s", field, value)
	return nil
}

// MetadataPreview represents metadata with field indicators for preview
type MetadataPreview struct {
	Filename   string                 `json:"filename"`
	SourceType string                 `json:"source_type"`
	RawFields  []RawFieldPreview      `json:"raw_fields"`
	Mapping    organizer.FieldMapping `json:"mapping"`
}

// RawFieldPreview represents a raw metadata field with its indicator
type RawFieldPreview struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Indicator string `json:"indicator"` // "TITLE", "AUTHOR", "SERIES", "TRACK", "DISC", or ""
}

// lastScanResults stores the most recent scan results to avoid re-scanning
var lastScanResults []organizer.Metadata

// GetSampleMetadataPreviews returns sample metadata for up to 3 books with field indicators
func (a *App) GetSampleMetadataPreviews(dir string) ([]MetadataPreview, error) {
	a.log("GetSampleMetadataPreviews called with dir=%s", dir)
	a.log("Current config: UseEmbedded=%v, Flat=%v", a.config.UseEmbeddedMetadata, a.config.Flat)

	// Use cached scan results if available
	var audiobooks []organizer.Metadata
	if len(lastScanResults) > 0 {
		a.log("Using cached scan results (%d audiobooks)", len(lastScanResults))
		audiobooks = lastScanResults
	} else {
		// Scan to get audiobooks
		var err error
		audiobooks, err = organizer.ScanForAudiobooks(dir, a.config)
		if err != nil {
			a.log("ScanForAudiobooks error: %v", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		if len(audiobooks) == 0 {
			a.log("No audiobooks found")
			return nil, fmt.Errorf("no audiobooks found")
		}
	}

	// Get metadata for all books (frontend will slice as needed)
	var previews []MetadataPreview
	for i := 0; i < len(audiobooks); i++ {
		sample := audiobooks[i]
		a.log("Sample audiobook %d: Title=%s, Album=%s, SourceType=%s", i, sample.Title, sample.Album, sample.SourceType)
		a.log("RawData has %d fields", len(sample.RawData))

		// Build raw fields with indicators
		var rawFields []RawFieldPreview

		// Sort keys for consistent display
		keys := make([]string, 0, len(sample.RawData))
		for key := range sample.RawData {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := sample.RawData[key]

			// Skip nil or empty values
			if value == nil {
				continue
			}
			if strVal, ok := value.(string); ok && strVal == "" {
				continue
			}

			// Determine indicator based on field mapping
			indicator := ""
			if key == a.config.FieldMapping.TitleField {
				indicator = "TITLE"
			} else if key == a.config.FieldMapping.SeriesField {
				indicator = "SERIES"
			} else if key == a.config.FieldMapping.TrackField {
				indicator = "TRACK"
			} else if key == a.config.FieldMapping.DiscField {
				indicator = "DISC"
			} else {
				// Check if it's in author fields
				for _, af := range a.config.FieldMapping.AuthorFields {
					if key == af {
						indicator = "AUTHOR"
						break
					}
				}
			}

			// Convert value to string
			valueStr := fmt.Sprintf("%v", value)

			rawFields = append(rawFields, RawFieldPreview{
				Key:       key,
				Value:     valueStr,
				Indicator: indicator,
			})
		}

		previews = append(previews, MetadataPreview{
			Filename:   filepath.Base(sample.SourcePath),
			SourceType: sample.SourceType,
			RawFields:  rawFields,
			Mapping:    a.config.FieldMapping,
		})
	}

	return previews, nil
}

// ProgressUpdate represents the current operation progress
type ProgressUpdate struct {
	Status      string `json:"status"`
	Current     int    `json:"current"`
	Total       int    `json:"total"`
	CurrentFile string `json:"current_file"`
}

// PreviewItem represents a file operation preview
type PreviewItem struct {
	From       string `json:"from"`
	To         string `json:"to"`
	IsConflict bool   `json:"is_conflict"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return NewAppWithDirs("", "")
}

// NewAppWithDirs creates a new App with pre-set directories from CLI args
func NewAppWithDirs(inputDir, outputDir string) *App {
	return &App{
		config: &organizer.OrganizerConfig{
			Layout:              "author-series-title",
			ReplaceSpace:        " ",
			Verbose:             false,
			DryRun:              false,
			Undo:                false,
			Prompt:              false,
			RemoveEmpty:         false,
			UseEmbeddedMetadata: false, // Default to metadata.json mode
			Flat:                false,
			FieldMapping:        organizer.DefaultFieldMapping(), // Use default field mapping for metadata.json
			BaseDir:             inputDir,
			OutputDir:           outputDir,
		},
		scanning: false,
		progress: ProgressUpdate{
			Status: "idle",
		},
		initialDirs: InitialDirectories{
			InputDir:  inputDir,
			OutputDir: outputDir,
		},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetInitialDirectories returns the directories set via CLI arguments
func (a *App) GetInitialDirectories() InitialDirectories {
	return a.initialDirs
}

// SelectDirectory opens a directory picker dialog
func (a *App) SelectDirectory(title string) (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
	return dir, err
}

// ScanDirectory scans a directory for audiobooks and returns metadata
func (a *App) ScanDirectory(dir string) ([]organizer.Metadata, error) {
	a.log("ScanDirectory called with dir=%s", dir)

	if dir == "" {
		return nil, fmt.Errorf("directory path is required")
	}

	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("error resolving path: %w", err)
	}

	a.log("Absolute path: %s", absPath)
	a.log("UseEmbeddedMetadata=%v", a.config.UseEmbeddedMetadata)
	a.log("Flat=%v", a.config.Flat)
	a.log("FieldMapping=%+v", a.config.FieldMapping)

	// Use absolute path for scanning
	dir = absPath

	a.scanning = true
	a.progress = ProgressUpdate{
		Status: "scanning",
		Total:  0,
	}

	// Create a temporary config for scanning
	scanConfig := *a.config
	scanConfig.BaseDir = dir
	scanConfig.DryRun = true

	a.log("Calling ScanForAudiobooks with config: UseEmbeddedMetadata=%v", scanConfig.UseEmbeddedMetadata)

	// Use the public scanning API
	audiobooks, err := organizer.ScanForAudiobooks(dir, &scanConfig)
	if err != nil {
		a.log("ScanForAudiobooks returned error: %v", err)
		a.scanning = false
		a.progress = ProgressUpdate{
			Status: "error",
		}
		return nil, err
	}

	a.log("ScanForAudiobooks found %d audiobooks", len(audiobooks))

	// FALLBACK SCANNING: If we found 0 audiobooks, try other scan modes
	if len(audiobooks) == 0 {
		a.log("No audiobooks found with current mode, trying fallback modes...")

		// Try different scan mode combinations
		fallbackModes := []struct {
			name     string
			embedded bool
			flat     bool
		}{
			{"Hierarchical (embedded)", true, false},
			{"Flat (embedded)", true, true},
			{"Metadata.json only", false, false},
		}

		for _, mode := range fallbackModes {
			// Skip if this is the mode we just tried
			if mode.embedded == scanConfig.UseEmbeddedMetadata && mode.flat == scanConfig.Flat {
				continue
			}

			a.log("Trying fallback mode: %s (UseEmbedded=%v, Flat=%v)", mode.name, mode.embedded, mode.flat)

			fallbackConfig := scanConfig
			fallbackConfig.UseEmbeddedMetadata = mode.embedded
			fallbackConfig.Flat = mode.flat

			// Update field mapping based on mode
			if mode.embedded {
				fallbackConfig.FieldMapping = organizer.AudioFieldMapping()
			} else {
				fallbackConfig.FieldMapping = organizer.DefaultFieldMapping()
			}

			audiobooks, err = organizer.ScanForAudiobooks(dir, &fallbackConfig)
			if err != nil {
				a.log("Fallback mode %s returned error: %v", mode.name, err)
				continue
			}

			if len(audiobooks) > 0 {
				a.log("SUCCESS! Fallback mode %s found %d audiobooks", mode.name, len(audiobooks))
				// Update the app config to use this mode
				a.config.UseEmbeddedMetadata = mode.embedded
				a.config.Flat = mode.flat
				a.config.FieldMapping = fallbackConfig.FieldMapping
				break
			}
			a.log("Fallback mode %s found 0 audiobooks", mode.name)
		}
	}

	a.log("Final scan result: %d audiobooks", len(audiobooks))
	for i, ab := range audiobooks {
		a.log("Audiobook %d: Title=%s, Album=%s, Series=%v", i, ab.Title, ab.Album, ab.Series)
		fmt.Printf("  Authors: %v\n", ab.Authors)
		fmt.Printf("  SourcePath: %s\n", ab.SourcePath)
	}

	// Cache results for metadata preview
	lastScanResults = audiobooks

	a.scanning = false
	a.progress = ProgressUpdate{
		Status: "idle",
	}

	return audiobooks, nil
}

// GetSettings returns the current organizer settings
func (a *App) GetSettings() organizer.OrganizerConfig {
	return *a.config
}

// UpdateSettings updates the organizer configuration
func (a *App) UpdateSettings(config organizer.OrganizerConfig) error {
	a.config = &config
	return nil
}

// PreviewChanges generates a preview of file operations for selected books
func (a *App) PreviewChanges(inputDir string, outputDir string, selectedBooks []int) ([]PreviewItem, error) {
	if inputDir == "" || outputDir == "" {
		return nil, fmt.Errorf("input and output directories are required")
	}

	a.log("PreviewChanges called with %d selected books", len(selectedBooks))
	a.log("Selected indices: %v", selectedBooks)
	a.log("Cached scan results: %d books", len(lastScanResults))

	// Check if we have cached scan results
	if len(lastScanResults) == 0 {
		return nil, fmt.Errorf("no scan results available - please scan first")
	}

	a.config.BaseDir = inputDir
	a.config.OutputDir = outputDir
	a.config.DryRun = true

	// Filter the cached books by selected indices FIRST
	var selectedMetadata []organizer.Metadata
	if len(selectedBooks) > 0 {
		a.log("Filtering by selected indices")
		for _, idx := range selectedBooks {
			if idx >= 0 && idx < len(lastScanResults) {
				selectedMetadata = append(selectedMetadata, lastScanResults[idx])
				a.log("Including book %d: %s", idx, lastScanResults[idx].Title)
			} else {
				a.log("Skipping invalid index %d (out of range 0-%d)", idx, len(lastScanResults)-1)
			}
		}
	} else {
		// If no selection, use all books
		selectedMetadata = lastScanResults
	}

	a.log("Selected %d books for preview", len(selectedMetadata))

	// Populate AllowedSourcePaths so ExecuteOrganize only processes these books.
	// SourcePath points to metadata.json (hierarchical) or the audio file (flat);
	// in hierarchical mode the organizer filters by directory, so use Dir().
	allowedPaths := make([]string, 0, len(selectedMetadata))
	for _, meta := range selectedMetadata {
		if meta.SourcePath == "" {
			continue
		}
		absPath, err := filepath.Abs(meta.SourcePath)
		if err != nil {
			absPath = meta.SourcePath
		}
		// Use the directory when SourcePath is a file (e.g. metadata.json)
		info, statErr := os.Stat(absPath)
		if statErr == nil && !info.IsDir() {
			absPath = filepath.Dir(absPath)
		}
		allowedPaths = append(allowedPaths, absPath)
	}
	a.config.AllowedSourcePaths = allowedPaths
	a.log("Set AllowedSourcePaths: %v", allowedPaths)

	// Calculate target paths for the selected books
	moves, err := organizer.CalculateTargetPaths(selectedMetadata, a.config)
	if err != nil {
		return nil, fmt.Errorf("error calculating target paths: %w", err)
	}

	a.log("Calculated %d moves", len(moves))

	// Detect conflicts
	targetMap := make(map[string][]int)
	for i, move := range moves {
		cleanTarget := filepath.Clean(move.TargetPath)
		targetMap[cleanTarget] = append(targetMap[cleanTarget], i)
	}

	// Mark conflicts
	conflictCount := 0
	for targetPath, indices := range targetMap {
		if len(indices) > 1 {
			for _, idx := range indices {
				moves[idx].IsConflict = true
				moves[idx].ConflictReason = fmt.Sprintf("Multiple books (%d) would be moved to: %s", len(indices), targetPath)
				conflictCount++
			}
		}
	}

	a.log("Detected %d conflicts", conflictCount)

	// Convert PreviewMove to PreviewItem for the GUI
	items := make([]PreviewItem, len(moves))
	for i, move := range moves {
		items[i] = PreviewItem{
			From:       move.SourcePath,
			To:         move.TargetPath,
			IsConflict: move.IsConflict,
		}
		a.log("Preview %d: %s → %s (conflict=%v)", i, move.SourcePath, move.TargetPath, move.IsConflict)
	}

	// Create organizer for later execution
	org, err := organizer.NewOrganizer(a.config)
	if err != nil {
		return nil, err
	}
	a.organizer = org

	return items, nil
}

// ExecuteOrganize performs the actual file organization
func (a *App) ExecuteOrganize(dryRun bool) (*organizer.Summary, error) {
	if a.organizer == nil {
		return nil, fmt.Errorf("no organizer configured - run PreviewChanges first")
	}

	// Update dry-run setting
	a.config.DryRun = dryRun

	// Re-create organizer with updated config
	org, err := organizer.NewOrganizer(a.config)
	if err != nil {
		return nil, err
	}

	a.organizer = org

	a.progress = ProgressUpdate{
		Status: "organizing",
	}

	// Execute the organization
	err = org.Execute()
	if err != nil {
		a.progress = ProgressUpdate{
			Status: "error",
		}
		return nil, err
	}

	a.progress = ProgressUpdate{
		Status: "complete",
	}

	// Get the actual summary from the organizer
	summary := org.GetSummary()

	return &summary, nil
}

// GetProgress returns the current operation progress
func (a *App) GetProgress() ProgressUpdate {
	return a.progress
}

// GetLogPath returns the path to the operation log file
func (a *App) GetLogPath() string {
	if a.organizer != nil {
		return a.organizer.GetLogPath()
	}
	if a.config.OutputDir != "" {
		return filepath.Join(a.config.OutputDir, ".abook-org.log")
	}
	return ""
}

// SetLogLevel sets the logging level for the application
func (a *App) SetLogLevel(level string) {
	a.logLevel = level
	a.verbose = (level == "debug")
	if a.verbose {
		fmt.Printf("LOG LEVEL: %s (verbose=%v)\n", level, a.verbose)
	}
	// Also update the config
	if a.config != nil {
		a.config.Verbose = a.verbose
	}
}

// log prints a message if verbose logging is enabled
func (a *App) log(format string, args ...interface{}) {
	if a.verbose {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// ScanStatistics provides summary information about scan results
type ScanStatistics struct {
	TotalFiles      int                  `json:"total_files"`
	TotalAudiobooks int                  `json:"total_audiobooks"`
	MissingMetadata int                  `json:"missing_metadata"`
	AlbumGroups     []AlbumGroup         `json:"album_groups"`
	UngroupedFiles  []organizer.Metadata `json:"ungrouped_files"`
}

// AlbumGroup represents a group of files that belong to the same audiobook/album
type AlbumGroup struct {
	Name        string               `json:"name"`
	Author      string               `json:"author"`
	Series      string               `json:"series"`
	FileCount   int                  `json:"file_count"`
	FileIndices []int                `json:"file_indices"`
	Files       []organizer.Metadata `json:"files"`
}

// GetScanStatistics analyzes the last scan results and returns statistics
func (a *App) GetScanStatistics() (*ScanStatistics, error) {
	if len(lastScanResults) == 0 {
		return &ScanStatistics{}, nil
	}

	stats := &ScanStatistics{
		TotalFiles: len(lastScanResults),
	}

	// Group files by album/title combination
	albumMap := make(map[string]*AlbumGroup)
	var ungrouped []organizer.Metadata

	for i, book := range lastScanResults {
		// Check for missing metadata
		if book.Title == "" && book.Album == "" {
			stats.MissingMetadata++
		}

		// Create grouping key based on album and author
		album := book.Album
		if album == "" {
			album = book.Title
		}
		author := ""
		if len(book.Authors) > 0 {
			author = book.Authors[0]
		}

		// If no album/title, add to ungrouped
		if album == "" {
			ungrouped = append(ungrouped, book)
			continue
		}

		groupKey := fmt.Sprintf("%s|%s", author, album)

		if group, exists := albumMap[groupKey]; exists {
			group.FileCount++
			group.FileIndices = append(group.FileIndices, i)
			group.Files = append(group.Files, book)
		} else {
			series := ""
			if len(book.Series) > 0 {
				series = book.Series[0]
			}
			albumMap[groupKey] = &AlbumGroup{
				Name:        album,
				Author:      author,
				Series:      series,
				FileCount:   1,
				FileIndices: []int{i},
				Files:       []organizer.Metadata{book},
			}
		}
	}

	// Convert map to slice and sort by file count (largest first)
	for _, group := range albumMap {
		stats.AlbumGroups = append(stats.AlbumGroups, *group)
	}

	// Sort groups by file count (descending)
	sort.Slice(stats.AlbumGroups, func(i, j int) bool {
		return stats.AlbumGroups[i].FileCount > stats.AlbumGroups[j].FileCount
	})

	stats.TotalAudiobooks = len(stats.AlbumGroups)
	stats.UngroupedFiles = ungrouped

	a.log("Scan statistics: %d files, %d audiobooks, %d missing metadata, %d ungrouped",
		stats.TotalFiles, stats.TotalAudiobooks, stats.MissingMetadata, len(ungrouped))

	return stats, nil
}

// GetRenameConfig returns the current rename template configuration
func (a *App) GetRenameConfig() RenameConfig {
	return a.renameConfig
}

// UpdateRenameConfig updates the rename template configuration
func (a *App) UpdateRenameConfig(config RenameConfig) error {
	a.renameConfig = config
	a.log("Updated rename config: enabled=%v, template=%s, preset=%s",
		config.Enabled, config.Template, config.Preset)
	return nil
}

// PreviewRename generates a preview of what a file would be renamed to
func (a *App) PreviewRename(metadata organizer.Metadata) (string, error) {
	if !a.renameConfig.Enabled || a.renameConfig.Template == "" {
		return "", nil
	}

	// Parse template
	template, err := organizer.ParseTemplate(a.renameConfig.Template)
	if err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}

	// Determine author format
	var authorFormat organizer.AuthorFormat
	switch a.renameConfig.AuthorFormat {
	case "first-last":
		authorFormat = organizer.AuthorFormatFirstLast
	case "last-first":
		authorFormat = organizer.AuthorFormatLastFirst
	case "preserve":
		authorFormat = organizer.AuthorFormatPreserve
	default:
		authorFormat = organizer.AuthorFormatFirstLast
	}

	// Create renderer
	formatter := organizer.NewAuthorFormatter(authorFormat)
	renderer := organizer.NewTemplateRenderer(template, formatter)

	// Render template
	newFilename, err := renderer.Render(metadata)
	if err != nil {
		return "", err
	}

	// Apply space replacement if configured
	if a.renameConfig.ReplaceSpaces && a.renameConfig.SpaceChar != "" {
		newFilename = strings.ReplaceAll(newFilename, " ", a.renameConfig.SpaceChar)
	}

	// Add extension from original file
	ext := filepath.Ext(metadata.SourcePath)
	if !strings.HasSuffix(newFilename, ext) {
		newFilename += ext
	}

	return newFilename, nil
}

// GetRenamePresets returns available rename template presets
func (a *App) GetRenamePresets() []map[string]string {
	return []map[string]string{
		{"name": "Custom", "template": ""},
		{"name": "Track - Title", "template": "{track} - {title}"},
		{"name": "Author - Title", "template": "{author} - {title}"},
		{"name": "Author - Series - Title", "template": "{author} - {series} - {title}"},
		{"name": "Track - Author - Title", "template": "{track} - {author} - {title}"},
		{"name": "Series Number - Title", "template": "{series_number} - {title}"},
		{"name": "Author - Series #Number - Title", "template": "{author} - {series} #{series_number} - {title}"},
	}
}

// GetAvailableTemplateFields returns available template fields with descriptions
func (a *App) GetAvailableTemplateFields() []map[string]string {
	fields := organizer.GetAvailableFields()
	result := make([]map[string]string, len(fields))
	for i, field := range fields {
		result[i] = map[string]string{
			"name":        field.Name,
			"description": field.Description,
			"example":     field.Example,
		}
	}
	return result
}

// Greet returns a greeting (keeping for backwards compatibility with template)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s! Ready to organize some audiobooks? 📚", name)
}

// OrganizeFiles executes the organization process for selected files ONLY
func (a *App) OrganizeFiles(selectedPaths []string, outputDir string, copyMode bool) (map[string]interface{}, error) {
	a.log("Starting organization: %d files selected, output=%s, copy=%v", len(selectedPaths), outputDir, copyMode)

	if len(selectedPaths) == 0 {
		return nil, fmt.Errorf("no files selected")
	}

	if outputDir == "" {
		return nil, fmt.Errorf("output directory not specified")
	}

	// CRITICAL: This function is currently BROKEN and will organize ALL files
	// regardless of selection. DO NOT USE until fixed properly.
	// The organizer.Execute() processes all scanned files, not just selected ones.

	return map[string]interface{}{
		"success":        false,
		"filesProcessed": 0,
		"errors": []string{
			"OrganizeFiles is currently disabled due to a critical bug",
			"It was organizing ALL files instead of only selected files",
			"This function needs to be rewritten to filter files before organizing",
		},
	}, fmt.Errorf("function disabled - would organize all files instead of selected ones")
}
