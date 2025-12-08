package organizer

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCalculateTargetPathWithSeriesNumber(t *testing.T) {
	tests := []struct {
		name     string
		layout   string
		metadata Metadata
		expected string
	}{
		{
			name:   "author-series-title-number with series_index",
			layout: "author-series-title-number",
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
			name:   "author-series-title-number with series string",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "The Well of Ascension",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn #2"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "#2 - The Well of Ascension"),
		},
		{
			name:   "author-series-title-number with decimal series_index",
			layout: "author-series-title-number",
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
			name:   "author-series-title-number without series number",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "The Hero of Ages",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "The Hero of Ages"),
		},
		{
			name:   "author-series-title-number without series",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "Elantris",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Elantris"),
		},
		{
			name:   "regular author-series-title layout",
			layout: "author-series-title",
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn #1"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "The Final Empire"),
		},
		{
			name:   "author-series-title-number with large series number",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "Leviathan Wakes",
				Authors: []string{"James S.A. Corey"},
				Series:  []string{"The Expanse #1"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "James S.A. Corey", "The Expanse", "#1 - Leviathan Wakes"),
		},
		{
			name:   "author-series-title-number with double-digit series number",
			layout: "author-series-title-number",
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
			name:   "author-series-title-number with multiple authors",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "The Talisman",
				Authors: []string{"Stephen King", "Peter Straub"},
				Series:  []string{"The Talisman #1"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Stephen King,Peter Straub", "The Talisman", "#1 - The Talisman"),
		},
		{
			name:   "author-series-title-number with special characters in title",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "The Magician's Nephew",
				Authors: []string{"C.S. Lewis"},
				Series:  []string{"The Chronicles of Narnia"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("testbase", "C.S. Lewis", "The Chronicles of Narnia", "#1 - The Magician's Nephew"),
		},
		{
			name:   "author-series-title-number with invalid series marker",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "Some Book",
				Authors: []string{"Some Author"},
				Series:  []string{InvalidSeriesValue},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Some Author", "Some Book"),
		},
		{
			name:   "author-series-title-number series_index overrides series string",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "Conflict Test",
				Authors: []string{"Test Author"},
				Series:  []string{"Test Series #5"},
				RawData: map[string]interface{}{
					"series_index": 3.0,
				},
			},
			expected: filepath.Join("testbase", "Test Author", "Test Series", "#3 - Conflict Test"),
		},
		{
			name:   "author-series-title-number with zero series_index",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "Prequel",
				Authors: []string{"Test Author"},
				Series:  []string{"Test Series"},
				RawData: map[string]interface{}{
					"series_index": 0.0,
				},
			},
			expected: filepath.Join("testbase", "Test Author", "Test Series", "Prequel"),
		},
		{
			name:   "author-series-title-number with negative series_index",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "Before Time",
				Authors: []string{"Test Author"},
				Series:  []string{"Test Series"},
				RawData: map[string]interface{}{
					"series_index": -1.0,
				},
			},
			expected: filepath.Join("testbase", "Test Author", "Test Series", "Before Time"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OrganizerConfig{
				BaseDir: "testbase",
				Layout:  tt.layout,
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

func TestExtractSeriesNumber(t *testing.T) {
	tests := []struct {
		name     string
		series   string
		expected string
	}{
		{
			name:     "series with number",
			series:   "Mistborn #1",
			expected: "1",
		},
		{
			name:     "series with multi-digit number",
			series:   "The Expanse #12",
			expected: "12",
		},
		{
			name:     "series without number",
			series:   "Mistborn",
			expected: "",
		},
		{
			name:     "series with decimal number",
			series:   "Mistborn #0.5",
			expected: "0.5",
		},
		{
			name:     "empty series",
			series:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSeriesNumber(tt.series)
			if result != tt.expected {
				t.Errorf("ExtractSeriesNumber(%q) = %q, want %q", tt.series, result, tt.expected)
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
			name: "with series_index as integer",
			metadata: Metadata{
				Series: []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: "1",
		},
		{
			name: "with series_index as decimal",
			metadata: Metadata{
				Series: []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.5,
				},
			},
			expected: "1.5",
		},
		{
			name: "with series string containing number",
			metadata: Metadata{
				Series:  []string{"Mistborn #2"},
				RawData: map[string]interface{}{},
			},
			expected: "2",
		},
		{
			name: "without series number",
			metadata: Metadata{
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{},
			},
			expected: "",
		},
		{
			name: "empty metadata",
			metadata: Metadata{
				Series:  []string{},
				RawData: map[string]interface{}{},
			},
			expected: "",
		},
		{
			name: "series_index takes precedence over series string",
			metadata: Metadata{
				Series: []string{"Mistborn #2"},
				RawData: map[string]interface{}{
					"series_index": 3.0,
				},
			},
			expected: "3",
		},
		{
			name: "with zero series_index",
			metadata: Metadata{
				Series: []string{"Test Series"},
				RawData: map[string]interface{}{
					"series_index": 0.0,
				},
			},
			expected: "",
		},
		{
			name: "with negative series_index",
			metadata: Metadata{
				Series: []string{"Test Series"},
				RawData: map[string]interface{}{
					"series_index": -1.0,
				},
			},
			expected: "",
		},
		{
			name: "with very large series_index",
			metadata: Metadata{
				Series: []string{"Long Series"},
				RawData: map[string]interface{}{
					"series_index": 999.0,
				},
			},
			expected: "999",
		},
		{
			name: "with complex decimal series_index",
			metadata: Metadata{
				Series: []string{"Test Series"},
				RawData: map[string]interface{}{
					"series_index": 3.14159,
				},
			},
			expected: "3.1",
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

// TestLayoutWithSanitization tests that the layout works correctly with path sanitization
func TestLayoutWithSanitization(t *testing.T) {
	tests := []struct {
		name     string
		layout   string
		metadata Metadata
		expected string
	}{
		{
			name:   "sanitize special characters in numbered title",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "Book: The Beginning",
				Authors: []string{"Test Author"},
				Series:  []string{"Test Series"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("testbase", "Test Author", "Test Series", "#1 - Book_ The Beginning"),
		},
		{
			name:   "sanitize slashes in numbered title",
			layout: "author-series-title-number",
			metadata: Metadata{
				Title:   "Book/Part 1",
				Authors: []string{"Test Author"},
				Series:  []string{"Test Series"},
				RawData: map[string]interface{}{
					"series_index": 2.0,
				},
			},
			expected: filepath.Join("testbase", "Test Author", "Test Series", "#2 - Book_Part 1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OrganizerConfig{
				BaseDir: "testbase",
				Layout:  tt.layout,
			}

			// Use a simple sanitizer that replaces : and / with _
			sanitizer := func(s string) string {
				s = strings.ReplaceAll(s, ":", "_")
				s = strings.ReplaceAll(s, "/", "_")
				return s
			}

			lc := NewLayoutCalculator(config, sanitizer)
			result := lc.CalculateTargetPath(tt.metadata)
			if result != tt.expected {
				t.Errorf("CalculateTargetPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestLayoutWithOutputDir tests that output directory is respected
func TestLayoutWithOutputDir(t *testing.T) {
	tests := []struct {
		name      string
		layout    string
		outputDir string
		metadata  Metadata
		expected  string
	}{
		{
			name:      "output dir with series number layout",
			layout:    "author-series-title-number",
			outputDir: "/output/books",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
				Series:  []string{"Test Series"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("/output/books", "Test Author", "Test Series", "#1 - Test Book"),
		},
		{
			name:      "output dir with regular layout",
			layout:    "author-series-title",
			outputDir: "/output/books",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
				Series:  []string{"Test Series #1"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("/output/books", "Test Author", "Test Series", "Test Book"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OrganizerConfig{
				BaseDir:   "testbase",
				OutputDir: tt.outputDir,
				Layout:    tt.layout,
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

// TestAllLayoutOptions tests all layout options to ensure they work correctly
func TestAllLayoutOptions(t *testing.T) {
	metadata := Metadata{
		Title:   "The Final Empire",
		Authors: []string{"Brandon Sanderson"},
		Series:  []string{"Mistborn"},
		RawData: map[string]interface{}{
			"series_index": 1.0,
		},
	}

	tests := []struct {
		layout   string
		expected string
	}{
		{
			layout:   "author-only",
			expected: filepath.Join("testbase", "Brandon Sanderson"),
		},
		{
			layout:   "author-title",
			expected: filepath.Join("testbase", "Brandon Sanderson", "The Final Empire"),
		},
		{
			layout:   "author-series-title",
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "The Final Empire"),
		},
		{
			layout:   "author-series-title-number",
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "#1 - The Final Empire"),
		},
		{
			layout:   "series-title",
			expected: filepath.Join("testbase", "Mistborn", "The Final Empire"),
		},
		{
			layout:   "series-title-number",
			expected: filepath.Join("testbase", "Mistborn", "#1 - The Final Empire"),
		},
		{
			layout:   "", // Default should be author-series-title
			expected: filepath.Join("testbase", "Brandon Sanderson", "Mistborn", "The Final Empire"),
		},
		{
			layout:   "unknown-layout", // Unknown should default to author-title
			expected: filepath.Join("testbase", "Brandon Sanderson", "The Final Empire"),
		},
	}

	for _, tt := range tests {
		t.Run("layout_"+tt.layout, func(t *testing.T) {
			config := &OrganizerConfig{
				BaseDir: "testbase",
				Layout:  tt.layout,
			}

			sanitizer := func(s string) string { return s }
			lc := NewLayoutCalculator(config, sanitizer)

			result := lc.CalculateTargetPath(metadata)
			if result != tt.expected {
				t.Errorf("Layout %q: CalculateTargetPath() = %v, want %v", tt.layout, result, tt.expected)
			}
		})
	}
}

// TestSeriesNumberEdgeCases tests edge cases for series number extraction
func TestSeriesNumberEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		layout   string
		expected string
	}{
		{
			name: "series with multiple hash symbols",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
				Series:  []string{"Test #Series #3"},
				RawData: map[string]interface{}{},
			},
			layout:   "author-series-title-number",
			expected: filepath.Join("testbase", "Test Author", "Test #Series", "#3 - Test Book"),
		},
		{
			name: "series with hash but no number",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
				Series:  []string{"Test Series #"},
				RawData: map[string]interface{}{},
			},
			layout:   "author-series-title-number",
			expected: filepath.Join("testbase", "Test Author", "Test Series", "Test Book"),
		},
		{
			name: "series with whitespace around number",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
				Series:  []string{"Test Series #  5  "},
				RawData: map[string]interface{}{},
			},
			layout:   "author-series-title-number",
			expected: filepath.Join("testbase", "Test Author", "Test Series", "#5 - Test Book"),
		},
		{
			name: "empty series array",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
				Series:  []string{},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			layout:   "author-series-title-number",
			expected: filepath.Join("testbase", "Test Author", "Test Book"),
		},
		{
			name: "nil RawData",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
				Series:  []string{"Test Series #1"},
				RawData: nil,
			},
			layout:   "author-series-title-number",
			expected: filepath.Join("testbase", "Test Author", "Test Series", "#1 - Test Book"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OrganizerConfig{
				BaseDir: "testbase",
				Layout:  tt.layout,
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

// TestSeriesOnlyLayouts tests the series-title and series-title-number layouts
func TestSeriesOnlyLayouts(t *testing.T) {
	tests := []struct {
		name     string
		layout   string
		metadata Metadata
		expected string
	}{
		{
			name:   "series-title with series",
			layout: "series-title",
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Mistborn", "The Final Empire"),
		},
		{
			name:   "series-title without series",
			layout: "series-title",
			metadata: Metadata{
				Title:   "Elantris",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Elantris"),
		},
		{
			name:   "series-title-number with series_index",
			layout: "series-title-number",
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("testbase", "Mistborn", "#1 - The Final Empire"),
		},
		{
			name:   "series-title-number with series string number",
			layout: "series-title-number",
			metadata: Metadata{
				Title:   "The Well of Ascension",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn #2"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Mistborn", "#2 - The Well of Ascension"),
		},
		{
			name:   "series-title-number without series",
			layout: "series-title-number",
			metadata: Metadata{
				Title:   "Elantris",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Elantris"),
		},
		{
			name:   "series-title-number with series but no number",
			layout: "series-title-number",
			metadata: Metadata{
				Title:   "The Hero of Ages",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Mistborn", "The Hero of Ages"),
		},
		{
			name:   "series-title with multiple series (sorted)",
			layout: "series-title",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
				Series:  []string{"Z Series", "A Series"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "A Series", "Test Book"),
		},
		{
			name:   "series-title-number with decimal series_index",
			layout: "series-title-number",
			metadata: Metadata{
				Title:   "The Eleventh Metal",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 0.5,
				},
			},
			expected: filepath.Join("testbase", "Mistborn", "#0.5 - The Eleventh Metal"),
		},
		{
			name:   "series-title with invalid series marker",
			layout: "series-title",
			metadata: Metadata{
				Title:   "Test Book",
				Authors: []string{"Test Author"},
				Series:  []string{InvalidSeriesValue},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("testbase", "Test Book"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OrganizerConfig{
				BaseDir: "testbase",
				Layout:  tt.layout,
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
