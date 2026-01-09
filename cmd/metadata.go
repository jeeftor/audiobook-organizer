package cmd

import (
	"fmt"
	"os"

	"github.com/jeeftor/audiobook-organizer/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "View and explore audiobook metadata",
	Long: `Launch an interactive terminal UI to view audiobook metadata.

The metadata command provides a guided interface to:
  - Scan directories and view metadata from files
  - See metadata from both metadata.json and embedded tags
  - Explore available metadata fields
  - Build and test rename templates interactively
  - Customize field mappings for metadata.json files

This command helps you understand what metadata is available in your files
before organizing or renaming them.

Examples:
  # View metadata with default settings
  audiobook-organizer metadata --dir=/path/to/books

  # Force embedded metadata (ignore metadata.json)
  audiobook-organizer metadata --dir=/path --use-embedded-metadata

  # Flat mode (implies embedded metadata)
  audiobook-organizer metadata --dir=/path --flat

  # Custom field mapping for metadata.json
  audiobook-organizer metadata --dir=/path \
    --title-field=album \
    --author-fields=artist,album_artist`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check for required --dir flag - read directly from command flags
		inputDir, _ := cmd.Flags().GetString("dir")
		if inputDir == "" {
			inputDir, _ = cmd.Flags().GetString("input")
		}
		// Also check viper for config file values
		if inputDir == "" {
			inputDir = viper.GetString("dir")
		}
		if inputDir == "" {
			inputDir = viper.GetString("input")
		}
		if inputDir == "" {
			return fmt.Errorf("--dir must be specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get inputDir from command flags first, then viper
		inputDir, _ := cmd.Flags().GetString("dir")
		if inputDir == "" {
			inputDir, _ = cmd.Flags().GetString("input")
		}
		if inputDir == "" {
			inputDir = viper.GetString("dir")
		}
		if inputDir == "" {
			inputDir = viper.GetString("input")
		}

		// Also set viper values for the TUI to use
		viper.Set("dir", inputDir)

		// Get other flags and set them in viper for TUI
		useEmbedded, _ := cmd.Flags().GetBool("use-embedded-metadata")
		flat, _ := cmd.Flags().GetBool("flat")
		verbose, _ := cmd.Flags().GetBool("verbose")

		viper.Set("use-embedded-metadata", useEmbedded)
		viper.Set("flat", flat)
		viper.Set("verbose", verbose)

		// Field mapping flags
		if titleField, _ := cmd.Flags().GetString("title-field"); titleField != "" {
			viper.Set("title-field", titleField)
		}
		if seriesField, _ := cmd.Flags().GetString("series-field"); seriesField != "" {
			viper.Set("series-field", seriesField)
		}
		if authorFields, _ := cmd.Flags().GetString("author-fields"); authorFields != "" {
			viper.Set("author-fields", authorFields)
		}
		if trackField, _ := cmd.Flags().GetString("track-field"); trackField != "" {
			viper.Set("track-field", trackField)
		}

		if err := tui.RunRenameMode(inputDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(metadataCmd)

	// Basic flags
	metadataCmd.Flags().StringP("dir", "d", "", "Directory to scan for audiobooks (required)")
	metadataCmd.Flags().String("input", "", "Alias for --dir")
	metadataCmd.Flags().Bool("use-embedded-metadata", false, "Force use of embedded metadata (ignore metadata.json)")
	metadataCmd.Flags().Bool("flat", false, "Flat mode (implies --use-embedded-metadata)")
	metadataCmd.Flags().BoolP("verbose", "v", false, "Verbose output")

	// Field mapping flags (for metadata.json customization)
	metadataCmd.Flags().String("title-field", "", "Field to use for title (e.g., 'title', 'album')")
	metadataCmd.Flags().String("series-field", "", "Field to use for series (e.g., 'series', 'album')")
	metadataCmd.Flags().String("author-fields", "", "Comma-separated fields for authors (e.g., 'artist,album_artist')")
	metadataCmd.Flags().String("track-field", "", "Field to use for track number (e.g., 'track', 'track_number')")

	// Bind to viper
	viper.BindPFlag("dir", metadataCmd.Flags().Lookup("dir"))
	viper.BindPFlag("input", metadataCmd.Flags().Lookup("input"))
	viper.BindPFlag("use-embedded-metadata", metadataCmd.Flags().Lookup("use-embedded-metadata"))
	viper.BindPFlag("flat", metadataCmd.Flags().Lookup("flat"))
	viper.BindPFlag("verbose", metadataCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("title-field", metadataCmd.Flags().Lookup("title-field"))
	viper.BindPFlag("series-field", metadataCmd.Flags().Lookup("series-field"))
	viper.BindPFlag("author-fields", metadataCmd.Flags().Lookup("author-fields"))
	viper.BindPFlag("track-field", metadataCmd.Flags().Lookup("track-field"))
}
