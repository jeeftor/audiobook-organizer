package models

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents different screens in the TUI
type Screen int

const (
	DirPickerScreen Screen = iota
	ScanScreen
	BookListScreen
	SettingsScreen
	AdvancedSettingsScreen
	PreviewScreen
	ProcessScreen
	CommandOutputScreen
)

// MainModel is the main model for the TUI application
type MainModel struct {
	inputDir  string
	outputDir string
	screen    Screen
	width     int
	height    int

	// Sub-models for different screens
	dirPickerModel         *DirPickerModel
	scanModel              *ScanModel
	bookListModel          *BookListModel
	settingsModel          *SettingsTableModel
	advancedSettingsModel  *SettingsTableModel
	previewModel           *PreviewModel
	processModel           *ProcessModel
	commandOutputModel     *CommandOutputModel

	// Application state
	quitting bool
	err      error
}

// NewMainModel creates a new main model
func NewMainModel(inputDir, outputDir string) *MainModel {
	// If no directories provided, start with directory picker
	startScreen := ScanScreen
	if inputDir == "" || outputDir == "" {
		startScreen = DirPickerScreen
	}

	return &MainModel{
		inputDir:  inputDir,
		outputDir: outputDir,
		screen:    startScreen,
	}
}

// Init initializes the model
func (m *MainModel) Init() tea.Cmd {
	if m.screen == DirPickerScreen {
		// Initialize the directory picker
		m.dirPickerModel = NewDirPickerModel(PickingInput, "")
		return m.dirPickerModel.Init()
	}

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

			case AdvancedSettingsScreen:
				// Go back to basic settings
				m.screen = SettingsScreen
				return m, nil

			case PreviewScreen:
				// Go back to advanced settings
				m.screen = AdvancedSettingsScreen
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
	case DirPickerScreen:
		if m.dirPickerModel != nil {
			var dirPickerModel tea.Model
			dirPickerModel, cmd = m.dirPickerModel.Update(msg)

			// Check if model changed (directory was selected)
			if newPicker, ok := dirPickerModel.(*DirPickerModel); ok {
				m.dirPickerModel = newPicker
				cmds = append(cmds, cmd)

				// Check if both directories are selected
				if m.dirPickerModel.mode == PickingOutput && m.dirPickerModel.outputDir != "" {
					// Both directories selected, update main model and start scan
					m.inputDir = m.dirPickerModel.inputDir
					m.outputDir = m.dirPickerModel.outputDir
					m.screen = ScanScreen
					m.scanModel = NewScanModel(m.inputDir)
					return m, m.scanModel.Init()
				}
			}
		}

	case ScanScreen:
		if m.scanModel != nil {
			var scanModel tea.Model
			scanModel, cmd = m.scanModel.Update(msg)
			m.scanModel = scanModel.(*ScanModel)
			cmds = append(cmds, cmd)

			// Also check for scan complete message directly
			if scanMsg, ok := msg.(ScanCompleteMsg); ok {
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
					m.settingsModel = NewSettingsTableModel(selectedBooks, false)
					cmds = append(cmds, m.settingsModel.Init())
				}
			}
		}

	case SettingsScreen:
		if m.settingsModel != nil {
			var settingsModel tea.Model
			settingsModel, cmd = m.settingsModel.Update(msg)
			m.settingsModel = settingsModel.(*SettingsTableModel)
			cmds = append(cmds, cmd)

			// Check for c/n keys to proceed to preview (always works)
			// Or Enter key to proceed to advanced settings (only if model allows it)
			if msg, ok := msg.(tea.KeyMsg); ok {
				key := msg.String()

				// c/n advances directly to preview screen
				if key == "c" || key == "n" {
					m.screen = PreviewScreen
					if m.previewModel == nil {
						selectedBooks := m.bookListModel.GetSelectedBooks()
						// Get config and field mapping from unified settings model
						config := m.settingsModel.GetConfig()
						fieldMapping := m.settingsModel.GetFieldMapping()
						// Add input and output directories to config
						config["Input Directory"] = m.inputDir
						config["Output Directory"] = m.outputDir
						m.previewModel = NewPreviewModel(selectedBooks, config, fieldMapping)
						cmds = append(cmds, m.previewModel.Init())
					}
				} else if key == "enter" && m.settingsModel.ShouldAdvance() {
					// Enter goes to advanced settings (old flow - probably not needed anymore)
					m.screen = AdvancedSettingsScreen
					if m.advancedSettingsModel == nil {
						selectedBooks := m.bookListModel.GetSelectedBooks()
						m.advancedSettingsModel = NewSettingsTableModel(selectedBooks, true)
						cmds = append(cmds, m.advancedSettingsModel.Init())
					}
				}
			}
		}

	case AdvancedSettingsScreen:
		if m.advancedSettingsModel != nil {
			var advancedModel tea.Model
			advancedModel, cmd = m.advancedSettingsModel.Update(msg)
			m.advancedSettingsModel = advancedModel.(*SettingsTableModel)
			cmds = append(cmds, cmd)

			// Check for c/n keys to proceed to preview (always works)
			// Or Enter key (only if model allows it)
			if msg, ok := msg.(tea.KeyMsg); ok {
				key := msg.String()
				shouldAdvance := false

				// c/n always advances, regardless of popup state
				if key == "c" || key == "n" {
					shouldAdvance = true
				} else if key == "enter" && m.advancedSettingsModel.ShouldAdvance() {
					// Enter only advances if popup not showing
					shouldAdvance = true
				}

				if shouldAdvance {
					m.screen = PreviewScreen
					if m.previewModel == nil {
						selectedBooks := m.bookListModel.GetSelectedBooks()
						// Get config from basic settings, field mapping from advanced settings
						config := m.settingsModel.GetConfig()
						fieldMapping := m.advancedSettingsModel.GetFieldMapping()
						// Add input and output directories to config
						config["Input Directory"] = m.inputDir
						config["Output Directory"] = m.outputDir
						m.previewModel = NewPreviewModel(selectedBooks, config, fieldMapping)
						cmds = append(cmds, m.previewModel.Init())
					}
				}
			}
		}

	case PreviewScreen:
		if m.previewModel != nil {
			// Check for navigation keys first
			if msg, ok := msg.(tea.KeyMsg); ok {
				key := msg.String()
				if key == "b" || key == "backspace" || key == "q" {
					// Go back to settings screen
					m.screen = SettingsScreen
					// Don't reset the settings model - keep the existing one
					return m, nil
				}
			}

			var previewModel tea.Model
			previewModel, cmd = m.previewModel.Update(msg)

			// Check if we need to switch screens based on return type
			switch previewModel := previewModel.(type) {
			case *PreviewModel:
				// Continue in preview screen
				m.previewModel = previewModel
				cmds = append(cmds, cmd)
			case *ProcessModel:
				// Switch to process screen (when Enter is pressed)
				m.screen = ProcessScreen
				m.processModel = previewModel
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			case *CommandOutputModel:
				// Switch to command output screen (when 'c' is pressed)
				m.screen = CommandOutputScreen
				m.commandOutputModel = previewModel
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			default:
				// Handle other model types or quit command
				return m, cmd
			}
		}

	case ProcessScreen:
		if m.processModel != nil {
			// Check for 'r' key to return to main menu when complete
			if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "r" && m.processModel.complete {
				// Reset models and go back to scan screen
				m.bookListModel = nil
				m.settingsModel = nil
				m.advancedSettingsModel = nil
				m.previewModel = nil
				m.processModel = nil
				m.commandOutputModel = nil
				m.scanModel = NewScanModel(m.inputDir)
				m.screen = ScanScreen
				return m, m.scanModel.Init()
			}

			var processModel tea.Model
			processModel, cmd = m.processModel.Update(msg)
			m.processModel = processModel.(*ProcessModel)
			cmds = append(cmds, cmd)
		}

	case CommandOutputScreen:
		if m.commandOutputModel != nil {
			var commandOutputModel tea.Model
			commandOutputModel, cmd = m.commandOutputModel.Update(msg)

			// Check if we need to switch screens based on return type
			switch commandOutputModel := commandOutputModel.(type) {
			case *CommandOutputModel:
				// Continue in command output screen
				m.commandOutputModel = commandOutputModel
				cmds = append(cmds, cmd)
			case *PreviewModel:
				// Switch back to preview screen (when 'b' is pressed)
				m.screen = PreviewScreen
				m.previewModel = commandOutputModel
				m.commandOutputModel = nil
				return m, tea.Batch(cmds...)
			default:
				// Handle other model types or quit command
				return m, cmd
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
	case DirPickerScreen:
		if m.dirPickerModel != nil {
			content = m.dirPickerModel.View()
		} else {
			content = "Initializing directory picker..."
		}

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

	case AdvancedSettingsScreen:
		if m.advancedSettingsModel != nil {
			content = m.advancedSettingsModel.View()
		} else {
			content = "Loading advanced settings..."
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

	case CommandOutputScreen:
		if m.commandOutputModel != nil {
			content = m.commandOutputModel.View()
		} else {
			content = "Generating command..."
		}

	default:
		content = "Unknown screen"
	}

	return content
}
