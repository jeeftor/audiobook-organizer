package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// ScanForAudiobooks scans a directory for audiobook files and returns metadata.
// This is the main public API for scanning operations, used by the GUI and other external tools.
//
// Parameters:
//   - baseDir: Directory to scan for audiobooks
//   - config: Organizer configuration (determines scanning mode: flat vs hierarchical)
//
// Returns:
//   - []Metadata: List of audiobooks found with their metadata
//   - error: Any error encountered during scanning
//
// Behavior:
//   - In hierarchical mode (default): Scans directories for metadata.json or embedded metadata
//   - In flat mode (config.Flat=true): Scans individual files
//   - Applies field mapping configuration to all metadata
func ScanForAudiobooks(baseDir string, config *OrganizerConfig) ([]Metadata, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("base directory is required")
	}

	// Verify directory exists
	info, err := os.Stat(baseDir)
	if err != nil {
		return nil, fmt.Errorf("error accessing directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", baseDir)
	}

	var results []Metadata

	// Walk the directory tree
	err = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't access
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		// Skip the output directory to avoid scanning organized files
		if config.OutputDir != "" && (path == config.OutputDir || isSubPath(config.OutputDir, path)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if config.Flat {
			// Flat mode: scan individual files
			return scanFlatMode(path, info, config, &results)
		}

		// Hierarchical mode: scan directories
		return scanHierarchicalMode(path, info, config, &results)
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning directory: %w", err)
	}

	return results, nil
}

// ScanSingleFile scans a single audiobook file and returns its metadata.
// This is useful for flat mode processing or when you have a specific file to analyze.
//
// Parameters:
//   - filePath: Path to the audiobook file (EPUB, MP3, M4B, etc.)
//   - config: Organizer configuration (for field mapping)
//
// Returns:
//   - Metadata: Extracted metadata from the file
//   - error: Any error encountered during scanning
func ScanSingleFile(filePath string, config *OrganizerConfig) (Metadata, error) {
	if filePath == "" {
		return organizer.NewMetadata(), fmt.Errorf("file path is required")
	}

	// Verify file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return organizer.NewMetadata(), fmt.Errorf("error accessing file: %w", err)
	}
	if info.IsDir() {
		return organizer.NewMetadata(), fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	// Get appropriate metadata provider based on file type
	provider, err := getMetadataProviderForFile(filePath, config)
	if err != nil {
		return organizer.NewMetadata(), err
	}

	// Extract metadata
	metadata, err := provider.GetMetadata()
	if err != nil {
		return organizer.NewMetadata(), fmt.Errorf("error extracting metadata: %w", err)
	}

	// Apply field mapping
	metadata.ApplyFieldMapping(config.FieldMapping)

	return metadata, nil
}

// scanFlatMode processes a single file in flat mode
func scanFlatMode(path string, info os.FileInfo, config *OrganizerConfig, results *[]Metadata) error {
	// Skip directories in flat mode
	if info.IsDir() {
		return nil
	}

	// Check if this is a supported file
	ext := strings.ToLower(filepath.Ext(path))
	if !organizer.IsSupportedFile(ext) {
		return nil
	}

	// Extract metadata
	metadata, err := ScanSingleFile(path, config)
	if err != nil {
		// Skip files with extraction errors
		return nil
	}

	// Only add if metadata is valid
	if metadata.IsValid() {
		*results = append(*results, metadata)
	}

	return nil
}

// scanHierarchicalMode processes a directory in hierarchical mode
func scanHierarchicalMode(path string, info os.FileInfo, config *OrganizerConfig, results *[]Metadata) error {
	// Only process directories
	if !info.IsDir() {
		return nil
	}

	// Try to get metadata for this directory
	metadata, found, err := tryGetDirectoryMetadata(path, config)
	if err != nil {
		// Log error but continue scanning
		return nil
	}

	if found && metadata.IsValid() {
		*results = append(*results, metadata)
		// Skip subdirectories since we found metadata for this directory
		return filepath.SkipDir
	}

	// Continue scanning subdirectories
	return nil
}

// tryGetDirectoryMetadata attempts to extract metadata from a directory
// using available sources (JSON, EPUB, audio files)
func tryGetDirectoryMetadata(dirPath string, config *OrganizerConfig) (Metadata, bool, error) {
	// Try embedded metadata first if enabled
	if config.UseEmbeddedMetadata {
		// Try EPUB
		if metadata, found := tryEPUBMetadata(dirPath, config); found {
			return metadata, true, nil
		}

		// Try audio files
		if metadata, found := tryAudioMetadata(dirPath, config); found {
			return metadata, true, nil
		}
	}

	// Try JSON metadata
	metadataPath := filepath.Join(dirPath, organizer.MetadataFileName)
	if _, err := os.Stat(metadataPath); err == nil {
		provider := organizer.NewJSONMetadataProvider(metadataPath)
		metadata, err := provider.GetMetadata()
		if err != nil {
			return organizer.NewMetadata(), false, err
		}

		metadata.ApplyFieldMapping(config.FieldMapping)
		return metadata, true, nil
	}

	return organizer.NewMetadata(), false, nil
}

// tryEPUBMetadata attempts to extract metadata from EPUB files in the directory
func tryEPUBMetadata(dirPath string, config *OrganizerConfig) (Metadata, bool) {
	epubPath, err := organizer.FindEPUBInDirectory(dirPath)
	if err != nil {
		return organizer.NewMetadata(), false
	}

	provider := organizer.NewEPUBMetadataProvider(epubPath)
	metadata, err := provider.GetMetadata()
	if err != nil || !metadata.IsValid() {
		return organizer.NewMetadata(), false
	}

	metadata.ApplyFieldMapping(config.FieldMapping)
	return metadata, true
}

// tryAudioMetadata attempts to extract metadata from audio files in the directory
func tryAudioMetadata(dirPath string, config *OrganizerConfig) (Metadata, bool) {
	audioPath, err := organizer.FindAudioFileInDirectory(dirPath)
	if err != nil {
		return organizer.NewMetadata(), false
	}

	provider := organizer.NewAudioMetadataProvider(audioPath)
	metadata, err := provider.GetMetadata()
	if err != nil || !metadata.IsValid() {
		return organizer.NewMetadata(), false
	}

	metadata.ApplyFieldMapping(config.FieldMapping)
	return metadata, true
}

// getMetadataProviderForFile returns the appropriate metadata provider for a file
func getMetadataProviderForFile(filePath string, config *OrganizerConfig) (organizer.MetadataProvider, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".epub":
		return organizer.NewEPUBMetadataProvider(filePath), nil
	case ".mp3", ".m4b", ".m4a":
		return organizer.NewAudioMetadataProvider(filePath), nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// isSubPath checks if a child path is a subdirectory of a parent path
func isSubPath(parent, child string) bool {
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)

	parentParts := strings.Split(parent, string(filepath.Separator))
	childParts := strings.Split(child, string(filepath.Separator))

	if len(childParts) <= len(parentParts) {
		return false
	}

	for i := range parentParts {
		if parentParts[i] != childParts[i] {
			return false
		}
	}

	return true
}
