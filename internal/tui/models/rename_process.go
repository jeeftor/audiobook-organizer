package models

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// RenameProcessModel executes the rename operations
type RenameProcessModel struct {
	candidates []organizer.RenameCandidate
	config     *organizer.RenamerConfig
	renamer    *organizer.Renamer
	processing bool
	complete   bool
	summary    organizer.RenameSummary
	err        error
}

// NewRenameProcessModel creates a new process model
func NewRenameProcessModel(
	candidates []organizer.RenameCandidate,
	config *organizer.RenamerConfig,
) *RenameProcessModel {
	renamer, _ := organizer.NewRenamer(config)

	return &RenameProcessModel{
		candidates: candidates,
		config:     config,
		renamer:    renamer,
	}
}

// Init starts the processing
func (m *RenameProcessModel) Init() tea.Cmd {
	return m.executeRenames
}

// Update handles messages
func (m *RenameProcessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "q" && m.complete {
			// Allow quitting after completion
			return m, tea.Quit
		}

	case processStartMsg:
		m.processing = true
		return m, nil

	case processCompleteMsg:
		m.processing = false
		m.complete = true
		m.summary = msg.summary
		// Return completion message
		return m, nil

	case processErrorMsg:
		m.err = msg.err
		m.processing = false
		return m, func() tea.Msg {
			return RenameErrorMsg{Err: msg.err}
		}
	}

	return m, nil
}

// View renders the process screen
func (m *RenameProcessModel) View() string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	if m.processing {
		sb.WriteString(titleStyle.Render("⚙️  Processing Renames...") + "\n\n")
		sb.WriteString("Please wait...\n")
	} else if m.complete {
		sb.WriteString(successStyle.Render("✅ Rename Complete!") + "\n\n")

		// Summary
		sb.WriteString(fmt.Sprintf("Files scanned: %d\n", m.summary.FilesScanned))
		sb.WriteString(successStyle.Render(fmt.Sprintf("Files renamed: %d\n", m.summary.FilesRenamed)))

		if m.summary.FilesSkipped > 0 {
			sb.WriteString(fmt.Sprintf("Files skipped: %d\n", m.summary.FilesSkipped))
		}

		if m.summary.ConflictsFound > 0 {
			sb.WriteString(fmt.Sprintf("Conflicts resolved: %d\n", m.summary.ConflictsFound))
		}

		if len(m.summary.Errors) > 0 {
			sb.WriteString("\n" + errorStyle.Render("Errors:") + "\n")
			for _, errMsg := range m.summary.Errors {
				sb.WriteString(errorStyle.Render(fmt.Sprintf("  - %s\n", errMsg)))
			}
		}

		// Log file location
		if !m.config.DryRun && m.summary.FilesRenamed > 0 {
			logPath := filepath.Join(m.config.BaseDir, ".abook-rename.log")
			sb.WriteString(fmt.Sprintf("\n📝 Log file: %s\n", logPath))
			sb.WriteString(fmt.Sprintf("To undo: audiobook-organizer rename --dir=%s --undo\n", m.config.BaseDir))
		}

		// Show CLI command
		sb.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render("CLI Command:") + "\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("To rename these files from the command line, run:") + "\n\n")

		commandStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

		sb.WriteString(commandStyle.Render(m.generateCommand()) + "\n")

		sb.WriteString("\n" + helpStyle.Render("Press Q to exit"))
	} else if m.err != nil {
		sb.WriteString(errorStyle.Render("❌ Error") + "\n\n")
		sb.WriteString(m.err.Error() + "\n")
	} else {
		sb.WriteString("Initializing...\n")
	}

	return sb.String()
}

// generateCommand generates the CLI command to reproduce this rename operation
func (m *RenameProcessModel) generateCommand() string {
	var parts []string

	// Start with the base command
	parts = append(parts, "audiobook-organizer rename")

	// Add directory
	parts = append(parts, fmt.Sprintf("--dir=\"%s\"", m.config.BaseDir))

	// Add template
	if m.config.Template != "" {
		parts = append(parts, fmt.Sprintf("--template=\"%s\"", m.config.Template))
	}

	// Add author format
	if m.config.AuthorFormat != organizer.AuthorFormatFirstLast {
		switch m.config.AuthorFormat {
		case organizer.AuthorFormatLastFirst:
			parts = append(parts, "--author-format=last-first")
		case organizer.AuthorFormatPreserve:
			parts = append(parts, "--author-format=preserve")
		}
	}

	// Add embedded metadata flag
	if m.config.UseEmbeddedMetadata {
		parts = append(parts, "--use-embedded-metadata")
	}

	// Add field mapping if not using defaults
	if !m.config.FieldMapping.IsEmpty() {
		if m.config.FieldMapping.TitleField != "" && m.config.FieldMapping.TitleField != "title" {
			parts = append(parts, fmt.Sprintf("--title-field=%s", m.config.FieldMapping.TitleField))
		}
		if m.config.FieldMapping.SeriesField != "" &&
			m.config.FieldMapping.SeriesField != "series" {
			parts = append(
				parts,
				fmt.Sprintf("--series-field=%s", m.config.FieldMapping.SeriesField),
			)
		}
		if len(m.config.FieldMapping.AuthorFields) > 0 {
			// Check if it's not the default
			defaultAuthors := []string{"authors", "artist", "album_artist"}
			isDefault := len(m.config.FieldMapping.AuthorFields) == len(defaultAuthors)
			if isDefault {
				for i, field := range m.config.FieldMapping.AuthorFields {
					if field != defaultAuthors[i] {
						isDefault = false
						break
					}
				}
			}
			if !isDefault {
				parts = append(
					parts,
					fmt.Sprintf(
						"--author-fields=\"%s\"",
						strings.Join(m.config.FieldMapping.AuthorFields, ","),
					),
				)
			}
		}
		if m.config.FieldMapping.TrackField != "" && m.config.FieldMapping.TrackField != "track" {
			parts = append(parts, fmt.Sprintf("--track-field=%s", m.config.FieldMapping.TrackField))
		}
	}

	// Add other flags
	if m.config.Verbose {
		parts = append(parts, "--verbose")
	}

	if m.config.StrictMode {
		parts = append(parts, "--strict")
	}

	if !m.config.Recursive {
		parts = append(parts, "--no-recursive")
	}

	if m.config.PreservePath {
		parts = append(parts, "--preserve-path")
	}

	// Join with backslash for multi-line formatting
	return strings.Join(parts, " \\\n  ")
}

// executeRenames performs the actual rename operations
func (m *RenameProcessModel) executeRenames() tea.Msg {
	// Execute the rename operation
	if err := m.renamer.Execute(); err != nil {
		return processErrorMsg{err: err}
	}

	// Get summary
	summary := m.renamer.GetSummary()
	return processCompleteMsg{summary: summary}
}

// Internal messages

type processStartMsg struct{}

type processCompleteMsg struct {
	summary organizer.RenameSummary
}

type processErrorMsg struct {
	err error
}
