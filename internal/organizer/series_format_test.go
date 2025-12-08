package organizer

import (
	"path/filepath"
	"testing"
)

// TestFormatSeriesNumber tests the new FormatSeriesNumber function
func TestFormatSeriesNumber(t *testing.T) {
	tests := []struct {
		name          string
		seriesNumber  string
		format        string
		padding       int
		expected      string
		expectedError bool
	}{
		// Bracket format tests
		{
			name:         "bracket format - single digit with padding 2",
			seriesNumber: "1",
			format:       "bracket",
			padding:      2,
			expected:     "[01]",
		},
		{
			name:         "bracket format - double digit with padding 2",
			seriesNumber: "12",
			format:       "bracket",
			padding:      2,
			expected:     "[12]",
		},
		{
			name:         "bracket format - triple digit with padding 2",
			seriesNumber: "123",
			format:       "bracket",
			padding:      2,
			expected:     "[123]",
		},
		{
			name:         "bracket format - single digit with padding 3",
			seriesNumber: "5",
			format:       "bracket",
			padding:      3,
			expected:     "[005]",
		},
		{
			name:         "bracket format - decimal number",
			seriesNumber: "2.5",
			format:       "bracket",
			padding:      2,
			expected:     "[02.5]",
		},
		{
			name:         "bracket format - decimal with larger integer part",
			seriesNumber: "10.5",
			format:       "bracket",
			padding:      2,
			expected:     "[10.5]",
		},
		{
			name:         "bracket format - decimal with padding 3",
			seriesNumber: "2.5",
			format:       "bracket",
			padding:      3,
			expected:     "[002.5]",
		},
		{
			name:         "bracket format - zero padding 2",
			seriesNumber: "0",
			format:       "bracket",
			padding:      2,
			expected:     "[00]",
		},
		{
			name:         "bracket format - 0.5 with padding 2",
			seriesNumber: "0.5",
			format:       "bracket",
			padding:      2,
			expected:     "[00.5]",
		},

		// Hash format tests (legacy)
		{
			name:         "hash format - single digit",
			seriesNumber: "1",
			format:       "hash",
			padding:      2,
			expected:     "#1",
		},
		{
			name:         "hash format - double digit",
			seriesNumber: "12",
			format:       "hash",
			padding:      2,
			expected:     "#12",
		},
		{
			name:         "hash format - decimal",
			seriesNumber: "2.5",
			format:       "hash",
			padding:      2,
			expected:     "#2.5",
		},
		{
			name:         "hash format - padding ignored",
			seriesNumber: "1",
			format:       "hash",
			padding:      5,
			expected:     "#1",
		},

		// Edge cases
		{
			name:         "empty series number",
			seriesNumber: "",
			format:       "bracket",
			padding:      2,
			expected:     "",
		},
		{
			name:         "empty series number with hash",
			seriesNumber: "",
			format:       "hash",
			padding:      2,
			expected:     "",
		},
		{
			name:         "default format (should be bracket)",
			seriesNumber: "1",
			format:       "",
			padding:      2,
			expected:     "[01]",
		},
		{
			name:         "default padding (should be 2)",
			seriesNumber: "1",
			format:       "bracket",
			padding:      0,
			expected:     "[01]",
		},
		{
			name:         "padding 1",
			seriesNumber: "1",
			format:       "bracket",
			padding:      1,
			expected:     "[1]",
		},
		{
			name:         "large series number with small padding",
			seriesNumber: "999",
			format:       "bracket",
			padding:      2,
			expected:     "[999]",
		},
		{
			name:         "very large series number",
			seriesNumber: "1234",
			format:       "bracket",
			padding:      3,
			expected:     "[1234]",
		},
		{
			name:         "decimal with multiple decimal places",
			seriesNumber: "3.14",
			format:       "bracket",
			padding:      2,
			expected:     "[03.14]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSeriesNumber(tt.seriesNumber, tt.format, tt.padding)
			if result != tt.expected {
				t.Errorf("FormatSeriesNumber(%q, %q, %d) = %q, want %q",
					tt.seriesNumber, tt.format, tt.padding, result, tt.expected)
			}
		})
	}
}

// TestCalculateTargetPathWithBracketFormat tests path calculation with bracket format
func TestCalculateTargetPathWithBracketFormat(t *testing.T) {
	tests := []struct {
		name          string
		layout        string
		seriesFormat  string
		seriesPadding int
		metadata      Metadata
		expected      string
	}{
		{
			name:          "bracket format - single digit series",
			layout:        "author-series-title-number",
			seriesFormat:  "bracket",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "[01] The Final Empire"),
		},
		{
			name:          "bracket format - double digit series",
			layout:        "author-series-title-number",
			seriesFormat:  "bracket",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "Tiamat's Wrath",
				Authors: []string{"James S.A. Corey"},
				Series:  []string{"The Expanse"},
				RawData: map[string]interface{}{
					"series_index": 12.0,
				},
			},
			expected: filepath.Join("testbase", "James S.A. Corey", "The Expanse", "[12] Tiamat's Wrath"),
		},
		{
			name:          "bracket format - decimal series",
			layout:        "author-series-title-number",
			seriesFormat:  "bracket",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "The Eleventh Metal",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 0.5,
				},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "[00.5] The Eleventh Metal"),
		},
		{
			name:          "bracket format - padding 3",
			layout:        "author-series-title-number",
			seriesFormat:  "bracket",
			seriesPadding: 3,
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "[001] The Final Empire"),
		},
		{
			name:          "bracket format - series from string",
			layout:        "author-series-title-number",
			seriesFormat:  "bracket",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "The Well of Ascension",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn #2"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "[02] The Well of Ascension"),
		},
		{
			name:          "bracket format - no series number",
			layout:        "author-series-title-number",
			seriesFormat:  "bracket",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "The Hero of Ages",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "The Hero of Ages"),
		},
		{
			name:          "bracket format - no series",
			layout:        "author-series-title-number",
			seriesFormat:  "bracket",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "Elantris",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Elantris"),
		},
		{
			name:          "hash format - backward compatibility",
			layout:        "author-series-title-number",
			seriesFormat:  "hash",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "#1 - The Final Empire"),
		},
		{
			name:          "hash format - double digit",
			layout:        "author-series-title-number",
			seriesFormat:  "hash",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "Tiamat's Wrath",
				Authors: []string{"James S.A. Corey"},
				Series:  []string{"The Expanse"},
				RawData: map[string]interface{}{
					"series_index": 12.0,
				},
			},
			expected: filepath.Join("testbase", "James S.A. Corey", "The Expanse", "#12 - Tiamat's Wrath"),
		},
		{
			name:          "hash format - decimal",
			layout:        "author-series-title-number",
			seriesFormat:  "hash",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "The Eleventh Metal",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 0.5,
				},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "#0.5 - The Eleventh Metal"),
		},
		{
			name:          "default format should be bracket",
			layout:        "author-series-title-number",
			seriesFormat:  "",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "[01] The Final Empire"),
		},
		{
			name:          "default padding should be 2",
			layout:        "author-series-title-number",
			seriesFormat:  "bracket",
			seriesPadding: 0,
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "[01] The Final Empire"),
		},
		{
			name:          "large series with small padding",
			layout:        "author-series-title-number",
			seriesFormat:  "bracket",
			seriesPadding: 2,
			metadata: Metadata{
				Title:   "Book 999",
				Authors: []string{"Prolific Author"},
				Series:  []string{"Long Series"},
				RawData: map[string]interface{}{
					"series_index": 999.0,
				},
			},
			expected: filepath.Join("testbase", "Prolific Author", "Long Series", "[999] Book 999"),
		},
		{
			name:          "very large series with padding 3",
			layout:        "author-series-title-number",
			seriesFormat:  "bracket",
			seriesPadding: 3,
			metadata: Metadata{
				Title:   "Book 1234",
				Authors: []string{"Prolific Author"},
				Series:  []string{"Very Long Series"},
				RawData: map[string]interface{}{
					"series_index": 1234.0,
				},
			},
			expected: filepath.Join("testbase", "Prolific Author", "Very Long Series", "[1234] Book 1234"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OrganizerConfig{
				BaseDir:       "testbase",
				Layout:        tt.layout,
				SeriesFormat:  tt.seriesFormat,
				SeriesPadding: tt.seriesPadding,
			}

			sanitizer := func(s string) string { return s }
			lc := NewLayoutCalculator(config, sanitizer)

			result := lc.CalculateTargetPath(tt.metadata)
			if result != tt.expected {
				t.Errorf("CalculateTargetPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSeriesFormatSorting verifies that bracket format sorts correctly
func TestSeriesFormatSorting(t *testing.T) {
	// Test that bracket format provides correct lexicographic sorting
	tests := []struct {
		name     string
		paths    []string
		expected []string
	}{
		{
			name: "bracket format sorts correctly",
			paths: []string{
				"Author/Series/[02] Book Two",
				"Author/Series/[01] Book One",
				"Author/Series/[10] Book Ten",
				"Author/Series/[03] Book Three",
			},
			expected: []string{
				"Author/Series/[01] Book One",
				"Author/Series/[02] Book Two",
				"Author/Series/[03] Book Three",
				"Author/Series/[10] Book Ten",
			},
		},
		{
			name: "hash format sorts incorrectly (demonstrates problem)",
			paths: []string{
				"Author/Series/#2 - Book Two",
				"Author/Series/#1 - Book One",
				"Author/Series/#10 - Book Ten",
				"Author/Series/#3 - Book Three",
			},
			expected: []string{
				"Author/Series/#1 - Book One",
				"Author/Series/#10 - Book Ten", // This comes before #2!
				"Author/Series/#2 - Book Two",
				"Author/Series/#3 - Book Three",
			},
		},
		{
			name: "bracket format with decimals",
			paths: []string{
				"Author/Series/[02.5] Book Two Point Five",
				"Author/Series/[01] Book One",
				"Author/Series/[02] Book Two",
				"Author/Series/[03] Book Three",
			},
			expected: []string{
				"Author/Series/[01] Book One",
				"Author/Series/[02.5] Book Two Point Five",
				"Author/Series/[02] Book Two",
				"Author/Series/[03] Book Three",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Sort using string comparison
			sorted := make([]string, len(tt.paths))
			copy(sorted, tt.paths)

			// Simple bubble sort to demonstrate lexicographic ordering
			for i := 0; i < len(sorted); i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i] > sorted[j] {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}

			for i, path := range sorted {
				if path != tt.expected[i] {
					t.Errorf("Position %d: got %q, want %q", i, path, tt.expected[i])
				}
			}
		})
	}
}

// TestBracketFormatShellCompatibility documents shell compatibility improvements
func TestBracketFormatShellCompatibility(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		needsQuoting   bool
		description    string
	}{
		{
			name:         "bracket format doesn't need quoting",
			path:         "Author/Series/[01] Book",
			needsQuoting: false,
			description:  "Bracket format paths work without quotes in shell",
		},
		{
			name:         "hash format needs quoting",
			path:         "Author/Series/#1 - Book",
			needsQuoting: true,
			description:  "Hash format requires quoting due to # symbol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test is documentary - it shows the problem we're solving
			// In actual shell usage:
			// - cd Author/Series/[01]\ Book  (works with just space escaping)
			// - cd 'Author/Series/#1 - Book'  (requires full quoting)
			t.Logf("Path: %s", tt.path)
			t.Logf("Needs quoting: %v", tt.needsQuoting)
			t.Logf("Description: %s", tt.description)
		})
	}
}
