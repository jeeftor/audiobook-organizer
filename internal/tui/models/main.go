package models

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents different screens in the TUI
type Screen int

const (
	ScanScreen Screen = iota
	BookListScreen
	SettingsScreen
	PreviewScreen
	ProcessScreen
)

// MainModel is the main model for the TUI application
type MainModel struct {
	inputDir  string
	outputDir string
	screen    Screen
	width     int
	height    int

	// Sub-models for different screens
	scanModel     *ScanModel
	bookListModel *BookListModel
	settingsModel *SettingsModel
	previewModel  *PreviewModel
	processModel  *ProcessModel

	// Application state
	quitting bool
	err      error
}

// NewMainModel creates a new main model
func NewMainModel(inputDir, outputDir string) *MainModel {
	return &MainModel{
		inputDir:  inputDir,
		outputDir: outputDir,
		screen:    ScanScreen,
	}
}

// Init initializes the model
func (m *MainModel) Init() tea.Cmd {
	// Initialize the scan model
	m.scanModel = NewScanModel(m.inputDir)
	return m.scanModel.Init()
}

// Update handles messages and user input
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q":
			// Handle q differently based on screen
			switch m.screen {
			case ScanScreen:
				m.quitting = true
				return m, tea.Quit

			case BookListScreen:
				// Go back to scan screen
				m.screen = ScanScreen
				return m, nil

			case SettingsScreen:
				// Go back to book list
				m.screen = BookListScreen
				return m, nil

			case PreviewScreen:
				// Go back to settings
				m.screen = SettingsScreen
				return m, nil

			case ProcessScreen:
				// Only allow quitting if processing is complete
				if m.processModel != nil && m.processModel.complete {
					m.quitting = true
					return m, tea.Quit
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Pass window size to sub-models as needed
		// We'll handle this in each model's Update method
	}

	// Handle screen-specific updates
	switch m.screen {
	case ScanScreen:
		if m.scanModel != nil {
			var scanModel tea.Model
			scanModel, cmd = m.scanModel.Update(msg)
			m.scanModel = scanModel.(*ScanModel)
			cmds = append(cmds, cmd)

			// Also check for scan complete message directly
			if scanMsg, ok := msg.(ScanCompleteMsg); ok {
				// Log to a file for debugging
				logFile, _ := os.OpenFile("main_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				logFile.WriteString(fmt.Sprintf("Scan complete, found %d books\n", len(scanMsg.Books)))

				if len(scanMsg.Books) > 0 {
					// Create book list model with found books
					m.bookListModel = NewBookListModel(scanMsg.Books)

					// Automatically switch to book list screen
					m.screen = BookListScreen
					return m, m.bookListModel.Init()
				} else {
					// No books found, stay on scan screen
					return m, nil
				}
			}
		}

	case BookListScreen:
		if m.bookListModel != nil {
			var bookListModel tea.Model
			bookListModel, cmd = m.bookListModel.Update(msg)
			m.bookListModel = bookListModel.(*BookListModel)
			cmds = append(cmds, cmd)

			// Check for Enter key to proceed to settings
			if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
				m.screen = SettingsScreen
				if m.settingsModel == nil {
					// Pass selected books to settings model for preview
					selectedBooks := m.bookListModel.GetSelectedBooks()
					m.settingsModel = NewSettingsModel(selectedBooks)
					cmds = append(cmds, m.settingsModel.Init())
				}
			}
		}

	case SettingsScreen:
		if m.settingsModel != nil {
			var settingsModel tea.Model
			settingsModel, cmd = m.settingsModel.Update(msg)
			m.settingsModel = settingsModel.(*SettingsModel)
			cmds = append(cmds, cmd)

			// Check for Enter key to proceed to preview
			if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
				m.screen = PreviewScreen
				if m.previewModel == nil {
					selectedBooks := m.bookListModel.GetSelectedBooks()
					config := m.settingsModel.GetConfig()
					fieldMapping := m.settingsModel.GetFieldMapping()
					m.previewModel = NewPreviewModel(selectedBooks, config, fieldMapping)
					cmds = append(cmds, m.previewModel.Init())
				}
			}
		}

	case PreviewScreen:
		if m.previewModel != nil {
			var previewModel tea.Model
			previewModel, cmd = m.previewModel.Update(msg)

			// Check if we need to switch back to settings screen
			switch previewModel := previewModel.(type) {
			case *PreviewModel:
				// Continue in preview screen
				m.previewModel = previewModel
				cmds = append(cmds, cmd)
			case *SettingsModel:
				// Switch back to settings screen
				m.screen = SettingsScreen
				m.settingsModel = previewModel
				m.previewModel = nil
				return m, tea.Batch(cmds...)
			default:
				// Handle other model types or quit command
				return m, cmd
			}

			// Check for Enter key to proceed to processing
			if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
				m.screen = ProcessScreen
				if m.processModel == nil {
					selectedBooks := m.bookListModel.GetSelectedBooks()
					config := m.settingsModel.GetConfig()
					fieldMapping := m.settingsModel.GetFieldMapping()
					m.processModel = NewProcessModel(selectedBooks, config, m.previewModel.moves, fieldMapping)
					cmds = append(cmds, m.processModel.Init())
				}
			}
		}

	case ProcessScreen:
		if m.processModel != nil {
			var processModel tea.Model
			processModel, cmd = m.processModel.Update(msg)
			m.processModel = processModel.(*ProcessModel)
			cmds = append(cmds, cmd)

			// Check for 'r' key to return to main menu when complete
			if m.processModel.complete {
				if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "r" {
					// Reset models and go back to scan screen
					m.bookListModel = nil
					m.settingsModel = nil
					m.previewModel = nil
					m.processModel = nil
					m.scanModel = NewScanModel(m.inputDir)
					m.screen = ScanScreen
					cmds = append(cmds, m.scanModel.Init())
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m *MainModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var content string

	// Render the current screen
	switch m.screen {
	case ScanScreen:
		if m.scanModel != nil {
			content = m.scanModel.View()
		} else {
			content = "Initializing scanner..."
		}

	case BookListScreen:
		if m.bookListModel != nil {
			content = m.bookListModel.View()
		} else {
			content = "Loading book list..."
		}

	case SettingsScreen:
		if m.settingsModel != nil {
			content = m.settingsModel.View()
		} else {
			content = "Loading settings..."
		}

	case PreviewScreen:
		if m.previewModel != nil {
			content = m.previewModel.View()
		} else {
			content = "Generating preview..."
		}

	case ProcessScreen:
		if m.processModel != nil {
			content = m.processModel.View()
		} else {
			content = "Preparing to process files..."
		}

	default:
		content = "Unknown screen"
	}

	return content
}
