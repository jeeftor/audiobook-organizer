package models

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// PathPreviewWidget generates colorized output path previews
type PathPreviewWidget struct {
	books           []AudioBook
	fieldMapping    organizer.FieldMapping
	layout          string
	outputDir       string
	addTrackNumbers bool
	renameFiles     bool
	renamePattern   string
	currentIndex    int // Current book index for preview display
}

// NewPathPreviewWidget creates a new path preview widget
func NewPathPreviewWidget(books []AudioBook, fieldMapping organizer.FieldMapping) *PathPreviewWidget {
	return &PathPreviewWidget{
		books:           books,
		fieldMapping:    fieldMapping,
		layout:          "author-series-title",
		outputDir:       "output",
		addTrackNumbers: false,
		renameFiles:     false,
		renamePattern:   "{track} - {title}",
	}
}

// SetBooks updates the books
func (w *PathPreviewWidget) SetBooks(books []AudioBook) {
	w.books = books
}

// SetFieldMapping updates the field mapping
func (w *PathPreviewWidget) SetFieldMapping(fieldMapping organizer.FieldMapping) {
	w.fieldMapping = fieldMapping
}

// SetLayout sets the output layout
func (w *PathPreviewWidget) SetLayout(layout string) {
	w.layout = layout
}

// SetOutputDir sets the output directory
func (w *PathPreviewWidget) SetOutputDir(outputDir string) {
	w.outputDir = outputDir
}

// SetAddTrackNumbers enables/disables track number prefixing
func (w *PathPreviewWidget) SetAddTrackNumbers(enabled bool) {
	w.addTrackNumbers = enabled
}

// SetRenameFiles enables/disables file renaming
func (w *PathPreviewWidget) SetRenameFiles(enabled bool) {
	w.renameFiles = enabled
}

// SetRenamePattern sets the rename pattern
func (w *PathPreviewWidget) SetRenamePattern(pattern string) {
	w.renamePattern = pattern
}

// SetCurrentIndex sets the current book index for preview display
func (w *PathPreviewWidget) SetCurrentIndex(index int) {
	if index >= 0 && index < len(w.books) {
		w.currentIndex = index
	}
}

// GetTrackNumber extracts track number from a book using multiple sources
func (w *PathPreviewWidget) GetTrackNumber(book AudioBook) int {
	trackNum := book.TrackNumber
	if trackNum == 0 {
		trackNum = book.Metadata.TrackNumber
	}
	if trackNum == 0 {
		// Try raw metadata
		if rawTrack, ok := book.Metadata.RawData["track"].(float64); ok {
			trackNum = int(rawTrack)
		} else if rawTrack, ok := book.Metadata.RawData["track_number"].(float64); ok {
			trackNum = int(rawTrack)
		} else if rawTrack, ok := book.Metadata.RawData["track"].(int); ok {
			trackNum = rawTrack
		} else if rawTrack, ok := book.Metadata.RawData["track_number"].(int); ok {
			trackNum = rawTrack
		}
	}
	return trackNum
}

// GeneratePath generates the output path for a single book
func (w *PathPreviewWidget) GeneratePath(book AudioBook) string {
	// Generate base output path
	outputPath := GenerateOutputPath(book, w.layout, w.fieldMapping, w.outputDir)

	// Apply track number prefix or rename if enabled
	if w.addTrackNumbers || w.renameFiles {
		dir := filepath.Dir(outputPath)
		base := filepath.Base(outputPath)
		ext := filepath.Ext(base)
		nameWithoutExt := strings.TrimSuffix(base, ext)
		trackNum := w.GetTrackNumber(book)

		if w.renameFiles {
			// Apply rename pattern
			newName := w.renamePattern
			newName = strings.ReplaceAll(newName, "{track}", fmt.Sprintf("%02d", trackNum))
			newName = strings.ReplaceAll(newName, "{title}", book.Metadata.Title)
			newName = strings.ReplaceAll(newName, "{author}", book.Metadata.GetFirstAuthor("Unknown"))
			base = newName + ext
		} else if w.addTrackNumbers {
			// Just prefix with track number
			base = fmt.Sprintf("%02d - %s", trackNum, nameWithoutExt) + ext
		}

		outputPath = filepath.Join(dir, base)
	}

	return outputPath
}

// GenerateColorizedPath generates a colorized output path for a single book
func (w *PathPreviewWidget) GenerateColorizedPath(book AudioBook) string {
	outputPath := w.GeneratePath(book)
	return w.ColorizePath(outputPath)
}

// ColorizePath applies colors to path components based on layout
func (w *PathPreviewWidget) ColorizePath(path string) string {
	// Define color styles
	authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9500"))    // Orange
	seriesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF"))    // Cyan
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))     // Green
	fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))      // White
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")) // Gray

	// Split the path into components
	parts := strings.Split(path, string(filepath.Separator))

	// Skip the output directory (first component) and process the rest
	var coloredParts []string
	if len(parts) > 1 {
		parts = parts[1:] // Skip output directory
	}

	// Apply colors based on layout
	switch w.layout {
	case "author-only":
		if len(parts) >= 1 {
			coloredParts = []string{authorStyle.Render(parts[0])}
			for i := 1; i < len(parts); i++ {
				coloredParts = append(coloredParts, fileStyle.Render(parts[i]))
			}
		}
	case "author-title":
		if len(parts) >= 2 {
			coloredParts = []string{
				authorStyle.Render(parts[0]),
				titleStyle.Render(parts[1]),
			}
			for i := 2; i < len(parts); i++ {
				coloredParts = append(coloredParts, fileStyle.Render(parts[i]))
			}
		}
	case "author-series-title", "author-series-title-number":
		if len(parts) >= 3 {
			coloredParts = []string{
				authorStyle.Render(parts[0]),
				seriesStyle.Render(parts[1]),
				titleStyle.Render(parts[2]),
			}
			for i := 3; i < len(parts); i++ {
				coloredParts = append(coloredParts, fileStyle.Render(parts[i]))
			}
		} else if len(parts) == 2 {
			// Fallback when no series
			coloredParts = []string{
				authorStyle.Render(parts[0]),
				titleStyle.Render(parts[1]),
			}
		}
	default:
		// No colorization for unknown layouts
		for _, part := range parts {
			coloredParts = append(coloredParts, fileStyle.Render(part))
		}
	}

	if len(coloredParts) == 0 {
		return path
	}

	// Join with colorized separators
	return strings.Join(coloredParts, separatorStyle.Render("/"))
}

// RenderCompactPreview renders a compact preview (for settings screen)
// Shows files starting from currentIndex
func (w *PathPreviewWidget) RenderCompactPreview(maxBooks int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))
	content.WriteString(titleStyle.Render("Output Path Preview:") + "\n")

	if len(w.books) == 0 {
		return content.String()
	}

	// Start from current index
	startIdx := w.currentIndex
	if startIdx >= len(w.books) {
		startIdx = 0
	}

	// Calculate how many to show
	previewCount := maxBooks
	remaining := len(w.books) - startIdx
	if remaining < previewCount {
		previewCount = remaining
	}

	for i := 0; i < previewCount; i++ {
		book := w.books[startIdx+i]
		coloredPath := w.GenerateColorizedPath(book)
		content.WriteString("  " + coloredPath + "\n")
	}

	// Show count of remaining files
	totalRemaining := len(w.books) - startIdx - previewCount
	if totalRemaining > 0 {
		moreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
		content.WriteString(moreStyle.Render(fmt.Sprintf("  ... and %d more", totalRemaining)) + "\n")
	}

	return content.String()
}

// RenderGroupedPreview renders a grouped preview (for preview screen)
func (w *PathPreviewWidget) RenderGroupedPreview() ([]string, []int) {
	// Group by destination directory
	type moveGroup struct {
		destDir string
		books   []AudioBook
		paths   []string
	}

	groupMap := make(map[string]*moveGroup)
	var order []string

	for _, book := range w.books {
		outputPath := w.GeneratePath(book)
		destDir := filepath.Dir(outputPath)

		if _, exists := groupMap[destDir]; !exists {
			order = append(order, destDir)
			groupMap[destDir] = &moveGroup{destDir: destDir}
		}
		groupMap[destDir].books = append(groupMap[destDir].books, book)
		groupMap[destDir].paths = append(groupMap[destDir].paths, outputPath)
	}

	// Build lines and track group start positions
	var lines []string
	var groupStarts []int

	for _, dir := range order {
		group := groupMap[dir]
		groupStarts = append(groupStarts, len(lines))

		// Group header
		coloredDir := w.ColorizePath(dir + "/")
		lines = append(lines, fmt.Sprintf("📁 %s (%d files)", coloredDir, len(group.books)))

		// Files in group
		for i, book := range group.books {
			srcFile := filepath.Base(book.Path)
			dstFile := filepath.Base(group.paths[i])

			if srcFile == dstFile {
				lines = append(lines, fmt.Sprintf("   %s", srcFile))
			} else {
				lines = append(lines, fmt.Sprintf("   %s → %s", srcFile, dstFile))
			}
		}

		lines = append(lines, "") // Blank line between groups
	}

	return lines, groupStarts
}

// GetMoves returns source->target path pairs for all books
func (w *PathPreviewWidget) GetMoves() []MovePreview {
	var moves []MovePreview
	for _, book := range w.books {
		moves = append(moves, MovePreview{
			SourcePath: book.Path,
			TargetPath: w.GeneratePath(book),
		})
	}
	return moves
}
