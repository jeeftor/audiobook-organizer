// internal/organizer/organize.go
package organizer

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// processDirectory is the main entry point for processing each directory during filepath.Walk.
// It handles both flat and hierarchical processing modes based on configuration.
func (o *Organizer) processDirectory(path string, info os.FileInfo, err error) error {
	if err != nil {
		return o.handleDirectoryError(err, path)
	}

	if o.config.Flat {
		return o.handleFlatMode(path, info, nil)
	}

	return o.handleHierarchicalMode(path, info)
}

// handleDirectoryError processes errors encountered during directory traversal.
// It gracefully handles non-existent paths (which may have been moved) and reports other errors.
func (o *Organizer) handleDirectoryError(err error, path string) error {
	if os.IsNotExist(err) {
		if o.config.Verbose {
			PrintYellow("⏩ Skipping non-existent path (likely moved): %s", path)
		}
		return nil
	}
	return err
}

// handleFlatMode processes files in flat mode (each file independently)
func (o *Organizer) handleFlatMode(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// Skip output directory to avoid processing files we just organized
	if o.config.OutputDir != "" && (path == o.config.OutputDir || isSubPathOf(o.config.OutputDir, path)) {
		return nil
	}

	// Skip directories in flat mode, but don't skip traversal
	if info.IsDir() {
		// We still want to traverse subdirectories to find files
		// but we don't need to process the directory itself
		return nil
	}

	return o.processFlatDirectory(path, info)
}

// handleHierarchicalMode processes directories in standard hierarchical mode,
// organizing books into nested directory structures based on metadata.
func (o *Organizer) handleHierarchicalMode(path string, info os.FileInfo) error {
	if !info.IsDir() {
		return nil
	}

	if o.shouldSkipOutputDirectory(path) {
		return filepath.SkipDir
	}

	organized, err := o.tryOrganizeWithMetadata(path)
	if err != nil {
		PrintRed("❌ Error processing %s: %v", path, err)
		return nil
	}

	if organized {
		return filepath.SkipDir
	}

	o.handleMissingMetadata(path)
	return nil
}

// shouldSkipOutputDirectory checks if the current path is the output directory
// and should be skipped during processing to avoid recursive organization.
func (o *Organizer) shouldSkipOutputDirectory(path string) bool {
	return o.config.OutputDir != "" && path == o.config.OutputDir
}

// handleMissingMetadata logs directories that don't contain any usable metadata.
func (o *Organizer) handleMissingMetadata(path string) {
	if o.config.Verbose {
		o.summary.MetadataMissing = append(o.summary.MetadataMissing, path)
		PrintYellow("⚠️  No metadata found in %s", path)
	}
}

// processFlatDirectory processes a directory in flat mode, scanning for audio files
// and organizing them individually or as multi-file albums. Also handles special test environments.
func (o *Organizer) processFlatDirectory(path string, info os.FileInfo) error {
	// Skip non-directory files when trying to process them as directories
	if !info.IsDir() {
		// Check if this is a supported file type before processing
		ext := strings.ToLower(filepath.Ext(path))
		if IsSupportedFile(ext) {
			// Process individual file
			return o.OrganizeSingleFile(path, nil)
		} else {
			// Skip unsupported files silently
			if o.config.Verbose {
				PrintYellow("⏩ Skipping unsupported file: %s", path)
			}
			return nil
		}
	}

	if o.config.Verbose {
		PrintBlue("🔍 Processing directory in flat mode: %s", path)
	}

	// Handle test environment first
	if o.handleTestBookDirectory(path) {
		return nil
	}

	// Check if this directory contains multiple audio files that should be treated as an album
	if o.shouldProcessAsAlbum(path) {
		return o.ProcessMultiFileAlbum(path)
	}

	// Process audio files in the directory individually
	return o.processSupportedFilesInDirectory(path)
}

// handleTestBookDirectory checks for and processes special test_book directories
// used in testing environments. Returns true if a test book was found and processed.
func (o *Organizer) handleTestBookDirectory(path string) bool {
	testBookDir := filepath.Join(path, TestBookDirName)
	if !o.fileOps.DirectoryExists(testBookDir) {
		return false
	}

	metadataFile := filepath.Join(testBookDir, MetadataFileName)
	audioFile := filepath.Join(testBookDir, TestAudioFileName)

	if !o.fileOps.AllFilesExist(metadataFile, audioFile) {
		return false
	}

	if err := o.processTestBook(metadataFile, audioFile, testBookDir); err != nil {
		PrintRed("❌ Error processing test book: %v", err)
	}

	return true
}

// processTestBook processes a single test book with known metadata.json and audio.mp3 files.
// This is used for testing and development environments.
func (o *Organizer) processTestBook(metadataFile, audioFile, testBookDir string) error {
	metadata, err := o.readMetadataFromJSON(metadataFile)
	if err != nil {
		return fmt.Errorf("error reading metadata from JSON: %w", err)
	}

	targetDir := o.calculateTestBookTargetDir(metadata)
	targetAudioPath := filepath.Join(targetDir, metadata.Title+".mp3")

	if o.config.Verbose {
		message := o.formatTestBookMove(audioFile, targetAudioPath)
		fmt.Println(message)
	}

	if o.config.DryRun {
		// Add to plan script if configured
		if o.planWriter != nil {
			o.planWriter.AddMove(audioFile, targetAudioPath, &metadata)
		}
		// Add to plan file if configured
		if o.planFileWriter != nil {
			o.planFileWriter.AddMove(audioFile, targetAudioPath, &metadata)
		}
		return nil
	}

	if err := o.fileOps.CreateDirIfNotExists(targetDir); err != nil {
		return fmt.Errorf("error creating target directory: %w", err)
	}

	if err := o.moveFile(audioFile, targetAudioPath); err != nil {
		return fmt.Errorf("error moving file: %w", err)
	}

	o.summary.Moves = append(o.summary.Moves, MoveSummary{
		From: testBookDir,
		To:   targetDir,
	})

	return nil
}

// calculateTestBookTargetDir calculates the target directory for test books
// based on author and series information from metadata.
func (o *Organizer) calculateTestBookTargetDir(metadata Metadata) string {
	author := metadata.GetFirstAuthor("Unknown")

	if validSeries := metadata.GetValidSeries(); validSeries != "" {
		return filepath.Join(o.config.OutputDir, author, validSeries)
	}

	return filepath.Join(o.config.OutputDir, author)
}

// isSuspiciousTitle detects if a title looks like track numbers (e.g., "098/113", "98/99")
func isSuspiciousTitle(title string) bool {
	// Pattern: digits/digits or digits-digits
	if len(title) < 3 {
		return false
	}

	// Check for patterns like "98/99", "098/113", "1/10", etc.
	// Also check for "Part NN of NN"
	suspiciousPatterns := []string{
		`^\d+/\d+$`,         // 98/99, 098/113
		`^\d+-\d+$`,         // 98-99
		`^Part \d+ of \d+$`, // Part 98 of 99
		`^\d+\s*of\s*\d+$`,  // 98 of 99
	}

	for _, pattern := range suspiciousPatterns {
		matched, _ := regexp.MatchString(pattern, strings.TrimSpace(title))
		if matched {
			return true
		}
	}

	return false
}

// checkForSuspiciousTitles analyzes parsed files and warns about suspicious titles
func (o *Organizer) checkForSuspiciousTitles(parsedFiles []FileWithMetadata) (int, []string) {
	suspiciousCount := 0
	examples := []string{}

	for _, parsed := range parsedFiles {
		if parsed.Error != nil {
			continue
		}

		if isSuspiciousTitle(parsed.Metadata.Title) {
			suspiciousCount++
			if len(examples) < 5 {
				examples = append(examples, fmt.Sprintf("  • %s → Title: \"%s\"",
					filepath.Base(parsed.FilePath), parsed.Metadata.Title))
			}
		}
	}

	return suspiciousCount, examples
}

// parseMetadataParallel parses metadata for all files in parallel using a worker pool
func (o *Organizer) parseMetadataParallel(filePaths []string, workerCount int) []FileWithMetadata {
	if len(filePaths) == 0 {
		return []FileWithMetadata{}
	}

	// Create channels for work distribution
	jobs := make(chan string, len(filePaths))
	results := make(chan FileWithMetadata, len(filePaths))

	// Start worker pool
	for w := 0; w < workerCount; w++ {
		go func() {
			for filePath := range jobs {
				result := FileWithMetadata{
					FilePath: filePath,
				}

				// Get metadata provider for this file
				provider, err := o.getMetadataProvider(filePath)
				if err != nil {
					result.Error = err
					results <- result
					continue
				}

				// Parse metadata
				metadata, err := o.prepareMetadata(provider)
				if err != nil {
					result.Error = err
					results <- result
					continue
				}

				result.Metadata = metadata
				result.Provider = provider
				results <- result
			}
		}()
	}

	// Send jobs to workers
	for _, filePath := range filePaths {
		jobs <- filePath
	}
	close(jobs)

	// Collect results
	parsedFiles := make([]FileWithMetadata, 0, len(filePaths))
	for i := 0; i < len(filePaths); i++ {
		parsedFiles = append(parsedFiles, <-results)
	}
	close(results)

	return parsedFiles
}

// processAudioFilesInDirectory should be renamed to processSupportedFilesInDirectory
// and updated to handle all supported file types in flat mode
func (o *Organizer) processSupportedFilesInDirectory(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	// Collect all supported file paths
	var filePaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			if o.config.Verbose {
				PrintYellow("⏩ Skipping subdirectory in flat mode: %s", filepath.Join(path, entry.Name()))
			}
			continue
		}

		filePath := filepath.Join(path, entry.Name())
		ext := strings.ToLower(filepath.Ext(filePath))

		// Clean, centralized check for supported file types
		if IsSupportedFile(ext) {
			filePaths = append(filePaths, filePath)
		} else if o.config.Verbose {
			PrintYellow("⏩ Skipping unsupported file type: %s", filePath)
		}
	}

	// Parse metadata in parallel
	if o.config.Verbose {
		PrintCyan("📊 Parsing metadata for %d files in parallel...", len(filePaths))
	}

	// Use 20 workers for parallel parsing (good balance for I/O bound operations)
	parsedFiles := o.parseMetadataParallel(filePaths, 20)

	// Check for suspicious titles that look like track numbers
	suspiciousCount, examples := o.checkForSuspiciousTitles(parsedFiles)
	if suspiciousCount > 0 {
		PrintYellow("\n⚠️  WARNING: Found %d files with suspicious titles that look like track numbers!", suspiciousCount)
		PrintYellow("This may create nested folders like: Author/Series/098/113/filename.mp3\n")
		PrintYellow("Examples:")
		for _, example := range examples {
			PrintYellow(example)
		}
		if suspiciousCount > len(examples) {
			PrintYellow("  ... and %d more\n", suspiciousCount-len(examples))
		}

		PrintCyan("\n💡 Suggestions:")
		PrintCyan("  1. Use --layout author-title to skip the series/title nesting")
		PrintCyan("  2. Use --title-field album to use the Album field as title")
		PrintCyan("  3. Continue anyway if this is intentional\n")

		if !o.config.DryRun {
			response := o.PromptYesNo("Do you want to continue?")
			if !response {
				PrintYellow("⏩ Aborting organization")
				return nil
			}
		}
	}

	// Process files sequentially with pre-parsed metadata
	for _, parsed := range parsedFiles {
		if parsed.Error != nil {
			PrintRed("❌ Error parsing metadata for %s: %v", parsed.FilePath, parsed.Error)
			continue
		}

		// Pass the pre-parsed metadata to OrganizeSingleFile
		if err := o.OrganizeSingleFile(parsed.FilePath, parsed.Provider); err != nil {
			PrintRed("❌ Error organizing file %s: %v", parsed.FilePath, err)
		}
	}

	return nil
}

// tryOrganizeWithMetadata attempts to organize a directory using available metadata sources.
// It tries embedded metadata first (if enabled), then falls back to JSON metadata files.
func (o *Organizer) tryOrganizeWithMetadata(path string) (bool, error) {
	if o.config.UseEmbeddedMetadata {
		if organized, err := o.tryEmbeddedMetadata(path); organized || err != nil {
			return organized, err
		}
	}

	return o.tryJSONMetadata(path)
}

// tryEmbeddedMetadata attempts to extract and use metadata embedded within files.
// It tries EPUB files first, then audio files as fallback options.
func (o *Organizer) tryEmbeddedMetadata(path string) (bool, error) {
	// Try EPUB first
	if organized, err := o.tryEPUBMetadata(path); organized || err != nil {
		return organized, err
	}

	// Try audio files as fallback
	return o.tryAudioMetadata(path)
}

// tryEPUBMetadata attempts to extract metadata from EPUB files in the directory
// and organize the audiobook based on that metadata.
func (o *Organizer) tryEPUBMetadata(path string) (bool, error) {
	epubPath, err := FindEPUBInDirectory(path)
	if err != nil {
		if o.config.Verbose {
			PrintYellow("⚠️ No EPUB files found in %s", path)
		}
		return false, nil
	}

	epubProvider := NewEPUBMetadataProvider(epubPath)
	metadata, err := epubProvider.GetMetadata()

	if err != nil || !metadata.IsValid() {
		if o.config.Verbose {
			PrintYellow("⚠️ EPUB found but metadata extraction failed: %s", epubPath)
		}
		return false, nil
	}

	PrintGreen("📚 Found metadata in EPUB file: %s", epubPath)
	if err := o.OrganizeAudiobook(path, epubProvider); err != nil {
		return false, fmt.Errorf("error organizing with EPUB metadata: %v", err)
	}

	return true, nil
}

// tryAudioMetadata attempts to extract metadata from audio files in the directory
// and organize the audiobook based on that metadata.
func (o *Organizer) tryAudioMetadata(path string) (bool, error) {
	audioPath, err := FindAudioFileInDirectory(path)
	if err != nil {
		if o.config.Verbose {
			PrintYellow("⚠️ No supported audio files found in %s", path)
		}
		return false, nil
	}

	audioProvider := NewAudioMetadataProvider(audioPath)
	metadata, err := audioProvider.GetMetadata()

	if err != nil || !metadata.IsValid() {
		if o.config.Verbose {
			PrintYellow("⚠️ Audio file found but metadata extraction failed: %s", audioPath)
		}
		return false, nil
	}

	PrintGreen("🔊 Found metadata in audio file: %s", audioPath)
	if err := o.OrganizeAudiobook(path, audioProvider); err != nil {
		return false, fmt.Errorf("error organizing with audio metadata: %v", err)
	}

	return true, nil
}

// tryJSONMetadata attempts to find and use a metadata.json file in the directory
// for organizing the audiobook.
func (o *Organizer) tryJSONMetadata(path string) (bool, error) {
	metadataPath := filepath.Join(path, MetadataFileName)
	if !o.fileOps.FileExists(metadataPath) {
		return false, nil
	}

	o.summary.MetadataFound = append(o.summary.MetadataFound, metadataPath)
	if err := o.OrganizeAudiobook(path, NewJSONMetadataProvider(metadataPath)); err != nil {
		return false, fmt.Errorf("error organizing with JSON metadata: %v", err)
	}

	return true, nil
}

// OrganizeAudiobook is the main function for organizing a complete audiobook directory.
// It extracts metadata, validates it, calculates target paths, and moves files accordingly.
func (o *Organizer) OrganizeAudiobook(sourcePath string, provider MetadataProvider) error {
	metadata, err := o.prepareMetadata(provider)
	if err != nil {
		return err
	}

	o.logMetadataIfVerbose(metadata, provider)

	if err := metadata.Validate(); err != nil {
		return err
	}

	targetPath := o.layoutCalculator.CalculateTargetPath(metadata)

	if o.isAlreadyInCorrectLocation(sourcePath, targetPath) {
		return nil
	}

	if o.shouldSkipMove(metadata, sourcePath, targetPath) {
		return nil
	}

	return o.executeMove(sourcePath, targetPath, &metadata)
}

// prepareMetadata extracts metadata from a provider and applies field mapping
// configuration to ensure proper title, author, and series assignment.
func (o *Organizer) prepareMetadata(provider MetadataProvider) (Metadata, error) {
	metadata, err := provider.GetMetadata()
	if err != nil {
		return Metadata{}, fmt.Errorf("error getting metadata: %w", err)
	}

	metadata.ApplyFieldMapping(o.config.FieldMapping)

	return metadata, nil
}

// isAlreadyInCorrectLocation checks if the source path is already the same as
// the calculated target path, avoiding unnecessary moves.
func (o *Organizer) isAlreadyInCorrectLocation(sourcePath, targetPath string) bool {
	cleanSourcePath := filepath.Clean(sourcePath)
	cleanTargetPath := filepath.Clean(targetPath)

	if cleanSourcePath == cleanTargetPath {
		if o.config.Verbose {
			PrintGreen("✅ Book already in correct location: %s", cleanSourcePath)
		}
		return true
	}
	return false
}

// shouldSkipMove determines if a move operation should be skipped based on
// user prompts or other configuration settings.
func (o *Organizer) shouldSkipMove(metadata Metadata, sourcePath, targetPath string) bool {
	if o.config.Prompt && !o.promptForMoveConfirmation(metadata, sourcePath, targetPath) {
		PrintYellow("⏩ Skipping %s", metadata.Title)
		return true
	}
	return false
}

// executeMove performs the actual file moving operation for an audiobook directory,
// including logging and cleanup of empty directories.
func (o *Organizer) executeMove(sourcePath, targetPath string, metadata *Metadata) error {
	fileNames, err := o.moveFiles(sourcePath, targetPath, metadata)
	if err != nil {
		return err
	}

	if !o.config.DryRun {
		o.updateLogAndCleanup(sourcePath, targetPath, fileNames)
	}

	return nil
}

// OrganizeSingleFile organizes an individual file based on its embedded metadata.
// This is primarily used in flat mode where files are processed individually.
func (o *Organizer) OrganizeSingleFile(filePath string, provider MetadataProvider) error {
	if provider == nil {
		var err error
		provider, err = o.getMetadataProvider(filePath)
		if err != nil {
			return fmt.Errorf("error getting metadata provider: %w", err)
		}
	}

	metadata, err := o.prepareMetadata(provider)
	if err != nil {
		return err
	}

	o.logMetadataIfVerbose(metadata, provider)

	if err := metadata.Validate(); err != nil {
		return err
	}

	targetPath := o.calculateSingleFileTargetPath(filePath, metadata)

	if o.isAlreadyInCorrectLocation(filePath, targetPath) {
		return nil
	}

	if o.shouldSkipMove(metadata, filePath, targetPath) {
		return nil
	}

	return o.executeSingleFileMove(filePath, targetPath, metadata)
}

// calculateSingleFileTargetPath determines the complete target path for a single file
// including both directory and filename components.
func (o *Organizer) calculateSingleFileTargetPath(filePath string, metadata Metadata) string {
	targetDir := o.calculateSingleFileTargetDir(filePath, metadata)

	fileName := filepath.Base(filePath)

	// Only add track prefix if enabled
	if o.config.AddTrackNumbers && metadata.TrackNumber > 0 {
		// Get total tracks for proper padding
		totalTracks := o.getTotalTracksForFile(filePath, metadata)
		fileName = AddTrackPrefix(fileName, metadata.TrackNumber, totalTracks)
	}

	return filepath.Join(targetDir, fileName)
}

// getTotalTracksForFile determines the total number of tracks for proper padding
// It checks metadata first, then counts files in the parent directory
func (o *Organizer) getTotalTracksForFile(filePath string, metadata Metadata) int {
	// First try to get from metadata's RawData
	if totalTracks, ok := metadata.RawData["track_total"].(float64); ok && totalTracks > 0 {
		return int(totalTracks)
	}
	if totalTracks, ok := metadata.RawData["track_total"].(int); ok && totalTracks > 0 {
		return totalTracks
	}

	// If not in metadata, count audio files in the parent directory
	parentDir := filepath.Dir(filePath)
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		// Default to 100 if we can't read the directory (results in 3-digit padding)
		return 100
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if IsSupportedAudioFile(ext) {
				count++
			}
		}
	}

	// If we found files, use that count; otherwise default to 100
	if count > 0 {
		return count
	}
	return 100
}

// calculateSingleFileTargetDir determines the target directory for a single file
// based on the configured layout (author-only, author-title, author-series-title).
func (o *Organizer) calculateSingleFileTargetDir(filePath string, metadata Metadata) string {
	baseDir := o.getBaseDirForSingleFile(filePath)

	// Use PathBuilder for cleaner path construction
	pathBuilder := NewPathBuilder().WithSanitizer(o.SanitizePath)

	switch o.config.Layout {
	case "author-only":
		return pathBuilder.AddAuthor(strings.Join(metadata.Authors, ",")).Build(baseDir)
	case "author-title":
		return pathBuilder.
			AddAuthor(strings.Join(metadata.Authors, ",")).
			AddTitle(metadata.Title).
			Build(baseDir)
	case "author-series-title", "":
		pathBuilder.AddAuthor(strings.Join(metadata.Authors, ","))
		if validSeries := metadata.GetValidSeries(); validSeries != "" {
			pathBuilder.AddSeries(validSeries)
			// Only add title if it's different from the series
			if validSeries != metadata.Title {
				pathBuilder.AddTitle(metadata.Title)
			}
		} else {
			// No series, just add the title
			pathBuilder.AddTitle(metadata.Title)
		}
		return pathBuilder.Build(baseDir)
	default:
		return pathBuilder.
			AddAuthor(strings.Join(metadata.Authors, ",")).
			AddTitle(metadata.Title).
			Build(baseDir)
	}
}

// getBaseDirForSingleFile determines the base directory to use for single file operations,
// preferring the configured output directory over the source file's directory.
func (o *Organizer) getBaseDirForSingleFile(filePath string) string {
	if o.config.OutputDir != "" {
		return o.config.OutputDir
	}
	return filepath.Dir(filePath)
}

// executeSingleFileMove performs the actual moving of a single file, including
// directory creation, dry-run handling, and logging.
func (o *Organizer) executeSingleFileMove(filePath, targetPath string, metadata Metadata) error {
	targetDir := filepath.Dir(targetPath)

	if err := o.fileOps.CreateDirIfNotExists(targetDir); err != nil {
		return fmt.Errorf("error creating target directory: %w", err)
	}

	if o.config.DryRun {
		message := o.formatDryRunMove(filePath, targetPath)
		fmt.Println(message)
		// Add to summary even in dry-run mode
		o.addSingleFileMoveToSummary(filePath, targetPath)
		// Add to plan script if configured
		if o.planWriter != nil {
			o.planWriter.AddMove(filePath, targetPath, &metadata)
		}
		// Add to plan file if configured
		if o.planFileWriter != nil {
			o.planFileWriter.AddMove(filePath, targetPath, &metadata)
		}
		return nil
	}

	if o.config.Verbose {
		message := o.formatVerboseMove(filePath, targetPath)
		fmt.Println(message)
	}

	if err := o.moveFile(filePath, targetPath); err != nil {
		PrintRed("❌ Error moving %s: %v", filePath, err)
		return err
	}

	o.addSingleFileMoveToSummary(filePath, targetPath)
	o.updateLogAndCleanup(filepath.Dir(filePath), filepath.Dir(targetPath), []string{filepath.Base(filePath)})

	return nil
}

// addSingleFileMoveToSummary adds a single file move operation to the summary.
func (o *Organizer) addSingleFileMoveToSummary(filePath, targetPath string) {
	o.summary.Moves = append(o.summary.Moves, MoveSummary{
		From: filePath,
		To:   targetPath,
	})
}

// String formatting functions - return formatted strings instead of directly printing

// formatTestBookMove returns a formatted string for test book move operations.
func (o *Organizer) formatTestBookMove(audioFile, targetAudioPath string) string {
	return fmt.Sprintf("📦 Moving %s to %s", audioFile, targetAudioPath)
}

// formatDryRunMove returns a formatted string for dry-run move operations.
func (o *Organizer) formatDryRunMove(filePath, targetPath string) string {
	coloredPath := o.formatColoredPath(filePath, targetPath)
	return fmt.Sprintf("📦 [DRY-RUN] Would move %s", coloredPath)
}

// formatVerboseMove returns a formatted string for verbose move operations.
func (o *Organizer) formatVerboseMove(filePath, targetPath string) string {
	coloredPath := o.formatColoredPath(filePath, targetPath)
	return fmt.Sprintf("📦 Moving %s", coloredPath)
}

// formatColoredPath returns a complete formatted string showing source → target path transformation.
func (o *Organizer) formatColoredPath(filePath, targetPath string) string {
	sourcePath := o.formatColoredSourcePath(filePath)
	targetPath = o.formatTargetPathComponents(targetPath)
	return sourcePath + " to " + targetPath
}

// formatColoredSourcePath returns a formatted string representation of the source file path
// with track number and filename coloring.
func (o *Organizer) formatColoredSourcePath(filePath string) string {
	var result strings.Builder

	inputFilename := filepath.Base(filePath)
	inputDir := filepath.Dir(filePath)
	result.WriteString(inputDir + "/")

	trackPrefixRegex := regexp.MustCompile(`^(\d+)\s*-\s*`)
	if matches := trackPrefixRegex.FindStringSubmatch(inputFilename); len(matches) > 1 {
		trackNum := matches[1]
		restOfFilename := inputFilename[len(matches[0]):]
		result.WriteString(TrackNumberColor(trackNum + " - "))
		result.WriteString(FilenameColor(restOfFilename))
	} else {
		result.WriteString(FilenameColor(inputFilename))
	}

	return result.String()
}

// formatTargetPathComponents returns a formatted string representation of the target path
// with color coding for author, series, title, and filename components.
func (o *Organizer) formatTargetPathComponents(targetPath string) string {
	relPath := o.getRelativeTargetPath(targetPath)
	pathParts := strings.Split(filepath.ToSlash(relPath), "/")

	return o.formatPathComponentsWithColors(pathParts)
}

// formatPathComponentsWithColors returns a formatted string with path components colored appropriately.
func (o *Organizer) formatPathComponentsWithColors(pathParts []string) string {
	var result strings.Builder

	for i, part := range pathParts {
		if i > 0 {
			result.WriteString("/")
		}

		switch {
		case i == 0:
			// Author - Green
			result.WriteString(AuthorColor(part))
		case i == 1:
			// Series - Cyan
			result.WriteString(SeriesColor(part))
		case i == len(pathParts)-1:
			// Filename - White with track numbers colored
			result.WriteString(o.formatColoredFilename(part))
		default:
			// Title and any subdirectories - Yellow
			result.WriteString(TitleColor(part))
		}
	}

	return result.String()
}

// formatColoredFilename returns a formatted filename string with track number coloring if present.
func (o *Organizer) formatColoredFilename(filename string) string {
	trackPrefixRegex := regexp.MustCompile(`^(\d+)\s*-\s*`)
	if matches := trackPrefixRegex.FindStringSubmatch(filename); len(matches) > 1 {
		trackNum := matches[1]
		restOfFilename := filename[len(matches[0]):]
		return TrackNumberColor(trackNum+" - ") + FilenameColor(restOfFilename)
	}
	return FilenameColor(filename)
}

// formatFileMove returns a formatted string for file move operations in moveFiles.
func (o *Organizer) formatFileMove(sourceName, targetFullPath string, isDryRun bool) string {
	var result strings.Builder

	if isDryRun {
		result.WriteString(TrackNumberColor("📦 [DRY-RUN] Moving "))
	} else {
		result.WriteString(TrackNumberColor("📦 Moving "))
	}

	result.WriteString(FilenameColor(sourceName))
	result.WriteString(" to ")
	result.WriteString(TargetPathColor(targetFullPath))

	return result.String()
}

// formatDirectoryMoveHeader returns a formatted string for directory move headers.
func (o *Organizer) formatDirectoryMoveHeader(sourcePath, targetPath string) string {
	return fmt.Sprintf("🔄 Moving contents from %s to %s", sourcePath, targetPath)
}

// getRelativeTargetPath converts an absolute target path to a relative path for display.
func (o *Organizer) getRelativeTargetPath(targetPath string) string {
	relPath, err := filepath.Rel(filepath.Dir(o.config.BaseDir), targetPath)
	if err != nil {
		return targetPath
	}
	return relPath
}

// logMetadataIfVerbose displays formatted metadata information when verbose mode is enabled.
func (o *Organizer) logMetadataIfVerbose(metadata Metadata, provider MetadataProvider) {
	if !o.config.Verbose {
		return
	}

	providerIcon, providerType := getProviderTypeDisplay(provider)
	fmt.Printf("\n%s Found %s\n", providerIcon, providerType)
	formatter := NewMetadataFormatter(metadata, o.config.FieldMapping)
	fmt.Print(formatter.FormatMetadataWithMapping())
	fmt.Println()
}

// getProviderTypeDisplay returns appropriate icon and description for different metadata providers.
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

// promptForMoveConfirmation asks the user for confirmation before moving files.
func (o *Organizer) promptForMoveConfirmation(metadata Metadata, sourcePath, targetPath string) bool {
	return o.PromptForConfirmation(metadata, sourcePath, targetPath)
}

// getMetadataProvider creates an appropriate metadata provider based on file extension.
func (o *Organizer) getMetadataProvider(filePath string) (MetadataProvider, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".epub":
		// Track metadata file in summary
		o.summary.MetadataFound = append(o.summary.MetadataFound, filePath)
		return NewEPUBMetadataProvider(filePath), nil
	case ".mp3", ".m4b", ".m4a":
		// Track metadata file in summary
		o.summary.MetadataFound = append(o.summary.MetadataFound, filePath)
		return NewAudioMetadataProvider(filePath), nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// updateLogAndCleanup records the move operation in logs and cleans up empty directories.
func (o *Organizer) updateLogAndCleanup(sourcePath, targetPath string, fileNames []string) {
	o.logEntries = append(o.logEntries, LogEntry{
		Timestamp:  time.Now(),
		SourcePath: sourcePath,
		TargetPath: targetPath,
		Files:      fileNames,
	})

	if err := o.saveLog(); err != nil {
		PrintYellow("⚠️  Warning: couldn't save log: %v", err)
	}

}

// readMetadataFromJSON reads and processes metadata from a JSON file,
// applying field mapping configuration.
func (o *Organizer) readMetadataFromJSON(filePath string) (Metadata, error) {
	provider := NewJSONMetadataProvider(filePath)
	metadata, err := provider.GetMetadata()
	if err != nil {
		return Metadata{}, err
	}

	metadata.ApplyFieldMapping(o.config.FieldMapping)

	return metadata, nil
}

// cleanEmptyParents recursively removes empty parent directories up to a specified boundary.
// It ensures that empty directories created during file moves are cleaned up properly.
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
				PrintYellow("⏩ Skipping removal of parent directory %s", dir)
			}
			return nil
		}
	}

	if o.config.Verbose {
		PrintYellow("🗑️  Removing newly empty parent directory: %s", dir)
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

		// After removing the directory, check if parent is now empty,
		// but don't go beyond the input directory
		if parentDir != o.config.BaseDir {
			if err := o.cleanEmptyParents(parentDir, o.config.BaseDir); err != nil {
				PrintRed("❌ Error cleaning parent directories: %v", err)
			}
		}
	}

	// Skip further processing of this directory since it's been removed
	return filepath.SkipDir
}

// isSubPathOf checks if a child path is a subdirectory of a parent path.
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

// isEmptyDir is defined in organizer.go - removing duplicate

// debugLog logs debug messages only if verbose mode is enabled
func (o *Organizer) debugLog(format string, args ...interface{}) {
	if o.config.Verbose {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// moveFile moves a file from source to target, handling cross-device moves
// by falling back to copy-and-delete when necessary.
func (o *Organizer) moveFile(source, target string) error {
	// Check if source and target are the same
	if filepath.Clean(source) == filepath.Clean(target) {
		return nil
	}

	o.debugLog("moveFile: source=%s, target=%s", source, target)

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(target)
	if err := o.fileOps.CreateDirIfNotExists(targetDir); err != nil {
		return fmt.Errorf("error creating target directory: %w", err)
	}

	// Try to use os.Rename first (most efficient)
	err := os.Rename(source, target)
	if err != nil {
		// If rename fails (e.g., cross-device link), fall back to copy and delete
		o.debugLog("Rename failed, falling back to copy and delete: %v", err)
		return o.copyAndDeleteFile(source, target, targetDir)
	}

	o.debugLog("Successfully renamed file from %s to %s", source, target)
	return nil
}

// copyAndDeleteFile performs a copy-and-delete operation when os.Rename fails.
func (o *Organizer) copyAndDeleteFile(source, target, targetDir string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer sourceFile.Close()

	// Create target file
	targetFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("error creating target file: %w", err)
	}
	defer targetFile.Close()

	// Copy the contents
	data, err := io.ReadAll(sourceFile)
	if err != nil {
		return fmt.Errorf("error reading source file: %w", err)
	}
	o.debugLog("Read %d bytes from source file %s", len(data), source)

	n, err := targetFile.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to target file: %w", err)
	}
	o.debugLog("Successfully wrote %d bytes to target file %s", n, target)

	// Remove source file
	if err := os.Remove(source); err != nil {
		return fmt.Errorf("error removing source file: %w", err)
	}
	o.debugLog("Successfully removed source file %s", source)

	// Sync the target directory to ensure all changes are written to disk
	return o.syncTargetDirectory(targetDir)
}

// syncTargetDirectory ensures that directory changes are written to disk.
func (o *Organizer) syncTargetDirectory(targetDir string) error {
	targetDirFile, err := os.Open(targetDir)
	if err != nil {
		return fmt.Errorf("error opening target directory: %w", err)
	}
	defer targetDirFile.Close()

	if err := targetDirFile.Sync(); err != nil {
		return fmt.Errorf("error syncing target directory: %w", err)
	}
	o.debugLog("Successfully synced target directory %s", targetDir)
	return nil
}

// moveFiles moves all files from a source directory to a target directory,
// handling track number prefixes and maintaining a list of moved files.
func (o *Organizer) moveFiles(sourcePath, targetPath string, dirMetadata *Metadata) ([]string, error) {
	if o.config.Verbose {
		message := o.formatDirectoryMoveHeader(sourcePath, targetPath)
		PrintCyan(message)
	}

	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("error reading source directory: %w", err)
	}

	// Create target directory if it doesn't exist
	if err := o.fileOps.CreateDirIfNotExists(targetPath); err != nil {
		return nil, fmt.Errorf("error creating target directory: %w", err)
	}

	o.summary.Moves = append(o.summary.Moves, MoveSummary{
		From: sourcePath,
		To:   targetPath,
	})

	// Get metadata if not provided
	if dirMetadata == nil {
		dirMetadata = o.getDirectoryMetadata(sourcePath)
	}

	return o.processDirectoryFiles(entries, sourcePath, targetPath, dirMetadata)
}

// getDirectoryMetadata attempts to load metadata from a metadata.json file in the directory.
func (o *Organizer) getDirectoryMetadata(sourcePath string) *Metadata {
	metadataPath := filepath.Join(sourcePath, MetadataFileName)
	if o.fileOps.FileExists(metadataPath) {
		provider := NewJSONMetadataProvider(metadataPath)
		if md, err := provider.GetMetadata(); err == nil {
			md.ApplyFieldMapping(o.config.FieldMapping) // Changed from 'metadata' to 'md'
			return &md
		}
	}
	return nil
}

// processDirectoryFiles processes individual files in a directory for moving.
func (o *Organizer) processDirectoryFiles(entries []os.DirEntry, sourcePath, targetPath string, dirMetadata *Metadata) ([]string, error) {
	var fileNames []string

	// Count total files (excluding directories) for proper track number padding
	totalFiles := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			totalFiles++
		}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		sourceName := filepath.Join(sourcePath, entry.Name())
		targetName := o.calculateFileTargetName(entry.Name(), dirMetadata, totalFiles)
		targetFullPath := filepath.Join(targetPath, targetName)
		fileNames = append(fileNames, targetName)

		if o.config.Verbose || o.config.DryRun {
			message := o.formatFileMove(sourceName, targetFullPath, o.config.DryRun)
			fmt.Println(message)
		}

		if o.config.DryRun {
			// Add to plan script if configured
			if o.planWriter != nil {
				o.planWriter.AddMove(sourceName, targetFullPath, dirMetadata)
			}
			// Add to plan file if configured
			if o.planFileWriter != nil {
				o.planFileWriter.AddMove(sourceName, targetFullPath, dirMetadata)
			}
		} else {
			if err := o.moveFile(sourceName, targetFullPath); err != nil {
				PrintRed("❌ Error moving %s: %v", sourceName, err)
			}
		}
	}

	return fileNames, nil
}

// calculateFileTargetName determines the target filename, adding track prefixes when appropriate.
// totalFiles is used to calculate proper padding (e.g., 2 digits for <100 files, 3 for 100+)
func (o *Organizer) calculateFileTargetName(fileName string, dirMetadata *Metadata, totalFiles int) string {
	// If rename-files is enabled, apply the pattern completely
	if o.config.RenameFiles && dirMetadata != nil {
		ext := filepath.Ext(fileName)
		newBaseName := ApplyFilenamePattern(o.config.RenamePattern, *dirMetadata, totalFiles)

		// Apply space replacement if configured
		if o.config.ReplaceSpace != "" {
			newBaseName = strings.ReplaceAll(newBaseName, " ", o.config.ReplaceSpace)
		}

		return newBaseName + ext
	}

	// Otherwise, use the FilenameNormalizer for standard processing
	normalizer := NewFilenameNormalizer()

	// Add track prefix if enabled and available in metadata
	if o.config.AddTrackNumbers && dirMetadata != nil && dirMetadata.TrackNumber > 0 {
		normalizer = normalizer.WithTrackPrefix(dirMetadata.TrackNumber, totalFiles)
	}

	// Apply space replacement if configured
	if o.config.ReplaceSpace != "" {
		normalizer = normalizer.WithSpaceReplacement(o.config.ReplaceSpace)
	}

	return normalizer.Normalize(fileName)
}
