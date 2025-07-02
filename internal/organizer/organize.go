// internal/organizer/organize.go
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

	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".epub" && ext != ".mp3" && ext != ".m4b" {
		return nil
	}

	provider := &FileMetadataProvider{
		filePath: path,
	}
	_, err := provider.GetMetadata()
	if err != nil {
		if o.config.Verbose {
			color.Yellow("⚠️ Could not extract metadata from %s: %v", path, err)
		}
		return nil
	}

	if err := o.OrganizeSingleFile(path, provider); err != nil {
		color.Red("❌ Error organizing %s: %v", path, err)
	}

	return nil
}

func (o *Organizer) tryOrganizeWithMetadata(path string) (bool, error) {
	// First try embedded metadata if enabled
	if o.config.UseEmbeddedMetadata {
		// Try EPUB files first
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
		} else if o.config.Verbose {
			color.Yellow("⚠️ No EPUB files found in %s", path)
		}

		// Try audio files if no EPUB metadata found
		audioPath, err := FindAudioFileInDirectory(path)
		if err == nil {
			audioProvider := NewAudioMetadataProvider(audioPath)
			metadata, err := audioProvider.GetMetadata()

			if err == nil && metadata.Title != "" && len(metadata.Authors) > 0 {
				color.Green("🔊 Found metadata in audio file: %s", audioPath)
				if err := o.OrganizeAudiobook(path, audioProvider); err != nil {
					return false, fmt.Errorf("error organizing with audio metadata: %v", err)
				}
				return true, nil
			} else if o.config.Verbose {
				color.Yellow("⚠️ Audio file found but metadata extraction failed: %s", audioPath)
			}
		} else if o.config.Verbose {
			color.Yellow("⚠️ No supported audio files found in %s", path)
		}
	}

	// If no embedded metadata found or UseEmbeddedMetadata is false, try metadata.json
	metadataPath := filepath.Join(path, "metadata.json")
	if _, err := os.Stat(metadataPath); err == nil {
		o.summary.MetadataFound = append(o.summary.MetadataFound, metadataPath)
		if err := o.OrganizeAudiobook(path, NewJSONMetadataProvider(metadataPath)); err != nil {
			return false, fmt.Errorf("error organizing with JSON metadata: %v", err)
		}
		return true, nil
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

	// Apply the organizer's field mapping to the metadata before logging
	metadata.FieldMapping = o.config.FieldMapping
	metadata.ApplyFieldMapping()

	// Add layout information to the metadata for field mapping display
	metadata.RawMetadata["layout"] = o.config.Layout

	// Log the metadata with the applied field mapping
	o.logMetadataIfVerbose(metadata, provider)

	// Validate metadata
	if err := o.validateMetadata(metadata); err != nil {
		return err
	}

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
	if len(metadata.Authors) == 0 || metadata.Authors[0] == "" {
		return fmt.Errorf("missing author information")
	}

	if metadata.Title == "" {
		return fmt.Errorf("missing title information")
	}

	return nil
}

// OrganizeSingleFile organizes a single file based on its metadata
// This is used in flat mode to move only the specific file being processed
func (o *Organizer) OrganizeSingleFile(filePath string, provider MetadataProvider) error {
	// Get metadata from provider
	metadata, err := provider.GetMetadata()
	if err != nil {
		return fmt.Errorf("error getting metadata: %v", err)
	}

	// Store a copy of the raw metadata for display
	rawMetadata := metadata.Clone()

	// Add layout information to the raw metadata
	rawMetadata.RawMetadata["layout"] = o.config.Layout

	// Log the raw metadata first if verbose
	if o.config.Verbose {
		color.Cyan("\n🔍 Found Auto-detected metadata")
		fmt.Println(rawMetadata.FormatMetadata())
	}

	// Apply the organizer's field mapping to the metadata
	metadata.FieldMapping = o.config.FieldMapping
	metadata.ApplyFieldMapping()

	// Add layout information to the metadata for field mapping display
	metadata.RawMetadata["layout"] = o.config.Layout

	// Log the field mapping information
	if o.config.Verbose {
		fmt.Println(metadata.FormatFieldMappingAndValues())
	}

	// Validate metadata
	if err := o.validateMetadata(metadata); err != nil {
		return err
	}

	// Calculate target path - but in flat mode, we'll override this
	var targetDir string

	// In flat mode, we need to ensure we're not using the source file path as part of the target path
	if o.config.Flat {
		// Use the output directory if specified, otherwise use the directory containing the source file
		baseDir := filepath.Dir(filePath)
		if o.config.OutputDir != "" {
			baseDir = o.config.OutputDir
		}

		// Recalculate the target directory based on the layout
		authorDir := o.SanitizePath(strings.Join(metadata.Authors, ","))
		titleDir := o.SanitizePath(metadata.Title)

		switch o.config.Layout {
		case "author-only":
			targetDir = filepath.Join(baseDir, authorDir)
		case "author-title":
			targetDir = filepath.Join(baseDir, authorDir, titleDir)
		case "author-series-title", "":
			if len(metadata.Series) > 0 && metadata.Series[0] != "__INVALID_SERIES__" {
				cleanedSeries := cleanSeriesName(metadata.Series[0])
				seriesDir := o.SanitizePath(cleanedSeries)
				targetDir = filepath.Join(baseDir, authorDir, seriesDir, titleDir)
			} else {
				targetDir = filepath.Join(baseDir, authorDir, titleDir)
			}
		default:
			targetDir = filepath.Join(baseDir, authorDir, titleDir)
		}
	} else {
		// For non-flat mode, use the standard target path calculation
		var err error
		targetDir, err = o.calculateTargetPath(metadata)
		if err != nil {
			return err
		}
	}

	// Generate the target filename
	ext := filepath.Ext(filePath)
	baseName := filepath.Base(filePath)
	baseName = strings.TrimSuffix(baseName, ext)

	// If we have a track number, prepend it to the filename
	var targetName string
	if metadata.TrackNumber > 0 {
		targetName = fmt.Sprintf("%02d - %s%s", metadata.TrackNumber, baseName, ext)
	} else {
		targetName = baseName + ext
	}

	targetPath := filepath.Join(targetDir, targetName)

	// Check if already in correct location
	cleanSourcePath := filepath.Clean(filePath)
	cleanTargetPath := filepath.Clean(targetPath)
	if cleanSourcePath == cleanTargetPath {
		if o.config.Verbose {
			color.Green("✅ File already in correct location: %s", cleanSourcePath)
		}
		return nil
	}

	// Prompt for confirmation if needed
	if o.config.Prompt && !o.promptForMoveConfirmation(metadata, filePath, targetPath) {
		color.Yellow("⏩ Skipping %s", metadata.Title)
		return nil
	}

	// Create target directory if it doesn't exist
	if !o.config.DryRun {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("error creating target directory: %v", err)
		}
	}

	// If we're in dry-run mode, show what would happen
	if o.config.DryRun {
		// Extract components from the target path for colorizing
		// Format: /path/to/Author/Series/Title/File.ext
		pathParts := strings.Split(targetPath, string(os.PathSeparator))

		// Define consistent colors for each component (matching the field mapping colors)
		titleColor := color.New(color.FgHiBlue).SprintFunc()
		seriesColor := color.New(color.FgGreen).SprintFunc()
		authorColor := color.New(color.FgMagenta).SprintFunc()
		fileColor := color.New(color.FgCyan).SprintFunc()

		// Color the path components if we have enough parts
		coloredPath := targetPath
		if len(pathParts) >= 4 { // We need at least 4 parts to have author/series/title/file
			// Find the indices of the components in the path
			// We work backwards from the end since those positions are more predictable
			fileIndex := len(pathParts) - 1
			titleIndex := len(pathParts) - 2
			seriesIndex := -1
			authorIndex := -1

			// If we have series info, it should be one level up from title
			if len(metadata.Series) > 0 && titleIndex > 0 {
				seriesIndex = titleIndex - 1
				// Author should be one level up from series
				if seriesIndex > 0 {
					authorIndex = seriesIndex - 1
				}
			} else if titleIndex > 0 {
				// If no series, author is directly above title
				authorIndex = titleIndex - 1
			}

			// Build colored path
			var coloredParts []string
			for i, part := range pathParts {
				switch {
				case i == fileIndex:
					coloredParts = append(coloredParts, fileColor(part))
				case i == titleIndex:
					coloredParts = append(coloredParts, titleColor(part))
				case i == seriesIndex && seriesIndex >= 0:
					coloredParts = append(coloredParts, seriesColor(part))
				case i == authorIndex && authorIndex >= 0:
					coloredParts = append(coloredParts, authorColor(part))
				default:
					coloredParts = append(coloredParts, part)
				}
			}

			// Reconstruct the path with colored components
			coloredPath = strings.Join(coloredParts, string(os.PathSeparator))
		}

		fmt.Printf("📦 [DRY-RUN] Moving %s to %s\n\n",
			color.YellowString("\"%s\"", filePath),
			coloredPath)
		return nil
	}

	// Move the single file
	if o.config.Verbose || o.config.DryRun {
		prefix := "[DRY-RUN] "
		if !o.config.DryRun {
			prefix = ""
		}
		color.Blue("📦 %sMoving %s to %s", prefix, filePath, targetPath)
	}

	if !o.config.DryRun {
		if err := o.moveFile(filePath, targetPath); err != nil {
			color.Red("❌ Error moving %s: %v", filePath, err)
			return err
		}

		// Add to move log
		o.summary.Moves = append(o.summary.Moves, MoveSummary{
			From: filePath,
			To:   targetPath,
		})

		// Update the undo log
		o.updateLogAndCleanup(filepath.Dir(filePath), filepath.Dir(targetPath), []string{filepath.Base(filePath)})
	}

	return nil
}

func (o *Organizer) logMetadataIfVerbose(metadata Metadata, provider MetadataProvider) {
	if !o.config.Verbose {
		return
	}

	// Get provider type and icon for logging purposes
	providerIcon, providerType := getProviderTypeDisplay(provider)

	fmt.Printf("\n%s Found %s\n", providerIcon, providerType)

	// Display the formatted metadata
	fmt.Print(metadata.FormatMetadataWithMapping())
	fmt.Println()
}

func getProviderTypeDisplay(provider MetadataProvider) (string, string) {
	switch provider.(type) {
	case *JSONMetadataProvider:
		return "📋", "JSON metadata file"
	case *EPUBMetadataProvider:
		return "📚", "EPUB embedded metadata"
	case *AudioMetadataProvider:
		return "🎵", "Audio embedded metadata"
	case *FileMetadataProvider:
		return "🔍", "Auto-detected metadata"
	default:
		return "📄", "Metadata provider"
	}
}

func (o *Organizer) calculateTargetPath(metadata Metadata) (string, error) {
	authorDir := o.SanitizePath(strings.Join(metadata.Authors, ","))
	titleDir := o.SanitizePath(metadata.Title)

	targetBase := o.config.BaseDir
	if o.config.OutputDir != "" {
		targetBase = o.config.OutputDir
	}

	// Standard directory structure based on layout flag
	var targetPath string
	switch o.config.Layout {
	case "author-only":
		// Just put everything under author directory
		targetPath = filepath.Join(targetBase, authorDir)

	case "author-title":
		// Skip series level, use author/title structure
		targetPath = filepath.Join(targetBase, authorDir, titleDir)

	case "author-series-title", "": // Default to author-series-title if not specified
		// Use full author/series/title structure if series exists
		if len(metadata.Series) > 0 && metadata.Series[0] != "__INVALID_SERIES__" {
			cleanedSeries := cleanSeriesName(metadata.Series[0])
			seriesDir := o.SanitizePath(cleanedSeries)
			targetPath = filepath.Join(targetBase, authorDir, seriesDir, titleDir)
		} else {
			targetPath = filepath.Join(targetBase, authorDir, titleDir)
		}

	default:
		// Default to author/title if unknown layout
		targetPath = filepath.Join(targetBase, authorDir, titleDir)
	}

	return targetPath, nil
}

func (o *Organizer) promptForMoveConfirmation(metadata Metadata, sourcePath, targetPath string) bool {
	return o.PromptForConfirmation(metadata, sourcePath, targetPath)
}

// Create a file metadata provider for auto-detection
func (o *Organizer) getMetadataProvider(filePath string) (MetadataProvider, error) {
	// Use the same logic as in processFlatDirectory to get the appropriate provider
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".epub":
		return NewEPUBMetadataProvider(filePath), nil
	case ".mp3", ".m4b", ".m4a":
		return &AudioMetadataProvider{filePath: filePath}, nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
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
		sourceName := filepath.Join(sourcePath, entry.Name())
		ext := filepath.Ext(entry.Name())
		baseName := strings.TrimSuffix(entry.Name(), ext)

		// Get metadata for this file to check for track number
		var targetName string
		if provider, err := o.getMetadataProvider(sourceName); err == nil {
			if metadata, err := provider.GetMetadata(); err == nil && metadata.TrackNumber > 0 {
				targetName = fmt.Sprintf("%02d - %s%s", metadata.TrackNumber, baseName, ext)
			} else {
				targetName = entry.Name()
			}
		} else {
			targetName = entry.Name()
		}

		targetFullPath := filepath.Join(targetPath, targetName)
		fileNames = append(fileNames, targetName)

		if o.config.Verbose || o.config.DryRun {
			prefix := "[DRY-RUN] "
			if !o.config.DryRun {
				prefix = ""
			}
			color.Blue("📦 %sMoving %s to %s", prefix, sourceName, targetFullPath)
		}

		if !o.config.DryRun {
			if err := o.moveFile(sourceName, targetFullPath); err != nil {
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
