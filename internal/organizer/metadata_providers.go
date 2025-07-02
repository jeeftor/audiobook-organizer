// internal/organizer/metadata_providers.go
package organizer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"archive/zip"
	"encoding/xml"
	"github.com/dhowden/tag"
	"github.com/pirmd/epub"
)

// JSONMetadataProvider extracts metadata from a JSON file
type JSONMetadataProvider struct {
	filePath string
}

func NewJSONMetadataProvider(path string) *JSONMetadataProvider {
	return &JSONMetadataProvider{filePath: path}
}

func (p *JSONMetadataProvider) GetMetadata() (Metadata, error) {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error reading metadata: %v", err)
	}

	metadata := NewMetadata()
	if err := json.Unmarshal(data, &metadata); err != nil {
		return Metadata{}, fmt.Errorf("error parsing metadata: %v", err)
	}

	metadata.SourcePath = p.filePath
	return metadata, nil
}

// GetMetadataFromEPUBFile extracts and maps metadata from an EPUB file
func GetMetadataFromEPUBFile(epubPath string) (Metadata, error) {
	// Use pirmd/epub library to extract metadata
	info, err := epub.GetMetadataFromFile(epubPath)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error opening EPUB: %v", err)
	}

	metadata := NewMetadataWithSourceType("epub")
	metadata.SourcePath = epubPath

	// Extract values from the Information struct
	var title string
	var authors []string
	var series string
	var seriesIndex float64 = 1.0
	var publisher string
	var language string
	var identifier string
	var subjects []string

	// Get title
	if len(info.Title) > 0 && len(info.Title[0]) > 0 {
		title = info.Title[0]
	}

	// Get authors from creators
	for _, creator := range info.Creator {
		authors = append(authors, creator.FullName)
	}

	// Get series information
	series = info.Series
	if info.SeriesIndex != "" {
		if idx, err := strconv.ParseFloat(info.SeriesIndex, 64); err == nil {
			seriesIndex = idx
		}
	}

	// If series is empty, try to extract from the OPF file directly
	// This is a fallback in case the library doesn't handle all EPUB3 metadata formats
	if series == "" {
		var found bool
		series, seriesIndex, found = extractCalibreSeriesFromOPF(epubPath)
		if !found {
			series = ""
		}
	}

	// Get publisher
	if len(info.Publisher) > 0 {
		publisher = info.Publisher[0]
	}

	// Get language
	if len(info.Language) > 0 {
		language = info.Language[0]
	}

	// Get identifier
	if len(info.Identifier) > 0 {
		identifier = info.Identifier[0].Value
	}

	// Get subjects
	subjects = info.Subject

	// Set the basic fields
	metadata.Title = title
	metadata.Authors = authors
	if series != "" {
		metadata.Series = []string{series}
		// Store series index in raw metadata
		metadata.RawMetadata["series_index"] = seriesIndex
	}

	// Store comprehensive metadata in raw metadata for field mapping and display
	metadata.RawMetadata["title"] = title
	metadata.RawMetadata["series"] = series
	metadata.RawMetadata["authors"] = authors
	metadata.RawMetadata["publisher"] = publisher
	metadata.RawMetadata["language"] = language
	metadata.RawMetadata["identifier"] = identifier
	metadata.RawMetadata["subjects"] = subjects

	return metadata, nil
}

// extractCalibreSeriesFromOPF extracts series information from Calibre metadata in EPUB
func extractCalibreSeriesFromOPF(epubPath string) (string, float64, bool) {
	// Open the EPUB file as a zip archive
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
	var seriesIndex float64 = 1.0 // Default to 1 if not specified
	var foundSeries bool

	for _, meta := range doc.Metadata.Meta {
		// Check for belongs-to-collection property (EPUB3 series)
		if meta.Property == "belongs-to-collection" {
			seriesName = meta.Value
			foundSeries = true

			// Find the corresponding group-position for this collection
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

	return seriesName, seriesIndex, foundSeries
}

// opfDocument represents the structure of an OPF file
type opfDocument struct {
	XMLName  xml.Name `xml:"package"`
	Metadata struct {
		Meta []struct {
			Property string `xml:"property,attr"`
			Refines  string `xml:"refines,attr"`
			ID       string `xml:"id,attr"`
			Value    string `xml:",chardata"`
		} `xml:"meta"`
	} `xml:"metadata"`
}

// GetMetadataFromAudioFile extracts and maps metadata from an audio file
func GetMetadataFromAudioFile(filePath string) (Metadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error opening audio file: %v", err)
	}
	defer file.Close()

	m, err := tag.ReadFrom(file)
	if err != nil {
		return NewMetadata(), fmt.Errorf("error reading audio metadata: %v", err)
	}

	metadata := NewMetadataWithSourceType("audio")
	metadata.SourcePath = filePath

	// Extract comprehensive audio metadata
	title := strings.TrimSpace(m.Title())
	album := strings.TrimSpace(m.Album())
	artist := strings.TrimSpace(m.Artist())
	albumArtist := strings.TrimSpace(m.AlbumArtist())
	composer := strings.TrimSpace(m.Composer())
	genre := strings.TrimSpace(m.Genre())
	comment := strings.TrimSpace(m.Comment())
	lyrics := strings.TrimSpace(m.Lyrics())

	// Get track and disc numbers
	trackNum, trackTotal := m.Track()
	discNum, discTotal := m.Disc()
	year := m.Year()

	// Check for additional fields in raw tags that might contain useful info
	rawTags := m.Raw()

	// Look for additional useful fields
	var narrator, series, trackTitle, contentGroup string

	// Common fields that might contain narrator information
	if val, ok := rawTags["TXXX:NARRATOR"]; ok {
		if str, ok := val.(string); ok {
			narrator = strings.TrimSpace(str)
		}
	}
	if val, ok := rawTags["TXXX:Narrator"]; ok {
		if str, ok := val.(string); ok {
			narrator = strings.TrimSpace(str)
		}
	}

	// Series information might be in various fields
	if val, ok := rawTags["TXXX:SERIES"]; ok {
		if str, ok := val.(string); ok {
			series = strings.TrimSpace(str)
		}
	}
	if val, ok := rawTags["TXXX:Series"]; ok {
		if str, ok := val.(string); ok {
			series = strings.TrimSpace(str)
		}
	}

	// Content group might contain series or book info
	if val, ok := rawTags["TIT1"]; ok { // Content group
		if str, ok := val.(string); ok {
			contentGroup = strings.TrimSpace(str)
		}
	}
	if val, ok := rawTags["CONTENTGROUP"]; ok {
		if str, ok := val.(string); ok {
			contentGroup = strings.TrimSpace(str)
		}
	}

	// Track title might be different from main title
	if val, ok := rawTags["TXXX:TRACK_TITLE"]; ok {
		if str, ok := val.(string); ok {
			trackTitle = strings.TrimSpace(str)
		}
	}

	// Set the basic fields
	metadata.Title = title
	metadata.Album = album
	metadata.TrackNumber = trackNum

	// Set authors based on available artist information
	if artist != "" {
		metadata.Authors = []string{artist}
	} else if albumArtist != "" {
		metadata.Authors = []string{albumArtist}
	} else if narrator != "" {
		metadata.Authors = []string{narrator}
	}

	// For audio files, use series if available, otherwise fall back to album
	if series != "" {
		metadata.Series = []string{series}
	} else if album != "" {
		metadata.Series = []string{album}
	}

	// Store comprehensive metadata in raw metadata for field mapping and display
	metadata.RawMetadata["title"] = title
	metadata.RawMetadata["album"] = album
	metadata.RawMetadata["artist"] = artist
	metadata.RawMetadata["album_artist"] = albumArtist
	metadata.RawMetadata["composer"] = composer
	metadata.RawMetadata["narrator"] = narrator
	metadata.RawMetadata["genre"] = genre
	metadata.RawMetadata["comment"] = comment
	metadata.RawMetadata["lyrics"] = lyrics
	metadata.RawMetadata["series"] = series
	metadata.RawMetadata["track_title"] = trackTitle
	metadata.RawMetadata["content_group"] = contentGroup
	metadata.RawMetadata["track"] = trackNum
	metadata.RawMetadata["track_total"] = trackTotal
	metadata.RawMetadata["disc"] = discNum
	metadata.RawMetadata["disc_total"] = discTotal
	metadata.RawMetadata["year"] = year

	// Set additional fields that might be useful
	if trackTitle != "" {
		metadata.TrackTitle = trackTitle
	}

	return metadata, nil
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
			strings.HasSuffix(lowerName, ".m4a") {
			return filepath.Join(dirPath, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no supported audio files found in directory")
}

func IsEPUBFile(path string) bool {
	if !strings.HasSuffix(strings.ToLower(path), ".epub") {
		return false
	}

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	return true
}

// Provider wrappers for backward compatibility
type EPUBMetadataProvider struct {
	path string
}

func NewEPUBMetadataProvider(path string) *EPUBMetadataProvider {
	return &EPUBMetadataProvider{path: path}
}

func (p *EPUBMetadataProvider) GetMetadata() (Metadata, error) {
	epubPath := p.path
	if !IsEPUBFile(p.path) {
		var err error
		epubPath, err = FindEPUBInDirectory(p.path)
		if err != nil {
			return NewMetadata(), err
		}
	}
	return GetMetadataFromEPUBFile(epubPath)
}

type AudioMetadataProvider struct {
	filePath string
}

func NewAudioMetadataProvider(path string) *AudioMetadataProvider {
	return &AudioMetadataProvider{filePath: path}
}

func (p *AudioMetadataProvider) GetMetadata() (Metadata, error) {
	return GetMetadataFromAudioFile(p.filePath)
}

// Auto-detecting provider
type FileMetadataProvider struct {
	filePath string
}

func NewFileMetadataProvider(path string) *FileMetadataProvider {
	return &FileMetadataProvider{filePath: path}
}

func (p *FileMetadataProvider) GetMetadata() (Metadata, error) {
	ext := strings.ToLower(filepath.Ext(p.filePath))

	switch ext {
	case ".epub":
		return GetMetadataFromEPUBFile(p.filePath)
	case ".json":
		provider := NewJSONMetadataProvider(p.filePath)
		return provider.GetMetadata()
	case ".mp3", ".m4b", ".m4a":
		return GetMetadataFromAudioFile(p.filePath)
	default:
		metadata := NewMetadata()
		metadata.SourcePath = p.filePath
		metadata.SourceType = "unknown"
		return metadata, fmt.Errorf("unsupported file type: %s", ext)
	}
}
