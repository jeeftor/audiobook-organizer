package cmd

import (
	"audiobook-organizer/internal/organizer"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
)

var (
	baseDir      string
	outputDir    string
	replaceSpace string
	verbose      bool
	dryRun       bool
	undo         bool
	prompt       bool
)

var rootCmd = &cobra.Command{
	Use:   "audiobook-organizer",
	Short: "Organize audiobooks based on metadata.json files",
	Run: func(cmd *cobra.Command, args []string) {
		org := organizer.New(
			baseDir,
			outputDir,
			replaceSpace,
			verbose,
			dryRun,
			undo,
			prompt,
		)

		if err := org.Execute(); err != nil {
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}

		// Print log file location if not in dry-run mode
		if !dryRun {
			logPath := org.GetLogPath()
			color.Cyan("\n📝 Log file location: %s", logPath)
			color.Cyan("To undo these changes, run:")
			color.White("  audiobook-organizer --dir=%s --undo", baseDir)
			if outputDir != "" {
				color.White("  audiobook-organizer --dir=%s --out=%s --undo", baseDir, outputDir)
			}
		}
	},
}

func Execute() error {
	color.Cyan("🎧 Audiobook Organizer")
	color.Cyan("=====================")
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVar(&baseDir, "dir", "", "Base directory to scan")
	rootCmd.Flags().StringVar(&outputDir, "out", "", "Output directory (if different from base directory)")
	rootCmd.Flags().StringVar(&replaceSpace, "replace_space", "", "Character to replace spaces")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without making changes")
	rootCmd.Flags().BoolVar(&undo, "undo", false, "Restore files to their original locations")
	rootCmd.Flags().BoolVar(&prompt, "prompt", false, "Prompt for confirmation before moving each book")
	rootCmd.MarkFlagRequired("dir")
}
