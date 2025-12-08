package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PickerMode indicates which directory we're picking
type PickerMode int

const (
	PickingInput PickerMode = iota
	PickingOutput
)

// DirPickerModel represents the directory picker screen
type DirPickerModel struct {
	filepicker   filepicker.Model
	mode         PickerMode
	inputDir     string
	outputDir    string
	err          error
	width        int
	height       int
	filterText   string
	filterActive bool
	filteredDirs []string
	filterCursor int
	scrollOffset int
	allDirs      []string // All directories in current location
	cursor       int      // Cursor for non-filtered navigation
	creatingDir  bool     // Whether we're in "create new directory" mode
	newDirName   string   // Name of the new directory being created
}

// NewDirPickerModel creates a new directory picker model
func NewDirPickerModel(mode PickerMode, inputDir string) *DirPickerModel {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false // Only allow directory selection
	fp.ShowHidden = false
	fp.ShowSize = false
	fp.ShowPermissions = false
	fp.AllowedTypes = nil // Allow all types since we're filtering by directory only
	fp.AutoHeight = false // We'll set height manually
	fp.Height = 20        // Set a default height

	// Set starting directory
	startDir := inputDir
	if startDir == "" {
		startDir = "."
	}

	// Resolve to absolute path
	absDir, err := filepath.Abs(startDir)
	if err == nil {
		fp.CurrentDirectory = absDir
	} else {
		fp.CurrentDirectory = startDir
	}

	model := &DirPickerModel{
		filepicker: fp,
		mode:       mode,
		inputDir:   inputDir,
	}

	// Load all directories
	model.loadDirectories()

	return model
}

// Init initializes the model
func (m *DirPickerModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

// loadDirectories reads all directories in the current path
func (m *DirPickerModel) loadDirectories() {
	m.allDirs = nil
	m.cursor = 0

	// Add ".." at the top if not at root
	if m.filepicker.CurrentDirectory != "/" {
		m.allDirs = append(m.allDirs, "..")
	}

	entries, err := os.ReadDir(m.filepicker.CurrentDirectory)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip hidden files if not showing them
		if !m.filepicker.ShowHidden && strings.HasPrefix(name, ".") {
			continue
		}

		m.allDirs = append(m.allDirs, name)
	}
}

// Update handles messages and user input
func (m *DirPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Set filepicker height to leave room for header and footer
		// Header (1) + Description (1) + Blank (1) + Current dir (1) + Blank (1) + Help (2) + Filter input (2) = 9 lines
		m.filepicker.Height = msg.Height - 14 // Reserve more space for all UI elements
		// Re-init filepicker after height change to refresh the view
		return m, m.filepicker.Init()

	case tea.KeyMsg:
		key := msg.String()

		// Handle "create directory" mode input
		if m.creatingDir {
			switch key {
			case "enter":
				// Create the directory
				if m.newDirName != "" {
					newPath := filepath.Join(m.filepicker.CurrentDirectory, m.newDirName)
					if err := os.MkdirAll(newPath, 0755); err == nil {
						// Navigate into the new directory
						m.filepicker.CurrentDirectory = newPath
						m.loadDirectories()
					}
				}
				m.creatingDir = false
				m.newDirName = ""
				return m, m.filepicker.Init()
			case "esc":
				// Cancel directory creation
				m.creatingDir = false
				m.newDirName = ""
				return m, nil
			case "backspace":
				if len(m.newDirName) > 0 {
					m.newDirName = m.newDirName[:len(m.newDirName)-1]
				}
				return m, nil
			default:
				// Add character to directory name
				if len(key) == 1 && key >= " " && key <= "~" {
					m.newDirName += key
				}
				return m, nil
			}
		}

		// Handle special keys first
		switch key {
		case "ctrl+c":
			return m, tea.Quit

		case "ctrl+q":
			return m, tea.Quit

		case "ctrl+n":
			// Create new directory (only for output picker)
			if m.mode == PickingOutput {
				m.creatingDir = true
				m.newDirName = ""
				return m, nil
			}

		case "ctrl+b":
			// Navigate up one directory level
			if m.filepicker.CurrentDirectory != "/" {
				parent := filepath.Dir(m.filepicker.CurrentDirectory)
				m.filepicker.CurrentDirectory = parent
				m.filterText = ""
				m.filterActive = false
				m.filterCursor = 0
				m.scrollOffset = 0
				m.loadDirectories()
				return m, m.filepicker.Init()
			}
			return m, nil

		case "ctrl+h":
			// Jump to home directory
			if homeDir, err := os.UserHomeDir(); err == nil {
				m.filepicker.CurrentDirectory = homeDir
				m.filterText = ""
				m.filterActive = false
				m.filterCursor = 0
				m.scrollOffset = 0
				m.loadDirectories()
				return m, m.filepicker.Init()
			}
			return m, nil

		case "ctrl+r":
			// Jump to root directory
			m.filepicker.CurrentDirectory = "/"
			m.filterText = ""
			m.filterActive = false
			m.filterCursor = 0
			m.scrollOffset = 0
			m.loadDirectories()
			return m, m.filepicker.Init()

		case "esc":
			// Clear filter
			m.filterText = ""
			m.filterActive = false
			m.filterCursor = 0
			m.scrollOffset = 0
			return m, nil

		case "backspace":
			// Remove last character from filter
			if m.filterActive && len(m.filterText) > 0 {
				m.filterText = m.filterText[:len(m.filterText)-1]
				if m.filterText == "" {
					m.filterActive = false
					m.filterCursor = 0
					m.scrollOffset = 0
				} else {
					// Update filtered list
					m.filteredDirs, _ = m.getFilteredDirectories()
					if m.filterCursor >= len(m.filteredDirs) && len(m.filteredDirs) > 0 {
						m.filterCursor = len(m.filteredDirs) - 1
					}
				}
				return m, nil
			}

		case "up", "k":
			if m.filterActive {
				if len(m.filteredDirs) > 0 && m.filterCursor > 0 {
					m.filterCursor--
					// Adjust scroll
					if m.filterCursor < m.scrollOffset {
						m.scrollOffset = m.filterCursor
					}
				}
			} else {
				if len(m.allDirs) > 0 && m.cursor > 0 {
					m.cursor--
					// Adjust scroll
					if m.cursor < m.scrollOffset {
						m.scrollOffset = m.cursor
					}
				}
			}
			return m, nil

		case "down", "j":
			if m.filterActive {
				if len(m.filteredDirs) > 0 && m.filterCursor < len(m.filteredDirs)-1 {
					m.filterCursor++
					// Adjust scroll
					maxVisible := m.height - 14
					if m.filterCursor >= m.scrollOffset+maxVisible {
						m.scrollOffset = m.filterCursor - maxVisible + 1
					}
				}
			} else {
				if len(m.allDirs) > 0 && m.cursor < len(m.allDirs)-1 {
					m.cursor++
					// Adjust scroll
					maxVisible := m.height - 14
					if m.cursor >= m.scrollOffset+maxVisible {
						m.scrollOffset = m.cursor - maxVisible + 1
					}
				}
			}
			return m, nil

		case "enter":
			var selectedDir string
			if m.filterActive {
				if len(m.filteredDirs) == 0 {
					return m, nil
				}
				selectedDir = m.filteredDirs[m.filterCursor]
			} else {
				if len(m.allDirs) == 0 {
					return m, nil
				}
				selectedDir = m.allDirs[m.cursor]
			}

			// Handle ".." specially
			if selectedDir == ".." {
				if m.filepicker.CurrentDirectory != "/" {
					parent := filepath.Dir(m.filepicker.CurrentDirectory)
					m.filepicker.CurrentDirectory = parent
					m.filterText = ""
					m.filterActive = false
					m.filterCursor = 0
					m.scrollOffset = 0
					m.loadDirectories()
					return m, m.filepicker.Init()
				}
				return m, nil
			}

			selectedPath := filepath.Join(m.filepicker.CurrentDirectory, selectedDir)

			// Check if it's a valid directory
			if info, err := os.Stat(selectedPath); err == nil && info.IsDir() {
				m.filepicker.CurrentDirectory = selectedPath
				m.filterText = ""
				m.filterActive = false
				m.filterCursor = 0
				m.scrollOffset = 0
				m.loadDirectories()
				return m, m.filepicker.Init()
			}
			return m, nil

		case "ctrl+s", "ctrl+d":
			// Select current directory (Ctrl+S for "select" or Ctrl+D for "done")
			currentDir := m.filepicker.CurrentDirectory
			if m.mode == PickingInput {
				m.inputDir = currentDir
				newModel := NewDirPickerModel(PickingOutput, currentDir)
				// Pass current dimensions to new model
				newModel.width = m.width
				newModel.height = m.height
				newModel.filepicker.Height = m.height - 14
				return newModel, newModel.filepicker.Init()
			} else {
				m.outputDir = currentDir
				return m, nil
			}

		default:
			// Check if it's a printable character for filtering
			if len(key) == 1 && key >= " " && key <= "~" {
				m.filterText += key
				m.filterActive = true
				m.filterCursor = 0
				m.scrollOffset = 0
				// Update filtered list
				m.filteredDirs, _ = m.getFilteredDirectories()
				return m, nil
			}
		}
	}

	// Only pass non-filtering keys to the filepicker
	var cmd tea.Cmd
	if !m.filterActive {
		m.filepicker, cmd = m.filepicker.Update(msg)
	}

	// Check if a directory was selected (final selection, not just navigation)
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect && !m.filterActive {
		// Check if this is actually a directory we're navigating into or a final selection
		// When Enter is pressed on a directory in the filepicker, it navigates into it
		// We need to distinguish between "navigating into" vs "selecting"
		// The filepicker will handle navigation, we just handle final selection

		// For now, any selection is treated as final
		if m.mode == PickingInput {
			m.inputDir = path
			// Move to output directory picker, starting from the selected input directory
			return NewDirPickerModel(PickingOutput, path), nil
		} else {
			// Both directories selected - this will be handled in main.go
			m.outputDir = path
			return m, nil
		}
	}

	return m, cmd
}

// View renders the UI
func (m *DirPickerModel) View() string {
	var content string

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Italic(true)

	var header, description string
	if m.mode == PickingInput {
		header = "📁 SELECT INPUT DIRECTORY"
		description = "This is where the audiobook files you want to process are located."
	} else {
		header = "📁 SELECT OUTPUT DIRECTORY"
		description = "This is where the organized audiobook files will be moved to."
	}

	content = headerStyle.Render(header) + "\n"
	content += descStyle.Render(description) + "\n\n"

	// Show current directory
	currentDirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00AAFF"))
	content += currentDirStyle.Render(fmt.Sprintf("Current: %s", m.filepicker.CurrentDirectory)) + "\n\n"

	// Show currently selected input dir if we're picking output
	if m.mode == PickingOutput && m.inputDir != "" {
		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
		content += infoStyle.Render(fmt.Sprintf("✓ Input Directory: %s", m.inputDir)) + "\n\n"
	}

	// Show directory list (either filtered or all)
	var dirsToShow []string
	var currentCursor int

	if m.filterActive {
		dirsToShow = m.filteredDirs
		currentCursor = m.filterCursor
	} else {
		dirsToShow = m.allDirs
		currentCursor = m.cursor
	}

	maxVisible := m.height - 14
	endIdx := m.scrollOffset + maxVisible
	if endIdx > len(dirsToShow) {
		endIdx = len(dirsToShow)
	}

	if len(dirsToShow) == 0 {
		noResultsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		if m.filterActive {
			content += noResultsStyle.Render("No matching directories") + "\n"
		} else {
			content += noResultsStyle.Render("No directories found") + "\n"
		}
	} else {
		for i := m.scrollOffset; i < endIdx; i++ {
			dir := dirsToShow[i]
			cursor := "  "
			style := lipgloss.NewStyle()

			if i == currentCursor {
				cursor = "> "
				style = style.Bold(true).Foreground(lipgloss.Color("#00FF00"))
			}

			content += cursor + style.Render(dir) + "\n"
		}
	}

	// Show "create directory" input if active
	content += "\n\n"
	if m.creatingDir {
		createStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Bold(true)
		content += createStyle.Render(fmt.Sprintf("📁 New directory name: %s_", m.newDirName)) + "\n"
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("Enter: Create • ESC: Cancel") + "\n"
	} else if m.filterActive {
		// Show filter text if active
		filterStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)
		content += filterStyle.Render(fmt.Sprintf("Filter: %s_ (%d matches)", m.filterText, len(m.filteredDirs))) + "\n"
	} else {
		content += "\n"
	}

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))
	content += helpStyle.Render("↑/↓: Navigate • Enter: Open Directory • Ctrl+S: Select Current Directory")
	if m.mode == PickingOutput {
		content += "\n" + helpStyle.Render("Ctrl+N: Create New Directory • Type to filter • ESC: Clear filter")
	} else {
		content += "\n" + helpStyle.Render("Type to filter • ESC: Clear filter • Ctrl+B: Up • Ctrl+H: Home • Ctrl+R: Root • Ctrl+Q: Quit")
	}

	return content
}

// GetInputDir returns the selected input directory
func (m *DirPickerModel) GetInputDir() string {
	return m.inputDir
}

// GetOutputDir returns the selected output directory
func (m *DirPickerModel) GetOutputDir() string {
	return m.outputDir
}

// getFilteredDirectories reads directories from current path and filters them
func (m *DirPickerModel) getFilteredDirectories() ([]string, error) {
	entries, err := os.ReadDir(m.filepicker.CurrentDirectory)
	if err != nil {
		return nil, err
	}

	var dirs []string
	filterLower := strings.ToLower(m.filterText)

	// Add ".." at the top if not at root
	if m.filepicker.CurrentDirectory != "/" {
		// Only show .. if filter doesn't exclude it
		if !m.filterActive || strings.Contains("..", filterLower) {
			dirs = append(dirs, "..")
		}
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip hidden files if not showing them (but always show ..)
		if !m.filepicker.ShowHidden && strings.HasPrefix(name, ".") && name != ".." {
			continue
		}

		// Apply filter if active
		if m.filterActive {
			nameLower := strings.ToLower(name)
			if !strings.Contains(nameLower, filterLower) {
				continue
			}
		}

		dirs = append(dirs, name)
	}

	return dirs, nil
}
