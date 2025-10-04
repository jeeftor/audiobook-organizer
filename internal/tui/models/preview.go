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
	// Request window size on init
	return tea.WindowSize()
}

// generatePreviews generates previews of file move operations
func (m *PreviewModel) generatePreviews() {
	m.moves = []MovePreview{}

	// In a real implementation, we would use the organizer package
	// For now, we'll use our own implementation

	// Generate previews for each book
	for _, book := range m.books {
		// Calculate target path using universal function
		layout := m.config["Layout"]
		outputDir := m.config["Output Directory"]
		if outputDir == "" {
			outputDir = "output"
		}
		targetPath := GenerateOutputPath(book, layout, m.fieldMapping, outputDir)

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
			// Don't handle here - let main.go handle going back to settings
			// Just consume the key
			return m, nil

		case "q", "ctrl+c":
			// Don't handle here - let main.go handle going back to settings
			// Just consume the key
			return m, nil

		case "enter":
			// Process files - transition to the processing screen
			return NewProcessModel(m.books, m.config, m.moves, m.fieldMapping), nil

		case "c":
			// Show CLI command instead of processing
			return NewCommandOutputModel(m.books, m.config, m.fieldMapping), nil
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

		// Style for source path based on cursor position
		var sourceStyle lipgloss.Style
		if i == m.cursor {
			sourceStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
		} else {
			sourceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
		}

		// Add the move preview
		content.WriteString(fmt.Sprintf("%s From: %s/%s\n",
			cursor,
			sourceStyle.Render(sourceDir),
			sourceStyle.Render(sourceName)))

		// Colorize the output path
		coloredTarget := m.colorizeOutputPath(move.TargetPath, m.config["Layout"])
		content.WriteString(fmt.Sprintf("  To:   %s\n\n", coloredTarget))
	}

	// Show scroll indicator if needed
	if endIdx < len(m.moves) {
		content.WriteString("↓ Scroll down for more\n")
	}

	// Footer with help text
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("\n↑/↓: Navigate • Enter: Process Files • c: Show CLI Command • b: Back • q: Quit")

	content.WriteString(footer)

	return content.String()
}

// GetMoves returns the current move previews
func (m *PreviewModel) GetMoves() []MovePreview {
	return m.moves
}

// colorizeOutputPath colorizes the output path components based on the layout
func (m *PreviewModel) colorizeOutputPath(path string, layout string) string {
	// Define color styles
	authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9500"))   // Orange
	seriesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF"))   // Cyan
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))    // Green
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")) // Gray

	// Split the path into components
	parts := strings.Split(path, string(filepath.Separator))

	// Skip the output directory (first component) and process the rest
	var coloredParts []string
	if len(parts) > 1 {
		parts = parts[1:] // Skip output directory
	}

	// Apply colors based on layout
	switch layout {
	case "author-only":
		if len(parts) >= 1 {
			coloredParts = []string{
				authorStyle.Render(parts[0]),
			}
			if len(parts) > 1 {
				coloredParts = append(coloredParts, parts[1:]...)
			}
		}
	case "author-title":
		if len(parts) >= 2 {
			coloredParts = []string{
				authorStyle.Render(parts[0]),
				titleStyle.Render(parts[1]),
			}
			if len(parts) > 2 {
				coloredParts = append(coloredParts, parts[2:]...)
			}
		}
	case "author-series-title":
		if len(parts) >= 3 {
			coloredParts = []string{
				authorStyle.Render(parts[0]),
				seriesStyle.Render(parts[1]),
				titleStyle.Render(parts[2]),
			}
			if len(parts) > 3 {
				coloredParts = append(coloredParts, parts[3:]...)
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
		coloredParts = parts
	}

	if len(coloredParts) == 0 {
		return path
	}

	// Join with colorized separators
	return strings.Join(coloredParts, separatorStyle.Render("/"))
}
