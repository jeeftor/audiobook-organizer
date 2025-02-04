package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

const LogFileName = ".abook-org.log"

type Organizer struct {
	baseDir      string
	outputDir    string
	replaceSpace string
	verbose      bool
	dryRun       bool
	undo         bool
	prompt       bool
	removeEmpty  bool
	summary      Summary
	logEntries   []LogEntry
}

func New(baseDir, outputDir, replaceSpace string, verbose, dryRun, undo, prompt, removeEmpty bool) *Organizer {
	return &Organizer{
		baseDir:      baseDir,
		outputDir:    outputDir,
		replaceSpace: replaceSpace,
		verbose:      verbose,
		dryRun:       dryRun,
		undo:         undo,
		prompt:       prompt,
		removeEmpty:  removeEmpty,
	}
}

func (o *Organizer) GetLogPath() string {
	logBase := o.baseDir
	if o.outputDir != "" {
		logBase = o.outputDir
	}
	return filepath.Join(logBase, LogFileName)
}

func (o *Organizer) Execute() error {
	// Clean and resolve the paths
	color.Blue("🔍 Resolving paths...")
	resolvedBaseDir, err := filepath.EvalSymlinks(filepath.Clean(o.baseDir))
	if err != nil {
		return fmt.Errorf("error resolving base directory path: %v", err)
	}
	o.baseDir = resolvedBaseDir

	if o.outputDir != "" {
		resolvedOutputDir, err := filepath.EvalSymlinks(filepath.Clean(o.outputDir))
		if err != nil {
			return fmt.Errorf("error resolving output directory path: %v", err)
		}
		o.outputDir = resolvedOutputDir
	}

	if o.undo {
		color.Yellow("↩️  Undoing previous operations...")
		return o.undoMoves()
	}

	if o.dryRun {
		color.Yellow("🔍 Running in dry-run mode - no files will be moved")
	}

	startTime := time.Now()
	color.Blue("📚 Scanning for audiobooks...")
	err = filepath.Walk(o.baseDir, o.processDirectory)
	if err != nil {
		return fmt.Errorf("error walking directory: %v", err)
	}

	if !o.dryRun && len(o.logEntries) > 0 {
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
// It stops when it encounters a non-empty directory or reaches the baseDir
func (o *Organizer) removeEmptyDirs(dir string) error {
	if !o.removeEmpty || dir == o.baseDir {
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

	if o.verbose {
		color.Yellow("🗑️  Removing empty directory: %s", dir)
	}

	if !o.dryRun {
		if err := os.Remove(dir); err != nil {
			return fmt.Errorf("failed to remove directory %s: %v", dir, err)
		}
	}

	// Recursively check parent directory
	parent := filepath.Dir(dir)
	if parent != o.baseDir {
		return o.removeEmptyDirs(parent)
	}

	return nil
}

// removeEmptySourceDirs scans the source directory for empty directories
func (o *Organizer) removeEmptySourceDirs() error {
	if !o.removeEmpty {
		return nil
	}

	if o.verbose {
		color.Blue("🔍 Scanning for empty directories...")
	}

	var emptyDirs []string
	err := filepath.Walk(o.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		// Skip the base directory itself and the output directory
		if path == o.baseDir || (o.outputDir != "" && path == o.outputDir) {
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
		if o.verbose {
			color.Yellow("🗑️  Removing empty directory: %s", dir)
		}
		if !o.dryRun {
			if err := os.Remove(dir); err != nil {
				color.Red("❌ Error removing directory %s: %v", dir, err)
			}
		}
	}

	return nil
}
