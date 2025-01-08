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
	rootCmd := &cobra.Command{
		Use:   "audiobook-organizer",
		Short: "Organize audiobooks based on metadata.json files",
		Run:   organize,
	}

	rootCmd.Flags().StringVar(&baseDir, "dir", "", "Base directory to scan")
	rootCmd.Flags().StringVar(&replaceSpace, "replace_space", "", "Character to replace spaces")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without making changes")
	rootCmd.Flags().BoolVar(&undo, "undo", false, "Restore files to their original locations")
	rootCmd.Flags().BoolVar(&prompt, "prompt", false, "Prompt for confirmation before moving each book")
	rootCmd.MarkFlagRequired("dir")

	if err := rootCmd.Execute(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

func organize(cmd *cobra.Command, args []string) {
	if undo {
		if err := undoMoves(); err != nil {
			color.Red("Error undoing moves: %v", err)
			os.Exit(1)
		}
		return
	}

	startTime := time.Now()
	err := filepath.Walk(baseDir, processDirectory)
	if err != nil {
		color.Red("Error walking directory: %v", err)
		os.Exit(1)
	}

	if !dryRun && len(logEntries) > 0 {
		if err := saveLog(); err != nil {
			color.Red("Error saving log: %v", err)
		}
	}

	printSummary(startTime)
}

func saveLog() error {
	logPath := filepath.Join(baseDir, logFileName)
	data, err := json.MarshalIndent(logEntries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(logPath, data, 0644)
}

func undoMoves() error {
	logPath := filepath.Join(baseDir, logFileName)
	data, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("no log file found at %s", logPath)
	}

	var entries []LogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("error parsing log: %v", err)
	}

	for _, entry := range entries {
		color.Yellow("Restoring files from %s to %s", entry.TargetPath, entry.SourcePath)
		if err := os.MkdirAll(entry.SourcePath, 0755); err != nil {
			color.Red("Error creating source directory: %v", err)
			continue
		}

		for _, file := range entry.Files {
			oldPath := filepath.Join(entry.TargetPath, file)
			newPath := filepath.Join(entry.SourcePath, file)
			if verbose {
				color.Blue("Moving %s to %s", oldPath, newPath)
			}
			if err := os.Rename(oldPath, newPath); err != nil {
				color.Red("Error moving %s: %v", oldPath, err)
			}
		}
	}

	if err := os.Remove(logPath); err != nil {
		color.Yellow("Warning: couldn't remove log file: %v", err)
	}

	return nil
}

func processDirectory(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		metadataPath := filepath.Join(path, "metadata.json")
		if _, err := os.Stat(metadataPath); err == nil {
			summary.MetadataFound = append(summary.MetadataFound, metadataPath)
			if err := organizeAudiobook(path, metadataPath); err != nil {
				color.Red("Error organizing %s: %v", path, err)
			}
		} else if verbose {
			summary.MetadataMissing = append(summary.MetadataMissing, path)
			color.Yellow("No metadata.json found in %s", path)
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
	color.Yellow("\nBook found:")
	color.White("  Title: %s", metadata.Title)
	color.White("  Authors: %s", strings.Join(metadata.Authors, ", "))
	if len(metadata.Series) > 0 {
		color.White("  Series: %s", metadata.Series[0])
	}
	color.Cyan("\nProposed move:")
	color.White("  From: %s", sourcePath)
	color.White("  To: %s", targetPath)

	fmt.Print("\nProceed with move? [y/N] ")
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
			color.White("  Series: %v", metadata.Series)
		}
	}

	authorDir := processPath(strings.Join(metadata.Authors, ","))
	titleDir := processPath(metadata.Title)

	var targetPath string
	if len(metadata.Series) > 0 {
		seriesDir := processPath(metadata.Series[0])
		targetPath = filepath.Join(baseDir, authorDir, seriesDir, titleDir)
	} else {
		targetPath = filepath.Join(baseDir, authorDir, titleDir)
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
			color.Yellow("Skipping %s", metadata.Title)
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
			color.Blue("%sMoving %s to %s", prefix, sourceName, targetName)
		}

		if !dryRun {
			if err := os.Rename(sourceName, targetName); err != nil {
				color.Red("Error moving %s: %v", sourceName, err)
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
	color.White("Duration: %v", duration.Round(time.Millisecond))

	color.Green("\nMetadata files found: %d", len(summary.MetadataFound))
	if len(summary.MetadataFound) > 0 {
		fmt.Println("\nValid Audiobooks Found:")
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
					color.Green("     Series: %s", metadata.Series[0])
				}
			}
		}
	}

	if len(summary.MetadataMissing) > 0 {
		color.Yellow("\nDirectories without metadata: %d", len(summary.MetadataMissing))
		if verbose {
			for _, path := range summary.MetadataMissing {
				fmt.Printf("  - %s\n", path)
			}
		}
	}

	color.Cyan("\nMoves planned/executed: %d", len(summary.Moves))
	for _, move := range summary.Moves {
		fmt.Printf("  From: %s\n  To: %s\n\n", move.From, move.To)
	}

	if dryRun {
		color.Yellow("\nThis was a dry run - no files were actually moved")
	}
}
