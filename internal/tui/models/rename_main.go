package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// RenameScreen represents different screens in the rename TUI
type RenameScreen int

const (
	RenameScanScreen RenameScreen = iota
	RenameFieldMappingScreen
	RenameCommandScreen
	RenameTemplateScreen
	RenamePreviewScreen
	RenameProcessScreen
)

// RenameMainModel is the main coordinator for the rename TUI
type RenameMainModel struct {
	inputDir string
	screen   RenameScreen
	width    int
	height   int

	// Sub-models for different screens
	scanModel         *RenameScanModel
	fieldMappingModel *RenameFieldMappingModel
	commandModel      string
	dryRunCmd         string
	outCmd            string
	templateModel     *RenameTemplateModel
	previewModel      *RenamePreviewModel
	processModel      *RenameProcessModel

	// Shared state
	candidates []organizer.RenameCandidate
	config     *organizer.RenamerConfig
	quitting   bool
	err        error
}

// NewRenameMainModel creates a new rename main model
func NewRenameMainModel(inputDir string, useEmbeddedMetadata bool) *RenameMainModel {
	return &RenameMainModel{
		inputDir: inputDir,
		screen:   RenameScanScreen,
		config: &organizer.RenamerConfig{
			BaseDir:             inputDir,
			Template:            "{author} - {series} {series_number} - {title}",
			AuthorFormat:        organizer.AuthorFormatFirstLast,
			Recursive:           true,
			PreservePath:        true,
			DryRun:              false,
			Verbose:             false,
			StrictMode:          false,
			PromptEnabled:       false,
			ReplaceSpace:        "",
			FieldMapping:        organizer.FieldMapping{},
			UseEmbeddedMetadata: useEmbeddedMetadata,
		},
	}
}

// Init initializes the model
func (m *RenameMainModel) Init() tea.Cmd {
	// Start with scan screen
	m.scanModel = NewRenameScanModel(m.inputDir)
	return m.scanModel.Init()
}

// Update handles messages and user input
func (m *RenameMainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q":
			// Handle quit based on screen
			if m.screen == RenameScanScreen {
				m.quitting = true
				return m, tea.Quit
			}
			// On other screens, go back
			return m.handleBack()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// Screen transition messages
	case RenameScanCompleteMsg:
		// Transition to field mapping screen
		m.candidates = msg.Candidates
		m.screen = RenameFieldMappingScreen
		m.fieldMappingModel = NewRenameFieldMappingModel(m.candidates, m.config)
		return m, m.fieldMappingModel.Init()

	case ShowCommandMsg:
		// Show the command to run instead of going to template screen
		m.screen = RenameCommandScreen
		m.commandModel = msg.Command
		m.dryRunCmd = msg.DryRunCmd
		m.outCmd = msg.OutCmd
		return m, nil

	case RenameFieldMappingConfirmedMsg:
		// Update config with field mapping and transition to template builder
		m.config.FieldMapping = msg.FieldMapping
		m.screen = RenameTemplateScreen
		m.templateModel = NewRenameTemplateModel(m.candidates, m.config)
		return m, m.templateModel.Init()

	case RenameTemplateConfirmedMsg:
		// Update config with new template and settings
		m.config.Template = msg.Template
		m.config.AuthorFormat = msg.AuthorFormat

		// Transition to preview screen
		m.screen = RenamePreviewScreen
		m.previewModel = NewRenamePreviewModel(m.candidates, m.config)
		return m, m.previewModel.Init()

	case RenamePreviewConfirmedMsg:
		// Transition to process screen
		m.screen = RenameProcessScreen
		m.processModel = NewRenameProcessModel(m.candidates, m.config)
		return m, m.processModel.Init()

	case RenameProcessCompleteMsg:
		// Done - quit
		m.quitting = true
		return m, tea.Quit

	case RenameErrorMsg:
		m.err = msg.Err
		m.quitting = true
		return m, tea.Quit
	}

	// Delegate to current screen
	switch m.screen {
	case RenameScanScreen:
		if m.scanModel != nil {
			var newModel tea.Model
			newModel, cmd = m.scanModel.Update(msg)
			m.scanModel = newModel.(*RenameScanModel)
			cmds = append(cmds, cmd)
		}

	case RenameFieldMappingScreen:
		if m.fieldMappingModel != nil {
			var newModel tea.Model
			newModel, cmd = m.fieldMappingModel.Update(msg)
			m.fieldMappingModel = newModel.(*RenameFieldMappingModel)
			cmds = append(cmds, cmd)
		}

	case RenameTemplateScreen:
		if m.templateModel != nil {
			var newModel tea.Model
			newModel, cmd = m.templateModel.Update(msg)
			m.templateModel = newModel.(*RenameTemplateModel)
			cmds = append(cmds, cmd)
		}

	case RenamePreviewScreen:
		if m.previewModel != nil {
			var newModel tea.Model
			newModel, cmd = m.previewModel.Update(msg)
			m.previewModel = newModel.(*RenamePreviewModel)
			cmds = append(cmds, cmd)
		}

	case RenameProcessScreen:
		if m.processModel != nil {
			var newModel tea.Model
			newModel, cmd = m.processModel.Update(msg)
			m.processModel = newModel.(*RenameProcessModel)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the current screen
func (m *RenameMainModel) View() string {
	if m.quitting {
		if m.err != nil {
			return "Error: " + m.err.Error() + "\n"
		}
		return ""
	}

	switch m.screen {
	case RenameScanScreen:
		if m.scanModel != nil {
			return m.scanModel.View()
		}
		return "Scanning..."

	case RenameFieldMappingScreen:
		if m.fieldMappingModel != nil {
			return m.fieldMappingModel.View()
		}
		return "Loading field mapping..."

	case RenameCommandScreen:
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))
		commandStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

		var sb strings.Builder
		sb.WriteString(titleStyle.Render("📋 Generated Commands") + "\n\n")

		// Normal command
		sb.WriteString(labelStyle.Render("1. Normal (Execute Rename):") + "\n")
		sb.WriteString(commandStyle.Render(m.commandModel) + "\n\n")

		// Dry-run command
		sb.WriteString(labelStyle.Render("2. Dry-Run (Preview Only):") + "\n")
		sb.WriteString(commandStyle.Render(m.dryRunCmd) + "\n\n")

		// Output directory command
		sb.WriteString(labelStyle.Render("3. Copy to New Directory:") + "\n")
		sb.WriteString(commandStyle.Render(m.outCmd) + "\n")
		sb.WriteString(helpStyle.Render("   (Change './organized' to your desired output directory)") + "\n\n")

		sb.WriteString(helpStyle.Render("b: Back • q: Quit"))

		return lipgloss.NewStyle().Padding(2).Render(sb.String())

	case RenameTemplateScreen:
		if m.templateModel != nil {
			return m.templateModel.View()
		}
		return "Loading template..."

	case RenamePreviewScreen:
		if m.previewModel != nil {
			return m.previewModel.View()
		}

	case RenameProcessScreen:
		if m.processModel != nil {
			return m.processModel.View()
		}
	}

	return "Loading..."
}

// handleBack handles going back to previous screen
func (m *RenameMainModel) handleBack() (*RenameMainModel, tea.Cmd) {
	switch m.screen {
	case RenameScanScreen:
		// Can't go back from scan
		m.quitting = true
		return m, tea.Quit

	case RenameFieldMappingScreen:
		// Go back to scan
		m.screen = RenameScanScreen
		return m, nil

	case RenameTemplateScreen:
		// Go back to field mapping
		m.screen = RenameFieldMappingScreen
		return m, nil

	case RenamePreviewScreen:
		// Go back to template
		m.screen = RenameTemplateScreen
		return m, nil

	default:
		return m, nil
	}
}

// Messages for screen transitions

type RenameScanCompleteMsg struct {
	Candidates []organizer.RenameCandidate
}

type RenameTemplateConfirmedMsg struct {
	Template     string
	AuthorFormat organizer.AuthorFormat
}

type RenamePreviewConfirmedMsg struct{}

type RenameProcessCompleteMsg struct {
	Summary organizer.RenameSummary
}

type RenameErrorMsg struct {
	Err error
}

type RenameFieldMappingConfirmedMsg struct {
	FieldMapping organizer.FieldMapping
}
