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
			tt.metadata.FieldMapping = tt.fieldMapping
			tt.metadata.ApplyFieldMapping()

			// Verify the results
			assert.Equal(t, tt.expectedTitle, tt.metadata.Title, "title should match")

			if tt.expectedSeries != "" {
				assert.NotEmpty(t, tt.metadata.Series, "series should not be empty")
				assert.Equal(t, tt.expectedSeries, tt.metadata.Series[0], "series should match")
			} else {
				assert.Empty(t, tt.metadata.Series, "series should be empty")
			}

			if tt.expectedAuthor != "" {
				assert.NotEmpty(t, tt.metadata.Authors, "authors should not be empty")
				assert.Equal(t, tt.expectedAuthor, tt.metadata.Authors[0], "author should match")
			} else {
				assert.Empty(t, tt.metadata.Authors, "authors should be empty")
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
		Authors: []string{"First Author", "Second Author"},
	}

	fieldMapping := FieldMapping{
		TitleField:   "title",
		AuthorFields: []string{"authors"},
	}

	metadata.FieldMapping = fieldMapping
	metadata.ApplyFieldMapping()

	assert.Equal(t, 2, len(metadata.Authors), "should have two authors")
	assert.Equal(t, "First Author", metadata.Authors[0], "first author should match")
	assert.Equal(t, "Second Author", metadata.Authors[1], "second author should match")
}

func TestFieldMappingWithCustomFields(t *testing.T) {
	metadata := Metadata{
		RawMetadata: map[string]interface{}{
			"custom_title":  "Custom Title",
			"custom_series": "Custom Series",
		},
	}

	fieldMapping := FieldMapping{
		TitleField:   "custom_title",
		SeriesField:  "custom_series",
		AuthorFields: []string{"authors"},
	}

	metadata.FieldMapping = fieldMapping
	metadata.ApplyFieldMapping()

	assert.Equal(t, "Custom Title", metadata.Title, "should use custom title field")
	if assert.NotEmpty(t, metadata.Series, "series should not be empty") {
		assert.Equal(t, "Custom Series", metadata.Series[0], "should use custom series field")
	}
}
