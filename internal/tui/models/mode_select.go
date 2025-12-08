package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ScanMode represents different scanning modes
type ScanMode int

const (
	NormalMode ScanMode = iota
	FlatMode
	EmbeddedMetadataMode
)

// ModeSelectModel represents the scan mode selection screen
type ModeSelectModel struct {
	modes    []ModeOption
	cursor   int
	width    int
	height   int
	selected ScanMode
	done     bool
}

// ModeOption represents a selectable scan mode
type ModeOption struct {
	Mode        ScanMode
	Name        string
	Description string
	Icon        string
}

// ModeSelectedMsg is sent when a mode is selected
type ModeSelectedMsg struct {
	Mode        ScanMode
	Flat        bool
	UseEmbedded bool
}

// NewModeSelectModel creates a new mode selection model
func NewModeSelectModel() *ModeSelectModel {
	modes := []ModeOption{
		{
			Mode:        NormalMode,
			Name:        "Normal Mode (Directory-based)",
			Description: "Processes audiobooks organized in directories.\nLooks for metadata.json files or extracts from audio files.\nBest for: Audiobooks already in Author/Title folders",
			Icon:        "📁",
		},
		{
			Mode:        FlatMode,
			Name:        "Flat Mode (Individual Files)",
			Description: "Processes each audio file individually.\nExtracts metadata from each file's embedded tags.\nBest for: Loose files not organized in folders",
			Icon:        "📄",
		},
		{
			Mode:        EmbeddedMetadataMode,
			Name:        "Embedded Metadata Only",
			Description: "Uses only metadata embedded in audio files.\nIgnores metadata.json files even if present.\nBest for: Files with good ID3/M4B tags",
			Icon:        "🏷️",
		},
	}

	return &ModeSelectModel{
		modes:    modes,
		cursor:   0,
		selected: NormalMode,
	}
}

// Init initializes the model
func (m *ModeSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and user input
func (m *ModeSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.modes)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.modes[m.cursor].Mode
			m.done = true
			// Return the selected mode
			return m, func() tea.Msg {
				return ModeSelectedMsg{
					Mode:        m.selected,
					Flat:        m.selected == FlatMode,
					UseEmbedded: m.selected == EmbeddedMetadataMode || m.selected == FlatMode,
				}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI
func (m *ModeSelectModel) View() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	content.WriteString(headerStyle.Render("🎧 SELECT SCAN MODE") + "\n\n")

	// Description
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Italic(true)
	content.WriteString(descStyle.Render("How should the audiobook files be processed?") + "\n\n")

	// Mode options
	for i, mode := range m.modes {
		var optionStyle lipgloss.Style
		cursor := "  "

		if i == m.cursor {
			cursor = "▶ "
			optionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00FF00"))
		} else {
			optionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCCCCC"))
		}

		// Mode name with icon
		content.WriteString(cursor + optionStyle.Render(mode.Icon+" "+mode.Name) + "\n")

		// Description (indented)
		descLines := strings.Split(mode.Description, "\n")
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		if i == m.cursor {
			descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
		}
		for _, line := range descLines {
			content.WriteString("    " + descStyle.Render(line) + "\n")
		}
		content.WriteString("\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))
	content.WriteString("\n" + helpStyle.Render("↑/↓: Navigate • Enter/Space: Select • q: Quit"))

	return content.String()
}

// IsDone returns whether a mode has been selected
func (m *ModeSelectModel) IsDone() bool {
	return m.done
}

// GetSelectedMode returns the selected scan mode
func (m *ModeSelectModel) GetSelectedMode() ScanMode {
	return m.selected
}
