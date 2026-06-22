package organizer

import (
	"path/filepath"
	"testing"
)

func seriesBookMetadata() Metadata {
	return Metadata{
		Title:   "Book Title",
		Authors: []string{"Test Author"},
		Series:  []string{"Test Series #2"},
		RawData: map[string]interface{}{
			"series_index": 2.0,
			"narrator":     "Test Narrator",
		},
	}
}

func standaloneBookMetadata() Metadata {
	return Metadata{
		Title:   "Book Title",
		Authors: []string{"Test Author"},
		RawData: map[string]interface{}{
			"narrator": "Test Narrator",
		},
	}
}

func renderTemplate(t *testing.T, template string, metadata Metadata) string {
	t.Helper()

	tmpl, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate() error: %v", err)
	}

	renderer := NewTemplateRenderer(tmpl, NewAuthorFormatter(AuthorFormatFirstLast))
	got, err := renderer.Render(metadata)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	return got
}

func TestCompositeOptionalSegmentRender(t *testing.T) {
	tests := []struct {
		name     string
		template string
		metadata Metadata
		want     string
	}{
		{
			name:     "composite renders literal and zero-padded series number",
			template: "{Vol series_number:02}",
			metadata: seriesBookMetadata(),
			want:     "Vol 02",
		},
		{
			name:     "composite omits when series number missing",
			template: "{Vol series_number:02}",
			metadata: standaloneBookMetadata(),
			want:     "",
		},
		{
			name:     "composite omits trailing literal when field missing",
			template: "{Vol series_number:02 - }",
			metadata: standaloneBookMetadata(),
			want:     "",
		},
		{
			name:     "composite narrator brackets render when present",
			template: "{title}{ [narrator]}",
			metadata: seriesBookMetadata(),
			want:     "Book Title [Test Narrator]",
		},
		{
			name:     "composite narrator brackets omit when narrator missing",
			template: "{title}{ [narrator]}",
			metadata: Metadata{
				Title:   "Book Title",
				Authors: []string{"Test Author"},
			},
			want: "Book Title",
		},
		{
			name:     "composite trailing literal preserved before next token",
			template: "{Vol series_number:02 - }{title}",
			metadata: seriesBookMetadata(),
			want:     "Vol 02 - Book Title",
		},
		{
			name:     "simple field format renders padded value",
			template: "{series_number:02}",
			metadata: seriesBookMetadata(),
			want:     "02",
		},
		{
			name:     "simple field format renders empty when missing",
			template: "{series_number:02}",
			metadata: standaloneBookMetadata(),
			want:     "",
		},
		{
			name:     "fallback behavior unchanged",
			template: "{series|Standalone}",
			metadata: standaloneBookMetadata(),
			want:     "Standalone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderTemplate(t, tt.template, tt.metadata)
			if got != tt.want {
				t.Fatalf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateTemplateCompositeSyntax(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "rejects empty placeholder",
			template: "{}",
			wantErr:  true,
		},
		{
			name:     "accepts composite optional segment",
			template: "{Vol series_number:02 - }",
			wantErr:  false,
		},
		{
			name:     "accepts simple formatted field",
			template: "{series_number:02}",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.template)
			if tt.wantErr && err == nil {
				t.Fatal("ValidateTemplate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateTemplate() unexpected error: %v", err)
			}
		})
	}
}

func TestCustomLayoutTemplatePathSegmentOmission(t *testing.T) {
	tests := []struct {
		name     string
		template string
		metadata Metadata
		expected string
	}{
		{
			name:     "omits empty series path segment for standalone",
			template: "{author}/{series}/{title}",
			metadata: standaloneBookMetadata(),
			expected: filepath.Join("testbase", "Test Author", "Book Title"),
		},
		{
			name:     "keeps series path segment when present",
			template: "{author}/{series}/{title}",
			metadata: seriesBookMetadata(),
			expected: filepath.Join("testbase", "Test Author", "Test Series", "Book Title"),
		},
		{
			name:     "standalone fallback keeps series segment",
			template: "{author}/{series|Standalone}/{title}",
			metadata: standaloneBookMetadata(),
			expected: filepath.Join("testbase", "Test Author", "Standalone", "Book Title"),
		},
		{
			name:     "full user example with series book",
			template: "{author}/{series|Standalone}/{Vol series_number:02 - }{title}{ [narrator]}",
			metadata: seriesBookMetadata(),
			expected: filepath.Join(
				"testbase",
				"Test Author",
				"Test Series",
				"Vol 02 - Book Title [Test Narrator]",
			),
		},
		{
			name:     "full user example with standalone book",
			template: "{author}/{series|Standalone}/{Vol series_number:02 - }{title}{ [narrator]}",
			metadata: standaloneBookMetadata(),
			expected: filepath.Join(
				"testbase",
				"Test Author",
				"Standalone",
				"Book Title [Test Narrator]",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OrganizerConfig{
				BaseDir:        "testbase",
				LayoutTemplate: tt.template,
			}
			lc := NewLayoutCalculator(config, pathTestSanitizer)

			got, err := lc.CalculateTargetPathE(tt.metadata)
			if err != nil {
				t.Fatalf("CalculateTargetPathE() error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("CalculateTargetPathE() = %q, want %q", got, tt.expected)
			}
		})
	}
}
