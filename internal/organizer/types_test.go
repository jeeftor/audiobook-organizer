package organizer

import (
	"testing"
)

func TestMetadataGetFullValidSeries(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		want     string
	}{
		{
			name: "single_series",
			metadata: Metadata{
				Series: []string{"Mistborn #1"},
			},
			want: "Mistborn #1",
		},
		{
			name: "multiple_series_sorted_alphabetically",
			metadata: Metadata{
				Series: []string{"Z Series #3", "A Series #1", "M Series #2"},
			},
			want: "A Series #1",
		},
		{
			name: "empty_series",
			metadata: Metadata{
				Series: []string{},
			},
			want: "",
		},
		{
			name: "invalid_series_marker",
			metadata: Metadata{
				Series: []string{InvalidSeriesValue},
			},
			want: "",
		},
		{
			name: "empty_string_in_series",
			metadata: Metadata{
				Series: []string{""},
			},
			want: "",
		},
		{
			name: "multiple_series_with_invalid",
			metadata: Metadata{
				Series: []string{InvalidSeriesValue, "Valid Series #1"},
			},
			want: "Valid Series #1",
		},
		{
			name: "series_without_number",
			metadata: Metadata{
				Series: []string{"The Expanse"},
			},
			want: "The Expanse",
		},
		{
			name: "multiple_series_with_numbers",
			metadata: Metadata{
				Series: []string{"Series B #2", "Series A #1", "Series C #3"},
			},
			want: "Series A #1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metadata.GetFullValidSeries()
			if got != tt.want {
				t.Errorf("GetFullValidSeries() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMetadataGetValidSeries(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		want     string
	}{
		{
			name: "series_with_number",
			metadata: Metadata{
				Series: []string{"Mistborn #1"},
			},
			want: "Mistborn",
		},
		{
			name: "series_without_number",
			metadata: Metadata{
				Series: []string{"The Expanse"},
			},
			want: "The Expanse",
		},
		{
			name: "multiple_series_sorted_and_cleaned",
			metadata: Metadata{
				Series: []string{"Z Series #3", "A Series #1"},
			},
			want: "A Series",
		},
		{
			name: "empty_series",
			metadata: Metadata{
				Series: []string{},
			},
			want: "",
		},
		{
			name: "series_with_multiple_hash_symbols",
			metadata: Metadata{
				Series: []string{"Test #Series #3"},
			},
			want: "Test #Series",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metadata.GetValidSeries()
			if got != tt.want {
				t.Errorf("GetValidSeries() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMetadataGetFirstAuthor(t *testing.T) {
	tests := []struct {
		name         string
		metadata     Metadata
		defaultValue string
		want         string
	}{
		{
			name: "single_author",
			metadata: Metadata{
				Authors: []string{"Brandon Sanderson"},
			},
			defaultValue: "Unknown",
			want:         "Brandon Sanderson",
		},
		{
			name: "multiple_authors",
			metadata: Metadata{
				Authors: []string{"Stephen King", "Peter Straub"},
			},
			defaultValue: "Unknown",
			want:         "Stephen King",
		},
		{
			name: "no_authors",
			metadata: Metadata{
				Authors: []string{},
			},
			defaultValue: "Unknown Author",
			want:         "Unknown Author",
		},
		{
			name: "empty_first_author",
			metadata: Metadata{
				Authors: []string{""},
			},
			defaultValue: "Unknown",
			want:         "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metadata.GetFirstAuthor(tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetFirstAuthor(%q) = %q, want %q", tt.defaultValue, got, tt.want)
			}
		})
	}
}

func TestMetadataIsValid(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		want     bool
	}{
		{
			name: "valid_metadata",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
			},
			want: true,
		},
		{
			name: "missing_title",
			metadata: Metadata{
				Title:   "",
				Authors: []string{"Test Author"},
			},
			want: false,
		},
		{
			name: "missing_authors",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{},
			},
			want: false,
		},
		{
			name: "empty_author",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{""},
			},
			want: false,
		},
		{
			name: "missing_both",
			metadata: Metadata{
				Title:   "",
				Authors: []string{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metadata.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetadataValidate(t *testing.T) {
	tests := []struct {
		name      string
		metadata  Metadata
		wantError bool
	}{
		{
			name: "valid_metadata",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
			},
			wantError: false,
		},
		{
			name: "missing_title",
			metadata: Metadata{
				Title:   "",
				Authors: []string{"Test Author"},
			},
			wantError: true,
		},
		{
			name: "missing_authors",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{},
			},
			wantError: true,
		},
		{
			name: "empty_author",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{""},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metadata.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
