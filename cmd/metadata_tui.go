package cmd

import (
	"github.com/jeeftor/audiobook-organizer/internal/tui"
	"github.com/spf13/cobra"
)

// metadataTuiCmd represents the metadata-tui command for interactive metadata exploration.
var metadataTuiCmd = &cobra.Command{
	Use:   "metadata-tui",
	Short: "Start the Terminal User Interface (TUI) for metadata exploration",
	Long: `Launch a Terminal User Interface (TUI) for exploring audiobook metadata.

This interactive mode allows you to:
- Scan directories for audiobook files
- Preview metadata fields from both metadata.json and embedded sources
- Configure field mappings interactively
- Build and test rename templates
- Inspect the metadata available before organizing or renaming files`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if metadataInputDir(cmd) == "" {
			return errMetadataDirRequired()
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		inputDir := metadataInputDir(cmd)
		syncMetadataFlagsToViper(cmd, inputDir)
		return tui.RunRenameMode(inputDir)
	},
}

func init() {
	rootCmd.AddCommand(metadataTuiCmd)

	metadataTuiCmd.Flags().StringP("dir", "d", "", "Directory to scan for audiobooks")
	metadataTuiCmd.Flags().String("input", "", "Alias for --dir")
	metadataTuiCmd.Flags().
		Bool("use-embedded-metadata", false, "Force use of embedded metadata (ignore metadata.json)")
	metadataTuiCmd.Flags().Bool("flat", false, "Flat mode (implies --use-embedded-metadata)")
	metadataTuiCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	metadataTuiCmd.Flags().
		String("title-field", "", "Field to use for title (e.g., 'title', 'album')")
	metadataTuiCmd.Flags().
		String("series-field", "", "Field to use for series (e.g., 'series', 'album')")
	metadataTuiCmd.Flags().
		String("author-fields", "", "Comma-separated fields for authors (e.g., 'artist,album_artist')")
	metadataTuiCmd.Flags().
		String("track-field", "", "Field to use for track number (e.g., 'track', 'track_number')")
}
