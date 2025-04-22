// organizer/metadata_providers.go
package organizer

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dhowden/tag"
	"github.com/fatih/color"
	"github.com/meskio/epubgo"
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
		return Metadata{}, fmt.Errorf("error reading metadata: %v", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return Metadata{}, fmt.Errorf("error parsing metadata: %v", err)
	}

	return metadata, nil
}

// EPUBMetadataProvider extracts metadata from an EPUB file
type EPUBMetadataProvider struct {
	path string // Can be either a directory path or direct path to an EPUB file
}

// NewEPUBMetadataProvider creates a provider that extracts metadata from EPUB files
// path can be either a directory containing EPUB files or a direct path to an EPUB file
func NewEPUBMetadataProvider(path string) *EPUBMetadataProvider {
	return &EPUBMetadataProvider{path: path}
}

// FindEPUBInDirectory searches for EPUB files in a directory and returns the path to the first one found
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

// IsEPUBFile checks if the given path is an EPUB file
func IsEPUBFile(path string) bool {
	// Check file extension
	if !strings.HasSuffix(strings.ToLower(path), ".epub") {
		return false
	}

	// Check if it exists and is a file
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	return true
}

// GetMetadata returns book metadata from an EPUB file
func (p *EPUBMetadataProvider) GetMetadata() (Metadata, error) {
	// Find the EPUB file path
	epubPath, err := p.resolveEPUBPath()
	if err != nil {
		return Metadata{}, err
	}

	// Open the EPUB file
	book, err := epubgo.Open(epubPath)
	if err != nil {
		return Metadata{}, fmt.Errorf("error opening EPUB: %v", err)
	}
	defer book.Close()

	// Create and populate metadata
	metadata := Metadata{}
	p.extractBasicMetadata(book, &metadata)
	p.extractSeriesMetadata(book, &metadata)

	// Print all metadata fields when in verbose mode
	if IsVerboseMode() {
		PrintAllMetadataFields(book, epubPath)
	}

	return metadata, nil
}

// PrintAllMetadataFields prints all available metadata fields from an EPUB file when in verbose mode
func PrintAllMetadataFields(book *epubgo.Epub, epubPath string) {
	color.Cyan("📖 EPUB Metadata: %s", filepath.Base(epubPath))

	// Get all available metadata fields
	fields := book.MetadataFields()

	if len(fields) == 0 {
		color.Yellow("  No metadata fields found")
		return
	}

	// Print important fields first with better formatting
	printMetadataField(book, "title", "Title")
	printMetadataField(book, "creator", "Author(s)")
	printMetadataField(book, "language", "Language")

	// Try to print series information from various sources
	printSeriesInfo(book)

	// Print other fields in a more compact format
	color.White("\n  Other metadata:")
	for _, field := range fields {
		// Skip fields we've already shown
		if field == "title" || field == "creator" || field == "language" {
			continue
		}

		values, err := book.Metadata(field)
		if err != nil || len(values) == 0 {
			continue
		}

		// Print field name and first value (most fields only have one value)
		if len(values) == 1 {
			color.White("    %s: %s", field, values[0])
		} else {
			color.White("    %s: %v", field, values)
		}
	}

	// Print meta tags separately if they contain interesting information
	printMetaTags(book)
}

// printMetadataField prints a specific metadata field with nice formatting
func printMetadataField(book *epubgo.Epub, field, label string) {
	values, err := book.Metadata(field)
	if err != nil || len(values) == 0 || values[0] == "" {
		return
	}

	if len(values) == 1 {
		color.White("  %s: %s", label, values[0])
	} else {
		color.White("  %s: %v", label, strings.Join(values, ", "))
	}
}

// printSeriesInfo attempts to find and print series information from various sources
func printSeriesInfo(book *epubgo.Epub) {
	// Check standard series fields
	series, _ := book.Metadata("calibre:series")
	if len(series) > 0 && series[0] != "" {
		color.White("  Series: %s", series[0])
		return
	}

	series, _ = book.Metadata("series")
	if len(series) > 0 && series[0] != "" {
		color.White("  Series: %s", series[0])
		return
	}

	// Check for series in meta tags
	meta, err := book.MetadataAttr("meta")
	if err == nil && len(meta) > 0 {
		for _, attr := range meta {
			if name, ok := attr["name"]; ok && (name == "calibre:series" || strings.Contains(name, "series")) {
				if content, ok := attr["content"]; ok && content != "" {
					color.White("  Series: %s (from meta tag)", content)
					return
				}
			}
		}
	}

	// Try to extract from OPF package metadata (Calibre sometimes puts it here)
	// This is a more direct approach to access calibre:series
	// Unfortunately, epubgo doesn't expose this directly, so we'll need to use a different approach
	// for production code
}

// printMetaTags prints meta tags that contain interesting information
func printMetaTags(book *epubgo.Epub) {
	meta, err := book.MetadataAttr("meta")
	if err != nil || len(meta) == 0 {
		return
	}

	// Check if there are any interesting meta tags
	interestingTags := false
	for _, attr := range meta {
		if name, ok := attr["name"]; ok && name != "" {
			if content, ok := attr["content"]; ok && content != "" {
				if strings.Contains(name, "calibre") || strings.Contains(name, "series") {
					interestingTags = true
					break
				}
			}
		}
	}

	if !interestingTags {
		return
	}

	color.White("\n  Calibre metadata:")
	for _, attr := range meta {
		if name, ok := attr["name"]; ok && name != "" {
			if content, ok := attr["content"]; ok && content != "" {
				if strings.Contains(name, "calibre") || strings.Contains(name, "series") {
					color.White("    %s: %s", name, content)
				}
			}
		}
	}
}

// resolveEPUBPath determines the actual EPUB file path
func (p *EPUBMetadataProvider) resolveEPUBPath() (string, error) {
	if IsEPUBFile(p.path) {
		return p.path, nil
	}

	// Treat as directory and find an EPUB file
	epubPath, err := FindEPUBInDirectory(p.path)
	if err != nil {
		return "", err
	}

	return epubPath, nil
}

// extractBasicMetadata gets title and author information
func (p *EPUBMetadataProvider) extractBasicMetadata(book *epubgo.Epub, metadata *Metadata) {
	// Get title
	title, err := book.Metadata("title")
	if err == nil && len(title) > 0 && title[0] != "" {
		metadata.Title = title[0]
	}

	// Get authors
	creators, err := book.Metadata("creator")
	if err == nil && len(creators) > 0 && creators[0] != "" {
		metadata.Authors = creators
	}
}

// extractSeriesMetadata attempts to find series information from various metadata fields
func (p *EPUBMetadataProvider) extractSeriesMetadata(book *epubgo.Epub, metadata *Metadata) {
	invalidSeries := false

	// Try standard series fields first
	if !p.tryStandardSeriesFields(book, metadata) {
		// Try subject fields next
		if !p.trySubjectFields(book, metadata) {
			// Try direct OPF parsing for Calibre metadata
			if series, ok := extractCalibreSeriesFromOPF(p.path); ok {
				if isValidSeries(series) {
					metadata.Series = []string{series}
					return
				} else if series != "" {
					invalidSeries = true
				}
			}

			// Finally, try to extract from title
			if metadata.Title != "" {
				seriesInfo := extractSeriesFromTitle(metadata.Title)
				if isValidSeries(seriesInfo) {
					metadata.Series = []string{seriesInfo}
					return
				} else if seriesInfo != "" {
					invalidSeries = true
				}
			}
		}
	}

	// Scan meta tags for invalid series-like data if we didn't find a valid series
	if len(metadata.Series) == 0 || metadata.Series[0] == "" {
		meta, err := book.MetadataAttr("meta")
		if err == nil && len(meta) > 0 {
			for _, attr := range meta {
				if name, ok := attr["name"]; ok && (name == "calibre:series" || strings.Contains(name, "series")) {
					if content, ok := attr["content"]; ok && content != "" {
						if !isValidSeries(content) {
							invalidSeries = true
						}
					}
				}
			}
		}
	}

	// If we found an invalid series anywhere, set Series to a sentinel value for logging
	if invalidSeries && (len(metadata.Series) == 0 || metadata.Series[0] == "") {
		metadata.Series = []string{"__INVALID_SERIES__"}
	}
}

// isValidSeries returns true if the series string is a valid, human-readable name
func isValidSeries(series string) bool {
	series = strings.TrimSpace(series)
	if series == "" {
		return false
	}
	if strings.HasPrefix(series, "{") || strings.HasPrefix(series, "[") {
		return false
	}
	if strings.Contains(series, "is_category") || strings.Contains(series, "kind") {
		return false
	}
	if len(series) > 100 {
		return false
	}
	return true
}

// tryStandardSeriesFields checks dedicated series metadata fields
func (p *EPUBMetadataProvider) tryStandardSeriesFields(book *epubgo.Epub, metadata *Metadata) bool {
	// Check for calibre:series metadata
	calibreSeries, err := book.Metadata("calibre:series")
	if err == nil && len(calibreSeries) > 0 && calibreSeries[0] != "" {
		if isValidSeries(calibreSeries[0]) {
			metadata.Series = []string{calibreSeries[0]}
			return true
		}
	}

	// Check for series metadata
	series, err := book.Metadata("series")
	if err == nil && len(series) > 0 && series[0] != "" {
		if isValidSeries(series[0]) {
			metadata.Series = []string{series[0]}
			return true
		}
		return false
	}

	// Check for belongs-to-collection metadata (used in some EPUB 3.0 files)
	collection, err := book.Metadata("belongs-to-collection")
	if err == nil && len(collection) > 0 && collection[0] != "" {
		if isValidSeries(collection[0]) {
			metadata.Series = []string{collection[0]}
			return true
		}
	}

	// Check for dc:description that might contain series info (Calibre sometimes puts it here)
	description, err := book.Metadata("description")
	if err == nil && len(description) > 0 && description[0] != "" {
		// Look for series pattern in description
		if strings.Contains(description[0], "Series:") {
			parts := strings.Split(description[0], "Series:")
			if len(parts) > 1 {
				seriesInfo := strings.TrimSpace(parts[1])
				// Extract up to the next period or end of string
				if idx := strings.Index(seriesInfo, "."); idx > 0 {
					seriesInfo = strings.TrimSpace(seriesInfo[:idx])
				}
				if isValidSeries(seriesInfo) {
					metadata.Series = []string{seriesInfo}
					return true
				}
			}
		}
	}

	// Check for content metadata (Calibre sometimes puts series info in content)
	content, err := book.Metadata("content")
	if err == nil && len(content) > 0 && content[0] != "" {
		if strings.Contains(content[0], "Series") {
			parts := strings.Split(content[0], "Series")
			if len(parts) > 1 {
				seriesInfo := strings.TrimSpace(parts[1])
				// Remove leading colon or other punctuation
				seriesInfo = strings.TrimLeft(seriesInfo, ":- ")
				if isValidSeries(seriesInfo) {
					metadata.Series = []string{seriesInfo}
					return true
				}
			}
		}
	}

	// Check for meta tags that might contain series info
	// This is where Calibre often stores series information
	meta, err := book.Metadata("meta")
	if err == nil && len(meta) > 0 {
		// Get the attributes for each meta tag
		attrs, err := book.MetadataAttr("meta")
		if err == nil && len(attrs) > 0 {
			for _, attr := range attrs {
				// Check for Calibre series metadata
				if name, ok := attr["name"]; ok && name == "calibre:series" {
					if content, ok := attr["content"]; ok && content != "" {
						if isValidSeries(content) {
							metadata.Series = []string{content}
							return true
						}
					}
				}

				// Check for generic series metadata
				if name, ok := attr["name"]; ok && (name == "series" || strings.Contains(name, "series")) {
					if content, ok := attr["content"]; ok && content != "" {
						if isValidSeries(content) {
							metadata.Series = []string{content}
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// trySubjectFields checks subject fields for potential series information
func (p *EPUBMetadataProvider) trySubjectFields(book *epubgo.Epub, metadata *Metadata) bool {
	subjects, err := book.Metadata("subject")
	if err == nil && len(subjects) > 0 {
		for _, subject := range subjects {
			// Look for subjects that might be series names
			if strings.Contains(strings.ToLower(subject), "series") {
				metadata.Series = []string{subject}
				return true
			}
		}
	}

	return false
}

// extractSeriesFromTitle attempts to extract series information from a book title
func extractSeriesFromTitle(title string) string {
	// Common patterns for series in titles:
	// - "Series Name, Book X"
	// - "Series Name #X"
	// - "Title (Series Name, Book X)"

	// Check for parentheses pattern
	if strings.Contains(title, "(") && strings.Contains(title, ")") {
		start := strings.LastIndex(title, "(")
		end := strings.LastIndex(title, ")")
		if start < end {
			potentialSeries := title[start+1 : end]
			if strings.Contains(potentialSeries, "Book") ||
				strings.Contains(potentialSeries, "#") ||
				strings.Contains(potentialSeries, "Series") {
				return potentialSeries
			}
		}
	}

	// Check for comma followed by "Book" pattern
	if strings.Contains(title, ",") {
		parts := strings.Split(title, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if strings.HasPrefix(strings.ToLower(trimmed), "book") ||
				strings.Contains(trimmed, "#") {
				// This is likely a book number indicator, so the series is the part before this
				return strings.TrimSpace(parts[0])
			}
		}
	}

	return ""
}

// extractCalibreSeriesFromOPF extracts Calibre series metadata directly from the OPF file
// Returns the series name and true if found, or empty string and false if not found
func extractCalibreSeriesFromOPF(epubPath string) (string, bool) {
	// Open the EPUB file as a ZIP archive
	readCloser, err := zip.OpenReader(epubPath)
	if err != nil {
		return "", false
	}
	defer readCloser.Close()

	// Find the OPF file
	var opfFile *zip.File
	for _, file := range readCloser.File {
		if strings.HasSuffix(file.Name, ".opf") {
			opfFile = file
			break
		}
	}

	if opfFile == nil {
		return "", false
	}

	// Open and read the OPF file
	opfReader, err := opfFile.Open()
	if err != nil {
		return "", false
	}
	defer opfReader.Close()

	opfContent, err := io.ReadAll(opfReader)
	if err != nil {
		return "", false
	}

	// Convert to string for easier debugging
	opfString := string(opfContent)

	// Look for EPUB 3.0 belongs-to-collection metadata (this is what we see in your OPF)
	// Format: <meta property="belongs-to-collection" id="id-3">Test Books</meta>
	collectionRegex := regexp.MustCompile(`<meta\s+property="belongs-to-collection"[^>]*>([^<]+)</meta>`)
	matches := collectionRegex.FindStringSubmatch(opfString)
	if len(matches) > 1 {
		if IsVerboseMode() {
			color.Green("📚 Found series '%s' in belongs-to-collection metadata", matches[1])
		}
		return matches[1], true
	}

	// Look for Calibre series metadata
	// Format: <meta name="calibre:series" content="Series Name"/>
	calibreSeriesRegex := regexp.MustCompile(`<meta\s+name="calibre:series"\s+content="([^"]+)"\s*/?>`)
	matches = calibreSeriesRegex.FindStringSubmatch(opfString)
	if len(matches) > 1 {
		if IsVerboseMode() {
			color.Green("📚 Found series '%s' in calibre:series metadata", matches[1])
		}
		return matches[1], true
	}

	// If we're in verbose mode, print a snippet of the OPF for debugging
	if IsVerboseMode() {
		color.Yellow("⚠️ No series metadata found in OPF file")
		// Print a small snippet of the OPF file to help with debugging
		if len(opfString) > 200 {
			color.White("OPF snippet: %s...", opfString[:200])
		} else {
			color.White("OPF snippet: %s", opfString)
		}
	}

	return "", false
}

// IsVerboseMode returns true if verbose mode is enabled
// This is a helper function to avoid passing the verbose flag around
var isVerboseMode bool

// SetVerboseMode sets the verbose mode flag
func SetVerboseMode(verbose bool) {
	isVerboseMode = verbose
}

// IsVerboseMode returns true if verbose mode is enabled
func IsVerboseMode() bool {
	return isVerboseMode
}

// AudioMetadataProvider extracts metadata from audio files (MP3, M4B, M4A)
type AudioMetadataProvider struct {
	path string
}

func NewAudioMetadataProvider(path string) *AudioMetadataProvider {
	return &AudioMetadataProvider{path: path}
}

func (p *AudioMetadataProvider) GetMetadata() (Metadata, error) {
	f, err := os.Open(p.path)
	if err != nil {
		return Metadata{}, fmt.Errorf("error opening audio file: %v", err)
	}
	defer f.Close()

	// Use github.com/dhowden/tag for reading audio metadata
	m, err := tag.ReadFrom(f)
	if err != nil {
		return Metadata{}, fmt.Errorf("error reading audio metadata: %v", err)
	}

	return Metadata{
		Title:   m.Title(),
		Authors: []string{m.Artist()},
		Series:  []string{m.Album()},
		// You can add more fields as needed
	}, nil
}

// FileMetadataProvider auto-detects file type and delegates to the appropriate provider
type FileMetadataProvider struct {
	path string
}

func NewFileMetadataProvider(path string) *FileMetadataProvider {
	return &FileMetadataProvider{path: path}
}

func (p *FileMetadataProvider) GetMetadata() (Metadata, error) {
	ext := strings.ToLower(filepath.Ext(p.path))
	switch ext {
	case ".epub":
		return NewEPUBMetadataProvider(p.path).GetMetadata()
	case ".json":
		return NewJSONMetadataProvider(p.path).GetMetadata()
	case ".mp3", ".m4b", ".m4a":
		return NewAudioMetadataProvider(p.path).GetMetadata()
	// Add more formats as needed
	default:
		return Metadata{}, fmt.Errorf("unsupported file type: %s", ext)
	}
}
