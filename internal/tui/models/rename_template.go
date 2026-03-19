package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// RenameTemplateModel handles interactive template building
type RenameTemplateModel struct {
	candidates      []organizer.RenameCandidate
	config          *organizer.RenamerConfig
	templateInput   textinput.Model
	authorFormat    organizer.AuthorFormat
	preview         []string
	showHelp        bool
	showMetadata    bool
	availableFields map[string]bool
	width           int
	height          int
}

// NewRenameTemplateModel creates a new template model
func NewRenameTemplateModel(
	candidates []organizer.RenameCandidate,
	config *organizer.RenamerConfig,
) *RenameTemplateModel {
	ti := textinput.New()
	ti.Placeholder = "Enter template..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60
	ti.SetValue(config.Template)

	// Detect available fields from candidates
	availableFields := make(map[string]bool)
	for _, candidate := range candidates {
		m := candidate.Metadata
		if m.Title != "" {
			availableFields["title"] = true
		}
		if len(m.Authors) > 0 {
			availableFields["author"] = true
			availableFields["authors"] = true
		}
		if len(m.Series) > 0 {
			availableFields["series"] = true
			availableFields["series_number"] = true
		}
		if m.TrackNumber > 0 {
			availableFields["track"] = true
		}
		if m.Album != "" {
			availableFields["album"] = true
		}
		if val, ok := m.RawData["year"]; ok && val != nil {
			availableFields["year"] = true
		}
		if val, ok := m.RawData["narrator"]; ok && val != nil {
			availableFields["narrator"] = true
		}
	}

	return &RenameTemplateModel{
		candidates:      candidates,
		config:          config,
		templateInput:   ti,
		authorFormat:    config.AuthorFormat,
		showHelp:        true,
		showMetadata:    false,
		availableFields: availableFields,
	}
}

// Init initializes the template model
func (m *RenameTemplateModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m *RenameTemplateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q", "esc":
			// Go back
			return m, nil

		case "enter":
			// Confirm template and proceed
			return m, func() tea.Msg {
				return RenameTemplateConfirmedMsg{
					Template:     m.templateInput.Value(),
					AuthorFormat: m.authorFormat,
				}
			}

		case "?":
			// Toggle help
			m.showHelp = !m.showHelp
			return m, nil

		case "m":
			// Toggle metadata display
			m.showMetadata = !m.showMetadata
			return m, nil

		case "tab":
			// Cycle through author formats
			switch m.authorFormat {
			case organizer.AuthorFormatFirstLast:
				m.authorFormat = organizer.AuthorFormatLastFirst
			case organizer.AuthorFormatLastFirst:
				m.authorFormat = organizer.AuthorFormatPreserve
			case organizer.AuthorFormatPreserve:
				m.authorFormat = organizer.AuthorFormatFirstLast
			}
			m.updatePreview()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update text input
	m.templateInput, cmd = m.templateInput.Update(msg)
	m.updatePreview()

	return m, cmd
}

// View renders the template builder screen
func (m *RenameTemplateModel) View() string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	sb.WriteString(titleStyle.Render("📝 Template Builder") + "\n\n")

	// Template input
	sb.WriteString("Template: " + m.templateInput.View() + "\n\n")

	// Author format selector
	formatStr := m.getAuthorFormatString()
	sb.WriteString(fmt.Sprintf("Author Format (Tab to cycle): %s\n\n", formatStr))

	// Preview
	sb.WriteString(titleStyle.Render("Preview:") + "\n")
	if len(m.preview) > 0 {
		for i, p := range m.preview {
			if i >= 10 {
				break // Limit to 10 preview items
			}
			sb.WriteString(fmt.Sprintf("  %s\n", p))
		}
	} else {
		sb.WriteString("  (No preview available)\n")
	}
	sb.WriteString("\n")

	// Help
	if m.showHelp {
		sb.WriteString(titleStyle.Render("Available Fields:") + "\n")
		sb.WriteString("  {author} {authors} {title} {series} {series_number}\n")
		sb.WriteString("  {track} {album} {year} {narrator}\n\n")
	}

	// Metadata display
	if m.showMetadata && len(m.candidates) > 0 {
		sb.WriteString(titleStyle.Render("Sample Metadata (First File):") + "\n")
		sample := m.candidates[0].Metadata
		sb.WriteString(fmt.Sprintf("  Title: %s\n", sample.Title))
		if len(sample.Authors) > 0 {
			sb.WriteString(fmt.Sprintf("  Authors: %v\n", sample.Authors))
		}
		if len(sample.Series) > 0 {
			sb.WriteString(fmt.Sprintf("  Series: %v\n", sample.Series))
		}
		if sample.TrackNumber > 0 {
			sb.WriteString(fmt.Sprintf("  Track: %d\n", sample.TrackNumber))
		}
		if sample.Album != "" {
			sb.WriteString(fmt.Sprintf("  Album: %s\n", sample.Album))
		}
		if val, ok := sample.RawData["year"]; ok {
			sb.WriteString(fmt.Sprintf("  Year: %v\n", val))
		}
		if val, ok := sample.RawData["narrator"]; ok {
			sb.WriteString(fmt.Sprintf("  Narrator: %v\n", val))
		}
		sb.WriteString(fmt.Sprintf("  Source: %s\n\n", sample.SourceType))
	}

	// Controls
	sb.WriteString(
		helpStyle.Render("Enter: Continue | Tab: Change format | ?: Help | m: Metadata | Q: Back"),
	)

	return sb.String()
}

// updatePreview regenerates the preview
func (m *RenameTemplateModel) updatePreview() {
	m.preview = []string{}

	// Parse template
	template, err := organizer.ParseTemplate(m.templateInput.Value())
	if err != nil {
		m.preview = []string{"Error: " + err.Error()}
		return
	}

	// Create renderer
	formatter := organizer.NewAuthorFormatter(m.authorFormat)
	renderer := organizer.NewTemplateRenderer(template, formatter)

	// Generate preview for first 10 candidates
	for i, candidate := range m.candidates {
		if i >= 10 {
			break
		}

		newFilename, err := renderer.Render(candidate.Metadata)
		if err != nil {
			continue
		}

		// Use the current filename from the path
		currentFilename := candidate.CurrentPath
		if len(currentFilename) > 50 {
			currentFilename = "..." + currentFilename[len(currentFilename)-47:]
		}

		m.preview = append(m.preview, fmt.Sprintf("%s → %s",
			currentFilename, newFilename))
	}
}

// getAuthorFormatString returns a display string for the author format
func (m *RenameTemplateModel) getAuthorFormatString() string {
	switch m.authorFormat {
	case organizer.AuthorFormatFirstLast:
		return "First Last"
	case organizer.AuthorFormatLastFirst:
		return "Last, First"
	case organizer.AuthorFormatPreserve:
		return "Preserve Original"
	default:
		return "Unknown"
	}
}
