// internal/organizer/organizer.go
package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Constants
const (
	LogFileName        = ".abook-org.log"
	TestBookDirName    = "test_book"
	MetadataFileName   = "metadata.json"
	TestAudioFileName  = "audio.mp3"
	TrackPrefixFormat  = "%02d - "
	InvalidSeriesValue = "__INVALID_SERIES__"
)

// OrganizerConfig contains all configuration parameters for an Organizer
type OrganizerConfig struct {
	BaseDir             string
	OutputDir           string
	ReplaceSpace        string
	Verbose             bool
	DryRun              bool
	Undo                bool
	Prompt              bool
	RemoveEmpty         bool
	UseEmbeddedMetadata bool
	Flat                bool
	Layout              string       // Directory structure layout (author-series-title, author-title, author-only)
	FieldMapping        FieldMapping // Configuration for mapping metadata fields
}

// Validate checks if the configuration is valid and returns helpful error messages
func (c *OrganizerConfig) Validate() error {
	// Check base directory
	if c.BaseDir == "" {
		return fmt.Errorf("base directory is required\n\nPlease specify an input directory:\n  --dir=/path/to/audiobooks\n  or\n  --input=/path/to/audiobooks")
	}

	// Check if base directory exists
	info, err := os.Stat(c.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("base directory does not exist: %s\n\nPlease check the path and try again", c.BaseDir)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied accessing: %s\n\nTry running with appropriate permissions:\n  sudo audiobook-organizer --dir=%s", c.BaseDir, c.BaseDir)
		}
		return fmt.Errorf("error accessing base directory %s: %w", c.BaseDir, err)
	}

	// Verify it's a directory
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory\n\nPlease specify a directory, not a file", c.BaseDir)
	}

	// If output directory is specified, validate it
	if c.OutputDir != "" {
		// Check if output directory exists or can be created
		if _, err := os.Stat(c.OutputDir); err != nil {
			if os.IsNotExist(err) && !c.DryRun {
				// Try to create it
				if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
					return fmt.Errorf("cannot create output directory %s: %w\n\nPlease check permissions or create it manually", c.OutputDir, err)
				}
			} else if os.IsPermission(err) {
				return fmt.Errorf("permission denied accessing output directory: %s\n\nTry running with appropriate permissions", c.OutputDir)
			}
		}
	}

	// Validate layout option
	validLayouts := map[string]bool{
		"author-series-title":        true,
		"author-series-title-number": true,
		"author-series":              true,
		"author-title":               true,
		"author-only":                true,
		"series-title":               true,
		"series-title-number":        true,
	}
	if c.Layout != "" && !validLayouts[c.Layout] {
		return fmt.Errorf("invalid layout: %s\n\nValid options are:\n  author-series-title (default)\n  author-series-title-number\n  author-series\n  author-title\n  author-only\n  series-title\n  series-title-number", c.Layout)
	}

	// Validate replace_space character (should be single char or empty)
	if len(c.ReplaceSpace) > 1 {
		return fmt.Errorf("replace_space must be a single character, got: %q\n\nExamples:\n  --replace_space=_\n  --replace_space=.\n  --replace_space=-", c.ReplaceSpace)
	}

	return nil
}

// FileOps handles file system operations with dry-run support
type FileOps struct {
	dryRun bool
}

// NewFileOps creates a new file operations handler
func NewFileOps(dryRun bool) *FileOps {
	return &FileOps{dryRun: dryRun}
}

// CreateDirIfNotExists creates a directory if it doesn't exist, respecting dry-run mode
func (f *FileOps) CreateDirIfNotExists(dir string) error {
	if f.dryRun {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

// FileExists checks if a file exists on the filesystem
func (f *FileOps) FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

// DirectoryExists checks if a directory exists on the filesystem
func (f *FileOps) DirectoryExists(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// AllFilesExist checks if all specified files exist on the filesystem
func (f *FileOps) AllFilesExist(files ...string) bool {
	for _, file := range files {
		if !f.FileExists(file) {
			return false
		}
	}
	return true
}

// LayoutCalculator handles path calculations based on layout configuration
type LayoutCalculator struct {
	config    *OrganizerConfig
	sanitizer func(string) string
}

// NewLayoutCalculator creates a new layout calculator
func NewLayoutCalculator(config *OrganizerConfig, sanitizer func(string) string) *LayoutCalculator {
	return &LayoutCalculator{
		config:    config,
		sanitizer: sanitizer,
	}
}

// CalculateTargetPath determines the target directory path based on metadata and layout
func (lc *LayoutCalculator) CalculateTargetPath(metadata Metadata) string {
	authorDir := lc.sanitizer(strings.Join(metadata.Authors, ","))
	titleDir := lc.sanitizer(metadata.Title)
	targetBase := lc.getTargetBase()

	switch lc.config.Layout {
	case "author-only":
		return filepath.Join(targetBase, authorDir)
	case "author-series":
		// Author/Series layout (no title subdirectory)
		// Used for multi-file audiobooks where each file is a chapter
		if validSeries := metadata.GetValidSeries(); validSeries != "" {
			seriesDir := lc.sanitizer(validSeries)
			return filepath.Join(targetBase, authorDir, seriesDir)
		}
		// If no series, fall back to author/title
		return filepath.Join(targetBase, authorDir, titleDir)
	case "author-title":
		return filepath.Join(targetBase, authorDir, titleDir)
	case "author-series-title", "":
		return filepath.Join(targetBase, authorDir, lc.calculateSeriesPath(titleDir, metadata))
	case "author-series-title-number":
		return filepath.Join(targetBase, authorDir, lc.calculateSeriesPathWithNumber(titleDir, metadata))
	case "series-title":
		return filepath.Join(targetBase, lc.calculateSeriesPath(titleDir, metadata))
	case "series-title-number":
		return filepath.Join(targetBase, lc.calculateSeriesPathWithNumber(titleDir, metadata))
	default:
		return filepath.Join(targetBase, authorDir, titleDir)
	}
}

// getTargetBase returns the base directory for organizing files
func (lc *LayoutCalculator) getTargetBase() string {
	if lc.config.OutputDir != "" {
		return lc.config.OutputDir
	}
	return lc.config.BaseDir
}

// calculateSeriesPath handles series-based path calculation
// Returns the series/title portion of the path (e.g., "Series/Title" or just "Title")
func (lc *LayoutCalculator) calculateSeriesPath(titleDir string, metadata Metadata) string {
	if validSeries := metadata.GetValidSeries(); validSeries != "" {
		seriesDir := lc.sanitizer(validSeries)
		return filepath.Join(seriesDir, titleDir)
	}
	return titleDir
}

// calculateSeriesPathWithNumber handles series-based path calculation with series number in title
// Returns the series/title portion of the path (e.g., "Series/#1 - Title" or just "Title")
func (lc *LayoutCalculator) calculateSeriesPathWithNumber(titleDir string, metadata Metadata) string {
	if validSeries := metadata.GetValidSeries(); validSeries != "" {
		seriesDir := lc.sanitizer(validSeries)

		// Get series number and prefix the title with it
		seriesNumber := GetSeriesNumberFromMetadata(metadata)
		if seriesNumber != "" {
			numberedTitle := fmt.Sprintf("#%s - %s", seriesNumber, titleDir)
			return filepath.Join(seriesDir, numberedTitle)
		}

		// If no series number, fall back to regular series path
		return filepath.Join(seriesDir, titleDir)
	}
	return titleDir
}

// Organizer is the main struct that performs audiobook organization
type Organizer struct {
	config           OrganizerConfig
	summary          Summary
	logEntries       []LogEntry
	fileOps          *FileOps
	layoutCalculator *LayoutCalculator
}

// NewOrganizer creates a new Organizer with the provided configuration
func NewOrganizer(config *OrganizerConfig) (*Organizer, error) {
	// Validate configuration first
	if err := config.Validate(); err != nil {
		return nil, err
	}

	org := &Organizer{
		config:  *config,
		fileOps: NewFileOps(config.DryRun),
	}

	org.layoutCalculator = NewLayoutCalculator(config, org.SanitizePath)

	// Set the verbose mode flag for the metadata providers
	SetVerboseMode(config.Verbose)

	// Initialize default field mapping if not provided
	if config.FieldMapping.IsEmpty() {
		config.FieldMapping = DefaultFieldMapping()
	}

	return org, nil
}

// GetLogPath returns the path where operation logs are stored
func (o *Organizer) GetLogPath() string {
	logBase := o.config.BaseDir
	if o.config.OutputDir != "" {
		logBase = o.config.OutputDir
	}
	return filepath.Join(logBase, LogFileName)
}

// GetSummary returns the current operation summary
func (o *Organizer) GetSummary() Summary {
	return o.summary
}

// Execute runs the main organization process
func (o *Organizer) Execute() error {
	// Clean and resolve the paths
	color.Blue("🔍 Resolving paths...")
	resolvedBaseDir, err := filepath.EvalSymlinks(filepath.Clean(o.config.BaseDir))
	if err != nil {
		return fmt.Errorf("error resolving base directory path: %v", err)
	}
	o.config.BaseDir = resolvedBaseDir

	if o.config.OutputDir != "" {
		resolvedOutputDir, err := filepath.EvalSymlinks(filepath.Clean(o.config.OutputDir))
		if err != nil {
			return fmt.Errorf("error resolving output directory path: %v", err)
		}
		o.config.OutputDir = resolvedOutputDir
	}

	// Check if the base path is a file rather than a directory
	fileInfo, err := os.Stat(o.config.BaseDir)
	if err != nil {
		return fmt.Errorf("error checking base path: %v", err)
	}

	// If it's a single file, process it directly
	if !fileInfo.IsDir() {
		if o.config.Verbose {
			color.Blue("🔍 Processing single file: %s", o.config.BaseDir)
		}

		// In flat mode, we need embedded metadata
		if o.config.Flat && !o.config.UseEmbeddedMetadata {
			return fmt.Errorf("flat mode requires embedded metadata to be enabled")
		}

		// Process the single file
		return o.OrganizeSingleFile(o.config.BaseDir, nil)
	}

	if o.config.Undo {
		color.Yellow("↩️  Undoing previous operations...")
		return o.undoMoves()
	}

	if o.config.DryRun {
		color.Yellow("🔍 Running in dry-run mode - no files will be moved")
	}

	startTime := time.Now()
	color.Blue("📚 Scanning for audiobooks...")
	err = filepath.Walk(o.config.BaseDir, o.processDirectory)
	if err != nil {
		return fmt.Errorf("error walking directory: %v", err)
	}

	if !o.config.DryRun && len(o.logEntries) > 0 {
		color.Blue("💾 Saving operation log...")
		if err := o.saveLog(); err != nil {
			return fmt.Errorf("error saving log: %v", err)
		}
	}

	// Remove empty directories after all moves are complete
	if err := o.removeEmptySourceDirs(); err != nil {
		color.Red("❌ Error removing empty directories: %v", err)
	}

	o.printSummary(startTime)
	return nil
}

// isEmptyDir checks if a directory is empty
func isEmptyDir(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) == 0
}

// removeEmptyDirs removes empty directories recursively up the tree
// It stops when it encounters a non-empty directory or reaches the BaseDir
func (o *Organizer) removeEmptyDirs(dir string) error {
	if !o.config.RemoveEmpty || dir == o.config.BaseDir {
		return nil
	}

	// Check if directory exists
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil
	}

	// If directory is not empty, stop here
	if !isEmptyDir(dir) {
		return nil
	}

	if o.config.Verbose {
		color.Yellow("🗑️  Removing empty directory: %s", dir)
	}

	if !o.config.DryRun {
		if err := os.Remove(dir); err != nil {
			return fmt.Errorf("failed to remove directory %s: %v", dir, err)
		}
	}

	// Recursively check parent directory
	parent := filepath.Dir(dir)
	if parent != o.config.BaseDir {
		return o.removeEmptyDirs(parent)
	}

	return nil
}
func (o *Organizer) removeEmptySourceDirs() error {
	if !o.config.RemoveEmpty {
		return nil
	}

	if o.config.Verbose {
		PrintBlue("🔍 Scanning for empty directories...")
	}

	// Keep removing empty directories until no more are found
	for {
		emptyDirs, err := o.findEmptyDirectories()
		if err != nil {
			return err
		}

		// If no empty directories found, we're done
		if len(emptyDirs) == 0 {
			break
		}

		// Sort by depth (deepest first) for safe removal
		sort.Slice(emptyDirs, func(i, j int) bool {
			depthI := strings.Count(emptyDirs[i], string(filepath.Separator))
			depthJ := strings.Count(emptyDirs[j], string(filepath.Separator))
			return depthI > depthJ
		})

		// Remove empty directories in this iteration
		var removedAny bool
		for _, dir := range emptyDirs {
			if err := o.removeEmptyDir(dir); err != nil {
				PrintRed("❌ Error removing directory %s: %v", dir, err)
			} else {
				removedAny = true
			}
		}

		// If we couldn't remove any directories, break to avoid infinite loop
		if !removedAny {
			break
		}
	}

	return nil
}

// Helper function to find empty directories in a single pass
func (o *Organizer) findEmptyDirectories() ([]string, error) {
	var emptyDirs []string

	err := filepath.Walk(o.config.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-directories
		if !info.IsDir() {
			return nil
		}

		// Skip the base directory itself
		if path == o.config.BaseDir {
			return nil
		}

		// Skip the output directory if it's different from base
		if o.config.OutputDir != "" && path == o.config.OutputDir {
			return filepath.SkipDir
		}

		// Check if directory is empty
		if o.isEmptyDir(path) {
			emptyDirs = append(emptyDirs, path)
		}

		return nil
	})

	return emptyDirs, err
}

func (o *Organizer) removeEmptyDir(dir string) error {
	// Double-check it's still empty (might have been removed already)
	if !o.isEmptyDir(dir) {
		return nil
	}

	// Prompt if enabled
	if o.config.Prompt {
		if !o.PromptForDirectoryRemoval(dir, false) {
			if o.config.Verbose {
				PrintYellow("⏩ Skipping removal of directory %s", dir)
			}
			return nil
		}
	}

	if o.config.Verbose {
		PrintYellow("🗑️  Removing empty directory: %s", dir)
	}

	if !o.config.DryRun {
		if err := os.Remove(dir); err != nil {
			return fmt.Errorf("failed to remove directory: %v", err)
		}
		// Add to summary
		o.summary.EmptyDirsRemoved = append(o.summary.EmptyDirsRemoved, dir)
	}

	return nil
}

func (o *Organizer) isEmptyDir(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) == 0
}
