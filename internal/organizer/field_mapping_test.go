//go:build !integration

package organizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldMapping(t *testing.T) {
	tests := []struct {
		name           string
		metadata       Metadata
		fieldMapping   FieldMapping
		expectedTitle  string
		expectedSeries string
		expectedAuthor string
	}{
		{
			name: "default_mapping",
			metadata: Metadata{
				Title:       "Book Title",
				Series:      []string{"Series Name"},
				Authors:     []string{"Author Name"},
				Album:       "Album Name",
				TrackTitle:  "Track Title",
				TrackNumber: 1,
				RawData: map[string]interface{}{
					"title":   "Book Title",
					"series":  "Series Name",
					"authors": "Author Name",
				},
			},
			fieldMapping: FieldMapping{
				TitleField:   "title",
				SeriesField:  "series",
				AuthorFields: []string{"authors"},
				TrackField:   "track",
			},
			expectedTitle:  "Book Title",
			expectedSeries: "Series Name",
			expectedAuthor: "Author Name",
		},
		{
			name: "series_as_title",
			metadata: Metadata{
				Title:   "Book Title",
				Series:  []string{"Series Name"},
				Authors: []string{"Author Name"},
			},
			fieldMapping: FieldMapping{
				TitleField:   "series", // Series should be used as title
				SeriesField:  "title",  // Title will be used as series
				AuthorFields: []string{"authors"},
			},
			expectedTitle:  "Series Name",
			expectedSeries: "Book Title",
			expectedAuthor: "Author Name",
		},
		{
			name: "album_as_title",
			metadata: Metadata{
				Title:   "Book Title",
				Album:   "Album Name",
				Authors: []string{"Author Name"},
			},
			fieldMapping: FieldMapping{
				TitleField:   "album",
				SeriesField:  "",
				AuthorFields: []string{"authors"},
			},
			expectedTitle:  "Album Name",
			expectedSeries: "",
			expectedAuthor: "Author Name",
		},
		{
			name: "track_title_as_title",
			metadata: Metadata{
				Title:      "Book Title",
				TrackTitle: "Track Title",
				Authors:    []string{"Author Name"},
			},
			fieldMapping: FieldMapping{
				TitleField:   "track_title",
				SeriesField:  "",
				AuthorFields: []string{"authors"},
			},
			expectedTitle:  "Track Title",
			expectedSeries: "",
			expectedAuthor: "Author Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply the field mapping
			tt.metadata.ApplyFieldMapping(tt.fieldMapping)

			// Verify the results
			assert.Equal(t, tt.expectedTitle, tt.metadata.Title, "title should match")
			if len(tt.metadata.Series) > 0 {
				assert.Equal(t, tt.expectedSeries, tt.metadata.Series[0], "first series should match")
			} else {
				assert.Empty(t, tt.expectedSeries, "expected no series")
			}
			if len(tt.metadata.Authors) > 0 {
				assert.Equal(t, tt.expectedAuthor, tt.metadata.Authors[0], "first author should match")
			} else {
				assert.Empty(t, tt.expectedAuthor, "expected no authors")
			}

			// Verify the verbose output shows the correct field usage
			output := tt.metadata.FormatFieldMappingAndValues()
			assert.Contains(t, output, "Title")
			assert.Contains(t, output, tt.expectedTitle)

			if tt.expectedSeries != "" {
				assert.Contains(t, output, "Series")
				if tt.name == "series_as_title" {
					assert.Contains(t, output, "Series Name")
				} else {
					assert.Contains(t, output, tt.expectedSeries)
				}
			}

			// Verify the field mapping is correctly indicated in the output
			if tt.fieldMapping.TitleField != "title" && tt.fieldMapping.TitleField != "" {
				// Check for the field name in the output
				fieldName := tt.fieldMapping.TitleField
				assert.Contains(t, output, fieldName)
			}
		})
	}
}

func TestFieldMappingWithMultipleAuthors(t *testing.T) {
	metadata := Metadata{
		Title:   "Book Title",
		Series:  []string{"Series Name"},
		Authors: []string{"Author One", "Author Two"},
		RawData: map[string]interface{}{
			"title":   "Book Title",
			"series":  "Series Name",
			"authors": "Author One; Author Two",
		},
	}

	fieldMapping := FieldMapping{
		TitleField:   "title",
		SeriesField:  "series",
		AuthorFields: []string{"authors"},
	}

	metadata.ApplyFieldMapping(fieldMapping)

	assert.Equal(t, "Book Title", metadata.Title)
	assert.Equal(t, "Series Name", metadata.GetValidSeries())
	assert.Len(t, metadata.Authors, 2)
	assert.Equal(t, "Author One", metadata.Authors[0])
	assert.Equal(t, "Author Two", metadata.Authors[1])
}

func TestFieldMappingWithCustomFields(t *testing.T) {
	metadata := Metadata{
		Title:   "Book Title",
		Series:  []string{"Series Name"},
		Authors: []string{"Author Name"},
		RawData: map[string]interface{}{
			"custom_title":  "Custom Title",
			"custom_series": "Custom Series",
			"custom_author": "Custom Author",
		},
	}

	fieldMapping := FieldMapping{
		TitleField:   "custom_title",
		SeriesField:  "custom_series",
		AuthorFields: []string{"custom_author"},
	}

	metadata.ApplyFieldMapping(fieldMapping)

	assert.Equal(t, "Custom Title", metadata.Title)
	assert.Equal(t, "Custom Series", metadata.GetValidSeries())
	assert.Equal(t, "Custom Author", metadata.GetFirstAuthor(""))
}
