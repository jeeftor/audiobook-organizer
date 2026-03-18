package organizer

// Common field options for metadata field mapping
// These are shared between CLI and GUI implementations

// TextFieldOptions returns common text field names for title/series/album mapping
func TextFieldOptions() []string {
	return []string{"title", "album", "series", "name", "book", "work", "track_title"}
}

// AuthorFieldOptions returns common author field names
func AuthorFieldOptions() []string {
	return []string{
		"authors",
		"artist",
		"album_artist",
		"narrator",
		"narrators",
		"creator",
		"author",
		"writer",
		"composer",
	}
}

// TrackFieldOptions returns common track number field names
func TrackFieldOptions() []string {
	return []string{
		"track",
		"track_number",
		"trck",
		"trk",
		"tracknumber",
	}
}

// DiscFieldOptions returns common disc number field names
func DiscFieldOptions() []string {
	return []string{
		"disc",
		"discnumber",
		"disk",
		"tpos",
		"disc_number",
	}
}

// FieldMappingConstants defines field names for consistency
const (
	TitleFieldKey   = "title"
	SeriesFieldKey  = "series"
	AuthorsFieldKey = "authors"
	TrackFieldKey   = "track"
	DiscFieldKey    = "disc"
)
