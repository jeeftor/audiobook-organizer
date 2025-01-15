package organizer

import (
	"fmt"
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
	summary      Summary
	logEntries   []LogEntry
}

func New(baseDir, outputDir, replaceSpace string, verbose, dryRun, undo, prompt bool) *Organizer {
	return &Organizer{
		baseDir:      baseDir,
		outputDir:    outputDir,
		replaceSpace: replaceSpace,
		verbose:      verbose,
		dryRun:       dryRun,
		undo:         undo,
		prompt:       prompt,
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

	o.printSummary(startTime)
	return nil
}
