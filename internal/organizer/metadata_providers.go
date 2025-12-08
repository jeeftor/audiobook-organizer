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
	filePath   string
	sourceType string
}

// NewMetadataProvider creates a unified metadata provider that auto-detects file type
func NewMetadataProvider(path string) *UnifiedMetadataProvider {
	return &UnifiedMetadataProvider{
		filePath:   path,
		sourceType: detectSourceType(path),
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
func detectSourceType(path string) string {
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
			if _, err := os.Stat(filepath.Join(path, "metadata.json")); err == nil {
				return "json"
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
func (p *UnifiedMetadataProvider) extractJSONMetadata() (Metadata, error) {
	var jsonPath string

	// If path is a directory, look for metadata.json inside it
	if info, err := os.Stat(p.filePath); err == nil && info.IsDir() {
		jsonPath = filepath.Join(p.filePath, "metadata.json")
	} else {
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

	// Extract basic fields
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

	// Store all raw data for field mapping
	metadata.RawData = rawData

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
func (p *UnifiedMetadataProvider) extractAudioMetadata() (Metadata, error) {
	var audioPath string

	// If path is a directory, find audio file inside it
	if info, err := os.Stat(p.filePath); err == nil && info.IsDir() {
		var err error
		audioPath, err = FindAudioFileInDirectory(p.filePath)
		if err != nil {
			return NewMetadata(), err
		}
	} else {
		audioPath = p.filePath
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

	// Get track numbers
	trackNum, _ := m.Track()
	metadata.TrackNumber = trackNum

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

	trackTotal, discNum, discTotal := 0, 0, 0
	if _, total := m.Track(); total > 0 {
		trackTotal = total
	}
	if num, total := m.Disc(); num > 0 {
		discNum, discTotal = num, total
	}

	metadata.RawData["track_total"] = trackTotal
	metadata.RawData["disc"] = discNum
	metadata.RawData["disc_total"] = discTotal

	// Check for additional fields in raw tags
	rawTags := m.Raw()

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
type JSONMetadataProvider struct {
	*UnifiedMetadataProvider
}

func NewJSONMetadataProvider(path string) *JSONMetadataProvider {
	return &JSONMetadataProvider{NewMetadataProvider(path)}
}

type EPUBMetadataProvider struct {
	*UnifiedMetadataProvider
}

func NewEPUBMetadataProvider(path string) *EPUBMetadataProvider {
	return &EPUBMetadataProvider{NewMetadataProvider(path)}
}

type AudioMetadataProvider struct {
	*UnifiedMetadataProvider
}

func NewAudioMetadataProvider(path string) *AudioMetadataProvider {
	return &AudioMetadataProvider{NewMetadataProvider(path)}
}

// GetMetadata extracts metadata only from embedded audio tags (ignores metadata.json)
func (p *AudioMetadataProvider) GetMetadata() (Metadata, error) {
	return p.extractAudioMetadata()
}

type FileMetadataProvider struct {
	*UnifiedMetadataProvider
}

func NewFileMetadataProvider(path string) *FileMetadataProvider {
	return &FileMetadataProvider{NewMetadataProvider(path)}
}
