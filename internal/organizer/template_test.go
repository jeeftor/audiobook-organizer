package organizer

import (
	"testing"
)

func TestParseTemplate(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantErr       bool
		wantNumTokens int
	}{
		{
			name:          "simple single field",
			input:         "{author}",
			wantErr:       false,
			wantNumTokens: 1,
		},
		{
			name:          "multiple fields with separators",
			input:         "{author} - {title}",
			wantErr:       false,
			wantNumTokens: 3, // author, " - ", title
		},
		{
			name:          "complex template",
			input:         "{author} - {series} #{series_number} - {title}",
			wantErr:       false,
			wantNumTokens: 7,
		},
		{
			name:          "field with fallback",
			input:         "{series|Standalone}",
			wantErr:       false,
			wantNumTokens: 1,
		},
		{
			name:          "empty template",
			input:         "",
			wantErr:       false,
			wantNumTokens: 0,
		},
		{
			name:          "literal text only",
			input:         "prefix-suffix",
			wantErr:       false,
			wantNumTokens: 1,
		},
		{
			name:          "unclosed brace",
			input:         "{author - {title}",
			wantErr:       false, // We'll be lenient - matches inner {title}
			wantNumTokens: 1,     // Matches {title}, treats "{author - " as literal wasn't captured
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTemplate() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTemplate() unexpected error: %v", err)
				return
			}

			if tmpl == nil {
				t.Fatal("ParseTemplate() returned nil template")
			}

			if len(tmpl.tokens) != tt.wantNumTokens {
				t.Errorf("ParseTemplate() got %d tokens, want %d", len(tmpl.tokens), tt.wantNumTokens)
			}
		})
	}
}

func TestTemplateRender(t *testing.T) {
	tests := []struct {
		name     string
		template string
		metadata Metadata
		want     string
		wantErr  bool
	}{
		{
			name:     "simple title",
			template: "{title}",
			metadata: Metadata{
				Title: "The Final Empire",
			},
			want:    "The Final Empire",
			wantErr: false,
		},
		{
			name:     "author and title",
			template: "{author} - {title}",
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
			},
			want:    "Brandon Sanderson - The Final Empire",
			wantErr: false,
		},
		{
			name:     "full template with series",
			template: "{author} - {series} #{series_number} - {title}",
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn #1"},
			},
			want:    "Brandon Sanderson - Mistborn #1 - The Final Empire",
			wantErr: false,
		},
		{
			name:     "missing field skipped",
			template: "{author} - {series} - {title}",
			metadata: Metadata{
				Title:   "The Way of Kings",
				Authors: []string{"Brandon Sanderson"},
			},
			want:    "Brandon Sanderson -  - The Way of Kings", // Empty space where series would be
			wantErr: false,
		},
		{
			name:     "fallback used for missing field",
			template: "{author} - {series|Standalone} - {title}",
			metadata: Metadata{
				Title:   "Warbreaker",
				Authors: []string{"Brandon Sanderson"},
			},
			want:    "Brandon Sanderson - Standalone - Warbreaker",
			wantErr: false,
		},
		{
			name:     "track number formatting",
			template: "{track} - {title}",
			metadata: Metadata{
				Title:       "Chapter One",
				TrackNumber: 1,
			},
			want:    "01 - Chapter One",
			wantErr: false,
		},
		{
			name:     "series number extraction",
			template: "{author} - {series} {series_number} - {title}",
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn #1"},
			},
			want:    "Brandon Sanderson - Mistborn 1 - The Final Empire",
			wantErr: false,
		},
		{
			name:     "multiple authors",
			template: "{authors} - {title}",
			metadata: Metadata{
				Title:   "The Talisman",
				Authors: []string{"Stephen King", "Peter Straub"},
			},
			want:    "Stephen King, Peter Straub - The Talisman",
			wantErr: false,
		},
		{
			name:     "literal text preserved",
			template: "prefix_{title}_suffix",
			metadata: Metadata{
				Title: "Test",
			},
			want:    "prefix_Test_suffix",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("ParseTemplate() error: %v", err)
			}

			renderer := NewTemplateRenderer(tmpl, NewAuthorFormatter(AuthorFormatFirstLast))
			got, err := renderer.Render(tt.metadata)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Render() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Render() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTemplateWithAuthorFormatting(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		metadata     Metadata
		authorFormat AuthorFormat
		want         string
	}{
		{
			name:     "first last format",
			template: "{author} - {title}",
			metadata: Metadata{
				Title:   "Test",
				Authors: []string{"Sanderson, Brandon"},
			},
			authorFormat: AuthorFormatFirstLast,
			want:         "Brandon Sanderson - Test",
		},
		{
			name:     "last first format",
			template: "{author} - {title}",
			metadata: Metadata{
				Title:   "Test",
				Authors: []string{"Brandon Sanderson"},
			},
			authorFormat: AuthorFormatLastFirst,
			want:         "Sanderson, Brandon - Test",
		},
		{
			name:     "preserve format",
			template: "{author} - {title}",
			metadata: Metadata{
				Title:   "Test",
				Authors: []string{"Brandon Sanderson"},
			},
			authorFormat: AuthorFormatPreserve,
			want:         "Brandon Sanderson - Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("ParseTemplate() error: %v", err)
			}

			renderer := NewTemplateRenderer(tmpl, NewAuthorFormatter(tt.authorFormat))
			got, err := renderer.Render(tt.metadata)

			if err != nil {
				t.Errorf("Render() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetAvailableFields(t *testing.T) {
	fields := GetAvailableFields()

	if len(fields) == 0 {
		t.Error("GetAvailableFields() returned empty list")
	}

	// Check for essential fields
	requiredFields := []string{"author", "title", "series", "track"}
	for _, required := range requiredFields {
		found := false
		for _, field := range fields {
			if field.Name == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetAvailableFields() missing required field: %s", required)
		}
	}

	// Verify fields have descriptions and examples
	for _, field := range fields {
		if field.Name == "" {
			t.Error("Field missing name")
		}
		if field.Description == "" {
			t.Errorf("Field %s missing description", field.Name)
		}
		if field.Example == "" {
			t.Errorf("Field %s missing example", field.Name)
		}
	}
}

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "valid simple",
			template: "{author} - {title}",
			wantErr:  false,
		},
		{
			name:     "valid complex",
			template: "{author} - {series} #{series_number} - {title}",
			wantErr:  false,
		},
		{
			name:     "empty template valid",
			template: "",
			wantErr:  false,
		},
		{
			name:     "literal only valid",
			template: "prefix-suffix",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.template)

			if tt.wantErr && err == nil {
				t.Error("ValidateTemplate() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ValidateTemplate() unexpected error: %v", err)
			}
		})
	}
}
