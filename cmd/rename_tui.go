package cmd

import (
	"fmt"
	"os"

	"github.com/jeeftor/audiobook-organizer/internal/tui"
	"github.com/spf13/cobra"
)

// renameTuiCmd represents the rename-tui command for the rename TUI interface
var renameTuiCmd = &cobra.Command{
	Use:   "rename-tui",
	Short: "Start the Terminal User Interface (TUI) for renaming audiobooks",
	Long: `Launch a Terminal User Interface (TUI) for renaming audiobooks.

This interactive mode allows you to:
- Scan directories for audiobook files
- Preview metadata fields from both metadata.json and embedded sources
- Configure field mappings interactively
- Design custom filename templates
- Preview rename operations before executing
- Execute renames with visual progress

The TUI supports hybrid metadata extraction, showing both JSON and embedded
metadata with visual indicators (📁 for JSON fields, 🎵 for embedded fields).`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// For TUI mode, directory is optional - we can browse if not provided
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get input directory from either flag
		inputDir := cmd.Flags().Lookup("input").Value.String()
		if inputDir == "" {
			inputDir = cmd.Flags().Lookup("dir").Value.String()
		}

		// Initialize and run the rename TUI
		if err := tui.RunRenameMode(inputDir); err != nil {
			fmt.Printf("Error running rename TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(renameTuiCmd)

	// Define flags with aliases matching the root command
	renameTuiCmd.Flags().String("dir", "", "Base directory to scan (alias for --input)")
	renameTuiCmd.Flags().StringP("input", "i", "", "Base directory to scan (alias for --dir)")

	// Note: rename-tui doesn't need output directory since files are renamed in-place
}
