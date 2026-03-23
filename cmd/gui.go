package cmd

import (
	"fmt"
	"os"

	"github.com/jeeftor/audiobook-organizer/internal/guiapp"
	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the desktop GUI application",
	Long: `Launch the Audiobook Organizer graphical interface.

The GUI provides a three-pane layout for browsing, editing metadata,
and previewing file organization changes before applying them.

On macOS, the binary must be allowed in System Settings > Privacy & Security
the first time it runs.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputDir, _ := cmd.Flags().GetString("input")
		outputDir, _ := cmd.Flags().GetString("output")
		if err := guiapp.Run(inputDir, outputDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	guiCmd.Flags().StringP("input", "i", "", "Input directory to pre-load in the GUI")
	guiCmd.Flags().StringP("output", "o", "", "Output directory to pre-load in the GUI")
	rootCmd.AddCommand(guiCmd)
}
