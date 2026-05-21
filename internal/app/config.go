package app

import (
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/abs"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// WebConfig contains the initial local web UI configuration.
type WebConfig struct {
	Host      string             `json:"host"`
	Port      int                `json:"port"`
	Open      bool               `json:"open"`
	Initial   InitialConfigDTO   `json:"initial"`
	Organizer OrganizerConfigDTO `json:"organizer"`
	Rename    RenameConfigDTO    `json:"rename"`
	ABS       ABSConfigDTO       `json:"abs"`
}

// InitialConfigDTO contains initial path values passed from the CLI.
type InitialConfigDTO struct {
	InputDir  string `json:"input_dir"`
	OutputDir string `json:"output_dir"`
}

// OrganizerConfigDTO is the JSON-safe organization configuration.
type OrganizerConfigDTO struct {
	BaseDir             string          `json:"base_dir"`
	OutputDir           string          `json:"output_dir"`
	ReplaceSpace        string          `json:"replace_space"`
	DryRun              bool            `json:"dry_run"`
	RemoveEmpty         bool            `json:"remove_empty"`
	UseEmbeddedMetadata bool            `json:"use_embedded_metadata"`
	Flat                bool            `json:"flat"`
	SkipErrors          bool            `json:"skip_errors"`
	Layout              string          `json:"layout"`
	LayoutTemplate      string          `json:"layout_template"`
	AuthorFormat        string          `json:"author_format"`
	FieldMapping        FieldMappingDTO `json:"field_mapping"`
	AllowedSourcePaths  []string        `json:"allowed_source_paths,omitempty"`
}

// RenameConfigDTO is the JSON-safe rename configuration.
type RenameConfigDTO struct {
	BaseDir             string          `json:"base_dir"`
	Template            string          `json:"template"`
	DryRun              bool            `json:"dry_run"`
	AuthorFormat        string          `json:"author_format"`
	Recursive           bool            `json:"recursive"`
	FieldMapping        FieldMappingDTO `json:"field_mapping"`
	ReplaceSpace        string          `json:"replace_space"`
	StrictMode          bool            `json:"strict_mode"`
	PreservePath        bool            `json:"preserve_path"`
	UseEmbeddedMetadata bool            `json:"use_embedded_metadata"`
	AllowedCurrentPaths []string        `json:"allowed_current_paths,omitempty"`
}

// FieldMappingDTO is the JSON-safe metadata field mapping.
type FieldMappingDTO struct {
	TitleField   string   `json:"title_field,omitempty"`
	SeriesField  string   `json:"series_field,omitempty"`
	AuthorFields []string `json:"author_fields,omitempty"`
	TrackField   string   `json:"track_field,omitempty"`
	DiscField    string   `json:"disc_field,omitempty"`
}

// ABSConfigDTO is the JSON-safe Audiobookshelf configuration.
type ABSConfigDTO struct {
	URL          string           `json:"url"`
	Token        string           `json:"token,omitempty"`
	LibraryID    string           `json:"library_id"`
	SQLitePath   string           `json:"sqlite_path,omitempty"`
	PathMappings []PathMappingDTO `json:"path_mappings,omitempty"`
	AllLibraries bool             `json:"all_libraries"`
	HeaderFile   string           `json:"header_file,omitempty"`
	Headers      []HeaderDTO      `json:"headers,omitempty"`
}

// PathMappingDTO maps Audiobookshelf paths to local filesystem paths.
type PathMappingDTO struct {
	ABSPrefix   string `json:"abs_prefix"`
	LocalPrefix string `json:"local_prefix"`
}

// HeaderDTO represents a custom ABS proxy/auth header.
type HeaderDTO struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

// Redacted returns a copy safe to send to the browser.
func (d ABSConfigDTO) Redacted() ABSConfigDTO {
	if d.Token != "" {
		d.Token = "redacted"
	}
	for i := range d.Headers {
		if d.Headers[i].Value != "" {
			d.Headers[i].Value = "redacted"
		}
	}
	return d
}

// ToOrganizerConfig converts DTOs to the organizer core config.
func (d OrganizerConfigDTO) ToOrganizerConfig() organizer.OrganizerConfig {
	return organizer.OrganizerConfig{
		BaseDir:             d.BaseDir,
		OutputDir:           d.OutputDir,
		ReplaceSpace:        d.ReplaceSpace,
		DryRun:              d.DryRun,
		RemoveEmpty:         d.RemoveEmpty,
		UseEmbeddedMetadata: d.UseEmbeddedMetadata,
		Flat:                d.Flat,
		SkipErrors:          d.SkipErrors,
		Layout:              d.Layout,
		LayoutTemplate:      d.LayoutTemplate,
		AuthorFormat:        d.AuthorFormat,
		FieldMapping:        d.FieldMapping.ToFieldMapping(),
		AllowedSourcePaths:  d.AllowedSourcePaths,
	}
}

// ToRenamerConfig converts DTOs to the rename core config.
func (d RenameConfigDTO) ToRenamerConfig() organizer.RenamerConfig {
	return organizer.RenamerConfig{
		BaseDir:             d.BaseDir,
		Template:            d.Template,
		DryRun:              d.DryRun,
		AuthorFormat:        ParseAuthorFormat(d.AuthorFormat),
		Recursive:           d.Recursive,
		FieldMapping:        d.FieldMapping.ToFieldMapping(),
		ReplaceSpace:        d.ReplaceSpace,
		StrictMode:          d.StrictMode,
		PreservePath:        d.PreservePath,
		UseEmbeddedMetadata: d.UseEmbeddedMetadata,
		AllowedCurrentPaths: d.AllowedCurrentPaths,
	}
}

// ToFieldMapping converts DTOs to the organizer field mapping.
func (d FieldMappingDTO) ToFieldMapping() organizer.FieldMapping {
	return organizer.FieldMapping{
		TitleField:   d.TitleField,
		SeriesField:  d.SeriesField,
		AuthorFields: d.AuthorFields,
		TrackField:   d.TrackField,
		DiscField:    d.DiscField,
	}
}

// ToPathMappings converts DTOs to ABS path mappings.
func (d ABSConfigDTO) ToPathMappings() []abs.PathMapping {
	mappings := make([]abs.PathMapping, 0, len(d.PathMappings))
	for _, mapping := range d.PathMappings {
		if mapping.ABSPrefix == "" && mapping.LocalPrefix == "" {
			continue
		}
		mappings = append(mappings, abs.PathMapping{
			ABSPrefix:   mapping.ABSPrefix,
			LocalPrefix: mapping.LocalPrefix,
		})
	}
	return mappings
}

// ParseAuthorFormat converts a user-facing string into the organizer enum.
func ParseAuthorFormat(value string) organizer.AuthorFormat {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "last-first":
		return organizer.AuthorFormatLastFirst
	case "preserve":
		return organizer.AuthorFormatPreserve
	default:
		return organizer.AuthorFormatFirstLast
	}
}

// DefaultWebConfig returns conservative defaults for the local web UI.
func DefaultWebConfig(host string, port int, open bool, inputDir, outputDir string) WebConfig {
	return WebConfig{
		Host: host,
		Port: port,
		Open: open,
		Initial: InitialConfigDTO{
			InputDir:  inputDir,
			OutputDir: outputDir,
		},
		Organizer: OrganizerConfigDTO{
			BaseDir:             inputDir,
			OutputDir:           outputDir,
			DryRun:              true,
			Layout:              "author-series-title",
			LayoutTemplate:      "",
			UseEmbeddedMetadata: false,
			FieldMapping:        FieldMappingFromOrganizer(organizer.DefaultFieldMapping()),
		},
		Rename: RenameConfigDTO{
			BaseDir:             inputDir,
			Template:            "{author} - {series} {series_number} - {title}",
			DryRun:              true,
			AuthorFormat:        "first-last",
			Recursive:           true,
			PreservePath:        true,
			UseEmbeddedMetadata: false,
			FieldMapping:        FieldMappingFromOrganizer(organizer.DefaultFieldMapping()),
		},
		ABS: ABSConfigDTO{
			LibraryID: "main",
		},
	}
}

// FieldMappingFromOrganizer converts an organizer field mapping to a DTO.
func FieldMappingFromOrganizer(mapping organizer.FieldMapping) FieldMappingDTO {
	return FieldMappingDTO{
		TitleField:   mapping.TitleField,
		SeriesField:  mapping.SeriesField,
		AuthorFields: mapping.AuthorFields,
		TrackField:   mapping.TrackField,
		DiscField:    mapping.DiscField,
	}
}
