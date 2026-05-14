package app

import (
	"context"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// Service coordinates app use cases for CLI, TUI, and web transports.
type Service struct {
	config WebConfig
}

// NewService creates a reusable app service.
func NewService(config WebConfig) *Service {
	config.ABS = config.ABS.Redacted()
	return &Service{config: config}
}

// Config returns the redacted service configuration.
func (s *Service) Config() WebConfig {
	return s.config
}

// Option describes a selectable UI option.
type Option struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// OptionsResponse contains frontend bootstrap options.
type OptionsResponse struct {
	Layouts       []Option                   `json:"layouts"`
	ScanModes     []Option                   `json:"scan_modes"`
	AuthorFormats []Option                   `json:"author_formats"`
	FieldMappings map[string]FieldMappingDTO `json:"field_mappings"`
}

// Options returns static organizer option metadata for the UI.
func (s *Service) Options(_ context.Context) OptionsResponse {
	return OptionsResponse{
		Layouts: []Option{
			{Value: "author-series-title", Label: "Author / Series / Title"},
			{Value: "author-series-title-number", Label: "Author / Series / # - Title"},
			{Value: "author-series", Label: "Author / Series"},
			{Value: "author-title", Label: "Author / Title"},
			{Value: "author-only", Label: "Author only"},
			{Value: "series-title", Label: "Series / Title"},
			{Value: "series-title-number", Label: "Series / # - Title"},
		},
		ScanModes: []Option{
			{Value: "json", Label: "metadata.json"},
			{Value: "embedded-directory", Label: "Embedded metadata by directory"},
			{Value: "embedded-file", Label: "Embedded metadata by file"},
			{Value: "abs", Label: "Audiobookshelf metadata"},
		},
		AuthorFormats: []Option{
			{Value: "first-last", Label: "First Last"},
			{Value: "last-first", Label: "Last, First"},
			{Value: "preserve", Label: "Preserve source"},
		},
		FieldMappings: map[string]FieldMappingDTO{
			"default": FieldMappingFromOrganizer(organizer.DefaultFieldMapping()),
			"audio":   FieldMappingFromOrganizer(organizer.AudioFieldMapping()),
			"epub":    FieldMappingFromOrganizer(organizer.EpubFieldMapping()),
		},
	}
}
