package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jeeftor/audiobook-organizer/pkg/organizer"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx         context.Context
	config      *organizer.OrganizerConfig
	organizer   *organizer.Organizer
	scanning    bool
	progress    ProgressUpdate
	initialDirs InitialDirectories
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

			fmt.Printf("DEBUG: Updated scan mode to %s (UseEmbedded=%v, Flat=%v)\n",
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
	fmt.Printf("DEBUG: Updated layout to %s\n", layout)
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
	fmt.Printf("DEBUG: Updated field mapping: %+v\n", mapping)
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
// Uses predefined common field lists
func (a *App) GetFieldMappingOptions() []FieldMappingOption {
	// Predefined common text fields
	textFieldOptions := []string{"title", "album", "series", "name", "book", "work"}

	// Predefined common author fields
	authorFieldOptions := []string{
		"authors",
		"artist",
		"album_artist",
		"narrator",
		"narrators",
		"creator",
		"author",
		"writer",
		"composer",
	}

	// Predefined numeric fields for track/disc
	numericFields := []string{
		"track",
		"track_number",
		"trck",
		"trk",
		"tracknumber",
		"disc",
		"discnumber",
		"disk",
		"tpos",
		"disc_number",
	}

	return []FieldMappingOption{
		{
			Field:       "title",
			Label:       "Title Field",
			Description: "Select field for title (title, album, series, or any available field)",
			Options:     textFieldOptions,
			Current:     a.config.FieldMapping.TitleField,
		},
		{
			Field:       "series",
			Label:       "Series Field",
			Description: "Select field for series (series, album, title, or any available field)",
			Options:     textFieldOptions,
			Current:     a.config.FieldMapping.SeriesField,
		},
		{
			Field:       "authors",
			Label:       "Author Fields (Priority Order)",
			Description: "Select fields in priority order (comma-separated: authors,artist,album_artist)",
			Options:     authorFieldOptions,
			Current:     strings.Join(a.config.FieldMapping.AuthorFields, ","),
		},
		{
			Field:       "track",
			Label:       "Track Field",
			Description: "Field to use for track number",
			Options:     numericFields,
			Current:     a.config.FieldMapping.TrackField,
		},
		{
			Field:       "disc",
			Label:       "Disc Field",
			Description: "Field to use for disc number",
			Options:     numericFields,
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
	case "title":
		a.config.FieldMapping.TitleField = value
	case "series":
		a.config.FieldMapping.SeriesField = value
	case "authors":
		// Parse comma-separated list
		a.config.FieldMapping.AuthorFields = strings.Split(value, ",")
	case "track":
		a.config.FieldMapping.TrackField = value
	case "disc":
		a.config.FieldMapping.DiscField = value
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	fmt.Printf("DEBUG: Updated field mapping field %s=%s\n", field, value)
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
	fmt.Printf("DEBUG: GetSampleMetadataPreviews called with dir=%s\n", dir)
	fmt.Printf(
		"DEBUG: Current config: UseEmbedded=%v, Flat=%v\n",
		a.config.UseEmbeddedMetadata,
		a.config.Flat,
	)

	// Use cached scan results if available
	var audiobooks []organizer.Metadata
	if len(lastScanResults) > 0 {
		fmt.Printf("DEBUG: Using cached scan results (%d audiobooks)\n", len(lastScanResults))
		audiobooks = lastScanResults
	} else {
		// Scan to get audiobooks
		var err error
		audiobooks, err = organizer.ScanForAudiobooks(dir, a.config)
		if err != nil {
			fmt.Printf("DEBUG: ScanForAudiobooks error: %v\n", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		if len(audiobooks) == 0 {
			fmt.Printf("DEBUG: No audiobooks found\n")
			return nil, fmt.Errorf("no audiobooks found")
		}
	}

	// Get metadata for all books (frontend will slice as needed)
	var previews []MetadataPreview
	for i := 0; i < len(audiobooks); i++ {
		sample := audiobooks[i]
		fmt.Printf(
			"DEBUG: Sample audiobook %d: Title=%s, Album=%s, SourceType=%s\n",
			i,
			sample.Title,
			sample.Album,
			sample.SourceType,
		)
		fmt.Printf("DEBUG: RawData has %d fields\n", len(sample.RawData))

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
	if dir == "" {
		return nil, fmt.Errorf("directory path is required")
	}

	// Convert to absolute path to ensure proper resolution
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("error resolving directory path: %w", err)
	}

	fmt.Printf("DEBUG: ScanDirectory called with dir=%s\n", dir)
	fmt.Printf("DEBUG: Absolute path: %s\n", absDir)
	fmt.Printf("DEBUG: UseEmbeddedMetadata=%v\n", a.config.UseEmbeddedMetadata)
	fmt.Printf("DEBUG: Flat=%v\n", a.config.Flat)
	fmt.Printf("DEBUG: FieldMapping=%+v\n", a.config.FieldMapping)

	// Use absolute path for scanning
	dir = absDir

	a.scanning = true
	a.progress = ProgressUpdate{
		Status: "scanning",
		Total:  0,
	}

	// Create a temporary config for scanning
	scanConfig := *a.config
	scanConfig.BaseDir = dir
	scanConfig.DryRun = true

	fmt.Printf(
		"DEBUG: Calling ScanForAudiobooks with config: UseEmbeddedMetadata=%v\n",
		scanConfig.UseEmbeddedMetadata,
	)

	// Use the public scanning API
	audiobooks, err := organizer.ScanForAudiobooks(dir, &scanConfig)
	if err != nil {
		fmt.Printf("DEBUG: ScanForAudiobooks returned error: %v\n", err)
		a.scanning = false
		a.progress = ProgressUpdate{
			Status: "error",
		}
		return nil, err
	}

	fmt.Printf("DEBUG: ScanForAudiobooks found %d audiobooks\n", len(audiobooks))
	for i, ab := range audiobooks {
		fmt.Printf("DEBUG: Audiobook %d:\n", i)
		fmt.Printf("  Title: %s\n", ab.Title)
		fmt.Printf("  Album: %s\n", ab.Album)
		fmt.Printf("  Series: %v\n", ab.Series)
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
func (a *App) PreviewChanges(
	inputDir string,
	outputDir string,
	selectedBooks []int,
) ([]PreviewItem, error) {
	if inputDir == "" || outputDir == "" {
		return nil, fmt.Errorf("input and output directories are required")
	}

	fmt.Printf("DEBUG: PreviewChanges called with %d selected books\n", len(selectedBooks))
	fmt.Printf("DEBUG: Selected indices: %v\n", selectedBooks)
	fmt.Printf("DEBUG: Cached scan results: %d books\n", len(lastScanResults))

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
		fmt.Printf("DEBUG: Filtering by selected indices\n")
		for _, idx := range selectedBooks {
			if idx >= 0 && idx < len(lastScanResults) {
				selectedMetadata = append(selectedMetadata, lastScanResults[idx])
				fmt.Printf("DEBUG: Including book %d: %s\n", idx, lastScanResults[idx].Title)
			} else {
				fmt.Printf("DEBUG: Skipping invalid index %d (out of range 0-%d)\n", idx, len(lastScanResults)-1)
			}
		}
	} else {
		// If no selection, use all books
		selectedMetadata = lastScanResults
	}

	fmt.Printf("DEBUG: Selected %d books for preview\n", len(selectedMetadata))

	// Calculate target paths for the selected books
	moves, err := organizer.CalculateTargetPaths(selectedMetadata, a.config)
	if err != nil {
		return nil, fmt.Errorf("error calculating target paths: %w", err)
	}

	fmt.Printf("DEBUG: Calculated %d moves\n", len(moves))

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
				moves[idx].ConflictReason = fmt.Sprintf(
					"Multiple books (%d) would be moved to: %s",
					len(indices),
					targetPath,
				)
				conflictCount++
			}
		}
	}

	fmt.Printf("DEBUG: Detected %d conflicts\n", conflictCount)

	// Convert PreviewMove to PreviewItem for the GUI
	items := make([]PreviewItem, len(moves))
	for i, move := range moves {
		items[i] = PreviewItem{
			From:       move.SourcePath,
			To:         move.TargetPath,
			IsConflict: move.IsConflict,
		}
		fmt.Printf(
			"DEBUG: Preview %d: %s → %s (conflict=%v)\n",
			i,
			move.SourcePath,
			move.TargetPath,
			move.IsConflict,
		)
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

// Greet returns a greeting (keeping for backwards compatibility with template)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s! Ready to organize some audiobooks? 📚", name)
}
