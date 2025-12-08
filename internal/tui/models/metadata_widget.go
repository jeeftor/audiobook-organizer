package models

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// MetadataWidget is a reusable component for displaying audiobook metadata
// It shows raw metadata fields with annotations indicating which fields are being used
type MetadataWidget struct {
	books        []AudioBook
	fieldMapping organizer.FieldMapping
	currentIndex int
	width        int
	height       int
}

// NewMetadataWidget creates a new metadata widget
func NewMetadataWidget(books []AudioBook, fieldMapping organizer.FieldMapping) *MetadataWidget {
	return &MetadataWidget{
		books:        books,
		fieldMapping: fieldMapping,
		currentIndex: 0,
	}
}

// SetBooks updates the books displayed by the widget
func (w *MetadataWidget) SetBooks(books []AudioBook) {
	w.books = books
	if w.currentIndex >= len(books) {
		w.currentIndex = 0
	}
}

// SetFieldMapping updates the field mapping used for annotations
func (w *MetadataWidget) SetFieldMapping(fieldMapping organizer.FieldMapping) {
	w.fieldMapping = fieldMapping
}

// SetSize sets the widget dimensions
func (w *MetadataWidget) SetSize(width, height int) {
	w.width = width
	w.height = height
}

// CurrentIndex returns the current book index
func (w *MetadataWidget) CurrentIndex() int {
	return w.currentIndex
}

// SetCurrentIndex sets the current book index
func (w *MetadataWidget) SetCurrentIndex(index int) {
	if index >= 0 && index < len(w.books) {
		w.currentIndex = index
	}
}

// NextBook moves to the next book (file)
func (w *MetadataWidget) NextBook() {
	if w.currentIndex < len(w.books)-1 {
		w.currentIndex++
	}
}

// PrevBook moves to the previous book (file)
func (w *MetadataWidget) PrevBook() {
	if w.currentIndex > 0 {
		w.currentIndex--
	}
}

// NextBookGroup moves to the first file of the next book group (different album)
func (w *MetadataWidget) NextBookGroup() {
	if len(w.books) == 0 {
		return
	}
	currentAlbum := w.books[w.currentIndex].Metadata.Album
	for i := w.currentIndex + 1; i < len(w.books); i++ {
		if w.books[i].Metadata.Album != currentAlbum {
			w.currentIndex = i
			return
		}
	}
}

// PrevBookGroup moves to the first file of the previous book group (different album)
func (w *MetadataWidget) PrevBookGroup() {
	if len(w.books) == 0 || w.currentIndex == 0 {
		return
	}
	currentAlbum := w.books[w.currentIndex].Metadata.Album
	// First, find start of current group
	startOfCurrent := w.currentIndex
	for startOfCurrent > 0 && w.books[startOfCurrent-1].Metadata.Album == currentAlbum {
		startOfCurrent--
	}
	// If we're not at the start of current group, go there
	if startOfCurrent < w.currentIndex {
		w.currentIndex = startOfCurrent
		return
	}
	// Otherwise find start of previous group
	if startOfCurrent > 0 {
		prevAlbum := w.books[startOfCurrent-1].Metadata.Album
		for i := startOfCurrent - 1; i >= 0; i-- {
			if i == 0 || w.books[i-1].Metadata.Album != prevAlbum {
				w.currentIndex = i
				return
			}
		}
	}
}

// BookCount returns the total number of files
func (w *MetadataWidget) BookCount() int {
	return len(w.books)
}

// BookGroupCount returns the total number of unique book groups (by album)
func (w *MetadataWidget) BookGroupCount() int {
	if len(w.books) == 0 {
		return 0
	}
	seen := make(map[string]bool)
	for _, book := range w.books {
		seen[book.Metadata.Album] = true
	}
	return len(seen)
}

// CurrentBookGroupIndex returns the 1-indexed book group number for the current file
func (w *MetadataWidget) CurrentBookGroupIndex() int {
	if len(w.books) == 0 {
		return 0
	}
	currentAlbum := w.books[w.currentIndex].Metadata.Album
	seen := make(map[string]bool)
	for i := 0; i <= w.currentIndex; i++ {
		seen[w.books[i].Metadata.Album] = true
	}
	// Count how many unique albums we've seen up to and including current
	// But we want the index of the current album, not total seen
	groupIndex := 0
	seenAlbums := make(map[string]bool)
	for i := 0; i < len(w.books); i++ {
		album := w.books[i].Metadata.Album
		if !seenAlbums[album] {
			seenAlbums[album] = true
			groupIndex++
			if album == currentAlbum {
				return groupIndex
			}
		}
	}
	return groupIndex
}

// CurrentBook returns the currently selected book
func (w *MetadataWidget) CurrentBook() *AudioBook {
	if len(w.books) == 0 || w.currentIndex >= len(w.books) {
		return nil
	}
	return &w.books[w.currentIndex]
}

// View renders the metadata widget
func (w *MetadataWidget) View() string {
	if len(w.books) == 0 {
		return "No books selected"
	}

	if w.currentIndex >= len(w.books) {
		w.currentIndex = 0
	}

	book := w.books[w.currentIndex]
	return w.renderMetadata(book)
}

// renderMetadata renders the metadata for a single book
func (w *MetadataWidget) renderMetadata(book AudioBook) string {
	var content strings.Builder

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFAAFF"))
	authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9500")).Bold(true) // Orange, bold
	seriesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Bold(true) // Cyan, bold
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)  // Green, bold
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF66FF")).Bold(true)  // Pink, bold
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))           // Dim
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Usage annotation styles - bright and bold to stand out
	authorUsageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9500")).Bold(true)
	seriesUsageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Bold(true)
	titleUsageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	trackUsageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF66FF")).Bold(true)

	// Header with book group and file index
	pathParts := strings.Split(book.Path, "/")
	displayPath := book.Path
	if len(pathParts) > 2 {
		displayPath = ".../" + strings.Join(pathParts[len(pathParts)-2:], "/")
	}
	bookGroupIdx := w.CurrentBookGroupIndex()
	bookGroupCount := w.BookGroupCount()
	content.WriteString(headerStyle.Render(fmt.Sprintf("Book %d/%d  File %d/%d: ", bookGroupIdx, bookGroupCount, w.currentIndex+1, len(w.books))))
	content.WriteString(valueStyle.Render(displayPath) + "\n\n")

	// Show ALL metadata fields with usage annotations (sorted alphabetically)
	// Include both raw data fields AND standard fields that might be mapped
	allFields := make(map[string]string)

	// Add all raw data fields
	for key, val := range book.Metadata.RawData {
		allFields[key] = fmt.Sprintf("%v", val)
	}

	// Add standard metadata fields that might not be in RawData
	// These are fields that can be selected in field mapping options
	if book.Metadata.Title != "" {
		allFields["title"] = book.Metadata.Title
	} else if _, exists := allFields["title"]; !exists {
		allFields["title"] = ""
	}
	if len(book.Metadata.Authors) > 0 {
		allFields["authors"] = strings.Join(book.Metadata.Authors, ", ")
	} else if _, exists := allFields["authors"]; !exists {
		allFields["authors"] = ""
	}
	if series := book.Metadata.GetValidSeries(); series != "" {
		allFields["series"] = series
	} else if _, exists := allFields["series"]; !exists {
		allFields["series"] = ""
	}
	if book.Metadata.TrackTitle != "" {
		allFields["track_title"] = book.Metadata.TrackTitle
	} else if _, exists := allFields["track_title"]; !exists {
		allFields["track_title"] = ""
	}
	if book.Metadata.TrackNumber != 0 {
		allFields["track_number"] = fmt.Sprintf("%d", book.Metadata.TrackNumber)
	} else if _, exists := allFields["track_number"]; !exists {
		allFields["track_number"] = ""
	}
	// Check for narrators in RawData (not a direct field on Metadata)
	if narrators, ok := book.Metadata.RawData["narrators"]; ok {
		allFields["narrators"] = fmt.Sprintf("%v", narrators)
	} else if _, exists := allFields["narrators"]; !exists {
		allFields["narrators"] = ""
	}

	if len(allFields) > 0 {
		// Get sorted keys for stable iteration order
		keys := make([]string, 0, len(allFields))
		for key := range allFields {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			valStr := allFields[key]

			// Determine styling and usage annotation based on field mapping
			labelStyle := defaultStyle
			usageAnnotation := ""
			usageStyle := defaultStyle // Will be overridden if field is used

			// Check if it's the title field
			if key == w.fieldMapping.TitleField {
				labelStyle = titleStyle
				usageStyle = titleUsageStyle
				usageAnnotation = " ← TITLE"
			}
			// Check if it's the series field
			if key == w.fieldMapping.SeriesField {
				labelStyle = seriesStyle
				usageStyle = seriesUsageStyle
				usageAnnotation = " ← SERIES"
			}
			// Check if it's an author field
			for _, af := range w.fieldMapping.AuthorFields {
				if key == af {
					labelStyle = authorStyle
					usageStyle = authorUsageStyle
					usageAnnotation = " ← AUTHOR"
					break
				}
			}
			// Check if it's the track field
			if key == w.fieldMapping.TrackField {
				labelStyle = trackStyle
				usageStyle = trackUsageStyle
				usageAnnotation = " ← TRACK"
			}

			// Show empty values as dimmed "(empty)"
			displayVal := valStr
			currentValueStyle := valueStyle
			if valStr == "" || valStr == "0" {
				displayVal = "(empty)"
				// Use even dimmer style for empty values
				currentValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Italic(true)
			}

			// Format: "field: value  <- USAGE"
			content.WriteString(fmt.Sprintf("  %s: %s%s\n",
				labelStyle.Render(key),
				currentValueStyle.Render(displayVal),
				usageStyle.Render(usageAnnotation)))
		}
	} else {
		// Fallback if no raw data - show processed metadata
		content.WriteString(defaultStyle.Render("  (No raw metadata available)\n"))
		if book.Metadata.Title != "" {
			content.WriteString(fmt.Sprintf("  %s: %s\n", titleStyle.Render("title"), valueStyle.Render(book.Metadata.Title)))
		}
		if len(book.Metadata.Authors) > 0 {
			content.WriteString(fmt.Sprintf("  %s: %s\n", authorStyle.Render("authors"), valueStyle.Render(strings.Join(book.Metadata.Authors, ", "))))
		}
		if series := book.Metadata.GetValidSeries(); series != "" {
			content.WriteString(fmt.Sprintf("  %s: %s\n", seriesStyle.Render("series"), valueStyle.Render(series)))
		}
	}

	return content.String()
}

// RenderCompact renders a more compact single-line summary
func (w *MetadataWidget) RenderCompact() string {
	if len(w.books) == 0 {
		return "No books"
	}

	book := w.books[w.currentIndex]
	author := book.Metadata.GetFirstAuthor("Unknown")
	title := book.Metadata.Title
	if title == "" {
		title = "Unknown"
	}

	return fmt.Sprintf("%s - %s (%d/%d)", author, title, w.currentIndex+1, len(w.books))
}

// GetFieldValue returns the value of a specific field from the current book's raw metadata
func (w *MetadataWidget) GetFieldValue(fieldName string) string {
	if len(w.books) == 0 || w.currentIndex >= len(w.books) {
		return ""
	}

	book := w.books[w.currentIndex]
	if val, ok := book.Metadata.RawData[fieldName]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}
