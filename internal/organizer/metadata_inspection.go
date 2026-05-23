package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MetadataInspectionConfig controls non-interactive metadata inspection.
type MetadataInspectionConfig struct {
	UseEmbeddedMetadata bool
	FieldMapping        FieldMapping
}

// MetadataInspectionOutput contains metadata inspection results and summary data.
type MetadataInspectionOutput struct {
	Files   []MetadataInspectionFile  `json:"files"`
	Summary MetadataInspectionSummary `json:"summary"`
}

// MetadataInspectionFile contains the extracted metadata for one inspected file.
type MetadataInspectionFile struct {
	Path        string                 `json:"path"`
	SourceType  string                 `json:"source_type"`
	Title       string                 `json:"title"`
	Authors     []string               `json:"authors"`
	Series      []string               `json:"series"`
	TrackNumber int                    `json:"track_number"`
	TrackTitle  string                 `json:"track_title,omitempty"`
	Album       string                 `json:"album"`
	RawData     map[string]interface{} `json:"raw_data,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    Metadata               `json:"-"`
}

// MetadataInspectionSummary contains aggregate metadata inspection counts.
type MetadataInspectionSummary struct {
	FilesScanned int `json:"files_scanned"`
	Errors       int `json:"errors"`
}

// ExtractMappedMetadata extracts metadata from a provider and applies field mapping.
func ExtractMappedMetadata(provider MetadataProvider, fieldMapping FieldMapping) (Metadata, error) {
	metadata, err := provider.GetMetadata()
	if err != nil {
		return Metadata{}, err
	}

	if !fieldMapping.IsEmpty() {
		metadata.ApplyFieldMapping(fieldMapping)
	}

	return metadata, nil
}

// InspectMetadataDirectory scans supported files under inputDir and extracts metadata per file.
func InspectMetadataDirectory(
	inputDir string,
	config MetadataInspectionConfig,
) (MetadataInspectionOutput, error) {
	info, err := os.Stat(inputDir)
	if err != nil {
		return MetadataInspectionOutput{}, fmt.Errorf(
			"error accessing metadata directory %s: %w",
			inputDir,
			err,
		)
	}
	if !info.IsDir() {
		return MetadataInspectionOutput{}, fmt.Errorf("%s is not a directory", inputDir)
	}

	output := MetadataInspectionOutput{
		Files: []MetadataInspectionFile{},
	}

	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !IsSupportedFile(filepath.Ext(path)) {
			return nil
		}

		output.Summary.FilesScanned++
		file := InspectMetadataFile(path, config)
		if file.Error != "" {
			output.Summary.Errors++
		}
		output.Files = append(output.Files, file)
		return nil
	})
	if err != nil {
		return MetadataInspectionOutput{}, fmt.Errorf(
			"error scanning metadata directory %s: %w",
			inputDir,
			err,
		)
	}

	return output, nil
}

// InspectMetadataFile extracts metadata for one supported file without aborting on extraction errors.
func InspectMetadataFile(path string, config MetadataInspectionConfig) MetadataInspectionFile {
	file := MetadataInspectionFile{
		Path:       path,
		SourceType: MetadataSourceTypeForPath(path),
		Authors:    []string{},
		Series:     []string{},
	}

	provider := NewMetadataProvider(path, config.UseEmbeddedMetadata)
	metadata, err := ExtractMappedMetadata(provider, config.FieldMapping)
	if err != nil {
		file.Error = fmt.Sprintf("failed to extract metadata: %v", err)
		return file
	}

	file.SourceType = metadata.SourceType
	file.Title = metadata.Title
	file.Authors = nonNilMetadataStrings(metadata.Authors)
	file.Series = nonNilMetadataStrings(metadata.Series)
	file.TrackNumber = metadata.TrackNumber
	file.TrackTitle = metadata.TrackTitle
	file.Album = metadata.Album
	file.RawData = metadata.RawData
	file.Metadata = metadata
	return file
}

// MetadataSourceTypeForPath returns the coarse metadata source type implied by a file path.
func MetadataSourceTypeForPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".epub":
		return "epub"
	case ".mp3", ".m4b", ".m4a", ".ogg", ".flac":
		return "audio"
	default:
		return "unknown"
	}
}

func nonNilMetadataStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
