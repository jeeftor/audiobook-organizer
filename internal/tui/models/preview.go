package models

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// MovePreview represents a preview of a file move operation
type MovePreview struct {
	SourcePath string
	TargetPath string
}

// PreviewModel represents the preview screen
type PreviewModel struct {
	books       []AudioBook
	config      map[string]string
	fieldMapping organizer.FieldMapping
	moves       []MovePreview
	cursor      int
	width       int
	height      int
	scrollOffset int
}

// NewPreviewModel creates a new preview model
func NewPreviewModel(books []AudioBook, config map[string]string, fieldMapping organizer.FieldMapping) *PreviewModel {
	m := &PreviewModel{
		books:        books,
		config:       config,
		fieldMapping: fieldMapping,
		cursor:       0,
	}

	// Generate move previews
	m.generatePreviews()

	return m
}

// Init initializes the model
func (m *PreviewModel) Init() tea.Cmd {
	return nil
}

// generatePreviews generates previews of file move operations
func (m *PreviewModel) generatePreviews() {
	m.moves = []MovePreview{}

	// In a real implementation, we would use the organizer package
	// For now, we'll use our own implementation

	// Generate previews for each book
	for _, book := range m.books {
		// Calculate target path using our helper method
		targetPath := m.CalculateTargetPath(book.Path, book.Metadata)

		// Add to moves
		m.moves = append(m.moves, MovePreview{
			SourcePath: book.Path,
			TargetPath: targetPath,
		})
	}
}

// Update handles messages and user input
func (m *PreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Adjust scroll if necessary
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case "down", "j":
			if m.cursor < len(m.moves)-1 {
				m.cursor++
				// Adjust scroll if necessary
				maxVisible := m.height - 10 // Approximate space for header and footer
				if m.cursor >= m.scrollOffset+maxVisible {
					m.scrollOffset = m.cursor - maxVisible + 1
				}
			}

		case "home":
			m.cursor = 0
			m.scrollOffset = 0

		case "end":
			m.cursor = len(m.moves) - 1
			maxVisible := m.height - 10
			if len(m.moves) > maxVisible {
				m.scrollOffset = len(m.moves) - maxVisible
			}

		case "b", "backspace":
			// Return to the previous screen (settings)
			return NewSettingsModel(m.books), nil

		case "q", "ctrl+c":
			// Quit the application
			return m, tea.Quit

		case "enter":
			// Process files - transition to the processing screen
			return NewProcessModel(m.books, m.config, m.moves, m.fieldMapping), nil
		}
	}

	return m, nil
}

// View renders the UI
func (m *PreviewModel) View() string {
	var content strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("👁️ File Organization Preview")

	content.WriteString(header + "\n\n")

	// Configuration summary
	configSummary := fmt.Sprintf("Layout: %s | Embedded Metadata: %s | Flat Mode: %s",
		m.config["Layout"],
		m.config["Use Embedded Metadata"],
		m.config["Flat Mode"])

	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Render(configSummary) + "\n\n")

	// Preview count
	content.WriteString(fmt.Sprintf("Previewing %d file moves:\n\n", len(m.moves)))

	// Calculate visible range based on height
	maxVisible := m.height - 12 // Approximate space for header and footer
	endIdx := m.scrollOffset + maxVisible
	if endIdx > len(m.moves) {
		endIdx = len(m.moves)
	}

	// Show scroll indicator if needed
	if m.scrollOffset > 0 {
		content.WriteString("↑ Scroll up for more\n")
	}

	// Moves preview
	for i := m.scrollOffset; i < endIdx; i++ {
		move := m.moves[i]

		// Cursor indicator
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		// Format paths
		sourceName := filepath.Base(move.SourcePath)
		sourceDir := filepath.Dir(move.SourcePath)
		targetName := filepath.Base(move.TargetPath)
		targetDir := filepath.Dir(move.TargetPath)

		// Style based on cursor position
		var sourceStyle, targetStyle lipgloss.Style
		if i == m.cursor {
			sourceStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
			targetStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
		} else {
			sourceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
			targetStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAFFAA"))
		}

		// Add the move preview
		content.WriteString(fmt.Sprintf("%s From: %s/%s\n",
			cursor,
			sourceStyle.Render(sourceDir),
			sourceStyle.Render(sourceName)))

		content.WriteString(fmt.Sprintf("  To:   %s/%s\n\n",
			targetStyle.Render(targetDir),
			targetStyle.Render(targetName)))
	}

	// Show scroll indicator if needed
	if endIdx < len(m.moves) {
		content.WriteString("↓ Scroll down for more\n")
	}

	// Footer with help text
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("\n↑/↓: Navigate • Enter: Process Files • b: Back • q: Quit")

	content.WriteString(footer)

	return content.String()
}

// CalculateTargetPath calculates the target path based on metadata and field mapping
func (m *PreviewModel) CalculateTargetPath(sourcePath string, metadata organizer.Metadata) string {
	// Use the field mapping to extract the correct metadata fields
	layout := m.config["Layout"]
	outputDir := m.config["Output Directory"]
	if outputDir == "" {
		outputDir = "output"
	}

	// Get author using field mapping
	author := "Unknown"
	for _, field := range m.fieldMapping.AuthorFields {
		if val, ok := metadata.RawData[field]; ok && val != nil {
			if strVal, ok := val.(string); ok && strVal != "" {
				author = strVal
				break
			}
		}
	}
	// Fallback to Authors field if not found in raw data
	if author == "Unknown" && len(metadata.Authors) > 0 {
		author = metadata.Authors[0]
	}

	// Get title using field mapping
	title := ""
	if val, ok := metadata.RawData[m.fieldMapping.TitleField]; ok && val != nil {
		if strVal, ok := val.(string); ok && strVal != "" {
			title = strVal
		}
	}
	// Fallback to metadata.Title if not found in raw data
	if title == "" {
		title = metadata.Title
	}
	// Fallback to filename if title is empty
	if title == "" {
		title = filepath.Base(sourcePath)
	}

	// Get series using field mapping
	series := ""
	if val, ok := metadata.RawData[m.fieldMapping.SeriesField]; ok && val != nil {
		if strVal, ok := val.(string); ok && strVal != "" {
			series = strVal
		}
	}
	// Fallback to GetValidSeries if not found in raw data
	if series == "" {
		series = metadata.GetValidSeries()
	}

	// Calculate path based on layout
	switch layout {
	case "author-only":
		return filepath.Join(outputDir, author, filepath.Base(sourcePath))
	case "author-title":
		return filepath.Join(outputDir, author, title, filepath.Base(sourcePath))
	case "author-series-title":
		if series != "" {
			return filepath.Join(outputDir, author, series, title, filepath.Base(sourcePath))
		}
		return filepath.Join(outputDir, author, title, filepath.Base(sourcePath))
	default:
		return filepath.Join(outputDir, filepath.Base(sourcePath))
	}
}

// GetMoves returns the current move previews
func (m *PreviewModel) GetMoves() []MovePreview {
	return m.moves
}
