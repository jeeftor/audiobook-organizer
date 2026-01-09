package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	renameTemplate     string
	renameAuthorFormat string
	renameRecursive    bool
	renameStrictMode   bool
	renamePreservePath bool
	renamePrompt       bool
)

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename audiobook files based on metadata templates",
	Long: `Rename audiobook files in-place using metadata templates.

The rename command allows you to rename audiobook files based on flexible
templates that use metadata fields from metadata.json files or embedded metadata.

By default, the command prefers metadata.json if present, falling back to embedded
metadata. Use --use-embedded-metadata to force embedded metadata extraction.

Examples:
  # Preview renames with default template
  audiobook-organizer rename --dir=/path/to/books --dry-run

  # Rename with custom template
  audiobook-organizer rename --dir=/path --template="{author} - {title}"

  # Use Last, First author format
  audiobook-organizer rename --dir=/path --author-format=last-first

  # Prompt before each rename
  audiobook-organizer rename --dir=/path --prompt

  # Undo previous rename operation
  audiobook-organizer rename --dir=/path --undo

  # Force embedded metadata (ignore metadata.json)
  audiobook-organizer rename --dir=/path --use-embedded-metadata

  # Use flat mode (implies embedded metadata)
  audiobook-organizer rename --dir=/path --flat

Template Fields:
  {author}         - First author (formatted)
  {authors}        - All authors (comma-separated)
  {title}          - Book title
  {series}         - Series name (without number)
  {series_number}  - Series number only
  {track}          - Track number (zero-padded)
  {album}          - Album field
  {year}           - Publication year
  {narrator}       - Narrator (if available)

Author Formats:
  first-last  - Brandon Sanderson (default)
  last-first  - Sanderson, Brandon
  preserve    - Keep original format`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check for required --dir flag - read directly from command flags
		inputDir, _ := cmd.Flags().GetString("dir")
		if inputDir == "" {
			// Also check viper for config file values
			inputDir = viper.GetString("dir")
		}
		if inputDir == "" {
			return fmt.Errorf("--dir must be specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		runRename()
	},
}

func runRename() {
	// Get directory - try viper first (includes flags, env vars, config)
	inputDir := viper.GetString("dir")
	if inputDir == "" {
		organizer.PrintRed("Error: --dir must be specified")
		os.Exit(1)
	}

	// Parse author format
	var authorFormat organizer.AuthorFormat
	switch renameAuthorFormat {
	case "first-last":
		authorFormat = organizer.AuthorFormatFirstLast
	case "last-first":
		authorFormat = organizer.AuthorFormatLastFirst
	case "preserve":
		authorFormat = organizer.AuthorFormatPreserve
	default:
		organizer.PrintRed("Invalid author format: %s. Valid options: first-last, last-first, preserve", renameAuthorFormat)
		os.Exit(1)
	}

	// Parse field mapping from flags
	authorFieldsList := []string{}
	if af := viper.GetString("author-fields"); af != "" {
		authorFieldsList = strings.Split(af, ",")
	}

	// Determine if we should use embedded metadata
	useEmbedded := viper.GetBool("use-embedded-metadata") || viper.GetBool("flat")

	config := &organizer.RenamerConfig{
		BaseDir:      inputDir,
		Template:     renameTemplate,
		DryRun:       viper.GetBool("dry-run"),
		Verbose:      viper.GetBool("verbose"),
		AuthorFormat: authorFormat,
		Recursive:    renameRecursive,
		FieldMapping: organizer.FieldMapping{
			TitleField:   viper.GetString("title-field"),
			SeriesField:  viper.GetString("series-field"),
			AuthorFields: authorFieldsList,
			TrackField:   viper.GetString("track-field"),
			DiscField:    viper.GetString("disc-field"),
		},
		ReplaceSpace:        viper.GetString("replace_space"),
		StrictMode:          renameStrictMode,
		PreservePath:        renamePreservePath,
		PromptEnabled:       renamePrompt,
		UseEmbeddedMetadata: useEmbedded,
	}

	renamer, err := organizer.NewRenamer(config)
	if err != nil {
		organizer.PrintRed("Error creating renamer: %v", err)
		os.Exit(1)
	}

	// Handle undo
	if viper.GetBool("rename-undo") {
		organizer.PrintYellow("Undoing previous rename operations...")
		if err := renamer.UndoRenames(); err != nil {
			organizer.PrintRed("Error during undo: %v", err)
			os.Exit(1)
		}
		organizer.PrintGreen("✅ Undo complete!")
		return
	}

	// Preview changes (especially useful for dry-run)
	if viper.GetBool("dry-run") || viper.GetBool("verbose") {
		candidates, err := renamer.ScanFiles()
		if err != nil {
			organizer.PrintRed("Error scanning files: %v", err)
			os.Exit(1)
		}

		organizer.PrintCyan("\n📋 Proposed Changes:")
		changesShown := 0
		skipped := 0
		for _, candidate := range candidates {
			// Skip if no change needed
			if candidate.CurrentPath == candidate.ProposedPath {
				skipped++
				continue
			}

			currentName := filepath.Base(candidate.CurrentPath)
			newName := filepath.Base(candidate.ProposedPath)

			if candidate.IsConflict {
				organizer.PrintYellow("  %s → %s (conflict resolved)", currentName, newName)
			} else {
				organizer.PrintGreen("  %s → %s", currentName, newName)
			}
			changesShown++
		}

		if skipped > 0 {
			organizer.PrintBase("\n  (%d files already have correct names)", skipped)
		}

		if changesShown == 0 {
			organizer.PrintYellow("\n  No changes needed - all files already have correct names!")
			return
		}
		organizer.PrintBase("")
	}

	// Execute rename
	if err := renamer.Execute(); err != nil {
		organizer.PrintRed("Error during rename: %v", err)
		os.Exit(1)
	}

	// Print summary
	summary := renamer.GetSummary()
	organizer.PrintCyan("\n📊 Rename Summary:")
	organizer.PrintBase("  Files scanned: %d", summary.FilesScanned)
	organizer.PrintGreen("  Files renamed: %d", summary.FilesRenamed)
	if summary.FilesSkipped > 0 {
		organizer.PrintYellow("  Files skipped: %d", summary.FilesSkipped)
	}
	if summary.ConflictsFound > 0 {
		organizer.PrintMagenta("  Conflicts resolved: %d", summary.ConflictsFound)
	}
	if len(summary.Errors) > 0 {
		organizer.PrintRed("  Errors: %d", len(summary.Errors))
		for _, errMsg := range summary.Errors {
			organizer.PrintRed("    - %s", errMsg)
		}
	}

	if !viper.GetBool("dry-run") && summary.FilesRenamed > 0 {
		logPath := filepath.Join(inputDir, ".abook-rename.log")
		organizer.PrintCyan("\n📝 Log file: %s", logPath)
		organizer.PrintCyan("To undo: audiobook-organizer rename --dir=%s --undo", inputDir)
	} else if viper.GetBool("dry-run") {
		organizer.PrintYellow("\n🔍 This was a dry run - no files were actually renamed")
	}
}

var renameHelpTemplateCmd = &cobra.Command{
	Use:   "help-template",
	Short: "Show detailed template field reference",
	Long: `Display comprehensive information about all available template fields
and their usage for renaming audiobook files.

This command provides detailed documentation on:
- All available template placeholders
- Field descriptions and examples
- Author formatting options
- Advanced template patterns`,
	Run: func(cmd *cobra.Command, args []string) {
		showTemplateHelp()
	},
}

func showTemplateHelp() {
	organizer.PrintCyan("═══════════════════════════════════════════════════════════════")
	organizer.PrintCyan("  📚 Audiobook Rename Template Field Reference")
	organizer.PrintCyan("═══════════════════════════════════════════════════════════════\n")

	organizer.PrintGreen("BASIC FIELDS:")
	organizer.PrintBase("  {author}         First author name (formatted per --author-format)")
	organizer.PrintBase("                   Example: 'Brandon Sanderson'")
	organizer.PrintBase("")
	organizer.PrintBase("  {authors}        All authors comma-separated")
	organizer.PrintBase("                   Example: 'Brandon Sanderson, Mary Robinette Kowal'")
	organizer.PrintBase("")
	organizer.PrintBase("  {title}          Book title")
	organizer.PrintBase("                   Example: 'The Way of Kings'")
	organizer.PrintBase("")

	organizer.PrintGreen("SERIES FIELDS:")
	organizer.PrintBase("  {series}         Series name (without number)")
	organizer.PrintBase("                   Example: 'The Stormlight Archive'")
	organizer.PrintBase("")
	organizer.PrintBase("  {series_number}  Series number only")
	organizer.PrintBase("                   Example: '1' or '2.5'")
	organizer.PrintBase("")

	organizer.PrintGreen("AUDIO FIELDS:")
	organizer.PrintBase("  {track}          Track number (zero-padded to 2 digits)")
	organizer.PrintBase("                   Example: '01', '02', '15'")
	organizer.PrintBase("")
	organizer.PrintBase("  {album}          Album field from metadata")
	organizer.PrintBase("                   Example: 'Mistborn Trilogy'")
	organizer.PrintBase("")
	organizer.PrintBase("  {narrator}       Narrator name (if available)")
	organizer.PrintBase("                   Example: 'Michael Kramer'")
	organizer.PrintBase("")

	organizer.PrintGreen("OTHER FIELDS:")
	organizer.PrintBase("  {year}           Publication year")
	organizer.PrintBase("                   Example: '2010'")
	organizer.PrintBase("")

	organizer.PrintCyan("\nAUTHOR FORMAT OPTIONS (--author-format):")
	organizer.PrintBase("  first-last       Brandon Sanderson (default)")
	organizer.PrintBase("  last-first       Sanderson, Brandon")
	organizer.PrintBase("  preserve         Keep original format from metadata")
	organizer.PrintBase("")

	organizer.PrintCyan("\nTEMPLATE EXAMPLES:")
	organizer.PrintBase("  Default:")
	organizer.PrintBase("    {author} - {series} {series_number} - {title}")
	organizer.PrintBase("    → Brandon Sanderson - Mistborn 1 - The Final Empire")
	organizer.PrintBase("")
	organizer.PrintBase("  Simple:")
	organizer.PrintBase("    {author} - {title}")
	organizer.PrintBase("    → Brandon Sanderson - The Way of Kings")
	organizer.PrintBase("")
	organizer.PrintBase("  With track number:")
	organizer.PrintBase("    {author} - {title} - Part {track}")
	organizer.PrintBase("    → Brandon Sanderson - The Way of Kings - Part 01")
	organizer.PrintBase("")
	organizer.PrintBase("  Podcast style:")
	organizer.PrintBase("    {series} S{series_number} - {title} [{narrator}]")
	organizer.PrintBase("    → The Stormlight Archive S1 - The Way of Kings [Michael Kramer]")
	organizer.PrintBase("")

	organizer.PrintCyan("\nUSAGE:")
	organizer.PrintBase("  audiobook-organizer rename --dir=/path/to/books \\")
	organizer.PrintBase("    --template=\"{author} - {title}\" \\")
	organizer.PrintBase("    --dry-run")
	organizer.PrintBase("")

	organizer.PrintCyan("\nTIPS:")
	organizer.PrintBase("  • Use --dry-run to preview changes before applying")
	organizer.PrintBase("  • Combine with --verbose for detailed output")
	organizer.PrintBase("  • Use --strict to error on missing fields")
	organizer.PrintBase("  • Missing fields are removed from the template")
	organizer.PrintBase("  • Use --replace_space to replace spaces (e.g., with '.')")
	organizer.PrintBase("")

	organizer.PrintCyan("═══════════════════════════════════════════════════════════════\n")
}

func init() {
	rootCmd.AddCommand(renameCmd)
	renameCmd.AddCommand(renameHelpTemplateCmd)

	// Rename-specific flags (inherit --dir, --out, --verbose, --dry-run, etc. from root)
	renameCmd.Flags().StringVar(&renameTemplate, "template", "{author} - {series} {series_number} - {title}", "Filename template with placeholders")
	renameCmd.Flags().StringVar(&renameAuthorFormat, "author-format", "first-last", "Author name format: first-last, last-first, preserve")
	renameCmd.Flags().BoolVar(&renameRecursive, "recursive", true, "Recursively process subdirectories")
	renameCmd.Flags().BoolVar(&renameStrictMode, "strict", false, "Error on missing template fields")
	renameCmd.Flags().BoolVar(&renamePreservePath, "preserve-path", true, "Only rename filename, preserve directory structure")
	renameCmd.Flags().BoolVar(&renamePrompt, "prompt", false, "Prompt before renaming each file")
	renameCmd.Flags().Bool("undo", false, "Undo previous rename operations")

	// Bind rename-specific flags to viper
	viper.BindPFlag("rename-template", renameCmd.Flags().Lookup("template"))
	viper.BindPFlag("rename-author-format", renameCmd.Flags().Lookup("author-format"))
	viper.BindPFlag("rename-recursive", renameCmd.Flags().Lookup("recursive"))
	viper.BindPFlag("rename-strict", renameCmd.Flags().Lookup("strict"))
	viper.BindPFlag("rename-preserve-path", renameCmd.Flags().Lookup("preserve-path"))
	viper.BindPFlag("rename-prompt", renameCmd.Flags().Lookup("prompt"))
	viper.BindPFlag("rename-undo", renameCmd.Flags().Lookup("undo"))
}
