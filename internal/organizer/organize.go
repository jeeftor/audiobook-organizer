package organizer

import (
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
			if o.config.Verbose {
				color.Yellow("⏩ Skipping non-existent path (likely moved): %s", path)
			}
			return nil
		}
		return err
	}

	// Handle flat mode differently - we process files directly
	if o.config.Flat {
		return o.processFlatDirectory(path, info)
	}

	// Standard hierarchical mode processing
	if !info.IsDir() {
		return nil
	}

	// Skip if this is the output directory
	if o.config.OutputDir != "" && path == o.config.OutputDir {
		return filepath.SkipDir
	}

	// Try to organize using available metadata
	organized, err := o.tryOrganizeWithMetadata(path)
	if err != nil {
		color.Red("❌ Error processing %s: %v", path, err)
		return nil
	}

	// If successfully organized, skip further processing of this directory
	if organized {
		return filepath.SkipDir
	}

	// Handle directories without metadata
	if o.config.Verbose {
		o.summary.MetadataMissing = append(o.summary.MetadataMissing, path)
		color.Yellow("⚠️  No metadata found in %s", path)
	}

	// Check if directory is empty and should be removed
	if o.config.RemoveEmpty && path != o.config.BaseDir {
		if isEmptyDir(path) {
			if o.config.Verbose {
				color.Yellow("🗑️  Found empty directory during scan: %s", path)
			}

			// Prompt for removal if enabled
			if o.config.Prompt {
				if !o.PromptForDirectoryRemoval(path, false) {
					if o.config.Verbose {
						color.Yellow("⏩ Skipping removal of %s", path)
					}
					return nil
				}
			}

			// Remove the empty directory
			if !o.config.DryRun {
				if o.config.Verbose {
					color.Yellow("🗑️  Removing empty directory: %s", path)
				}

				if err := os.Remove(path); err != nil {
					color.Red("❌ Error removing directory: %v", err)
					return nil
				}

				// Add to summary
				o.summary.EmptyDirsRemoved = append(o.summary.EmptyDirsRemoved, path)

				// After removing the directory, check if parent is now empty,
				// but don't go beyond the input directory
				parentDir := filepath.Dir(path)
				if parentDir != o.config.BaseDir {
					if err := o.cleanEmptyParents(parentDir, o.config.BaseDir); err != nil {
						color.Red("❌ Error cleaning parent directories: %v", err)
					}
				}
			}

			// Skip further processing of this directory since it's been removed
			return filepath.SkipDir
		}
	}

	return nil
}

func (o *Organizer) processFlatDirectory(path string, info os.FileInfo) error {
	// Only process directories in flat mode, except at the root level
	if info.IsDir() {
		// Skip output directory
		if o.config.OutputDir != "" && path == o.config.OutputDir {
			return filepath.SkipDir
		}

		// At the root level, continue processing to find files
		if path == o.config.BaseDir {
			return nil
		}

		// Skip subdirectories in flat mode
		if o.config.Verbose {
			color.Yellow("⏩ Skipping subdirectory in flat mode: %s", path)
		}
		return filepath.SkipDir
	}

	// Only process EPUB, MP3, and M4B files in flat mode
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".epub" && ext != ".mp3" && ext != ".m4b" {
		return nil
	}

	// Create a metadata provider for this specific file (audio or epub)
	provider := NewFileMetadataProvider(path)

	// Try to get metadata - we don't need the actual metadata here,
	// just checking if we can extract it successfully
	_, err := provider.GetMetadata()
	if err != nil {
		if o.config.Verbose {
			color.Yellow("⚠️ Could not extract metadata from %s: %v", path, err)
		}
		return nil
	}

	// Organize this individual file
	if err := o.OrganizeAudiobook(filepath.Dir(path), provider); err != nil {
		color.Red("❌ Error organizing %s: %v", path, err)
	}

	return nil
}

func (o *Organizer) tryOrganizeWithMetadata(path string) (bool, error) {
	// First try JSON metadata if it exists
	metadataPath := filepath.Join(path, "metadata.json")
	if _, err := os.Stat(metadataPath); err == nil {
		o.summary.MetadataFound = append(o.summary.MetadataFound, metadataPath)
		if err := o.OrganizeAudiobook(path, NewJSONMetadataProvider(metadataPath)); err != nil {
			return false, fmt.Errorf("error organizing with JSON metadata: %v", err)
		}
		return true, nil
	}

	// If JSON metadata not found and we should use embedded metadata, try EPUB
	if o.config.UseEmbeddedMetadata {
		// Check if there are EPUB files in the directory
		epubPath, err := FindEPUBInDirectory(path)
		if err == nil {
			epubProvider := NewEPUBMetadataProvider(epubPath)
			metadata, err := epubProvider.GetMetadata()

			if err == nil && metadata.Title != "" && len(metadata.Authors) > 0 {
				color.Green("📚 Found metadata in EPUB file: %s", epubPath)
				if err := o.OrganizeAudiobook(path, epubProvider); err != nil {
					return false, fmt.Errorf("error organizing with EPUB metadata: %v", err)
				}
				return true, nil
			} else if o.config.Verbose {
				color.Yellow("⚠️ EPUB found but metadata extraction failed: %s", epubPath)
			}
		} else if o.config.Verbose && o.config.UseEmbeddedMetadata {
			color.Yellow("⚠️ No EPUB files found in %s", path)
		}
	}

	return false, nil
}

func (o *Organizer) cleanEmptyParents(dir string, stopAt string) error {
	// Stop if we've reached the boundary directory
	if dir == stopAt || (o.config.OutputDir != "" && dir == o.config.OutputDir) {
		return nil
	}

	// Get absolute paths for more reliable comparison
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	absStopAt, err := filepath.Abs(stopAt)
	if err != nil {
		return err
	}

	// Extra safety check: make sure we don't go above the stop directory
	if absDir == absStopAt || isSubPathOf(absStopAt, absDir) {
		return nil
	}

	// Check if directory exists (it might have been removed by another operation)
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	// Check if directory is empty
	if !isEmptyDir(dir) {
		return nil
	}

	// Prompt for removal of parent directory if enabled
	if o.config.Prompt {
		if !o.PromptForDirectoryRemoval(dir, true) {
			if o.config.Verbose {
				color.Yellow("⏩ Skipping removal of parent directory %s", dir)
			}
			return nil
		}
	}

	if o.config.Verbose {
		color.Yellow("🗑️  Removing newly empty parent directory: %s", dir)
	}

	// Store parent before removing current directory
	parentDir := filepath.Dir(dir)

	// Remove the empty directory
	if !o.config.DryRun {
		if err := os.Remove(dir); err != nil {
			return fmt.Errorf("failed to remove empty parent directory %s: %v", dir, err)
		}

		// Add to summary
		o.summary.EmptyDirsRemoved = append(o.summary.EmptyDirsRemoved, dir)
	}

	// Recursively check the parent directory
	return o.cleanEmptyParents(parentDir, stopAt)
}

func isSubPathOf(parent, child string) bool {
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)

	// Split both paths into components
	parentParts := strings.Split(parent, string(filepath.Separator))
	childParts := strings.Split(child, string(filepath.Separator))

	// Child path must be longer than parent path
	if len(childParts) <= len(parentParts) {
		return false
	}

	// Check if child starts with parent
	for i := range parentParts {
		if parentParts[i] != childParts[i] {
			return false
		}
	}

	return true
}

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

func (o *Organizer) OrganizeAudiobook(sourcePath string, provider MetadataProvider) error {
	// Get metadata from provider
	metadata, err := provider.GetMetadata()
	if err != nil {
		return fmt.Errorf("error getting metadata: %v", err)
	}

	// Validate metadata
	if err := o.validateMetadata(metadata); err != nil {
		return err
	}

	// Log metadata if Verbose
	o.logMetadataIfVerbose(metadata, provider)

	// Calculate target path
	targetPath, err := o.calculateTargetPath(metadata)
	if err != nil {
		return err
	}

	// Check if already in correct location
	cleanSourcePath := filepath.Clean(sourcePath)
	cleanTargetPath := filepath.Clean(targetPath)
	if cleanSourcePath == cleanTargetPath {
		if o.config.Verbose {
			color.Green("✅ Book already in correct location: %s", cleanSourcePath)
		}
		return nil
	}

	// Prompt for confirmation if needed
	if o.config.Prompt && !o.promptForMoveConfirmation(metadata, sourcePath, targetPath) {
		color.Yellow("⏩ Skipping %s", metadata.Title)
		return nil
	}

	// Move files
	fileNames, err := o.moveFiles(sourcePath, targetPath)
	if err != nil {
		return err
	}

	// Update log and clean up
	if !o.config.DryRun {
		o.updateLogAndCleanup(sourcePath, targetPath, fileNames)
	}

	return nil
}

func (o *Organizer) validateMetadata(metadata Metadata) error {
	if len(metadata.Authors) == 0 || metadata.Title == "" {
		return fmt.Errorf("missing required metadata fields")
	}
	return nil
}

func (o *Organizer) logMetadataIfVerbose(metadata Metadata, provider MetadataProvider) {
	if !o.config.Verbose {
		return
	}

	// Get provider type for logging purposes
	providerType := "provider"
	switch provider.(type) {
	case *JSONMetadataProvider:
		providerType = "JSON metadata"
	case *EPUBMetadataProvider:
		providerType = "EPUB metadata"
	}

	color.Green("📚 Metadata detected from %s:", providerType)
	color.White("  Authors: %v", metadata.Authors)
	color.White("  Title: %s", metadata.Title)
	if len(metadata.Series) > 0 {
		cleanedSeries := cleanSeriesName(metadata.Series[0])
		color.White("  Series: %s (%s)", metadata.Series[0], cleanedSeries)
	}
}

func (o *Organizer) calculateTargetPath(metadata Metadata) (string, error) {
	authorDir := o.SanitizePath(strings.Join(metadata.Authors, ","))
	titleDir := o.SanitizePath(metadata.Title)

	targetBase := o.config.BaseDir
	if o.config.OutputDir != "" {
		targetBase = o.config.OutputDir
	}

	var targetPath string
	if len(metadata.Series) > 0 {
		cleanedSeries := cleanSeriesName(metadata.Series[0])
		seriesDir := o.SanitizePath(cleanedSeries)
		targetPath = filepath.Join(targetBase, authorDir, seriesDir, titleDir)
	} else {
		targetPath = filepath.Join(targetBase, authorDir, titleDir)
	}

	return targetPath, nil
}

func (o *Organizer) promptForMoveConfirmation(metadata Metadata, sourcePath, targetPath string) bool {
	return o.PromptForConfirmation(metadata, sourcePath, targetPath)
}

func (o *Organizer) moveFiles(sourcePath, targetPath string) ([]string, error) {
	if o.config.Verbose {
		color.Cyan("🔄 Moving contents from %s to %s", sourcePath, targetPath)
	}

	if !o.config.DryRun {
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return nil, fmt.Errorf("error creating target directory: %v", err)
		}
	}

	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("error reading source directory: %v", err)
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

		if o.config.Verbose || o.config.DryRun {
			prefix := "[DRY-RUN] "
			if !o.config.DryRun {
				prefix = ""
			}
			color.Blue("📦 %sMoving %s to %s", prefix, sourceName, targetName)
		}

		if !o.config.DryRun {
			if err := o.moveFile(sourceName, targetName); err != nil {
				color.Red("❌ Error moving %s: %v", sourceName, err)
			}
		}
	}

	return fileNames, nil
}

func (o *Organizer) updateLogAndCleanup(sourcePath, targetPath string, fileNames []string) {
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

	// Check and remove empty source directory
	if err := o.removeEmptyDirs(sourcePath); err != nil {
		color.Yellow("⚠️  Warning: couldn't remove empty directory: %v", err)
	}
}
