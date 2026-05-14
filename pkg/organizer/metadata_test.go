package organizer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewMetadataProvider(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		useEmbeddedOnly bool
	}{
		{
			name:            "audio file with embedded metadata",
			path:            "test.mp3",
			useEmbeddedOnly: true,
		},
		{
			name:            "directory with json metadata",
			path:            "/path/to/book",
			useEmbeddedOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewMetadataProvider(tt.path, tt.useEmbeddedOnly)
			if provider == nil {
				t.Error("NewMetadataProvider returned nil")
			}
			if provider.provider == nil {
				t.Error("Internal provider is nil")
			}
		})
	}
}

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"mp3 file", "book.mp3", "audio"},
		{"m4b file", "book.m4b", "audio"},
		{"m4a file", "book.m4a", "audio"},
		{"ogg file", "book.ogg", "audio"},
		{"flac file", "book.flac", "audio"},
		{"epub file", "book.epub", "epub"},
		{"json file", "metadata.json", "json"},
		{"unknown file", "book.txt", "unknown"},
		{"no extension", "book", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectFileType(tt.path)
			if result != tt.expected {
				t.Errorf("DetectFileType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCleanSeriesName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "series with number",
			input:    "The Stormlight Archive #1",
			expected: "The Stormlight Archive",
		},
		{
			name:     "series without number",
			input:    "The Stormlight Archive",
			expected: "The Stormlight Archive",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanSeriesName(tt.input)
			if result != tt.expected {
				t.Errorf("CleanSeriesName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatAuthorName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		format   AuthorFormat
		expected string
	}{
		{
			name:     "first-last to first-last",
			input:    "Brandon Sanderson",
			format:   AuthorFormatFirstLast,
			expected: "Brandon Sanderson",
		},
		{
			name:     "last-first to first-last",
			input:    "Sanderson, Brandon",
			format:   AuthorFormatFirstLast,
			expected: "Brandon Sanderson",
		},
		{
			name:     "first-last to last-first",
			input:    "Brandon Sanderson",
			format:   AuthorFormatLastFirst,
			expected: "Sanderson, Brandon",
		},
		{
			name:     "preserve format",
			input:    "Brandon Sanderson",
			format:   AuthorFormatPreserve,
			expected: "Brandon Sanderson",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAuthorName(tt.input, tt.format)
			if result != tt.expected {
				t.Errorf(
					"FormatAuthorName(%q, %v) = %q, want %q",
					tt.input,
					tt.format,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestConvertAuthorToFirstLast(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "last-first format",
			input:    "Sanderson, Brandon",
			expected: "Brandon Sanderson",
		},
		{
			name:     "already first-last",
			input:    "Brandon Sanderson",
			expected: "Brandon Sanderson",
		},
		{
			name:     "single name",
			input:    "Sanderson",
			expected: "Sanderson",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertAuthorToFirstLast(tt.input)
			if result != tt.expected {
				t.Errorf(
					"ConvertAuthorToFirstLast(%q) = %q, want %q",
					tt.input,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestConvertAuthorToLastFirst(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "first-last format",
			input:    "Brandon Sanderson",
			expected: "Sanderson, Brandon",
		},
		{
			name:     "already last-first",
			input:    "Sanderson, Brandon",
			expected: "Sanderson, Brandon",
		},
		{
			name:     "single name",
			input:    "Sanderson",
			expected: "Sanderson",
		},
		{
			name:     "multiple middle names",
			input:    "Brandon von Sanderson",
			expected: "Sanderson, Brandon von",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertAuthorToLastFirst(tt.input)
			if result != tt.expected {
				t.Errorf(
					"ConvertAuthorToLastFirst(%q) = %q, want %q",
					tt.input,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestDetectAuthorFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected AuthorFormat
	}{
		{
			name:     "first-last format",
			input:    "Brandon Sanderson",
			expected: AuthorFormatFirstLast,
		},
		{
			name:     "last-first format",
			input:    "Sanderson, Brandon",
			expected: AuthorFormatLastFirst,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectAuthorFormat(tt.input)
			if result != tt.expected {
				t.Errorf("DetectAuthorFormat(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateMetadata(t *testing.T) {
	tests := []struct {
		name      string
		metadata  Metadata
		wantError bool
	}{
		{
			name: "valid metadata",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
			},
			wantError: false,
		},
		{
			name: "missing title",
			metadata: Metadata{
				Authors: []string{"Test Author"},
			},
			wantError: true,
		},
		{
			name: "missing authors",
			metadata: Metadata{
				Title: "Test Book",
			},
			wantError: true,
		},
		{
			name: "empty author",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{""},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetadata(tt.metadata)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMetadata() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestIsMetadataValid(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		expected bool
	}{
		{
			name: "valid metadata",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
			},
			expected: true,
		},
		{
			name: "missing title",
			metadata: Metadata{
				Authors: []string{"Test Author"},
			},
			expected: false,
		},
		{
			name: "missing authors",
			metadata: Metadata{
				Title: "Test Book",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMetadataValid(tt.metadata)
			if result != tt.expected {
				t.Errorf("IsMetadataValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractMetadataWithMapping(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "metadata.json")

	jsonContent := `{
		"title": "Test Book",
		"authors": ["Test Author"],
		"album": "Test Album"
	}`

	err := os.WriteFile(testFile, []byte(jsonContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name            string
		path            string
		useEmbeddedOnly bool
		mapping         FieldMapping
		wantError       bool
	}{
		{
			name:            "extract with default mapping",
			path:            testFile,
			useEmbeddedOnly: false,
			mapping:         DefaultFieldMapping(),
			wantError:       false,
		},
		{
			name:            "extract with empty mapping",
			path:            testFile,
			useEmbeddedOnly: false,
			mapping:         FieldMapping{},
			wantError:       false,
		},
		{
			name:            "extract from non-existent file",
			path:            filepath.Join(tmpDir, "nonexistent.json"),
			useEmbeddedOnly: false,
			mapping:         DefaultFieldMapping(),
			wantError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := ExtractMetadataWithMapping(tt.path, tt.useEmbeddedOnly, tt.mapping)
			if (err != nil) != tt.wantError {
				t.Errorf("ExtractMetadataWithMapping() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && metadata.Title == "" {
				t.Error("Expected metadata to have a title")
			}
		})
	}
}

func TestGetSeriesNumberFromMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		expected string
	}{
		{
			name: "series with number",
			metadata: Metadata{
				Series: []string{"The Stormlight Archive #1"},
			},
			expected: "1",
		},
		{
			name: "series without number",
			metadata: Metadata{
				Series: []string{"The Stormlight Archive"},
			},
			expected: "",
		},
		{
			name:     "no series",
			metadata: Metadata{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSeriesNumberFromMetadata(tt.metadata)
			if result != tt.expected {
				t.Errorf("GetSeriesNumberFromMetadata() = %q, want %q", result, tt.expected)
			}
		})
	}
}
