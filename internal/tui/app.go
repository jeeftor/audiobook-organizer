package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jeeftor/audiobook-organizer/internal/tui/models"
)

// Run initializes and starts the TUI application
func Run(inputDir, outputDir string) error {
	// Create the initial model
	m := models.NewMainModel(inputDir, outputDir)

	// Initialize the program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
