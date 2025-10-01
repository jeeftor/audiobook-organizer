// internal/organizer/album_handler.go
package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// AlbumGroup represents a group of files that belong to the same album
type AlbumGroup struct {
	Metadata    Metadata
	Files       []string
	TrackOrder  map[string]int // Maps filenames to track numbers
	AlbumFolder string         // The folder name to use for this album
}

// NewAlbumGroup creates a new album group with the given metadata
func NewAlbumGroup(metadata Metadata) *AlbumGroup {
	return &AlbumGroup{
		Metadata:   metadata,
		Files:      []string{},
		TrackOrder: make(map[string]int),
	}
}

// AddFile adds a file to the album group and records its track number
func (ag *AlbumGroup) AddFile(filePath string, trackNumber int) {
	ag.Files = append(ag.Files, filePath)
	ag.TrackOrder[filePath] = trackNumber
}

// SortFilesByTrackNumber sorts the files in the album group by track number
func (ag *AlbumGroup) SortFilesByTrackNumber() {
	sort.Slice(ag.Files, func(i, j int) bool {
		// If track numbers are available, use them
		trackI := ag.TrackOrder[ag.Files[i]]
		trackJ := ag.TrackOrder[ag.Files[j]]

		// If both have track numbers, compare them
		if trackI > 0 && trackJ > 0 {
			return trackI < trackJ
		}

		// If only one has a track number, prioritize it
		if trackI > 0 {
			return true
		}
		if trackJ > 0 {
			return false
		}

		// If neither has a track number, sort by filename
		return ag.Files[i] < ag.Files[j]
	})
}

// ProcessMultiFileAlbum processes a directory containing multiple files that belong to the same album
func (o *Organizer) ProcessMultiFileAlbum(dirPath string) error {
	if o.config.Verbose {
		PrintBlue("🎵 Processing multi-file album in: %s", dirPath)
	}

	// Read all files in the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	// Group files by album metadata
	albumGroups, err := o.groupFilesByAlbum(dirPath, entries)
	if err != nil {
		return err
	}

	// Process each album group
	for _, albumGroup := range albumGroups {
		if err := o.organizeAlbumGroup(albumGroup); err != nil {
			PrintRed("❌ Error organizing album group: %v", err)
		}
	}

	return nil
}

// groupFilesByAlbum groups files in a directory by their album metadata
func (o *Organizer) groupFilesByAlbum(dirPath string, entries []os.DirEntry) (map[string]*AlbumGroup, error) {
	albumGroups := make(map[string]*AlbumGroup)

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		filePath := filepath.Join(dirPath, entry.Name())
		ext := strings.ToLower(filepath.Ext(filePath))

		// Skip non-audio files
		if !IsSupportedAudioFile(ext) {
			continue
		}

		// Extract metadata from the file
		provider := newAudioMetadataProviderFunc(filePath)
		metadata, err := provider.GetMetadata()
		if err != nil {
			if o.config.Verbose {
				PrintYellow("⚠️ Could not extract metadata from %s: %v", filePath, err)
			}
			continue
		}

		// Apply field mapping
		metadata.ApplyFieldMapping(o.config.FieldMapping)

		// Create a key for grouping files by album
		albumKey := o.createAlbumKey(metadata)

		// Add to existing group or create a new one
		group, exists := albumGroups[albumKey]
		if !exists {
			group = NewAlbumGroup(metadata)
			albumGroups[albumKey] = group
		}

		// Add file to the group
		group.AddFile(filePath, metadata.TrackNumber)
	}

	return albumGroups, nil
}

// createAlbumKey creates a unique key for grouping files by album
func (o *Organizer) createAlbumKey(metadata Metadata) string {
	// Normalize author names and title to handle special characters
	normalizedAuthors := normalizeStrings(metadata.Authors)
	normalizedTitle := normalizeString(metadata.Title)

	// Use album, artist, and series (if available) to group files
	key := strings.Join(normalizedAuthors, ",") + "|" + normalizedTitle

	if validSeries := metadata.GetValidSeries(); validSeries != "" {
		normalizedSeries := normalizeString(validSeries)
		key += "|" + normalizedSeries
	}

	return key
}

// normalizeString prepares a string for consistent comparison by removing/replacing special characters
func normalizeString(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Special handling for exactly double dollars with spaces around them
	s = strings.ReplaceAll(s, " $$ ", " doubledollar ")

	// First, consolidate repeated special characters
	var prev rune
	var result strings.Builder
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			// If this is a special character and it's the same as the previous one, skip it
			if r == prev {
				continue
			}
		}
		result.WriteRune(r)
		prev = r
	}

	s = result.String()

	// Convert special marker back to proper format
	s = strings.ReplaceAll(s, " doubledollar ", " dollar dollar ")

	// Replace common special characters and accents
	replacements := map[string]string{
		"&":  "and",
		"+":  "plus",
		"@":  "at",
		"#":  "number",
		"%":  "percent",
		"$":  "dollar",
		"*":  "",
		"\\":"",
		"/":  "",
		":":  "",
		"_":  " ",
		".":  " ",
	}

	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	// Remove extra whitespace
	s = strings.Join(strings.Fields(s), " ")

	return s
}

// normalizeStrings applies normalizeString to a slice of strings
func normalizeStrings(slice []string) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = normalizeString(s)
	}
	return result
}

// organizeAlbumGroup organizes a group of files that belong to the same album
func (o *Organizer) organizeAlbumGroup(albumGroup *AlbumGroup) error {
	if len(albumGroup.Files) == 0 {
		return nil // Nothing to do
	}

	// Sort files by track number
	albumGroup.SortFilesByTrackNumber()

	// Calculate target directory based on the album metadata
	targetDir := o.calculateAlbumTargetDir(albumGroup.Metadata)

	if o.config.Verbose {
		PrintGreen("📂 Organizing album: %s by %s to %s",
			albumGroup.Metadata.Title,
			strings.Join(albumGroup.Metadata.Authors, ", "),
			targetDir)
	}

	// Create target directory if it doesn't exist
	if !o.config.DryRun {
		if err := o.fileOps.CreateDirIfNotExists(targetDir); err != nil {
			return fmt.Errorf("error creating target directory: %w", err)
		}
	}

	// Move each file to the target directory with appropriate track numbering
	for i, filePath := range albumGroup.Files {
		// Get original track number or use index+1 if not available
		trackNum := albumGroup.TrackOrder[filePath]
		if trackNum <= 0 {
			trackNum = i + 1
		}

		// Calculate target filename with track prefix
		fileName := filepath.Base(filePath)
		targetName := AddTrackPrefix(fileName, trackNum)
		targetPath := filepath.Join(targetDir, targetName)

		if o.config.Verbose || o.config.DryRun {
			message := o.formatFileMove(filePath, targetPath, o.config.DryRun)
			fmt.Println(message)
		}

		// Move the file
		if !o.config.DryRun {
			if err := o.moveFile(filePath, targetPath); err != nil {
				PrintRed("❌ Error moving %s: %v", filePath, err)
			}
		}

		// Add to summary
		o.summary.Moves = append(o.summary.Moves, MoveSummary{
			From: filePath,
			To:   targetPath,
		})
	}

	return nil
}

// calculateAlbumTargetDir calculates the target directory for an album based on metadata
func (o *Organizer) calculateAlbumTargetDir(metadata Metadata) string {
	baseDir := o.config.OutputDir
	if baseDir == "" {
		// If no output directory is specified, use the current directory
		baseDir = "."
	}

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
