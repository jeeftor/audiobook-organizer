package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Constants for field names to avoid duplication
const ( // This makes sonar pass
	titleFieldKey      = "title-field"
	seriesFieldKey     = "series-field"
	authorFieldsKey    = "author-fields"
	trackFieldKey      = "track-field"
	discFieldKey       = "disc-field"
	useEmbeddedMetaKey = "use-embedded-metadata"
	removeEmptyKey     = "remove-empty"
	dryRunKey          = "dry-run"
)

var (
	inputDir            string // Combined input from --dir and --input
	outputDir           string // Combined output from --out and --output
	replaceSpace        string
	verbose             bool
	dryRun              bool
	undo                bool
	prompt              bool
	removeEmpty         bool
	useEmbeddedMetadata bool
	flat                bool
	skipErrors          bool
	layout              string // Directory structure layout

	// Field mapping flags
	titleField   string
	seriesField  string
	authorFields string // Comma-separated list
	trackField   string
	discField    string

	cfgFile string
)

// envAliases maps config keys to their possible environment variable names
var envAliases = map[string][]string{
	"dir": {
		"AO_DIR",
		"AO_INPUT",
		"AUDIOBOOK_ORGANIZER_DIR",
		"AUDIOBOOK_ORGANIZER_INPUT",
	},
	"input": {
		"AO_DIR",
		"AO_INPUT",
		"AUDIOBOOK_ORGANIZER_DIR",
		"AUDIOBOOK_ORGANIZER_INPUT",
	},
	"out": {
		"AO_OUT",
		"AO_OUTPUT",
		"AUDIOBOOK_ORGANIZER_OUT",
		"AUDIOBOOK_ORGANIZER_OUTPUT",
	},
	"output": {
		"AO_OUT",
		"AO_OUTPUT",
		"AUDIOBOOK_ORGANIZER_OUT",
		"AUDIOBOOK_ORGANIZER_OUTPUT",
	},
	"replace_space":    {"AO_REPLACE_SPACE", "AUDIOBOOK_ORGANIZER_REPLACE_SPACE"},
	"verbose":          {"AO_VERBOSE", "AUDIOBOOK_ORGANIZER_VERBOSE"},
	dryRunKey:          {"AO_DRY_RUN", "AUDIOBOOK_ORGANIZER_DRY_RUN"},
	"undo":             {"AO_UNDO", "AUDIOBOOK_ORGANIZER_UNDO"},
	"prompt":           {"AO_PROMPT", "AUDIOBOOK_ORGANIZER_PROMPT"},
	removeEmptyKey:     {"AO_REMOVE_EMPTY", "AUDIOBOOK_ORGANIZER_REMOVE_EMPTY"},
	useEmbeddedMetaKey: {"AO_USE_EMBEDDED_METADATA", "AUDIOBOOK_ORGANIZER_USE_EMBEDDED_METADATA"},
	"flat":             {"AO_FLAT", "AUDIOBOOK_ORGANIZER_FLAT"},
	"layout":           {"AO_LAYOUT", "AUDIOBOOK_ORGANIZER_LAYOUT"},

	// Field mapping environment variables
	titleFieldKey:   {"AO_TITLE_FIELD", "AUDIOBOOK_ORGANIZER_TITLE_FIELD"},
	seriesFieldKey:  {"AO_SERIES_FIELD", "AUDIOBOOK_ORGANIZER_SERIES_FIELD"},
	authorFieldsKey: {"AO_AUTHOR_FIELDS", "AUDIOBOOK_ORGANIZER_AUTHOR_FIELDS"},
	trackFieldKey:   {"AO_TRACK_FIELD", "AUDIOBOOK_ORGANIZER_TRACK_FIELD"},
	discFieldKey:    {"AO_DISC_FIELD", "AUDIOBOOK_ORGANIZER_DISC_FIELD"},

	// Rename command environment variables
	"rename-template":      {"AO_RENAME_TEMPLATE", "AUDIOBOOK_ORGANIZER_RENAME_TEMPLATE"},
	"rename-author-format": {"AO_RENAME_AUTHOR_FORMAT", "AUDIOBOOK_ORGANIZER_RENAME_AUTHOR_FORMAT"},
	"rename-recursive":     {"AO_RENAME_RECURSIVE", "AUDIOBOOK_ORGANIZER_RENAME_RECURSIVE"},
	"rename-strict":        {"AO_RENAME_STRICT", "AUDIOBOOK_ORGANIZER_RENAME_STRICT"},
	"rename-preserve-path": {"AO_RENAME_PRESERVE_PATH", "AUDIOBOOK_ORGANIZER_RENAME_PRESERVE_PATH"},
	"rename-prompt":        {"AO_RENAME_PROMPT", "AUDIOBOOK_ORGANIZER_RENAME_PROMPT"},
}

var rootCmd = &cobra.Command{
	Use:   "audiobook-organizer",
	Short: "Organize audiobooks based on metadata.json files",
	PreRun: func(cmd *cobra.Command, args []string) {
		// Store the original PreRun logic in a separate function
		handleInputAliases(cmd)
		// Handle input directory aliases
		if cmd.Flags().Changed("input") {
			viper.Set("dir", viper.GetString("input"))
		} else if cmd.Flags().Changed("dir") {
			viper.Set("input", viper.GetString("dir"))
		}

		// Handle output directory aliases
		if cmd.Flags().Changed("output") {
			viper.Set("out", viper.GetString("output"))
		} else if cmd.Flags().Changed("out") {
			viper.Set("output", viper.GetString("out"))
		}

		// If flat mode is enabled, automatically enable embedded metadata
		if viper.GetBool("flat") {
			viper.Set(useEmbeddedMetaKey, true)
			if viper.GetBool("verbose") {
				color.Cyan("ℹ️ Flat mode enabled: automatically using embedded metadata")
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get the final input directory (either from --dir or --input)
		inputDir := viper.GetString("dir")
		if inputDir == "" {
			inputDir = viper.GetString("input")
		}

		// Get the final output directory (either from --out or --output)
		outputDir := viper.GetString("out")
		if outputDir == "" {
			outputDir = viper.GetString("output")
		}

		// Parse author fields from comma-separated string
		authorFieldsList := []string{}
		if af := viper.GetString(authorFieldsKey); af != "" {
			authorFieldsList = strings.Split(af, ",")
		}

		org, err := organizer.NewOrganizer(
			&organizer.OrganizerConfig{
				BaseDir:             inputDir,
				OutputDir:           outputDir,
				ReplaceSpace:        viper.GetString("replace_space"),
				Verbose:             viper.GetBool("verbose"),
				DryRun:              viper.GetBool(dryRunKey),
				Undo:                viper.GetBool("undo"),
				Prompt:              viper.GetBool("prompt"),
				RemoveEmpty:         viper.GetBool(removeEmptyKey),
				UseEmbeddedMetadata: viper.GetBool(useEmbeddedMetaKey),
				Flat:                viper.GetBool("flat"),
				SkipErrors:          viper.GetBool("skip-errors"),
				Layout:              viper.GetString("layout"),
				FieldMapping: organizer.FieldMapping{
					TitleField:   viper.GetString(titleFieldKey),
					SeriesField:  viper.GetString(seriesFieldKey),
					AuthorFields: authorFieldsList,
					TrackField:   viper.GetString(trackFieldKey),
					DiscField:    viper.GetString(discFieldKey),
				},
			},
		)
		if err != nil {
			organizer.PrintRed("Configuration error: %v", err)
			os.Exit(1)
		}

		if err := org.Execute(); err != nil {
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}

		// Print log file location if not in dry-run mode
		if !viper.GetBool(dryRunKey) {
			logPath := org.GetLogPath()
			color.Cyan("\n📝 Log file location: %s", logPath)
			color.Cyan("To undo these changes, run:")
			color.White("  audiobook-organizer --input=%s --undo", inputDir)
			if outputDir != "" {
				color.White("  audiobook-organizer --input=%s --output=%s --undo",
					inputDir, outputDir)
			}
		}
	},
}

func Execute() error {
	color.Cyan("🎧 Audiobook Organizer")
	color.Cyan("=====================")
	return rootCmd.Execute()
}

// getEnvValue checks all possible environment variable names for a config key
// handleInputAliases handles the aliasing between dir/input and out/output flags
func handleInputAliases(cmd *cobra.Command) {
	// Handle input directory aliases
	if cmd.Flags().Changed("input") {
		viper.Set("dir", viper.GetString("input"))
	} else if cmd.Flags().Changed("dir") {
		viper.Set("input", viper.GetString("dir"))
	}

	// Handle output directory aliases
	if cmd.Flags().Changed("output") {
		viper.Set("out", viper.GetString("output"))
	} else if cmd.Flags().Changed("out") {
		viper.Set("output", viper.GetString("out"))
	}
}

func getEnvValue(key string) string {
	if aliases, ok := envAliases[key]; ok {
		for _, alias := range aliases {
			if value := os.Getenv(alias); value != "" {
				return value
			}
		}
	}
	return ""
}

func init() {
	cobra.OnInitialize(initConfig)

	// Config file flag
	rootCmd.PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.audiobook-organizer.yaml)")

	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().
		StringVar(&inputDir, "dir", "", "Base directory to scan (alias for --input)")
	rootCmd.PersistentFlags().
		StringVar(&inputDir, "input", "", "Base directory to scan (alias for --dir)")
	rootCmd.PersistentFlags().
		StringVar(&outputDir, "out", "", "Output directory (alias for --output)")
	rootCmd.PersistentFlags().
		StringVar(&outputDir, "output", "", "Output directory (alias for --out)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().
		BoolVar(&dryRun, dryRunKey, false, "Show what would happen without making changes")
	rootCmd.PersistentFlags().
		BoolVar(&useEmbeddedMetadata, useEmbeddedMetaKey, false, "Use metadata embedded in EPUB files if metadata.json is not found")
	rootCmd.PersistentFlags().
		BoolVar(&flat, "flat", false, "Process files in a flat directory structure (automatically enables --use-embedded-metadata)")
	rootCmd.PersistentFlags().
		BoolVar(&skipErrors, "skip-errors", false, "Skip files with missing/invalid metadata instead of stopping")

	// Local flags (only for root command)
	rootCmd.Flags().StringVar(&replaceSpace, "replace_space", "", "Character to replace spaces")
	rootCmd.Flags().BoolVar(&undo, "undo", false, "Restore files to their original locations")
	rootCmd.Flags().
		BoolVar(&prompt, "prompt", false, "Prompt for confirmation before moving each book")
	rootCmd.Flags().
		BoolVar(&removeEmpty, removeEmptyKey, false, "Remove empty directories after moving files")
	rootCmd.Flags().
		StringVarP(&layout, "layout", "l", "author-series-title", "Directory structure layout:\n  - author-series-title:        Author/Series/Title/ (default)\n  - author-series-title-number: Author/Series/#1 - Title/ (include series number in title)\n  - author-title:               Author/Title/ (ignore series)\n  - author-only:                Author/ (flatten all books)")

	// Field mapping flags (persistent for all commands)
	rootCmd.PersistentFlags().
		StringVar(&titleField, titleFieldKey, "", "Field to use as title (e.g., 'album', 'title', 'track_title')")
	rootCmd.PersistentFlags().
		StringVar(&seriesField, seriesFieldKey, "", "Field to use as series (e.g., 'series', 'album')")
	rootCmd.PersistentFlags().
		StringVar(&authorFields, authorFieldsKey, "", "Comma-separated list of fields to try for author (e.g., 'authors,narrators,album_artist,artist')")
	rootCmd.PersistentFlags().
		StringVar(&trackField, trackFieldKey, "", "Field to use for track number (e.g., 'track', 'track_number', 'trck', 'trk')")
	rootCmd.PersistentFlags().
		StringVar(&discField, discFieldKey, "", "Field to use for disc number (e.g., 'disc', 'discnumber', 'disk', 'tpos')")

	// Bind persistent flags to viper
	viper.BindPFlag("dir", rootCmd.PersistentFlags().Lookup("dir"))
	viper.BindPFlag("input", rootCmd.PersistentFlags().Lookup("input"))
	viper.BindPFlag("out", rootCmd.PersistentFlags().Lookup("out"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag(dryRunKey, rootCmd.PersistentFlags().Lookup(dryRunKey))
	viper.BindPFlag(useEmbeddedMetaKey, rootCmd.PersistentFlags().Lookup(useEmbeddedMetaKey))
	viper.BindPFlag("flat", rootCmd.PersistentFlags().Lookup("flat"))
	viper.BindPFlag("skip-errors", rootCmd.PersistentFlags().Lookup("skip-errors"))
	viper.BindPFlag(titleFieldKey, rootCmd.PersistentFlags().Lookup(titleFieldKey))
	viper.BindPFlag(seriesFieldKey, rootCmd.PersistentFlags().Lookup(seriesFieldKey))
	viper.BindPFlag(authorFieldsKey, rootCmd.PersistentFlags().Lookup(authorFieldsKey))
	viper.BindPFlag(trackFieldKey, rootCmd.PersistentFlags().Lookup(trackFieldKey))
	viper.BindPFlag(discFieldKey, rootCmd.PersistentFlags().Lookup(discFieldKey))

	// Bind local flags to viper
	viper.BindPFlag("replace_space", rootCmd.Flags().Lookup("replace_space"))
	viper.BindPFlag("undo", rootCmd.Flags().Lookup("undo"))
	viper.BindPFlag("prompt", rootCmd.Flags().Lookup("prompt"))
	viper.BindPFlag(removeEmptyKey, rootCmd.Flags().Lookup(removeEmptyKey))
	viper.BindPFlag("layout", rootCmd.Flags().Lookup("layout"))

	// Set up environment variable handling
	viper.SetEnvPrefix("AUDIOBOOK_ORGANIZER") // This will still be used for unmapped variables
	viper.AutomaticEnv()

	// Custom validation instead of using MarkFlagRequired
	// Only validate when running the root command itself, not subcommands
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip validation if a subcommand is being called
		// Check if we have any subcommands and if args indicate a subcommand
		if cmd.Name() != "audiobook-organizer" {
			// This is a subcommand, skip root validation
			return nil
		}

		// First run the existing PreRun function
		if cmd.PreRun != nil {
			cmd.PreRun(cmd, args)
		}

		// Check if input directory is set via flags, env vars, or config file
		inputDir := viper.GetString("dir")
		if inputDir == "" {
			inputDir = viper.GetString("input")
		}

		if inputDir == "" {
			return fmt.Errorf("either --dir or --input must be specified")
		}
		return nil
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".audiobook-organizer" (without extension)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".audiobook-organizer")
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			color.Cyan("Using config file: %s", viper.ConfigFileUsed())
		}
	} else {
		// Only show error if a config file was explicitly specified
		if cfgFile != "" {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
	}

	// Set up custom environment variable handling for our aliases
	// This needs to happen after config file is read but before validation
	for key := range envAliases {
		viper.RegisterAlias(key, strings.ToUpper(key))
		if value := getEnvValue(key); value != "" {
			// Only set if not already set by config file or flags
			if !viper.IsSet(key) {
				viper.Set(key, value)
			}
		}
	}
}
