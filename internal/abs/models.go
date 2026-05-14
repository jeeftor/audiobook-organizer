// internal/abs/models.go
// Data models for Audiobookshelf API

package abs

// Library represents an ABS library
type Library struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	MediaType    string   `json:"mediaType"` // "book" or "podcast"
	Folders      []Folder `json:"folders"`
	DisplayOrder int      `json:"displayOrder"`
	Icon         string   `json:"icon"`
	CreatedAt    int64    `json:"createdAt"`  // Milliseconds timestamp
	LastUpdate   int64    `json:"lastUpdate"` // Milliseconds timestamp
}

// Folder represents a library folder
type Folder struct {
	ID        string `json:"id"`
	Path      string `json:"path"`     // Relative path (e.g., "/audiobooks")
	FullPath  string `json:"fullPath"` // Full filesystem path
	LibraryID string `json:"libraryId,omitempty"`
}

// LibraryItemsResponse is the paginated response from ABS API
type LibraryItemsResponse struct {
	Results []LibraryItem `json:"results"`
	Total   int           `json:"total"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
}

// LibraryItem represents an audiobook/podcast in ABS
type LibraryItem struct {
	ID           string        `json:"id"`
	LibraryID    string        `json:"libraryId"`
	FolderID     string        `json:"folderId"`
	Path         string        `json:"path"` // Full path to the audiobook folder
	RelPath      string        `json:"relPath"`
	IsFile       bool          `json:"isFile"`
	MtimeMs      int64         `json:"mtimeMs"`
	CTimeMs      int64         `json:"ctimeMs"`
	BirthtimeMs  int64         `json:"birthtimeMs"`
	AddedAt      int64         `json:"addedAt"`
	UpdatedAt    int64         `json:"updatedAt"`
	IsMissing    bool          `json:"isMissing"`
	IsInvalid    bool          `json:"isInvalid"`
	MediaType    string        `json:"mediaType"`
	Media        Media         `json:"media"`
	LibraryFiles []LibraryFile `json:"libraryFiles,omitempty"`
}

// Media contains the book/podcast metadata
type Media struct {
	ID            string      `json:"id"`
	LibraryItemID string      `json:"libraryItemId"`
	Metadata      Metadata    `json:"metadata"`
	CoverPath     string      `json:"coverPath"`
	EbookFile     *EbookFile  `json:"ebookFile,omitempty"`
	AudioFiles    []AudioFile `json:"audioFiles,omitempty"`
	Tracks        []Track     `json:"tracks,omitempty"`
	Duration      float64     `json:"duration"`
	Size          int64       `json:"size"`
}

// Metadata contains book information
// Note: ABS API returns flattened fields (authorName, seriesName) not object arrays
type Metadata struct {
	Title             string   `json:"title"`
	TitleIgnorePrefix string   `json:"titleIgnorePrefix,omitempty"`
	Subtitle          string   `json:"subtitle,omitempty"`
	Authors           []Author `json:"authors,omitempty"`      // Array format (newer ABS versions)
	AuthorName        string   `json:"authorName,omitempty"`   // Flat format (common)
	AuthorNameLF      string   `json:"authorNameLF,omitempty"` // Last, First format
	Series            []Series `json:"series,omitempty"`       // Array format
	SeriesName        string   `json:"seriesName,omitempty"`   // Flat format
	SeriesSequence    string   `json:"seriesSequence,omitempty"`
	Description       string   `json:"description,omitempty"`
	Publisher         string   `json:"publisher,omitempty"`
	PublishedYear     string   `json:"publishedYear,omitempty"`
	PublishedDate     string   `json:"publishedDate,omitempty"`
	Language          string   `json:"language,omitempty"`
	Genres            []string `json:"genres,omitempty"`
	Tags              []string `json:"tags,omitempty"`
	ASIN              string   `json:"asin,omitempty"`
	ISBN              string   `json:"isbn,omitempty"`
	Explicit          bool     `json:"explicit"`
	Abridged          bool     `json:"abridged"`
	NarratorName      string   `json:"narratorName,omitempty"`
}

// Author represents a book author
type Author struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ImagePath   string `json:"imagePath,omitempty"`
}

// Series represents a book series
type Series struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// LibraryFile represents a file in ABS
type LibraryFile struct {
	Ino       string       `json:"ino"`
	Metadata  FileMetadata `json:"metadata"`
	AddedAt   int64        `json:"addedAt"`
	UpdatedAt int64        `json:"updatedAt"`
	FileType  string       `json:"fileType"`
}

// FileMetadata contains file information
type FileMetadata struct {
	Filename    string `json:"filename"`
	Path        string `json:"path"`
	RelPath     string `json:"relPath"`
	Size        int64  `json:"size"`
	MtimeMs     int64  `json:"mtimeMs"`
	CtimeMs     int64  `json:"ctimeMs"`
	BirthtimeMs int64  `json:"birthtimeMs"`
}

// AudioFile represents an audio file
type AudioFile struct {
	LibraryFile
	TrackNumberFromMeta int    `json:"trackNumFromMeta"`
	DiscNumberFromMeta  int    `json:"discNumFromMeta"`
	Bitrate             int    `json:"bitRate"`
	Codec               string `json:"codec"`
	TimeBase            string `json:"timeBase"`
}

// EbookFile represents an ebook file
type EbookFile struct {
	LibraryFile
}

// Track represents a track in an audiobook
type Track struct {
	Index       int     `json:"index"`
	StartOffset float64 `json:"startOffset"`
	Duration    float64 `json:"duration"`
	Title       string  `json:"title,omitempty"`
	ContentUrl  string  `json:"contentUrl"`
}
