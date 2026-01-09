package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jeeftor/audiobook-organizer/internal/tui/models"
	"github.com/spf13/viper"
)

// RunRenameMode initializes and starts the rename TUI application
func RunRenameMode(inputDir string) error {
	// Determine if we should use embedded metadata
	useEmbedded := viper.GetBool("use-embedded-metadata") || viper.GetBool("flat")

	// Create the main model
	m := models.NewRenameMainModel(inputDir, useEmbedded)

	// Create the Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running rename TUI: %w", err)
	}

	return nil
}
