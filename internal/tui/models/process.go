package models

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// ProcessStatus represents the status of a file processing operation
type ProcessStatus int

const (
	StatusPending ProcessStatus = iota
	StatusProcessing
	StatusSuccess
	StatusError
)

// ProcessItem represents a file being processed
type ProcessItem struct {
	SourcePath string
	TargetPath string
	Status     ProcessStatus
	Error      error
	Message    string // Additional information about the processing
}

// ProcessCompleteMsg is sent when processing is complete
type ProcessCompleteMsg struct {
	Success int
	Failed  int
	Items   []ProcessItem
	Elapsed time.Duration
}

// ProcessModel represents the processing screen
type ProcessModel struct {
	books        []AudioBook
	config       map[string]string
	fieldMapping organizer.FieldMapping
	items        []ProcessItem
	processing   bool
	complete     bool
	cursor       int
	width        int
	height       int
	scrollOffset int
	startTime    time.Time
	elapsedTime  time.Duration
	success      int
	failed       int
}

// NewProcessModel creates a new process model
func NewProcessModel(
	books []AudioBook,
	config map[string]string,
	moves []MovePreview,
	fieldMapping organizer.FieldMapping,
) *ProcessModel {
	items := make([]ProcessItem, len(moves))
	for i, move := range moves {
		items[i] = ProcessItem{
			SourcePath: move.SourcePath,
			TargetPath: move.TargetPath,
			Status:     StatusPending,
		}
	}

	return &ProcessModel{
		books:        books,
		config:       config,
		fieldMapping: fieldMapping,
		items:        items,
		processing:   false,
		complete:     false,
	}
}

// Init initializes the model
func (m *ProcessModel) Init() tea.Cmd {
	return m.startProcessing()
}

// startProcessing begins the processing of files
func (m *ProcessModel) startProcessing() tea.Cmd {
	m.processing = true
	m.startTime = time.Now()

	// Commands run outside the Bubble Tea update loop. Give the worker private
	// state and publish its result as a message instead of mutating the live model.
	worker := *m
	worker.items = append([]ProcessItem(nil), m.items...)
	worker.config = make(map[string]string, len(m.config))
	for key, value := range m.config {
		worker.config[key] = value
	}
	return func() tea.Msg {
		// Get directories from config
		baseDir := worker.config["Input Directory"]
		outputDir := worker.config["Output Directory"]

		// Fallback: if not in config, try to get from the first book
		if baseDir == "" && len(worker.books) > 0 {
			baseDir = filepath.Dir(worker.books[0].Path)
		}
		if outputDir == "" {
			outputDir = baseDir // Use same directory if not specified
		}

		// Get layout from settings
		layout := worker.config["Layout"]
		if layout == "" {
			layout = "author-series-title"
		}
		layoutTemplate := ""
		if layout == "custom" {
			layoutTemplate = strings.TrimSpace(worker.config["Layout Template"])
			layout = "author-series-title"
		}

		// Create configuration from settings
		config := &organizer.OrganizerConfig{
			BaseDir:             baseDir,
			OutputDir:           outputDir,
			Layout:              layout,
			LayoutTemplate:      layoutTemplate,
			UseEmbeddedMetadata: worker.config["Use Embedded Metadata"] == "Yes",
			Flat:                worker.config["Flat Mode"] == "Yes",
			DryRun:              worker.config["Dry Run"] == "Yes",
			Verbose:             false, // Always false in TUI mode - we have our own display
			FieldMapping:        worker.fieldMapping,
			RemoveEmpty:         false, // Don't remove empty directories in TUI mode
			Prompt:              false, // Don't prompt in TUI mode
		}

		// Process each item individually using OrganizeSingleFile
		org, err := organizer.NewOrganizer(config)
		if err != nil {
			// Mark all items as failed with the configuration error
			for i := range worker.items {
				worker.items[i].Status = StatusError
				worker.items[i].Error = fmt.Errorf("configuration error: %v", err)
			}
			worker.failed = len(worker.items)
			worker.complete = true
			worker.processing = false
			worker.elapsedTime = time.Since(worker.startTime)
			return ProcessCompleteMsg{
				Success: 0,
				Failed:  worker.failed,
				Items:   worker.items,
				Elapsed: time.Since(worker.startTime),
			}
		}

		for i := range worker.items {
			// Update status to processing
			worker.items[i].Status = StatusProcessing

			// Get the source path for this file
			sourcePath := worker.items[i].SourcePath

			// Process the file using the organizer
			// Pass nil as the provider to let the organizer create and configure it
			err := org.OrganizeSingleFile(sourcePath, nil)

			if err != nil {
				// Processing failed
				worker.items[i].Status = StatusError
				worker.items[i].Error = fmt.Errorf("failed to process: %v", err)
				worker.failed++
			} else {
				// Processing succeeded
				worker.items[i].Status = StatusSuccess

				// Build a descriptive message about what was done
				if config.DryRun {
					worker.items[i].Message = "Would move (dry-run mode)"
				} else {
					worker.items[i].Message = fmt.Sprintf("Moved to %s", filepath.Dir(worker.items[i].TargetPath))
				}
				worker.success++
			}
		}

		worker.complete = true
		worker.processing = false

		return ProcessCompleteMsg{
			Success: worker.success,
			Failed:  worker.failed,
			Items:   worker.items,
			Elapsed: time.Since(worker.startTime),
		}
	}
}

// Update handles messages and user input
func (m *ProcessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case ProcessCompleteMsg:
		m.success = msg.Success
		m.failed = msg.Failed
		m.items = msg.Items
		m.elapsedTime = msg.Elapsed
		m.complete = true
		m.processing = false
		return m, nil

	case tea.KeyMsg:
		if !m.processing && !m.complete {
			if msg.String() == "enter" {
				return m, m.startProcessing()
			}
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Adjust scroll if necessary
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				// Adjust scroll if necessary
				maxVisible := m.height - 10 // Approximate space for header and footer
				if m.cursor >= m.scrollOffset+maxVisible {
					m.scrollOffset = m.cursor - maxVisible + 1
				}
			}

		case "r":
			// Only allow returning to main menu if processing is complete
			// Don't handle here - let main.go handle it
			// Just consume the key
			return m, nil

		case "q", "ctrl+c":
			// Quit the application
			return m, tea.Quit
		}
	}

	// Update elapsed time if processing
	if m.processing {
		m.elapsedTime = time.Since(m.startTime)
	}

	return m, nil
}

// View renders the UI
func (m *ProcessModel) View() string {
	var content strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("⚙️ Processing Files")

	content.WriteString(header + "\n\n")

	if !m.processing && !m.complete {
		// Initial state
		content.WriteString(fmt.Sprintf("Ready to process %d files.\n\n", len(m.items)))
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Render("Press Enter to begin processing..."))
		return content.String()
	}

	// Status information
	if m.processing {
		content.WriteString(fmt.Sprintf("Processing %d files...\n", len(m.items)))
		content.WriteString(fmt.Sprintf("Elapsed time: %s\n\n", m.elapsedTime.Round(time.Second)))
	} else if m.complete {
		content.WriteString(fmt.Sprintf("Processing complete in %s\n", m.elapsedTime.Round(time.Second)))
		content.WriteString(fmt.Sprintf("Success: %d | Failed: %d\n\n", m.success, m.failed))
	}

	// Calculate visible range based on height
	maxVisible := m.height - 12 // Approximate space for header and footer
	endIdx := m.scrollOffset + maxVisible
	if endIdx > len(m.items) {
		endIdx = len(m.items)
	}

	// Show scroll indicator if needed
	if m.scrollOffset > 0 {
		content.WriteString("↑ Scroll up for more\n")
	}

	// Process items
	for i := m.scrollOffset; i < endIdx; i++ {
		item := m.items[i]

		// Cursor indicator
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		// Status indicator
		var statusStr string
		var statusStyle lipgloss.Style

		switch item.Status {
		case StatusPending:
			statusStr = "⏳ Pending"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
		case StatusProcessing:
			statusStr = "🔄 Processing"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
		case StatusSuccess:
			statusStr = "✅ Success"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
		case StatusError:
			statusStr = "❌ Error"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		}

		// Format paths
		sourceName := filepath.Base(item.SourcePath)
		targetName := filepath.Base(item.TargetPath)

		// Style based on cursor position
		var pathStyle lipgloss.Style
		if i == m.cursor {
			pathStyle = lipgloss.NewStyle().Bold(true)
		} else {
			pathStyle = lipgloss.NewStyle()
		}

		// Add the item
		content.WriteString(fmt.Sprintf("%s %s: %s → %s\n",
			cursor,
			statusStyle.Render(statusStr),
			pathStyle.Render(sourceName),
			pathStyle.Render(targetName)))

		// Add error message if any
		if item.Status == StatusError && item.Error != nil {
			content.WriteString(fmt.Sprintf(
				"  %s\n",
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FF0000")).
					Render(item.Error.Error()),
			))
		}

		// Add field mapping information if available
		if item.Status == StatusSuccess && item.Message != "" {
			content.WriteString(fmt.Sprintf("  %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("#00AAFF")).Render(item.Message)))
		}
	}

	// Show scroll indicator if needed
	if endIdx < len(m.items) {
		content.WriteString("↓ Scroll down for more\n")
	}

	// Footer with help text
	var footer string
	if m.complete {
		footer = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888")).
			Render("\n↑/↓: Navigate • r: Return to main menu • q: Quit")
	} else {
		footer = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888")).
			Render("\n↑/↓: Navigate • q: Quit")
	}

	content.WriteString(footer)

	return content.String()
}
