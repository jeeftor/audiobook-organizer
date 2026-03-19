package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// RenamePreviewModel shows all proposed renames
type RenamePreviewModel struct {
	candidates   []organizer.RenameCandidate
	config       *organizer.RenamerConfig
	renamer      *organizer.Renamer
	scrollOffset int
	width        int
	height       int
}

// NewRenamePreviewModel creates a new preview model
func NewRenamePreviewModel(
	candidates []organizer.RenameCandidate,
	config *organizer.RenamerConfig,
) *RenamePreviewModel {
	renamer, _ := organizer.NewRenamer(config)

	return &RenamePreviewModel{
		candidates: candidates,
		config:     config,
		renamer:    renamer,
	}
}

// Init initializes the preview model
func (m *RenamePreviewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *RenamePreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q", "esc":
			// Go back to template editor
			return m, nil

		case "enter", "y":
			// Confirm and proceed to processing
			return m, func() tea.Msg {
				return RenamePreviewConfirmedMsg{}
			}

		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}

		case "down", "j":
			maxScroll := len(m.candidates) - 10
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollOffset < maxScroll {
				m.scrollOffset++
			}

		case "pgup":
			m.scrollOffset -= 10
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}

		case "pgdown":
			m.scrollOffset += 10
			maxScroll := len(m.candidates) - 10
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollOffset > maxScroll {
				m.scrollOffset = maxScroll
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the preview screen
func (m *RenamePreviewModel) View() string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	sb.WriteString(titleStyle.Render("👀 Preview Changes") + "\n\n")

	// Summary
	totalFiles := len(m.candidates)
	conflicts := 0
	for _, c := range m.candidates {
		if c.IsConflict {
			conflicts++
		}
	}

	sb.WriteString(fmt.Sprintf("Total files: %d\n", totalFiles))
	if conflicts > 0 {
		sb.WriteString(
			warningStyle.Render(fmt.Sprintf("Conflicts: %d (will be auto-resolved)\n", conflicts)),
		)
	}
	sb.WriteString("\n")

	// File list (paginated)
	sb.WriteString(titleStyle.Render("Changes:") + "\n")
	visibleStart := m.scrollOffset
	visibleEnd := m.scrollOffset + 15
	if visibleEnd > len(m.candidates) {
		visibleEnd = len(m.candidates)
	}

	for i := visibleStart; i < visibleEnd; i++ {
		candidate := m.candidates[i]
		fromFile := candidate.CurrentPath
		toFile := candidate.ProposedPath

		conflictMarker := ""
		if candidate.IsConflict {
			conflictMarker = warningStyle.Render(" [conflict]")
		}

		sb.WriteString(fmt.Sprintf("  %s\n", fromFile))
		sb.WriteString(successStyle.Render(fmt.Sprintf("  → %s%s\n", toFile, conflictMarker)))
		sb.WriteString("\n")
	}

	if totalFiles > 15 {
		sb.WriteString(
			helpStyle.Render(
				fmt.Sprintf(
					"(Showing %d-%d of %d) ↑↓ to scroll\n\n",
					visibleStart+1,
					visibleEnd,
					totalFiles,
				),
			),
		)
	}

	// Controls
	sb.WriteString(helpStyle.Render("Enter/Y: Proceed | Q: Back to edit | ↑↓: Scroll"))

	return sb.String()
}
