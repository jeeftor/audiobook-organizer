package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// CommandOutputModel represents the command output screen
type CommandOutputModel struct {
	books            []AudioBook
	config           map[string]string
	fieldMapping     organizer.FieldMapping
	command          string
	width            int
	height           int
	partialSelection bool // True if not all available books were selected
}

// NewCommandOutputModel creates a new command output model
func NewCommandOutputModel(books []AudioBook, config map[string]string, fieldMapping organizer.FieldMapping) *CommandOutputModel {
	return NewCommandOutputModelWithSelection(books, config, fieldMapping, false)
}

// NewCommandOutputModelWithSelection creates a new command output model with partial selection info
func NewCommandOutputModelWithSelection(books []AudioBook, config map[string]string, fieldMapping organizer.FieldMapping, partialSelection bool) *CommandOutputModel {
	m := &CommandOutputModel{
		books:            books,
		config:           config,
		fieldMapping:     fieldMapping,
		partialSelection: partialSelection,
	}

	// Generate the CLI command
	m.command = m.generateCommand()

	return m
}

// Init initializes the model
func (m *CommandOutputModel) Init() tea.Cmd {
	return nil
}

// generateCommand generates the CLI command based on current settings
func (m *CommandOutputModel) generateCommand() string {
	var parts []string

	// Start with the base command
	parts = append(parts, "audiobook-organizer")

	// Add input directory
	if inputDir := m.config["Input Directory"]; inputDir != "" {
		parts = append(parts, fmt.Sprintf("--dir=\"%s\"", inputDir))
	}

	// Add output directory
	if outputDir := m.config["Output Directory"]; outputDir != "" {
		parts = append(parts, fmt.Sprintf("--out=\"%s\"", outputDir))
	}

	// Add layout
	if layout := m.config["Layout"]; layout != "" && layout != "author-series-title" {
		parts = append(parts, fmt.Sprintf("--layout=%s", layout))
	}

	// Add embedded metadata flag
	if useEmbedded := m.config["Use Embedded Metadata"]; useEmbedded == "Yes" {
		parts = append(parts, "--use-embedded-metadata")
	}

	// Add flat mode flag
	if flatMode := m.config["Flat Mode"]; flatMode == "Yes" {
		parts = append(parts, "--flat")
	}

	// Add verbose flag (but not dry-run yet - we'll add that last)
	if verbose := m.config["Verbose"]; verbose == "Yes" {
		parts = append(parts, "--verbose")
	}

	// Add field mapping if not using defaults
	if !m.fieldMapping.IsEmpty() {
		if m.fieldMapping.TitleField != "" && m.fieldMapping.TitleField != "title" {
			parts = append(parts, fmt.Sprintf("--title-field=%s", m.fieldMapping.TitleField))
		}
		if m.fieldMapping.SeriesField != "" && m.fieldMapping.SeriesField != "series" {
			parts = append(parts, fmt.Sprintf("--series-field=%s", m.fieldMapping.SeriesField))
		}
		if len(m.fieldMapping.AuthorFields) > 0 {
			// Check if it's not the default
			defaultAuthors := []string{"authors", "artist", "album_artist"}
			if !equalStringSlices(m.fieldMapping.AuthorFields, defaultAuthors) {
				parts = append(parts, fmt.Sprintf("--author-fields=\"%s\"", strings.Join(m.fieldMapping.AuthorFields, ",")))
			}
		}
		if m.fieldMapping.TrackField != "" && m.fieldMapping.TrackField != "track" {
			parts = append(parts, fmt.Sprintf("--track-field=%s", m.fieldMapping.TrackField))
		}
	}

	// ALWAYS add --dry-run as the last flag for safety
	// Users can remove it if they're ready to actually move files
	parts = append(parts, "--dry-run")

	// Join with backslash for multi-line formatting
	return strings.Join(parts, " \\\n  ")
}

// equalStringSlices compares two string slices for equality
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Update handles messages and user input
func (m *CommandOutputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			// Quit the application
			return m, tea.Quit

		case "b", "backspace":
			// Go back to preview
			return NewPreviewModel(m.books, m.config, m.fieldMapping), nil
		}
	}

	return m, nil
}

// View renders the UI
func (m *CommandOutputModel) View() string {
	var content strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("📋 Generated CLI Command")

	content.WriteString(header + "\n\n")

	// Description
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	content.WriteString(descStyle.Render("Copy and paste this command to run with the same settings:") + "\n")

	// Safety note
	noteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8800")).Italic(true)
	content.WriteString(noteStyle.Render("Note: --dry-run is always included for safety. Remove it to actually move files.") + "\n")

	// Warning for partial selection
	if m.partialSelection {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
		content.WriteString("\n" + warningStyle.Render("⚠️  WARNING: You selected only some books in the GUI.") + "\n")
		content.WriteString(noteStyle.Render("   The CLI command below will process ALL files in the input directory.") + "\n")
		content.WriteString(noteStyle.Render("   To process only specific files, use the GUI's Enter key instead.") + "\n")
	}
	content.WriteString("\n")

	// Separator line
	content.WriteString(strings.Repeat("─", 80) + "\n")

	// Command in plain text (easy to copy)
	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	content.WriteString(commandStyle.Render(m.command) + "\n")

	// Separator line
	content.WriteString(strings.Repeat("─", 80) + "\n\n")

	// Configuration summary
	content.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Render("Configuration Summary:") + "\n\n")

	summaryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAFF"))

	// Show key settings with proper alignment
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Input Directory:"), summaryStyle.Render(m.config["Input Directory"])))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Output Directory:"), summaryStyle.Render(m.config["Output Directory"])))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Layout:"), summaryStyle.Render(m.config["Layout"])))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Embedded Metadata:"), summaryStyle.Render(m.config["Use Embedded Metadata"])))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Flat Mode:"), summaryStyle.Render(m.config["Flat Mode"])))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Dry Run:"), summaryStyle.Render(m.config["Dry Run"])))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Verbose:"), summaryStyle.Render(m.config["Verbose"])))

	// Show field mapping if customized
	if !m.fieldMapping.IsEmpty() {
		content.WriteString("\n" + lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFFF")).
			Render("Field Mapping:") + "\n\n")

		if m.fieldMapping.TitleField != "" {
			content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Title Field:"), summaryStyle.Render(m.fieldMapping.TitleField)))
		}
		if m.fieldMapping.SeriesField != "" {
			content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Series Field:"), summaryStyle.Render(m.fieldMapping.SeriesField)))
		}
		if len(m.fieldMapping.AuthorFields) > 0 {
			content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Author Fields:"), summaryStyle.Render(strings.Join(m.fieldMapping.AuthorFields, ", "))))
		}
		if m.fieldMapping.TrackField != "" {
			content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Track Field:"), summaryStyle.Render(m.fieldMapping.TrackField)))
		}
	}

	// Statistics
	content.WriteString("\n" + lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Render("Statistics:") + "\n\n")

	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Books selected:"), summaryStyle.Render(fmt.Sprintf("%d", len(m.books)))))

	// Footer with help text
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("\n\nb: Back • q: Quit")

	content.WriteString(footer)

	return content.String()
}
