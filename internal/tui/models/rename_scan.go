package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// RenameScanModel handles scanning directories and extracting metadata
type RenameScanModel struct {
	inputDir   string
	scanning   bool
	complete   bool
	candidates []organizer.RenameCandidate
	err        error

	// Progress tracking
	filesScanned      int
	filesWithMetadata int
	errors            []string
}

// NewRenameScanModel creates a new scan model
func NewRenameScanModel(inputDir string) *RenameScanModel {
	return &RenameScanModel{
		inputDir: inputDir,
		scanning: false,
		complete: false,
	}
}

// Init starts the scanning process
func (m *RenameScanModel) Init() tea.Cmd {
	return m.scanFiles
}

// Update handles messages
func (m *RenameScanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	case scanStartMsg:
		m.scanning = true
		return m, nil

	case scanProgressMsg:
		m.filesScanned = msg.filesScanned
		m.filesWithMetadata = msg.filesWithMetadata
		return m, nil

	case scanCompleteMsg:
		m.scanning = false
		m.complete = true
		m.candidates = msg.candidates
		// Auto-transition to template screen
		return m, func() tea.Msg {
			return RenameScanCompleteMsg{Candidates: msg.candidates}
		}

	case scanErrorMsg:
		m.err = msg.err
		m.scanning = false
		return m, func() tea.Msg {
			return RenameErrorMsg{Err: msg.err}
		}
	}

	return m, nil
}

// View renders the scan screen
func (m *RenameScanModel) View() string {
	style := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	var content string

	if m.scanning {
		content = fmt.Sprintf(
			"🔍 Scanning directory...\n\n"+
				"Directory: %s\n"+
				"Files scanned: %d\n"+
				"Files with metadata: %d\n\n"+
				"Please wait...",
			m.inputDir,
			m.filesScanned,
			m.filesWithMetadata,
		)
	} else if m.complete {
		content = fmt.Sprintf(
			"✓ Scan complete!\n\n"+
				"Found %d files\n"+
				"Transitioning to template builder...",
			len(m.candidates),
		)
	} else {
		content = "Starting scan..."
	}

	return style.Render(content)
}

// scanFiles performs the actual scanning
func (m *RenameScanModel) scanFiles() tea.Msg {
	// Create a renamer to scan files
	config := &organizer.RenamerConfig{
		BaseDir:             m.inputDir,
		Template:            "{author} - {series} {series_number} - {title}",
		AuthorFormat:        organizer.AuthorFormatFirstLast,
		Recursive:           true,
		PreservePath:        true,
		DryRun:              true,
		Verbose:             false,
		StrictMode:          false,
		UseEmbeddedMetadata: false, // Will be set by parent model
	}

	renamer, err := organizer.NewRenamer(config)
	if err != nil {
		return scanErrorMsg{err: err}
	}

	// Scan files
	candidates, err := renamer.ScanFiles()
	if err != nil {
		return scanErrorMsg{err: err}
	}

	return scanCompleteMsg{candidates: candidates}
}

// Internal messages for scan model

type scanStartMsg struct{}

type scanProgressMsg struct {
	filesScanned      int
	filesWithMetadata int
}

type scanCompleteMsg struct {
	candidates []organizer.RenameCandidate
}

type scanErrorMsg struct {
	err error
}
