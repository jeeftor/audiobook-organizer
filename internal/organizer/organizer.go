package organizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

const logFileName = ".abook-org.log"

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
	return filepath.Join(logBase, logFileName)
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

func (o *Organizer) saveLog() error {
	logPath := o.GetLogPath()
	data, err := json.MarshalIndent(o.logEntries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(logPath, data, 0644)
}

func (o *Organizer) undoMoves() error {
	logPath := o.GetLogPath()
	data, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("no log file found at %s", logPath)
	}

	var entries []LogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("error parsing log: %v", err)
	}

	for _, entry := range entries {
		color.Yellow("↩️  Restoring files from %s to %s", entry.TargetPath, entry.SourcePath)
		if err := os.MkdirAll(entry.SourcePath, 0755); err != nil {
			color.Red("❌ Error creating source directory: %v", err)
			continue
		}

		for _, file := range entry.Files {
			oldPath := filepath.Join(entry.TargetPath, file)
			newPath := filepath.Join(entry.SourcePath, file)
			if o.verbose {
				color.Blue("📦 Moving %s to %s", oldPath, newPath)
			}
			if err := os.Rename(oldPath, newPath); err != nil {
				color.Red("❌ Error moving %s: %v", oldPath, err)
			}
		}
	}

	if err := os.Remove(logPath); err != nil {
		color.Yellow("⚠️  Warning: couldn't remove log file: %v", err)
	}

	return nil
}

func (o *Organizer) processDirectory(path string, info os.FileInfo, err error) error {
	if err != nil {
		if os.IsNotExist(err) {
			if o.verbose {
				color.Yellow("⏩ Skipping non-existent path (likely moved): %s", path)
			}
			return nil
		}
		return err
	}

	if info.IsDir() {
		metadataPath := filepath.Join(path, "metadata.json")
		if _, err := os.Stat(metadataPath); err == nil {
			o.summary.MetadataFound = append(o.summary.MetadataFound, metadataPath)
			if err := o.organizeAudiobook(path, metadataPath); err != nil {
				color.Red("❌ Error organizing %s: %v", path, err)
			}
			return filepath.SkipDir
		} else if o.verbose {
			o.summary.MetadataMissing = append(o.summary.MetadataMissing, path)
			color.Yellow("⚠️  No metadata.json found in %s", path)
		}
	}
	return nil
}

func (o *Organizer) processPath(s string) string {
	if o.replaceSpace != "" {
		return strings.ReplaceAll(s, " ", o.replaceSpace)
	}
	return s
}

func cleanSeriesName(series string) string {
	if idx := strings.LastIndex(series, " #"); idx != -1 {
		return strings.TrimSpace(series[:idx])
	}
	return series
}

func (o *Organizer) promptForConfirmation(metadata Metadata, sourcePath, targetPath string) bool {
	color.Yellow("\n📖 Book found:")
	color.White("  Title: %s", metadata.Title)
	color.White("  Authors: %s", strings.Join(metadata.Authors, ", "))
	if len(metadata.Series) > 0 {
		cleanedSeries := cleanSeriesName(metadata.Series[0])
		color.White("  Series: %s", cleanedSeries)
	}

	color.Cyan("\n📝 Proposed move:")
	color.White("  From: %s", sourcePath)
	color.White("  To: %s", targetPath)

	fmt.Print("\n❓ Proceed with move? [y/N] ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

func (o *Organizer) organizeAudiobook(sourcePath, metadataPath string) error {
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("error reading metadata: %v", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("error parsing metadata: %v", err)
	}

	if len(metadata.Authors) == 0 || metadata.Title == "" {
		return fmt.Errorf("missing required metadata fields")
	}

	if o.verbose {
		color.Green("📚 Metadata detected in %s:", metadataPath)
		color.White("  Authors: %v", metadata.Authors)
		color.White("  Title: %s", metadata.Title)
		if len(metadata.Series) > 0 {
			cleanedSeries := cleanSeriesName(metadata.Series[0])
			color.White("  Series: %s (%s)", metadata.Series[0], cleanedSeries)
		}
	}

	authorDir := o.processPath(strings.Join(metadata.Authors, ","))
	titleDir := o.processPath(metadata.Title)

	targetBase := o.baseDir
	if o.outputDir != "" {
		targetBase = o.outputDir
	}

	var targetPath string
	if len(metadata.Series) > 0 {
		cleanedSeries := cleanSeriesName(metadata.Series[0])
		seriesDir := o.processPath(cleanedSeries)
		targetPath = filepath.Join(targetBase, authorDir, seriesDir, titleDir)
	} else {
		targetPath = filepath.Join(targetBase, authorDir, titleDir)
	}

	cleanSourcePath := filepath.Clean(sourcePath)
	cleanTargetPath := filepath.Clean(targetPath)

	if cleanSourcePath == cleanTargetPath {
		if o.verbose {
			color.Green("✅ Book already in correct location: %s", cleanSourcePath)
		}
		return nil
	}

	if o.verbose {
		color.Cyan("🔄 Moving contents from %s to %s", sourcePath, targetPath)
	}

	if !o.dryRun {
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("error creating target directory: %v", err)
		}
	}

	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("error reading source directory: %v", err)
	}

	o.summary.Moves = append(o.summary.Moves, MoveSummary{
		From: sourcePath,
		To:   targetPath,
	})

	if o.prompt && !o.dryRun {
		if !o.promptForConfirmation(metadata, sourcePath, targetPath) {
			color.Yellow("⏩ Skipping %s", metadata.Title)
			return nil
		}
	}

	var fileNames []string
	for _, entry := range entries {
		fileNames = append(fileNames, entry.Name())
		sourceName := filepath.Join(sourcePath, entry.Name())
		targetName := filepath.Join(targetPath, entry.Name())

		if o.verbose || o.dryRun {
			prefix := "[DRY-RUN] "
			if !o.dryRun {
				prefix = ""
			}
			color.Blue("📦 %sMoving %s to %s", prefix, sourceName, targetName)
		}

		if !o.dryRun {
			if err := os.Rename(sourceName, targetName); err != nil {
				color.Red("❌ Error moving %s: %v", sourceName, err)
			}
		}
	}

	if !o.dryRun {
		o.logEntries = append(o.logEntries, LogEntry{
			Timestamp:  time.Now(),
			SourcePath: sourcePath,
			TargetPath: targetPath,
			Files:      fileNames,
		})
	}

	return nil
}

func (o *Organizer) printSummary(startTime time.Time) {
	duration := time.Since(startTime)

	fmt.Println("\n📊 Summary Report")
	color.White("⏱️  Duration: %v", duration.Round(time.Millisecond))

	color.Green("\n📚 Metadata files found: %d", len(o.summary.MetadataFound))
	if len(o.summary.MetadataFound) > 0 {
		fmt.Println("\n📖 Valid Audiobooks Found:")
		for _, path := range o.summary.MetadataFound {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var metadata Metadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				continue
			}
			if len(metadata.Authors) > 0 && metadata.Title != "" {
				color.Green("  📚 %s by %s", metadata.Title, strings.Join(metadata.Authors, ", "))
				if len(metadata.Series) > 0 {
					cleanedSeries := cleanSeriesName(metadata.Series[0])
					color.Green("     📖 Series: %s", cleanedSeries)
				}
			}
		}
	}

	if len(o.summary.MetadataMissing) > 0 {
		color.Yellow("\n⚠️  Directories without metadata: %d", len(o.summary.MetadataMissing))
		if o.verbose {
			for _, path := range o.summary.MetadataMissing {
				fmt.Printf("  - %s\n", path)
			}
		}
	}

	color.Cyan("\n🔄 Moves planned/executed: %d", len(o.summary.Moves))
	for _, move := range o.summary.Moves {
		fmt.Printf("  From: %s\n  To: %s\n\n", move.From, move.To)
	}

	if o.dryRun {
		color.Yellow("\n🔍 This was a dry run - no files were actually moved")
	} else {
		color.Green("\n✅ Organization complete!")
	}
}
