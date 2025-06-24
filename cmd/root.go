package cmd

import (
	"audiobook-organizer/internal/organizer"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
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
	layout              string // Directory structure layout
	useSeriesAsTitle    bool   // Use Series field as the main title directory
	cfgFile             string
)

// envAliases maps config keys to their possible environment variable names
var envAliases = map[string][]string{
	"dir":                   {"AO_DIR", "AO_INPUT", "AUDIOBOOK_ORGANIZER_DIR", "AUDIOBOOK_ORGANIZER_INPUT"},
	"input":                 {"AO_DIR", "AO_INPUT", "AUDIOBOOK_ORGANIZER_DIR", "AUDIOBOOK_ORGANIZER_INPUT"},
	"out":                   {"AO_OUT", "AO_OUTPUT", "AUDIOBOOK_ORGANIZER_OUT", "AUDIOBOOK_ORGANIZER_OUTPUT"},
	"output":                {"AO_OUT", "AO_OUTPUT", "AUDIOBOOK_ORGANIZER_OUT", "AUDIOBOOK_ORGANIZER_OUTPUT"},
	"replace_space":         {"AO_REPLACE_SPACE", "AUDIOBOOK_ORGANIZER_REPLACE_SPACE"},
	"verbose":               {"AO_VERBOSE", "AUDIOBOOK_ORGANIZER_VERBOSE"},
	"dry-run":               {"AO_DRY_RUN", "AUDIOBOOK_ORGANIZER_DRY_RUN"},
	"undo":                  {"AO_UNDO", "AUDIOBOOK_ORGANIZER_UNDO"},
	"prompt":                {"AO_PROMPT", "AUDIOBOOK_ORGANIZER_PROMPT"},
	"remove-empty":          {"AO_REMOVE_EMPTY", "AUDIOBOOK_ORGANIZER_REMOVE_EMPTY"},
	"use-embedded-metadata": {"AO_USE_EMBEDDED_METADATA", "AUDIOBOOK_ORGANIZER_USE_EMBEDDED_METADATA"},
	"flat":                  {"AO_FLAT", "AUDIOBOOK_ORGANIZER_FLAT"},
	"layout":                {"AO_LAYOUT", "AUDIOBOOK_ORGANIZER_LAYOUT"},
	"use-series-as-title":   {"AO_USE_SERIES_AS_TITLE", "AUDIOBOOK_ORGANIZER_USE_SERIES_AS_TITLE"},
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
			viper.Set("use-embedded-metadata", true)
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

		org := organizer.NewOrganizer(
			&organizer.OrganizerConfig{
				BaseDir:             inputDir,
				OutputDir:           outputDir,
				ReplaceSpace:        viper.GetString("replace_space"),
				Verbose:             viper.GetBool("verbose"),
				DryRun:              viper.GetBool("dry-run"),
				Undo:                viper.GetBool("undo"),
				Prompt:              viper.GetBool("prompt"),
				RemoveEmpty:         viper.GetBool("remove-empty"),
				UseEmbeddedMetadata: viper.GetBool("use-embedded-metadata"),
				Flat:                viper.GetBool("flat"),
				Layout:              viper.GetString("layout"),
				UseSeriesAsTitle:    viper.GetBool("use-series-as-title"),
			},
		)

		if err := org.Execute(); err != nil {
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}

		// Print log file location if not in dry-run mode
		if !viper.GetBool("dry-run") {
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.audiobook-organizer.yaml)")

	// Command line flags with aliases
	rootCmd.Flags().StringVar(&inputDir, "dir", "", "Base directory to scan (alias for --input)")
	rootCmd.Flags().StringVar(&inputDir, "input", "", "Base directory to scan (alias for --dir)")
	rootCmd.Flags().StringVar(&outputDir, "out", "", "Output directory (alias for --output)")
	rootCmd.Flags().StringVar(&outputDir, "output", "", "Output directory (alias for --out)")
	rootCmd.Flags().StringVar(&replaceSpace, "replace_space", "", "Character to replace spaces")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without making changes")
	rootCmd.Flags().BoolVar(&undo, "undo", false, "Restore files to their original locations")
	rootCmd.Flags().BoolVar(&prompt, "prompt", false, "Prompt for confirmation before moving each book")
	rootCmd.Flags().BoolVar(&removeEmpty, "remove-empty", false, "Remove empty directories after moving files")
	rootCmd.Flags().BoolVar(&useEmbeddedMetadata, "use-embedded-metadata", false, "Use metadata embedded in EPUB files if metadata.json is not found")
	rootCmd.Flags().BoolVar(&flat, "flat", false, "Process files in a flat directory structure (automatically enables --use-embedded-metadata)")
	rootCmd.Flags().StringVar(&layout, "layout", "author-series-title", "Directory structure layout (options: author-series-title, author-title, author-only)")
	rootCmd.Flags().BoolVar(&useSeriesAsTitle, "use-series-as-title", false, "Use Series field as the main title directory (useful for MP3 files where Series contains the book title)")

	// Bind flags to viper
	viper.BindPFlag("dir", rootCmd.Flags().Lookup("dir"))
	viper.BindPFlag("input", rootCmd.Flags().Lookup("input"))
	viper.BindPFlag("out", rootCmd.Flags().Lookup("out"))
	viper.BindPFlag("output", rootCmd.Flags().Lookup("output"))
	viper.BindPFlag("replace_space", rootCmd.Flags().Lookup("replace_space"))
	viper.BindPFlag("verbose", rootCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("dry-run", rootCmd.Flags().Lookup("dry-run"))
	viper.BindPFlag("undo", rootCmd.Flags().Lookup("undo"))
	viper.BindPFlag("prompt", rootCmd.Flags().Lookup("prompt"))
	viper.BindPFlag("remove-empty", rootCmd.Flags().Lookup("remove-empty"))
	viper.BindPFlag("use-embedded-metadata", rootCmd.Flags().Lookup("use-embedded-metadata"))
	viper.BindPFlag("flat", rootCmd.Flags().Lookup("flat"))
	viper.BindPFlag("layout", rootCmd.Flags().Lookup("layout"))
	viper.BindPFlag("use-series-as-title", rootCmd.Flags().Lookup("use-series-as-title"))

	// Set up environment variable handling
	viper.SetEnvPrefix("AUDIOBOOK_ORGANIZER") // This will still be used for unmapped variables
	viper.AutomaticEnv()

	// Set up custom environment variable handling for our aliases
	for key := range envAliases {
		viper.RegisterAlias(key, strings.ToUpper(key))
		if value := getEnvValue(key); value != "" {
			viper.Set(key, value)
		}
	}

	// Custom validation instead of using MarkFlagRequired
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// First run the existing PreRun function
		if cmd.PreRun != nil {
			cmd.PreRun(cmd, args)
		}

		// Check if either dir or input flag is set
		if !cmd.Flags().Changed("dir") && !cmd.Flags().Changed("input") {
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
	}
}
