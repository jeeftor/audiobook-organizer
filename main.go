package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type Metadata struct {
	Authors []string `json:"authors"`
	Title   string   `json:"title"`
}

var (
	baseDir      string
	replaceSpace string
	verbose      bool
	dryRun       bool
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
		fmt.Println(err)
		os.Exit(1)
	}
}

func organize(cmd *cobra.Command, args []string) {
	err := filepath.Walk(baseDir, processDirectory)
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
}

func processDirectory(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		metadataPath := filepath.Join(path, "metadata.json")
		if _, err := os.Stat(metadataPath); err == nil {
			if err := organizeAudiobook(path, metadataPath); err != nil {
				fmt.Printf("Error organizing %s: %v\n", path, err)
			}
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

	authorDir := strings.Join(metadata.Authors, ",")
	authorDir = strings.ReplaceAll(authorDir, " ", replaceSpace)
	titleDir := strings.ReplaceAll(metadata.Title, " ", replaceSpace)

	targetPath := filepath.Join(baseDir, authorDir, titleDir)

	if verbose {
		fmt.Printf("Moving contents from %s to %s\n", sourcePath, targetPath)
	}

	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("error creating target directory: %v", err)
	}

	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("error reading source directory: %v", err)
	}

	for _, entry := range entries {
		sourceName := filepath.Join(sourcePath, entry.Name())
		targetName := filepath.Join(targetPath, entry.Name())

		if verbose || dryRun {
			fmt.Printf("[DRY-RUN] Moving %s to %s\n", sourceName, targetName)
		}

		if !dryRun {
			if err := os.Rename(sourceName, targetName); err != nil {
				fmt.Printf("Error moving %s: %v\n", sourceName, err)
			}
		}
	}

	return nil
}