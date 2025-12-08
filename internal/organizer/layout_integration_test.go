package organizer

import (
	"path/filepath"
	"testing"
)

// TestSeriesNumberLayoutIntegration tests the complete flow from metadata to path
func TestSeriesNumberLayoutIntegration(t *testing.T) {
	tests := []struct {
		name        string
		description string
		config      OrganizerConfig
		metadata    Metadata
		expected    string
	}{
		{
			name:        "mistborn_series",
			description: "Brandon Sanderson's Mistborn series with series_index",
			config: OrganizerConfig{
				BaseDir: "/library",
				Layout:  "author-series-title-number",
			},
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("/library", "Brandon Sanderson", "Mistborn", "[01] The Final Empire"),
		},
		{
			name:        "expanse_series",
			description: "The Expanse series with double-digit number",
			config: OrganizerConfig{
				BaseDir: "/library",
				Layout:  "author-series-title-number",
			},
			metadata: Metadata{
				Title:   "Tiamat's Wrath",
				Authors: []string{"James S.A. Corey"},
				Series:  []string{"The Expanse"},
				RawData: map[string]interface{}{
					"series_index": 8.0,
				},
			},
			expected: filepath.Join("/library", "James S.A. Corey", "The Expanse", "[08] Tiamat's Wrath"),
		},
		{
			name:        "novella_with_decimal",
			description: "Novella with decimal series number (prequel)",
			config: OrganizerConfig{
				BaseDir: "/library",
				Layout:  "author-series-title-number",
			},
			metadata: Metadata{
				Title:   "The Eleventh Metal",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 0.5,
				},
			},
			expected: filepath.Join("/library", "Brandon Sanderson", "Mistborn", "[00.5] The Eleventh Metal"),
		},
		{
			name:        "series_from_string",
			description: "Series number extracted from series string",
			config: OrganizerConfig{
				BaseDir: "/library",
				Layout:  "author-series-title-number",
			},
			metadata: Metadata{
				Title:   "The Well of Ascension",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn #2"},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("/library", "Brandon Sanderson", "Mistborn", "[02] The Well of Ascension"),
		},
		{
			name:        "standalone_book",
			description: "Standalone book without series",
			config: OrganizerConfig{
				BaseDir: "/library",
				Layout:  "author-series-title-number",
			},
			metadata: Metadata{
				Title:   "Elantris",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{},
				RawData: map[string]interface{}{},
			},
			expected: filepath.Join("/library", "Brandon Sanderson", "Elantris"),
		},
		{
			name:        "with_output_dir",
			description: "Using custom output directory",
			config: OrganizerConfig{
				BaseDir:   "/input",
				OutputDir: "/organized",
				Layout:    "author-series-title-number",
			},
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("/organized", "Brandon Sanderson", "Mistborn", "[01] The Final Empire"),
		},
		{
			name:        "comparison_with_regular_layout",
			description: "Same book with regular layout (no series number)",
			config: OrganizerConfig{
				BaseDir: "/library",
				Layout:  "author-series-title",
			},
			metadata: Metadata{
				Title:   "The Final Empire",
				Authors: []string{"Brandon Sanderson"},
				Series:  []string{"Mistborn"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("/library", "Brandon Sanderson", "Mistborn", "The Final Empire"),
		},
		{
			name:        "narnia_publication_order",
			description: "Chronicles of Narnia in publication order",
			config: OrganizerConfig{
				BaseDir: "/library",
				Layout:  "author-series-title-number",
			},
			metadata: Metadata{
				Title:   "The Lion, the Witch and the Wardrobe",
				Authors: []string{"C.S. Lewis"},
				Series:  []string{"The Chronicles of Narnia"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("/library", "C.S. Lewis", "The Chronicles of Narnia", "[01] The Lion, the Witch and the Wardrobe"),
		},
		{
			name:        "narnia_chronological_order",
			description: "Chronicles of Narnia in chronological order (The Magician's Nephew is #1)",
			config: OrganizerConfig{
				BaseDir: "/library",
				Layout:  "author-series-title-number",
			},
			metadata: Metadata{
				Title:   "The Magician's Nephew",
				Authors: []string{"C.S. Lewis"},
				Series:  []string{"The Chronicles of Narnia"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("/library", "C.S. Lewis", "The Chronicles of Narnia", "[01] The Magician's Nephew"),
		},
		{
			name:        "collaborative_work",
			description: "Book with multiple authors",
			config: OrganizerConfig{
				BaseDir: "/library",
				Layout:  "author-series-title-number",
			},
			metadata: Metadata{
				Title:   "The Talisman",
				Authors: []string{"Stephen King", "Peter Straub"},
				Series:  []string{"The Talisman"},
				RawData: map[string]interface{}{
					"series_index": 1.0,
				},
			},
			expected: filepath.Join("/library", "Stephen King,Peter Straub", "The Talisman", "[01] The Talisman"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create organizer with test config
			sanitizer := func(s string) string { return s }
			lc := NewLayoutCalculator(&tt.config, sanitizer)

			// Calculate the target path
			result := lc.CalculateTargetPath(tt.metadata)

			// Verify the result
			if result != tt.expected {
				t.Errorf("\n%s\nCalculateTargetPath() = %v\nwant                  = %v",
					tt.description, result, tt.expected)
			} else {
				t.Logf("✓ %s\n  → %s", tt.description, result)
			}
		})
	}
}

// TestSeriesNumberLayoutRealWorldScenarios tests real-world usage scenarios
func TestSeriesNumberLayoutRealWorldScenarios(t *testing.T) {
	t.Run("mp3_collection_with_field_mapping", func(t *testing.T) {
		// Simulate MP3 files that use "album" for series
		config := OrganizerConfig{
			BaseDir: "/mp3_library",
			Layout:  "author-series-title-number",
			FieldMapping: FieldMapping{
				SeriesField:  "album",
				AuthorFields: []string{"artist"},
			},
		}

		metadata := Metadata{
			Title:   "Leviathan Wakes",
			Authors: []string{"James S.A. Corey"},
			Series:  []string{"The Expanse #1"},
			RawData: map[string]interface{}{},
		}

		sanitizer := func(s string) string { return s }
		lc := NewLayoutCalculator(&config, sanitizer)
		result := lc.CalculateTargetPath(metadata)

		expected := filepath.Join("/mp3_library", "James S.A. Corey", "The Expanse", "[01] Leviathan Wakes")
		if result != expected {
			t.Errorf("MP3 field mapping: got %v, want %v", result, expected)
		}
	})

	t.Run("epub_with_calibre_metadata", func(t *testing.T) {
		// Simulate EPUB with Calibre series metadata
		config := OrganizerConfig{
			BaseDir: "/epub_library",
			Layout:  "author-series-title-number",
		}

		metadata := Metadata{
			Title:   "The Final Empire",
			Authors: []string{"Brandon Sanderson"},
			Series:  []string{"Mistborn"},
			RawData: map[string]interface{}{
				"series_index": 1.0,
				"series":       "Mistborn",
			},
		}

		sanitizer := func(s string) string { return s }
		lc := NewLayoutCalculator(&config, sanitizer)
		result := lc.CalculateTargetPath(metadata)

		expected := filepath.Join("/epub_library", "Brandon Sanderson", "Mistborn", "[01] The Final Empire")
		if result != expected {
			t.Errorf("EPUB Calibre metadata: got %v, want %v", result, expected)
		}
	})

	t.Run("mixed_collection_with_and_without_series", func(t *testing.T) {
		config := OrganizerConfig{
			BaseDir: "/library",
			Layout:  "author-series-title-number",
		}

		sanitizer := func(s string) string { return s }
		lc := NewLayoutCalculator(&config, sanitizer)

		// Book with series
		metadata1 := Metadata{
			Title:   "The Final Empire",
			Authors: []string{"Brandon Sanderson"},
			Series:  []string{"Mistborn"},
			RawData: map[string]interface{}{
				"series_index": 1.0,
			},
		}
		result1 := lc.CalculateTargetPath(metadata1)
		expected1 := filepath.Join("/library", "Brandon Sanderson", "Mistborn", "[01] The Final Empire")

		// Standalone book
		metadata2 := Metadata{
			Title:   "Elantris",
			Authors: []string{"Brandon Sanderson"},
			Series:  []string{},
			RawData: map[string]interface{}{},
		}
		result2 := lc.CalculateTargetPath(metadata2)
		expected2 := filepath.Join("/library", "Brandon Sanderson", "Elantris")

		if result1 != expected1 {
			t.Errorf("Series book: got %v, want %v", result1, expected1)
		}
		if result2 != expected2 {
			t.Errorf("Standalone book: got %v, want %v", result2, expected2)
		}
	})
}
