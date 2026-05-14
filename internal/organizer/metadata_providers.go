// internal/organizer/metadata_providers.go
package organizer

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
	"github.com/pirmd/epub"
)

// UnifiedMetadataProvider handles all metadata extraction from different file types
type UnifiedMetadataProvider struct {
	filePath        string
	sourceType      string
	useEmbeddedOnly bool // If true, ignore metadata.json and use only embedded metadata
}

// NewMetadataProvider creates a unified metadata provider that auto-detects file type
// useEmbeddedOnly: if true, ignore metadata.json and use only embedded metadata from audio files
func NewMetadataProvider(path string, useEmbeddedOnly bool) *UnifiedMetadataProvider {
	return &UnifiedMetadataProvider{
		filePath:        path,
		sourceType:      detectSourceType(path, useEmbeddedOnly),
		useEmbeddedOnly: useEmbeddedOnly,
	}
}

// GetMetadata extracts metadata based on the detected file type
func (p *UnifiedMetadataProvider) GetMetadata() (Metadata, error) {
	switch p.sourceType {
	case "json":
		return p.extractJSONMetadata()
	case "epub":
		return p.extractEPUBMetadata()
	case "audio":
		return p.extractAudioMetadata()
	default:
		return NewMetadata(), fmt.Errorf("unsupported file type: %s", p.sourceType)
	}
}

// detectSourceType determines the file type based on extension
func detectSourceType(path string, useEmbeddedOnly bool) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return "json"
	case ".epub":
		return "epub"
	case ".mp3", ".m4b", ".m4a", ".ogg", ".flac":
		return "audio"
	default:
		// Try to detect if it's a directory with specific files
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			// If useEmbeddedOnly is true, skip metadata.json detection
			if !useEmbeddedOnly {
				if _, err := os.Stat(filepath.Join(path, "metadata.json")); err == nil {
					return "json"
				}
			}
			if _, err := FindEPUBInDirectory(path); err == nil {
				return "epub"
			}
			if _, err := FindAudioFileInDirectory(path); err == nil {
				return "audio"
			}
		}
		return "unknown"
	}
}

// extractJSONMetadata reads metadata from a JSON file
// HYBRID MODE: Also extracts file-level metadata (track#, disc#) from embedded audio tags
func (p *UnifiedMetadataProvider) extractJSONMetadata() (Metadata, error) {
	var jsonPath string
	var dirPath string

	// If path is a directory, look for metadata.json inside it
	if info, err := os.Stat(p.filePath); err == nil && info.IsDir() {
		dirPath = p.filePath
		jsonPath = filepath.Join(p.filePath, "metadata.json")
	} else {
		dirPath = filepath.Dir(p.filePath)
		jsonPath = p.filePath
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error reading metadata file: %v", err)
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return NewMetadata(), fmt.Errorf("error parsing metadata: %v", err)
	}

	metadata := NewMetadata()
	metadata.SourcePath = jsonPath
	metadata.SourceType = "json"

	// HYBRID: Extract file-level metadata from audio file if available
	// Track numbers and disc numbers come from the actual audio file, not metadata.json
	if audioPath, err := FindAudioFileInDirectory(dirPath); err == nil {
		if fileLevelMetadata, err := extractFileLevelMetadata(audioPath); err == nil {
			// Merge file-level metadata (track#, disc#) into book-level metadata
			metadata.TrackNumber = fileLevelMetadata.TrackNumber
			if fileLevelMetadata.RawData != nil {
				if metadata.RawData == nil {
					metadata.RawData = make(map[string]interface{})
				}
				// Copy file-level fields with a prefix to indicate source
				metadata.RawData["track"] = fileLevelMetadata.RawData["track"]
				metadata.RawData["track_total"] = fileLevelMetadata.RawData["track_total"]
				metadata.RawData["disc"] = fileLevelMetadata.RawData["disc"]
				metadata.RawData["disc_total"] = fileLevelMetadata.RawData["disc_total"]
				metadata.RawData["discnumber"] = fileLevelMetadata.RawData["discnumber"]
				metadata.RawData["_embedded_source"] = audioPath // Track where file-level data came from
			}
		}
	}

	// Extract basic fields from JSON (book-level metadata)
	if title, ok := rawData["title"].(string); ok {
		metadata.Title = title
	}

	if authors, ok := rawData["authors"].([]interface{}); ok {
		for _, author := range authors {
			if authorStr, ok := author.(string); ok {
				metadata.Authors = append(metadata.Authors, authorStr)
			}
		}
	}

	if series, ok := rawData["series"].([]interface{}); ok {
		for _, s := range series {
			if seriesStr, ok := s.(string); ok {
				metadata.Series = append(metadata.Series, seriesStr)
			}
		}
	}

	if trackNum, ok := rawData["track_number"].(float64); ok {
		metadata.TrackNumber = int(trackNum)
	}

	// Merge JSON data into RawData (preserving hybrid track/disc fields from above)
	// If metadata.RawData is nil, initialize it
	if metadata.RawData == nil {
		metadata.RawData = make(map[string]interface{})
	}

	// Save the hybrid fields before merging
	savedTrack := metadata.RawData["track"]
	savedTrackTotal := metadata.RawData["track_total"]
	savedDisc := metadata.RawData["disc"]
	savedDiscTotal := metadata.RawData["disc_total"]
	savedDiscNumber := metadata.RawData["discnumber"]
	savedEmbeddedSource := metadata.RawData["_embedded_source"]

	// Copy all JSON fields
	for key, val := range rawData {
		metadata.RawData[key] = val
	}

	// Restore hybrid fields (they take precedence over JSON)
	if savedTrack != nil {
		metadata.RawData["track"] = savedTrack
	}
	if savedTrackTotal != nil {
		metadata.RawData["track_total"] = savedTrackTotal
	}
	if savedDisc != nil {
		metadata.RawData["disc"] = savedDisc
	}
	if savedDiscTotal != nil {
		metadata.RawData["disc_total"] = savedDiscTotal
	}
	if savedDiscNumber != nil {
		metadata.RawData["discnumber"] = savedDiscNumber
	}
	if savedEmbeddedSource != nil {
		metadata.RawData["_embedded_source"] = savedEmbeddedSource
	}

	return metadata, nil
}

// extractBookLevelMetadataFromJSON extracts ONLY book-level metadata from metadata.json
// Does NOT perform any audio file lookups (used for hybrid mode)
func extractBookLevelMetadataFromJSON(jsonPath string) (Metadata, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error reading metadata file: %v", err)
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return NewMetadata(), fmt.Errorf("error parsing metadata: %v", err)
	}

	metadata := NewMetadata()
	metadata.SourcePath = jsonPath
	metadata.SourceType = "json"
	metadata.RawData = rawData // Store all JSON fields in RawData

	// Extract basic fields from JSON (book-level metadata only)
	if title, ok := rawData["title"].(string); ok {
		metadata.Title = title
	}

	if authors, ok := rawData["authors"].([]interface{}); ok {
		for _, author := range authors {
			if authorStr, ok := author.(string); ok {
				metadata.Authors = append(metadata.Authors, authorStr)
			}
		}
	}

	if series, ok := rawData["series"].([]interface{}); ok {
		for _, s := range series {
			if seriesStr, ok := s.(string); ok {
				metadata.Series = append(metadata.Series, seriesStr)
			}
		}
	}

	return metadata, nil
}

// extractEPUBMetadata extracts metadata from EPUB files
func (p *UnifiedMetadataProvider) extractEPUBMetadata() (Metadata, error) {
	var epubPath string

	// If path is a directory, find EPUB file inside it
	if info, err := os.Stat(p.filePath); err == nil && info.IsDir() {
		var err error
		epubPath, err = FindEPUBInDirectory(p.filePath)
		if err != nil {
			return NewMetadata(), err
		}
	} else {
		epubPath = p.filePath
	}

	// Use pirmd/epub library to extract metadata
	info, err := epub.GetMetadataFromFile(epubPath)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error opening EPUB: %v", err)
	}

	metadata := NewMetadata()
	metadata.SourcePath = epubPath
	metadata.SourceType = "epub"
	metadata.RawData = make(map[string]interface{})

	// Extract basic information
	if len(info.Title) > 0 && len(info.Title[0]) > 0 {
		metadata.Title = info.Title[0]
		metadata.RawData["title"] = metadata.Title
	}

	// Get authors from creators
	for _, creator := range info.Creator {
		metadata.Authors = append(metadata.Authors, creator.FullName)
	}
	metadata.RawData["authors"] = metadata.Authors

	// Get series information
	series := info.Series
	seriesIndex := 1.0
	if info.SeriesIndex != "" {
		if idx, err := strconv.ParseFloat(info.SeriesIndex, 64); err == nil {
			seriesIndex = idx
		}
	}

	// If series is empty, try to extract from the OPF file directly
	if series == "" {
		// Try to extract Calibre series metadata if available
		var found bool
		series, seriesIndex, found = ExtractCalibreSeriesFromOPF(epubPath)
		if !found {
			series = ""
		}
	}

	// Add series to metadata if found
	if series != "" {
		metadata.Series = []string{series}
		metadata.RawData["series"] = series
		metadata.RawData["series_index"] = seriesIndex
	}

	// Store additional metadata
	if len(info.Publisher) > 0 {
		metadata.RawData["publisher"] = info.Publisher[0]
	}
	if len(info.Language) > 0 {
		metadata.RawData["language"] = info.Language[0]
	}
	if len(info.Identifier) > 0 {
		metadata.RawData["identifier"] = info.Identifier[0].Value
	}
	metadata.RawData["subjects"] = info.Subject

	return metadata, nil
}

// extractAudioMetadata extracts metadata from audio files
// HYBRID MODE: If metadata.json exists in parent directory, merge book-level data from JSON with file-level data from audio
func (p *UnifiedMetadataProvider) extractAudioMetadata() (Metadata, error) {
	var audioPath string
	var dirPath string

	// If path is a directory, find audio file inside it
	if info, err := os.Stat(p.filePath); err == nil && info.IsDir() {
		dirPath = p.filePath
		var err error
		audioPath, err = FindAudioFileInDirectory(p.filePath)
		if err != nil {
			return NewMetadata(), err
		}
	} else {
		audioPath = p.filePath
		dirPath = filepath.Dir(audioPath)
	}

	// HYBRID MODE: Check if metadata.json exists in the same directory
	// BUT only if useEmbeddedOnly is false
	metadataJSONPath := filepath.Join(dirPath, "metadata.json")
	var bookMetadata *Metadata
	if !p.useEmbeddedOnly {
		if _, err := os.Stat(metadataJSONPath); err == nil {
			// metadata.json exists - extract ONLY book-level metadata from it (no audio file lookup)
			if jsonMeta, err := extractBookLevelMetadataFromJSON(metadataJSONPath); err == nil {
				bookMetadata = &jsonMeta
			}
		}
	}

	file, err := os.Open(audioPath)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error opening audio file: %v", err)
	}
	defer file.Close()

	m, err := tag.ReadFrom(file)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error reading audio metadata: %v", err)
	}

	metadata := NewMetadata()
	metadata.SourcePath = audioPath
	metadata.SourceType = "audio"
	metadata.RawData = make(map[string]interface{})

	// Extract basic fields
	metadata.Title = strings.TrimSpace(m.Title())
	metadata.Album = strings.TrimSpace(m.Album())

	// Set authors based on available artist information
	artist := strings.TrimSpace(m.Artist())
	albumArtist := strings.TrimSpace(m.AlbumArtist())

	if artist != "" {
		metadata.Authors = []string{artist}
	} else if albumArtist != "" {
		metadata.Authors = []string{albumArtist}
	}

	// For audio files, use album as series if no explicit series
	if metadata.Album != "" {
		metadata.Series = []string{metadata.Album}
	}

	// Check for additional fields in raw tags first (to support all variations)
	rawTags := m.Raw()

	// Get track numbers - check all variations (case-insensitive)
	// Audiobookshelf spec: track, trck, trk
	trackNum := 0
	if num, _ := m.Track(); num > 0 {
		trackNum = num
	} else {
		// Check raw tags for variations
		trackNum = getTrackNumberFromRaw(rawTags)
	}
	metadata.TrackNumber = trackNum

	// Get disc numbers - check all variations (case-insensitive)
	// Audiobookshelf spec: discnumber, disc, disk, tpos
	discNum, discTotal := 0, 0
	if num, total := m.Disc(); num > 0 {
		discNum, discTotal = num, total
	} else {
		// Check raw tags for variations
		discNum = getDiscNumberFromRaw(rawTags)
	}

	trackTotal := 0
	if _, total := m.Track(); total > 0 {
		trackTotal = total
	}

	// Store comprehensive raw metadata
	metadata.RawData["title"] = metadata.Title
	metadata.RawData["album"] = metadata.Album
	metadata.RawData["artist"] = artist
	metadata.RawData["album_artist"] = albumArtist
	metadata.RawData["composer"] = strings.TrimSpace(m.Composer())
	metadata.RawData["genre"] = strings.TrimSpace(m.Genre())
	metadata.RawData["comment"] = strings.TrimSpace(m.Comment())
	metadata.RawData["lyrics"] = strings.TrimSpace(m.Lyrics())
	metadata.RawData["year"] = m.Year()
	metadata.RawData["track"] = trackNum
	metadata.RawData["track_total"] = trackTotal
	metadata.RawData["disc"] = discNum
	metadata.RawData["disc_total"] = discTotal
	metadata.RawData["discnumber"] = discNum // Alias for disc

	// Look for narrator information
	if val, ok := rawTags["TXXX:NARRATOR"]; ok {
		if str, ok := val.(string); ok {
			metadata.RawData["narrator"] = strings.TrimSpace(str)
		}
	}
	if val, ok := rawTags["TXXX:Narrator"]; ok {
		if str, ok := val.(string); ok {
			metadata.RawData["narrator"] = strings.TrimSpace(str)
		}
	}

	// Look for series information
	if val, ok := rawTags["TXXX:SERIES"]; ok {
		if str, ok := val.(string); ok {
			metadata.RawData["series"] = strings.TrimSpace(str)
			metadata.Series = []string{strings.TrimSpace(str)}
		}
	}

	// Content group might contain series info
	if val, ok := rawTags["TIT1"]; ok {
		if str, ok := val.(string); ok {
			metadata.RawData["content_group"] = strings.TrimSpace(str)
		}
	}

	// HYBRID MODE: If we found metadata.json, merge book-level data with file-level data
	if bookMetadata != nil {
		// Save file-level fields BEFORE merge
		audioTrack := metadata.RawData["track"]
		audioDisc := metadata.RawData["disc"]
		audioTrackTotal := metadata.RawData["track_total"]
		audioDiscTotal := metadata.RawData["disc_total"]
		audioDiscNumber := metadata.RawData["discnumber"]

		// Use book-level metadata for these fields (from JSON)
		metadata.Title = bookMetadata.Title
		metadata.Authors = bookMetadata.Authors
		metadata.Series = bookMetadata.Series
		metadata.Album = bookMetadata.Album

		// Replace RawData with JSON data (book-level)
		metadata.RawData = make(map[string]interface{})
		for key, val := range bookMetadata.RawData {
			metadata.RawData[key] = val
		}

		// Restore file-level fields from audio
		if audioTrack != nil {
			metadata.RawData["track"] = audioTrack
		}
		if audioDisc != nil {
			metadata.RawData["disc"] = audioDisc
		}
		if audioTrackTotal != nil {
			metadata.RawData["track_total"] = audioTrackTotal
		}
		if audioDiscTotal != nil {
			metadata.RawData["disc_total"] = audioDiscTotal
		}
		if audioDiscNumber != nil {
			metadata.RawData["discnumber"] = audioDiscNumber
		}

		// Mark as JSON source type (hybrid mode) and track embedded source
		metadata.SourceType = "json"
		metadata.SourcePath = bookMetadata.SourcePath
		metadata.RawData["_embedded_source"] = audioPath
	}

	return metadata, nil
}

// ExtractCalibreSeriesFromOPF extracts series information from Calibre metadata in EPUB
func ExtractCalibreSeriesFromOPF(epubPath string) (string, float64, bool) {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return "", 0, false
	}
	defer r.Close()

	// Find the OPF file
	var opfFile *zip.File
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			opfFile = f
			break
		}
	}

	if opfFile == nil {
		return "", 0, false
	}

	// Read the OPF file
	rc, err := opfFile.Open()
	if err != nil {
		return "", 0, false
	}
	defer rc.Close()

	opfContent, err := io.ReadAll(rc)
	if err != nil {
		return "", 0, false
	}

	// Parse the OPF XML
	var doc opfDocument
	if err := xml.Unmarshal(opfContent, &doc); err != nil {
		return "", 0, false
	}

	// Look for series metadata
	var seriesName string
	var seriesIndex float64 = 1.0
	var foundSeries bool

	// Method 1: Check for belongs-to-collection property (EPUB3 standard)
	for _, meta := range doc.Metadata.Meta {
		if meta.Property == "belongs-to-collection" {
			seriesName = meta.Value
			foundSeries = true

			// Find the corresponding group-position
			collectionID := meta.ID
			if collectionID != "" {
				for _, refMeta := range doc.Metadata.Meta {
					if refMeta.Refines == "#"+collectionID && refMeta.Property == "group-position" {
						if idx, err := strconv.ParseFloat(refMeta.Value, 64); err == nil {
							seriesIndex = idx
						}
						break
					}
				}
			}
			break
		}
	}

	// If not found yet, try Method 2: Check for calibre:series metadata (Calibre format)
	if !foundSeries {
		for _, meta := range doc.Metadata.Meta {
			if meta.Name == "calibre:series" && meta.Content != "" {
				seriesName = meta.Content
				foundSeries = true

				// Look for series index
				for _, indexMeta := range doc.Metadata.Meta {
					if indexMeta.Name == "calibre:series_index" && indexMeta.Content != "" {
						if idx, err := strconv.ParseFloat(indexMeta.Content, 64); err == nil {
							seriesIndex = idx
						}
						break
					}
				}
				break
			}
		}
	}

	return seriesName, seriesIndex, foundSeries
}

// opfDocument represents the structure of an OPF file
type opfDocument struct {
	XMLName  xml.Name `xml:"package"`
	Metadata struct {
		Meta []struct {
			// EPUB3 standard attributes
			Property string `xml:"property,attr"`
			Refines  string `xml:"refines,attr"`
			ID       string `xml:"id,attr"`
			Value    string `xml:",chardata"`
			// Calibre specific attributes
			Name    string `xml:"name,attr"`
			Content string `xml:"content,attr"`
		} `xml:"meta"`
	} `xml:"metadata"`
}

// Helper functions for directory scanning
func FindEPUBInDirectory(dirPath string) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".epub") {
			return filepath.Join(dirPath, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no EPUB file found in directory")
}

func FindAudioFileInDirectory(dirPath string) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		lowerName := strings.ToLower(entry.Name())
		if strings.HasSuffix(lowerName, ".mp3") ||
			strings.HasSuffix(lowerName, ".m4b") ||
			strings.HasSuffix(lowerName, ".m4a") ||
			strings.HasSuffix(lowerName, ".ogg") ||
			strings.HasSuffix(lowerName, ".flac") {
			return filepath.Join(dirPath, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no supported audio files found in directory")
}

// Legacy provider interfaces for backward compatibility
// getTrackNumberFromRaw checks all track number field variations (case-insensitive)
// Audiobookshelf spec: track, trck, trk
func getTrackNumberFromRaw(rawTags map[string]interface{}) int {
	// Check common track field variations (case-insensitive)
	trackFields := []string{"TRCK", "TRK", "track", "trck", "trk", "TRACK"}

	for _, field := range trackFields {
		if val, ok := rawTags[field]; ok {
			return parseTrackNumber(val)
		}
	}

	return 0
}

// getDiscNumberFromRaw checks all disc number field variations (case-insensitive)
// Audiobookshelf spec: discnumber, disc, disk, tpos
func getDiscNumberFromRaw(rawTags map[string]interface{}) int {
	// Check common disc field variations (case-insensitive)
	discFields := []string{"TPOS", "discnumber", "disc", "disk", "DISC", "DISK", "DISCNUMBER"}

	for _, field := range discFields {
		if val, ok := rawTags[field]; ok {
			return parseTrackNumber(val) // Reuse same parsing logic
		}
	}

	return 0
}

// parseTrackNumber extracts track number from various formats
func parseTrackNumber(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case string:
		// Handle "track/total" format (e.g., "3/12")
		parts := strings.Split(v, "/")
		if num, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
			return num
		}
	}
	return 0
}

// extractFileLevelMetadata extracts ONLY file-level metadata (track#, disc#) from audio files
// Used in hybrid mode when metadata.json provides book-level data
func extractFileLevelMetadata(audioPath string) (Metadata, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error opening audio file: %v", err)
	}
	defer file.Close()

	m, err := tag.ReadFrom(file)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error reading audio metadata: %v", err)
	}

	metadata := NewMetadata()
	metadata.RawData = make(map[string]interface{})
	rawTags := m.Raw()

	// Get track numbers - check all variations
	trackNum := 0
	if num, _ := m.Track(); num > 0 {
		trackNum = num
	} else {
		trackNum = getTrackNumberFromRaw(rawTags)
	}
	metadata.TrackNumber = trackNum

	// Get disc numbers - check all variations
	discNum, discTotal := 0, 0
	if num, total := m.Disc(); num > 0 {
		discNum, discTotal = num, total
	} else {
		discNum = getDiscNumberFromRaw(rawTags)
	}

	trackTotal := 0
	if _, total := m.Track(); total > 0 {
		trackTotal = total
	}

	// Store only file-level metadata
	metadata.RawData["track"] = trackNum
	metadata.RawData["track_total"] = trackTotal
	metadata.RawData["disc"] = discNum
	metadata.RawData["disc_total"] = discTotal
	metadata.RawData["discnumber"] = discNum

	return metadata, nil
}

// JSONMetadataProvider is a convenience wrapper around UnifiedMetadataProvider.
// Deprecated: Use NewMetadataProvider(path, false) directly for automatic hybrid extraction.
type JSONMetadataProvider struct {
	*UnifiedMetadataProvider
}

// NewJSONMetadataProvider creates a metadata provider for JSON files with hybrid extraction.
// Deprecated: Use NewMetadataProvider(path, false) directly instead.
// This wrapper is kept for backwards compatibility but adds no additional functionality.
func NewJSONMetadataProvider(path string) *JSONMetadataProvider {
	return &JSONMetadataProvider{NewMetadataProvider(path, false)}
}

// EPUBMetadataProvider is a convenience wrapper around UnifiedMetadataProvider.
// Deprecated: Use NewMetadataProvider(path, false) directly for automatic file type detection.
type EPUBMetadataProvider struct {
	*UnifiedMetadataProvider
}

// NewEPUBMetadataProvider creates a metadata provider for EPUB files.
// Deprecated: Use NewMetadataProvider(path, false) directly instead.
// This wrapper is kept for backwards compatibility but adds no additional functionality.
func NewEPUBMetadataProvider(path string) *EPUBMetadataProvider {
	return &EPUBMetadataProvider{NewMetadataProvider(path, false)}
}

// AudioMetadataProvider is a convenience wrapper around UnifiedMetadataProvider.
// Deprecated: Use NewMetadataProvider(path, false) directly for automatic hybrid extraction.
type AudioMetadataProvider struct {
	*UnifiedMetadataProvider
}

// NewAudioMetadataProvider creates a metadata provider for audio files with hybrid extraction.
// Deprecated: Use NewMetadataProvider(path, false) directly instead.
// This wrapper is kept for backwards compatibility but adds no additional functionality.
func NewAudioMetadataProvider(path string) *AudioMetadataProvider {
	return &AudioMetadataProvider{NewMetadataProvider(path, false)}
}

// FileMetadataProvider is a convenience wrapper around UnifiedMetadataProvider.
// Deprecated: Use NewMetadataProvider(path, false) directly for automatic file type detection.
type FileMetadataProvider struct {
	*UnifiedMetadataProvider
}

// NewFileMetadataProvider creates a metadata provider for any supported file type.
// Deprecated: Use NewMetadataProvider(path, false) directly instead.
// This wrapper is kept for backwards compatibility but adds no additional functionality.
func NewFileMetadataProvider(path string) *FileMetadataProvider {
	return &FileMetadataProvider{NewMetadataProvider(path, false)}
}
