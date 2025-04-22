package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

const LogFileName = ".abook-org.log"

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
}

// Organizer is the main struct that performs audiobook organization
type Organizer struct {
	config     OrganizerConfig
	summary    Summary
	logEntries []LogEntry
}

// NewOrganizer creates a new Organizer with the provided configuration
func NewOrganizer(config *OrganizerConfig) *Organizer {
	// Set the verbose mode flag for the metadata providers
	SetVerboseMode(config.Verbose)

	return &Organizer{
		config: *config,
	}
}

func (o *Organizer) GetLogPath() string {
	logBase := o.config.BaseDir
	if o.config.OutputDir != "" {
		logBase = o.config.OutputDir
	}
	return filepath.Join(logBase, LogFileName)
}

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

// removeEmptySourceDirs scans the source directory for empty directories
func (o *Organizer) removeEmptySourceDirs() error {
	if !o.config.RemoveEmpty {
		return nil
	}

	if o.config.Verbose {
		color.Blue("🔍 Scanning for empty directories...")
	}

	var emptyDirs []string
	err := filepath.Walk(o.config.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		// Skip the base directory itself and the output directory
		if path == o.config.BaseDir || (o.config.OutputDir != "" && path == o.config.OutputDir) {
			return nil
		}

		if isEmptyDir(path) {
			emptyDirs = append(emptyDirs, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error scanning for empty directories: %v", err)
	}

	// Remove empty directories in reverse order (deepest first)
	for i := len(emptyDirs) - 1; i >= 0; i-- {
		dir := emptyDirs[i]
		if o.config.Verbose {
			color.Yellow("🗑️  Removing empty directory: %s", dir)
		}
		if !o.config.DryRun {
			if err := os.Remove(dir); err != nil {
				color.Red("❌ Error removing directory %s: %v", dir, err)
			}
		}
	}

	return nil
}
