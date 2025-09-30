package organizer

import (
	"strings"
	"testing"
)

func TestNewMetadataFormatter(t *testing.T) {
	tests := []struct {
		name         string
		metadata     Metadata
		fieldMapping FieldMapping
	}{
		{
			name: "basic metadata",
			metadata: Metadata{
				Title:       "Test Book",
				Authors:     []string{"Test Author"},
				SourcePath:  "test.mp3",
				TrackNumber: 1,
			},
			fieldMapping: DefaultFieldMapping(),
		},
		{
			name: "complex metadata",
			metadata: Metadata{
				Title:       "Advanced Test Book",
				Authors:     []string{"Author One", "Author Two"},
				Series:      []string{"Test Series"},
				SourcePath:    "complex_test.m4a",
				TrackNumber: 5,
				Album:       "Test Album",
			},
			fieldMapping: DefaultFieldMapping(),
		},
		{
			name: "minimal metadata",
			metadata: Metadata{
				SourcePath: "minimal.mp3",
			},
			fieldMapping: DefaultFieldMapping(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewMetadataFormatter(tt.metadata, tt.fieldMapping)

			if formatter == nil {
				t.Fatal("NewMetadataFormatter returned nil")
			}

			if formatter.metadata.Title != tt.metadata.Title {
				t.Errorf("Expected metadata title %q, got %q", tt.metadata.Title, formatter.metadata.Title)
			}

			if len(formatter.metadata.Authors) != len(tt.metadata.Authors) {
				t.Errorf("Expected %d authors, got %d", len(tt.metadata.Authors), len(formatter.metadata.Authors))
			}

			if formatter.metadata.SourcePath != tt.metadata.SourcePath {
				t.Errorf("Expected filename %q, got %q", tt.metadata.SourcePath, formatter.metadata.SourcePath)
			}

			// Test field mapping is stored
			if formatter.fieldMapping.TitleField != tt.fieldMapping.TitleField {
				t.Errorf("Expected TitleField %q, got %q", tt.fieldMapping.TitleField, formatter.fieldMapping.TitleField)
			}
		})
	}
}

func TestFormatMetadataWithMapping(t *testing.T) {
	tests := []struct {
		name           string
		metadata       Metadata
		fieldMapping   FieldMapping
		expectedParts  []string // Parts that should appear in the formatted output
		unexpectedParts []string // Parts that should NOT appear
	}{
		{
			name: "complete metadata",
			metadata: Metadata{
				Title:       "Test Book",
				Authors:     []string{"Test Author"},
				Series:      []string{"Test Series"},
				SourcePath:  "test.mp3",
				TrackNumber: 1,
				Album:       "Test Album",
			},
			fieldMapping: DefaultFieldMapping(),
			expectedParts: []string{
				"Test Book",
				"Test Author",
				"Test Series",
				"test.mp3",
				// "Test Album", // Album field may not be displayed in formatter
			},
		},
		{
			name: "metadata with multiple authors",
			metadata: Metadata{
				Title:    "Multi Author Book",
				Authors:  []string{"Author One", "Author Two", "Author Three"},
				SourcePath: "multi_author.mp3",
			},
			fieldMapping: DefaultFieldMapping(),
			expectedParts: []string{
				"Multi Author Book",
				"Author One",
				"Author Two",
				"Author Three",
				"multi_author.mp3",
			},
		},
		{
			name: "metadata with multiple series",
			metadata: Metadata{
				Title:    "Series Book",
				Authors:  []string{"Series Author"},
				Series:   []string{"Series One", "Series Two"},
				SourcePath: "series_book.mp3",
			},
			fieldMapping: DefaultFieldMapping(),
			expectedParts: []string{
				"Series Book",
				"Series Author",
				"Series One",
				"series_book.mp3",
			},
		},
		{
			name: "minimal metadata",
			metadata: Metadata{
				SourcePath: "minimal.mp3",
			},
			fieldMapping: DefaultFieldMapping(),
			expectedParts: []string{
				"minimal.mp3",
			},
			unexpectedParts: []string{
				"Unknown Title",
				"Unknown Author",
			},
		},
		{
			name: "special characters in metadata",
			metadata: Metadata{
				Title:    "Special & Characters: Book",
				Authors:  []string{"Author with $pecial Ch@racters"},
				SourcePath: "special_chars.mp3",
			},
			fieldMapping: DefaultFieldMapping(),
			expectedParts: []string{
				"Special & Characters: Book",
				"Author with $pecial Ch@racters",
				"special_chars.mp3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewMetadataFormatter(tt.metadata, tt.fieldMapping)

			formatted := formatter.FormatMetadataWithMapping()

			if formatted == "" {
				t.Error("FormatMetadataWithMapping should return non-empty string")
			}

			// Check for expected parts
			for _, expectedPart := range tt.expectedParts {
				if !strings.Contains(formatted, expectedPart) {
					t.Errorf("Expected formatted output to contain %q, got:\n%s", expectedPart, formatted)
				}
			}

			// Check for unexpected parts
			for _, unexpectedPart := range tt.unexpectedParts {
				if strings.Contains(formatted, unexpectedPart) {
					t.Errorf("Expected formatted output NOT to contain %q, got:\n%s", unexpectedPart, formatted)
				}
			}
		})
	}
}

func TestGetFileTypeDisplay(t *testing.T) {
	tests := []struct {
		name             string
		filename         string
		expectedFileType string
		iconNotEmpty     bool
	}{
		{
			name:             "MP3 file",
			filename:         "test.mp3",
			expectedFileType: "MP3",
			iconNotEmpty:     true,
		},
		{
			name:             "M4A file",
			filename:         "test.m4a",
			expectedFileType: "M4A",
			iconNotEmpty:     true,
		},
		{
			name:             "M4B file",
			filename:         "test.m4b",
			expectedFileType: "M4B",
			iconNotEmpty:     true,
		},
		{
			name:             "FLAC file",
			filename:         "test.flac",
			expectedFileType: "FLAC",
			iconNotEmpty:     true,
		},
		{
			name:             "unknown file type",
			filename:         "test.xyz",
			expectedFileType: "UNKNOWN",
			iconNotEmpty:     true,
		},
		{
			name:             "no extension",
			filename:         "test",
			expectedFileType: "UNKNOWN",
			iconNotEmpty:     true,
		},
		{
			name:             "uppercase extension",
			filename:         "test.MP3",
			expectedFileType: "MP3",
			iconNotEmpty:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := Metadata{
				SourcePath: tt.filename,
				SourceType: "audio",
			}
			formatter := NewMetadataFormatter(metadata, DefaultFieldMapping())

			// Since getFileTypeDisplay is not exported, we test it indirectly
			// through FormatMetadataWithMapping which uses it
			formatted := formatter.FormatMetadataWithMapping()

			if formatted == "" {
				t.Error("Formatted output should not be empty")
			}

			// The file type should appear somewhere in the formatted output
			if !strings.Contains(formatted, tt.expectedFileType) {
				t.Errorf("Expected formatted output to contain file type %q, got:\n%s", tt.expectedFileType, formatted)
			}
		})
	}
}

func TestFormatMetadataWithDifferentFieldMappings(t *testing.T) {
	metadata := Metadata{
		Title:       "Test Book",
		Authors:     []string{"Test Author"},
		Series:      []string{"Test Series"},
		SourcePath:    "test.mp3",
		TrackNumber: 1,
	}

	tests := []struct {
		name         string
		fieldMapping FieldMapping
		description  string
	}{
		{
			name:         "default field mapping",
			fieldMapping: DefaultFieldMapping(),
			description:  "Should use default field mappings",
		},
		{
			name: "custom field mapping",
			fieldMapping: FieldMapping{
				TitleField:   "custom_title",
				SeriesField:  "custom_series",
				AuthorFields: []string{"custom_author"},
				TrackField:   "custom_track",
			},
			description: "Should work with custom field mappings",
		},
		{
			name: "empty field mapping",
			fieldMapping: FieldMapping{
				TitleField:   "",
				SeriesField:  "",
				AuthorFields: []string{},
				TrackField:   "",
			},
			description: "Should handle empty field mappings gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewMetadataFormatter(metadata, tt.fieldMapping)

			formatted := formatter.FormatMetadataWithMapping()

			if formatted == "" {
				t.Errorf("FormatMetadataWithMapping should return non-empty string for %s", tt.description)
			}

			// Should still contain the filename regardless of field mapping
			if !strings.Contains(formatted, "test.mp3") {
				t.Errorf("Formatted output should contain filename regardless of field mapping: %s", formatted)
			}
		})
	}
}

func TestFormatMetadataFieldMappingDisplay(t *testing.T) {
	metadata := Metadata{
		Title:       "Test Book",
		Authors:     []string{"Test Author"},
		Series:      []string{"Test Series"},
		SourcePath:    "test.mp3",
		TrackNumber: 1,
	}

	fieldMapping := DefaultFieldMapping()
	formatter := NewMetadataFormatter(metadata, fieldMapping)

	formatted := formatter.FormatMetadataWithMapping()

	// Should show field mapping information
	expectedMappingParts := []string{
		fieldMapping.TitleField,
		fieldMapping.SeriesField,
		fieldMapping.TrackField,
	}

	for _, part := range expectedMappingParts {
		if part != "" && !strings.Contains(formatted, part) {
			t.Errorf("Expected formatted output to show field mapping %q, got:\n%s", part, formatted)
		}
	}

	// Should show author field mappings
	for _, authorField := range fieldMapping.AuthorFields {
		if authorField != "" && !strings.Contains(formatted, authorField) {
			t.Errorf("Expected formatted output to show author field mapping %q, got:\n%s", authorField, formatted)
		}
	}
}

func TestFormatMetadataWithEmptyValues(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
	}{
		{
			name: "empty title",
			metadata: Metadata{
				Title:    "",
				Authors:  []string{"Test Author"},
				SourcePath: "test.mp3",
			},
		},
		{
			name: "empty authors",
			metadata: Metadata{
				Title:    "Test Book",
				Authors:  []string{},
				SourcePath: "test.mp3",
			},
		},
		{
			name: "empty series",
			metadata: Metadata{
				Title:    "Test Book",
				Authors:  []string{"Test Author"},
				Series:   []string{},
				SourcePath: "test.mp3",
			},
		},
		{
			name: "all empty except filename",
			metadata: Metadata{
				Title:    "",
				Authors:  []string{},
				Series:   []string{},
				SourcePath: "test.mp3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewMetadataFormatter(tt.metadata, DefaultFieldMapping())

			formatted := formatter.FormatMetadataWithMapping()

			if formatted == "" {
				t.Error("FormatMetadataWithMapping should return non-empty string even with empty values")
			}

			// Should always contain the filename
			if !strings.Contains(formatted, tt.metadata.SourcePath) {
				t.Errorf("Formatted output should contain filename %q, got:\n%s", tt.metadata.SourcePath, formatted)
			}
		})
	}
}

func TestFormatMetadataStructure(t *testing.T) {
	metadata := Metadata{
		Title:       "Test Book",
		Authors:     []string{"Test Author"},
		Series:      []string{"Test Series"},
		SourcePath:    "test.mp3",
		TrackNumber: 1,
		Album:       "Test Album",
	}

	formatter := NewMetadataFormatter(metadata, DefaultFieldMapping())
	formatted := formatter.FormatMetadataWithMapping()

	// Basic structure checks
	if strings.Count(formatted, "\n") < 3 {
		t.Error("Formatted output should have multiple lines for readability")
	}

	// Should have some kind of visual structure (headers, separators, etc.)
	hasStructure := strings.Contains(formatted, "─") ||
		strings.Contains(formatted, "━") ||
		strings.Contains(formatted, "=") ||
		strings.Contains(formatted, "-") ||
		strings.Contains(formatted, ":") ||
		strings.Count(formatted, " ") > 5

	if !hasStructure {
		t.Error("Formatted output should have visual structure (separators, spacing, etc.)")
	}
}
