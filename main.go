package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type Metadata struct {
	Authors []string `json:"authors"`
	Title   string   `json:"title"`
	Series  []string `json:"series"`
}

type LogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	SourcePath string    `json:"source_path"`
	TargetPath string    `json:"target_path"`
	Files      []string  `json:"files"`
}

type Summary struct {
	MetadataFound   []string
	MetadataMissing []string
	Moves           []MoveSummary
}

type MoveSummary struct {
	From string
	To   string
}

var (
	baseDir      string
	outputDir    string
	replaceSpace string
	verbose      bool
	dryRun       bool
	undo         bool
	prompt       bool
	summary      Summary
	logEntries   []LogEntry
)

const logFileName = ".abook-org.log"

func main() {
	// Add some color to startup
	color.Cyan("🎧 Audiobook Organizer")
	color.Cyan("=====================")

	rootCmd := &cobra.Command{
		Use:   "audiobook-organizer",
		Short: "Organize audiobooks based on metadata.json files",
		Run:   organize,
	}

	rootCmd.Flags().StringVar(&baseDir, "dir", "", "Base directory to scan")
	rootCmd.Flags().StringVar(&outputDir, "out", "", "Output directory (if different from base directory)")
	rootCmd.Flags().StringVar(&replaceSpace, "replace_space", "", "Character to replace spaces")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without making changes")
	rootCmd.Flags().BoolVar(&undo, "undo", false, "Restore files to their original locations")
	rootCmd.Flags().BoolVar(&prompt, "prompt", false, "Prompt for confirmation before moving each book")
	rootCmd.MarkFlagRequired("dir")

	if err := rootCmd.Execute(); err != nil {
		color.Red("❌ Error: %v", err)
		os.Exit(1)
	}
}

func cleanSeriesName(series string) string {
	// Find the last occurrence of " #" and remove everything after it
	if idx := strings.LastIndex(series, " #"); idx != -1 {
		return strings.TrimSpace(series[:idx])
	}
	return series
}

func organize(cmd *cobra.Command, args []string) {
	// Clean and resolve the paths
	color.Blue("🔍 Resolving paths...")
	resolvedBaseDir, err := filepath.EvalSymlinks(filepath.Clean(baseDir))
	if err != nil {
		color.Red("❌ Error resolving base directory path: %v", err)
		os.Exit(1)
	}
	baseDir = resolvedBaseDir

	if outputDir != "" {
		resolvedOutputDir, err := filepath.EvalSymlinks(filepath.Clean(outputDir))
		if err != nil {
			color.Red("❌ Error resolving output directory path: %v", err)
			os.Exit(1)
		}
		outputDir = resolvedOutputDir
	}

	if undo {
		color.Yellow("↩️  Undoing previous operations...")
		if err := undoMoves(); err != nil {
			color.Red("❌ Error undoing moves: %v", err)
			os.Exit(1)
		}
		return
	}

	if dryRun {
		color.Yellow("🔍 Running in dry-run mode - no files will be moved")
	}

	startTime := time.Now()
	color.Blue("📚 Scanning for audiobooks...")
	err = filepath.Walk(baseDir, processDirectory)
	if err != nil {
		color.Red("❌ Error walking directory: %v", err)
		os.Exit(1)
	}

	if !dryRun && len(logEntries) > 0 {
		color.Blue("💾 Saving operation log...")
		if err := saveLog(); err != nil {
			color.Red("❌ Error saving log: %v", err)
		}
	}

	printSummary(startTime)
}

func saveLog() error {
	// Use output directory for log if specified, otherwise use base directory
	logBase := baseDir
	if outputDir != "" {
		logBase = outputDir
	}
	logPath := filepath.Join(logBase, logFileName)
	data, err := json.MarshalIndent(logEntries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(logPath, data, 0644)
}

func undoMoves() error {
	// Use output directory for log if specified, otherwise use base directory
	logBase := baseDir
	if outputDir != "" {
		logBase = outputDir
	}
	logPath := filepath.Join(logBase, logFileName)

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
			if verbose {
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

func processDirectory(path string, info os.FileInfo, err error) error {
	// Handle walk errors more gracefully
	if err != nil {
		// If the file doesn't exist, it might have been moved by a previous operation
		if os.IsNotExist(err) {
			if verbose {
				color.Yellow("⏩ Skipping non-existent path (likely moved): %s", path)
			}
			return nil
		}
		return err
	}

	if info.IsDir() {
		metadataPath := filepath.Join(path, "metadata.json")
		if _, err := os.Stat(metadataPath); err == nil {
			summary.MetadataFound = append(summary.MetadataFound, metadataPath)
			if err := organizeAudiobook(path, metadataPath); err != nil {
				color.Red("❌ Error organizing %s: %v", path, err)
			}
			// Skip processing subdirectories of organized books
			return filepath.SkipDir
		} else if verbose {
			summary.MetadataMissing = append(summary.MetadataMissing, path)
			color.Yellow("⚠️  No metadata.json found in %s", path)
		}
	}
	return nil
}

func processPath(s string) string {
	if replaceSpace != "" {
		return strings.ReplaceAll(s, " ", replaceSpace)
	}
	return s
}

func promptForConfirmation(metadata Metadata, sourcePath, targetPath string) bool {
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

func organizeAudiobook(sourcePath, metadataPath string) error {
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

	if verbose {
		color.Green("📚 Metadata detected in %s:", metadataPath)
		color.White("  Authors: %v", metadata.Authors)
		color.White("  Title: %s", metadata.Title)
		if len(metadata.Series) > 0 {
			cleanedSeries := cleanSeriesName(metadata.Series[0])
			color.White("  Series: %s (%s)", metadata.Series[0], cleanedSeries)
		}
	}

	authorDir := processPath(strings.Join(metadata.Authors, ","))
	titleDir := processPath(metadata.Title)

	// Determine which directory to use as base for output
	targetBase := baseDir
	if outputDir != "" {
		targetBase = outputDir
	}

	var targetPath string
	if len(metadata.Series) > 0 {
		cleanedSeries := cleanSeriesName(metadata.Series[0])
		seriesDir := processPath(cleanedSeries)
		targetPath = filepath.Join(targetBase, authorDir, seriesDir, titleDir)
	} else {
		targetPath = filepath.Join(targetBase, authorDir, titleDir)
	}

	// Clean both paths to ensure consistent comparison
	cleanSourcePath := filepath.Clean(sourcePath)
	cleanTargetPath := filepath.Clean(targetPath)

	// If the book is already in the correct location, skip it
	if cleanSourcePath == cleanTargetPath {
		if verbose {
			color.Green("✅ Book already in correct location: %s", cleanSourcePath)
		}
		return nil
	}

	if verbose {
		color.Cyan("🔄 Moving contents from %s to %s", sourcePath, targetPath)
	}

	if !dryRun {
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("error creating target directory: %v", err)
		}
	}

	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("error reading source directory: %v", err)
	}

	summary.Moves = append(summary.Moves, MoveSummary{
		From: sourcePath,
		To:   targetPath,
	})

	if prompt && !dryRun {
		if !promptForConfirmation(metadata, sourcePath, targetPath) {
			color.Yellow("⏩ Skipping %s", metadata.Title)
			return nil
		}
	}

	var fileNames []string
	for _, entry := range entries {
		fileNames = append(fileNames, entry.Name())
		sourceName := filepath.Join(sourcePath, entry.Name())
		targetName := filepath.Join(targetPath, entry.Name())

		if verbose || dryRun {
			prefix := "[DRY-RUN] "
			if !dryRun {
				prefix = ""
			}
			color.Blue("📦 %sMoving %s to %s", prefix, sourceName, targetName)
		}

		if !dryRun {
			if err := os.Rename(sourceName, targetName); err != nil {
				color.Red("❌ Error moving %s: %v", sourceName, err)
			}
		}
	}

	if !dryRun {
		logEntries = append(logEntries, LogEntry{
			Timestamp:  time.Now(),
			SourcePath: sourcePath,
			TargetPath: targetPath,
			Files:      fileNames,
		})
	}

	return nil
}

func printSummary(startTime time.Time) {
	duration := time.Since(startTime)

	fmt.Println("\n📊 Summary Report")
	color.White("⏱️  Duration: %v", duration.Round(time.Millisecond))

	color.Green("\n📚 Metadata files found: %d", len(summary.MetadataFound))
	if len(summary.MetadataFound) > 0 {
		fmt.Println("\n📖 Valid Audiobooks Found:")
		for _, path := range summary.MetadataFound {
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

	if len(summary.MetadataMissing) > 0 {
		color.Yellow("\n⚠️  Directories without metadata: %d", len(summary.MetadataMissing))
		if verbose {
			for _, path := range summary.MetadataMissing {
				fmt.Printf("  - %s\n", path)
			}
		}
	}

	color.Cyan("\n🔄 Moves planned/executed: %d", len(summary.Moves))
	for _, move := range summary.Moves {
		fmt.Printf("  From: %s\n  To: %s\n\n", move.From, move.To)
	}

	if dryRun {
		color.Yellow("\n🔍 This was a dry run - no files were actually moved")
	} else {
		color.Green("\n✅ Organization complete!")
	}
}
