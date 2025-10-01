package cmd

import (
	"fmt"
	"os"

	"github.com/jeeftor/audiobook-organizer/internal/tui"
	"github.com/spf13/cobra"
)

// guiCmd represents the gui command for the TUI interface
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Start the TUI interface for audiobook organization",
	Long: `Launch a Text User Interface (TUI) for organizing audiobooks.
This mode provides an interactive way to scan, select, and organize your audiobooks.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputDir, _ := cmd.Flags().GetString("input")
		outputDir, _ := cmd.Flags().GetString("output")

		// Validate required flags
		if inputDir == "" || outputDir == "" {
			fmt.Println("Error: input and output directories are required")
			cmd.Help()
			os.Exit(1)
		}

		// Initialize and run the TUI
		if err := tui.Run(inputDir, outputDir); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)

	// Define flags
	guiCmd.Flags().StringP("input", "i", "", "Input directory containing audiobooks (required)")
	guiCmd.Flags().StringP("output", "o", "", "Output directory for organized audiobooks (required)")

	// Mark flags as required
	guiCmd.MarkFlagRequired("input")
	guiCmd.MarkFlagRequired("output")
}
