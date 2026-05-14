package cmd

import "github.com/spf13/cobra"

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Alias for the local web UI",
	Long: `Start the Audiobook Organizer local web UI.

Use "audiobook-organizer web" directly for the canonical command.`,
	RunE: runWeb,
}

func init() {
	addWebFlags(guiCmd)
	rootCmd.AddCommand(guiCmd)
}
