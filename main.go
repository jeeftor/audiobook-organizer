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
	summary      Summary
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "audiobook-organizer",
		Short: "Organize audiobooks based on metadata.json files",
		Run:   organize,
	}

	rootCmd.Flags().StringVar(&baseDir, "dir", "", "Base directory to scan")
	rootCmd.Flags().StringVar(&replaceSpace, "replace_space", ".", "Character to replace spaces")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without making changes")
	rootCmd.MarkFlagRequired("dir")

	if err := rootCmd.Execute(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

func organize(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	
	err := filepath.Walk(baseDir, processDirectory)
	if err != nil {
		color.Red("Error walking directory: %v", err)
		os.Exit(1)
	}

	printSummary(startTime)
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
	}

	authorDir := strings.Join(metadata.Authors, ",")
	authorDir = strings.ReplaceAll(authorDir, " ", replaceSpace)
	titleDir := strings.ReplaceAll(metadata.Title, " ", replaceSpace)

	targetPath := filepath.Join(baseDir, authorDir, titleDir)

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

	for _, entry := range entries {
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

	return nil
}

func printSummary(startTime time.Time) {
	duration := time.Since(startTime)
	
	fmt.Println("\n📊 Summary Report")
	color.White("Duration: %v", duration.Round(time.Millisecond))
	
	color.Green("\nMetadata files found: %d", len(summary.MetadataFound))
	if verbose {
		for _, path := range summary.MetadataFound {
			fmt.Printf("  - %s\n", path)
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