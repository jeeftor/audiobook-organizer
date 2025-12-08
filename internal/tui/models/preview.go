package models

import (
	"fmt"
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
	books             []AudioBook
	config            map[string]string
	fieldMapping      organizer.FieldMapping
	pathPreviewWidget *PathPreviewWidget
	moves             []MovePreview
	lines             []string // Cached rendered lines
	groupStarts       []int    // Line numbers where groups start
	cursor            int      // For compatibility with tests
	width             int
	height            int
	scrollOffset      int
	totalAvailable    int // Total books available (for partial selection detection)
}

// NewPreviewModel creates a new preview model
func NewPreviewModel(books []AudioBook, config map[string]string, fieldMapping organizer.FieldMapping) *PreviewModel {
	return NewPreviewModelWithTotal(books, config, fieldMapping, len(books))
}

// NewPreviewModelWithTotal creates a new preview model with total available count
func NewPreviewModelWithTotal(books []AudioBook, config map[string]string, fieldMapping organizer.FieldMapping, totalAvailable int) *PreviewModel {
	// Create path preview widget
	widget := NewPathPreviewWidget(books, fieldMapping)
	widget.SetLayout(config["Layout"])
	widget.SetOutputDir(config["Output Directory"])

	// Check for rename settings
	if config["Rename Files"] == "Yes" {
		widget.SetRenameFiles(true)
		widget.SetRenamePattern(config["Rename Pattern"])
	}
	if config["Add Track Numbers"] == "Yes" {
		widget.SetAddTrackNumbers(true)
	}

	m := &PreviewModel{
		books:             books,
		config:            config,
		fieldMapping:      fieldMapping,
		pathPreviewWidget: widget,
		totalAvailable:    totalAvailable,
	}

	// Generate move previews and cached lines
	m.generatePreviews()

	return m
}

// Init initializes the model
func (m *PreviewModel) Init() tea.Cmd {
	// Request window size on init
	return tea.WindowSize()
}

// generatePreviews generates previews of file move operations using the widget
func (m *PreviewModel) generatePreviews() {
	// Use widget to generate moves
	m.moves = m.pathPreviewWidget.GetMoves()

	// Cache rendered lines and group starts for efficient scrolling
	m.lines, m.groupStarts = m.pathPreviewWidget.RenderGroupedPreview()
}

// Update handles messages and user input
func (m *PreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// Calculate total lines for scrolling
		totalLines := m.getTotalLines()
		maxVisible := m.height - 8

		switch msg.String() {
		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}

		case "down", "j":
			if m.scrollOffset < totalLines-maxVisible {
				m.scrollOffset++
			}

		case "pgup":
			m.scrollOffset -= maxVisible
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}

		case "pgdown":
			m.scrollOffset += maxVisible
			if m.scrollOffset > totalLines-maxVisible {
				m.scrollOffset = totalLines - maxVisible
			}
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}

		case "p", "{": // Jump to previous book group
			groupStarts := m.getGroupStartLines()
			for i := len(groupStarts) - 1; i >= 0; i-- {
				if groupStarts[i] < m.scrollOffset {
					m.scrollOffset = groupStarts[i]
					break
				}
			}
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}

		case "n", "}": // Jump to next book group
			groupStarts := m.getGroupStartLines()
			for _, start := range groupStarts {
				if start > m.scrollOffset {
					m.scrollOffset = start
					break
				}
			}
			// Clamp to max
			if m.scrollOffset > totalLines-maxVisible {
				m.scrollOffset = totalLines - maxVisible
			}
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}

		case "home", "g":
			m.scrollOffset = 0

		case "end", "G":
			m.scrollOffset = totalLines - maxVisible
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
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
			// Warn if partial selection (selected books != total available)
			partialSelection := len(m.books) < m.totalAvailable
			return NewCommandOutputModelWithSelection(m.books, m.config, m.fieldMapping, partialSelection), nil
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

	// Configuration summary - more compact
	configSummary := fmt.Sprintf("Layout: %s | %d files", m.config["Layout"], len(m.moves))
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Render(configSummary) + "\n\n")

	// Calculate visible lines based on height
	maxVisible := m.height - 8 // Space for header and footer

	// Use cached lines from widget
	lines := m.lines

	// Apply scrolling
	endIdx := m.scrollOffset + maxVisible
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	// Show scroll indicator if needed
	if m.scrollOffset > 0 {
		content.WriteString("↑ Scroll up for more\n")
	}

	// Render visible lines
	for i := m.scrollOffset; i < endIdx; i++ {
		content.WriteString(lines[i] + "\n")
	}

	// Show scroll indicator if needed
	if endIdx < len(lines) {
		content.WriteString("↓ Scroll down for more\n")
	}

	// Footer with help text
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("\n↑/↓: Scroll • n/p: Next/Prev Book • Enter: Process • c: CLI • b: Back • q: Quit")

	content.WriteString(footer)

	return content.String()
}

// getTotalLines returns the total number of cached lines
func (m *PreviewModel) getTotalLines() int {
	return len(m.lines)
}

// getGroupStartLines returns the cached group start line numbers
func (m *PreviewModel) getGroupStartLines() []int {
	return m.groupStarts
}

// GetMoves returns the current move previews
func (m *PreviewModel) GetMoves() []MovePreview {
	return m.moves
}
