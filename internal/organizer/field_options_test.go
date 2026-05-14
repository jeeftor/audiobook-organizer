package organizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextFieldOptions(t *testing.T) {
	opts := TextFieldOptions()
	assert.NotEmpty(t, opts)
	assert.Contains(t, opts, "title")
	assert.Contains(t, opts, "album")
	assert.Contains(t, opts, "series")
	assert.Contains(t, opts, "track_title")
}

func TestAuthorFieldOptions(t *testing.T) {
	opts := AuthorFieldOptions()
	assert.NotEmpty(t, opts)
	assert.Contains(t, opts, "authors")
	assert.Contains(t, opts, "artist")
	assert.Contains(t, opts, "album_artist")
	assert.Contains(t, opts, "narrator")
	assert.Contains(t, opts, "narrators")
}

func TestTrackFieldOptions(t *testing.T) {
	opts := TrackFieldOptions()
	assert.NotEmpty(t, opts)
	assert.Contains(t, opts, "track")
	assert.Contains(t, opts, "tracknumber")
	assert.Contains(t, opts, "trck")
}

func TestDiscFieldOptions(t *testing.T) {
	opts := DiscFieldOptions()
	assert.NotEmpty(t, opts)
	assert.Contains(t, opts, "disc")
	assert.Contains(t, opts, "discnumber")
	assert.Contains(t, opts, "disk")
	assert.Contains(t, opts, "tpos")
	assert.Contains(t, opts, "disc_number")
}

func TestFieldMappingConstants(t *testing.T) {
	assert.Equal(t, "title", TitleFieldKey)
	assert.Equal(t, "series", SeriesFieldKey)
	assert.Equal(t, "authors", AuthorsFieldKey)
	assert.Equal(t, "track", TrackFieldKey)
	assert.Equal(t, "disc", DiscFieldKey)
}

func TestFieldOptionsNoDuplicates(t *testing.T) {
	checkNoDuplicates := func(name string, opts []string) {
		seen := make(map[string]bool)
		for _, o := range opts {
			assert.False(t, seen[o], "%s: duplicate entry %q", name, o)
			seen[o] = true
		}
	}
	checkNoDuplicates("TextFieldOptions", TextFieldOptions())
	checkNoDuplicates("AuthorFieldOptions", AuthorFieldOptions())
	checkNoDuplicates("TrackFieldOptions", TrackFieldOptions())
	checkNoDuplicates("DiscFieldOptions", DiscFieldOptions())
}
