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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// For GUI mode, directories are optional - we'll use the file picker if not provided
		// No validation needed
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get input directory from either flag
		inputDir := cmd.Flags().Lookup("input").Value.String()
		if inputDir == "" {
			inputDir = cmd.Flags().Lookup("dir").Value.String()
		}

		// Get output directory from either flag
		outputDir := cmd.Flags().Lookup("output").Value.String()
		if outputDir == "" {
			outputDir = cmd.Flags().Lookup("out").Value.String()
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

	// Define flags with aliases matching the root command
	guiCmd.Flags().String("dir", "", "Base directory to scan (alias for --input)")
	guiCmd.Flags().StringP("input", "i", "", "Base directory to scan (alias for --dir)")
	guiCmd.Flags().String("out", "", "Output directory (alias for --output)")
	guiCmd.Flags().StringP("output", "o", "", "Output directory (alias for --out)")
}
