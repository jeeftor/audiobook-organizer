package organizer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

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
			if err := o.OrganizeAudiobook(path, metadataPath); err != nil {
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

// moveFile handles moving files between directories, even across devices
func (o *Organizer) moveFile(sourcePath, destPath string) error {
	// First attempt to rename (move) the file
	err := os.Rename(sourcePath, destPath)
	if err == nil {
		return nil
	}

	// If cross-device error, fallback to copy and delete
	if strings.Contains(err.Error(), "cross-device link") {
		// Open source file
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to open source file: %v", err)
		}
		defer sourceFile.Close()

		// Create destination file
		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create destination file: %v", err)
		}
		defer destFile.Close()

		// Copy the contents
		_, err = io.Copy(destFile, sourceFile)
		if err != nil {
			return fmt.Errorf("failed to copy file contents: %v", err)
		}

		// Close files before attempting removal
		sourceFile.Close()
		destFile.Close()

		// Remove the source file
		err = os.Remove(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to remove source file: %v", err)
		}

		return nil
	}

	// If it's not a cross-device error, return the original error
	return err
}

func (o *Organizer) OrganizeAudiobook(sourcePath, metadataPath string) error {
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

	authorDir := o.SanitizePath(strings.Join(metadata.Authors, ","))
	titleDir := o.SanitizePath(metadata.Title)

	targetBase := o.baseDir
	if o.outputDir != "" {
		targetBase = o.outputDir
	}

	var targetPath string
	if len(metadata.Series) > 0 {
		cleanedSeries := cleanSeriesName(metadata.Series[0])
		seriesDir := o.SanitizePath(cleanedSeries)
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

	if o.prompt {
		if !o.PromptForConfirmation(metadata, sourcePath, targetPath) {
			color.Yellow("⏩ Skipping %s", metadata.Title)
			return nil
		}
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
			if err := o.moveFile(sourceName, targetName); err != nil {
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

		// Save log after each successful move
		if err := o.saveLog(); err != nil {
			color.Yellow("⚠️  Warning: couldn't save log: %v", err)
		}
	}

	return nil
}
