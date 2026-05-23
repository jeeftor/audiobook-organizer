package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type metadataJSONOutput struct {
	Files   []metadataJSONFile  `json:"files"`
	Summary metadataJSONSummary `json:"summary"`
}

type metadataJSONFile struct {
	Path        string                 `json:"path"`
	SourceType  string                 `json:"source_type"`
	Title       string                 `json:"title"`
	Authors     []string               `json:"authors"`
	Series      []string               `json:"series"`
	TrackNumber int                    `json:"track_number"`
	TrackTitle  string                 `json:"track_title,omitempty"`
	Album       string                 `json:"album"`
	RawData     map[string]interface{} `json:"raw_data,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

type metadataJSONSummary struct {
	FilesScanned int `json:"files_scanned"`
	Errors       int `json:"errors"`
}

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Inspect audiobook metadata non-interactively",
	Long: `Inspect audiobook metadata without launching the interactive TUI.

The metadata command scans directories and prints extracted metadata in the
terminal. Use --json for machine-readable output, or metadata-tui for the guided
terminal workflow.

The output includes each file path, metadata source type, title, authors,
series, track number, album, and extraction errors when present.

Examples:
  # Inspect metadata in the terminal
  audiobook-organizer metadata --dir=/path/to/books

  # Inspect metadata as JSON for scripts and CI
  audiobook-organizer metadata --dir=/path/to/books --json

  # Force embedded metadata (ignore metadata.json)
  audiobook-organizer metadata --dir=/path --use-embedded-metadata

  # Flat mode (implies embedded metadata)
  audiobook-organizer metadata --dir=/path --flat

  # Launch the interactive metadata TUI
  audiobook-organizer metadata-tui --dir=/path/to/books

  # Custom field mapping for metadata.json
  audiobook-organizer metadata --dir=/path --json \
    --title-field=album \
    --author-fields=artist,album_artist`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if metadataInputDir(cmd) == "" {
			return errMetadataDirRequired()
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		inputDir := metadataInputDir(cmd)
		syncMetadataFlagsToViper(cmd, inputDir)

		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return runMetadataJSON(cmd, inputDir)
		}

		return runMetadataText(cmd, inputDir)
	},
}

func init() {
	rootCmd.AddCommand(metadataCmd)

	// Basic flags
	metadataCmd.Flags().StringP("dir", "d", "", "Directory to scan for audiobooks (required)")
	metadataCmd.Flags().String("input", "", "Alias for --dir")
	metadataCmd.Flags().
		Bool("use-embedded-metadata", false, "Force use of embedded metadata (ignore metadata.json)")
	metadataCmd.Flags().Bool("flat", false, "Flat mode (implies --use-embedded-metadata)")
	metadataCmd.Flags().Bool("json", false, "Write metadata scan results as JSON")
	metadataCmd.Flags().BoolP("verbose", "v", false, "Verbose output")

	// Field mapping flags (for metadata.json customization)
	metadataCmd.Flags().String("title-field", "", "Field to use for title (e.g., 'title', 'album')")
	metadataCmd.Flags().
		String("series-field", "", "Field to use for series (e.g., 'series', 'album')")
	metadataCmd.Flags().
		String("author-fields", "", "Comma-separated fields for authors (e.g., 'artist,album_artist')")
	metadataCmd.Flags().
		String("track-field", "", "Field to use for track number (e.g., 'track', 'track_number')")

	// Bind to viper
	viper.BindPFlag("dir", metadataCmd.Flags().Lookup("dir"))
	viper.BindPFlag("input", metadataCmd.Flags().Lookup("input"))
	viper.BindPFlag("use-embedded-metadata", metadataCmd.Flags().Lookup("use-embedded-metadata"))
	viper.BindPFlag("flat", metadataCmd.Flags().Lookup("flat"))
	viper.BindPFlag("json", metadataCmd.Flags().Lookup("json"))
	viper.BindPFlag("verbose", metadataCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("title-field", metadataCmd.Flags().Lookup("title-field"))
	viper.BindPFlag("series-field", metadataCmd.Flags().Lookup("series-field"))
	viper.BindPFlag("author-fields", metadataCmd.Flags().Lookup("author-fields"))
	viper.BindPFlag("track-field", metadataCmd.Flags().Lookup("track-field"))
}

func runMetadataJSON(cmd *cobra.Command, inputDir string) error {
	output, err := scanMetadataJSON(inputDir, metadataUseEmbedded(cmd), metadataFieldMapping(cmd))
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func runMetadataText(cmd *cobra.Command, inputDir string) error {
	output, err := scanMetadataJSON(inputDir, metadataUseEmbedded(cmd), metadataFieldMapping(cmd))
	if err != nil {
		return err
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		writeMetadataVerbose(cmd.OutOrStdout(), inputDir, output)
		return nil
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Metadata scan: %s\n", inputDir)
	fmt.Fprintf(out, "Files scanned: %d\n", output.Summary.FilesScanned)
	fmt.Fprintf(out, "Errors: %d\n\n", output.Summary.Errors)

	for i, file := range output.Files {
		if i > 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "Path: %s\n", file.Path)
		fmt.Fprintf(out, "  Source: %s\n", valueOrDash(file.SourceType))
		fmt.Fprintf(out, "  Title: %s\n", valueOrDash(file.Title))
		fmt.Fprintf(out, "  Authors: %s\n", joinedOrDash(file.Authors))
		fmt.Fprintf(out, "  Series: %s\n", joinedOrDash(file.Series))
		if file.TrackNumber > 0 {
			fmt.Fprintf(out, "  Track: %d\n", file.TrackNumber)
		} else {
			fmt.Fprintln(out, "  Track: -")
		}
		if file.TrackTitle != "" {
			fmt.Fprintf(out, "  Track Title: %s\n", file.TrackTitle)
		}
		fmt.Fprintf(out, "  Album: %s\n", valueOrDash(file.Album))
		if file.Error != "" {
			fmt.Fprintf(out, "  Error: %s\n", file.Error)
		}
		writeAdditionalMetadataFields(out, file.RawData)
	}

	return nil
}

func writeMetadataVerbose(out io.Writer, inputDir string, output metadataJSONOutput) {
	fmt.Fprintln(out, "🎧 Metadata scan")
	fmt.Fprintf(out, "  📁 Directory: %s\n", inputDir)
	fmt.Fprintf(out, "  📄 Files scanned: %d\n", output.Summary.FilesScanned)
	fmt.Fprintf(out, "  ⚠️ Errors: %d\n\n", output.Summary.Errors)

	for i, file := range output.Files {
		if i > 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "📄 %s\n", file.Path)
		fmt.Fprintf(out, "  🧭 Source: %s\n", valueOrDash(file.SourceType))
		fmt.Fprintf(out, "  📖 Title: %s\n", valueOrDash(file.Title))
		fmt.Fprintf(out, "  ✍️ Authors: %s\n", joinedOrDash(file.Authors))
		fmt.Fprintf(out, "  📚 Series: %s\n", joinedOrDash(file.Series))
		if file.TrackNumber > 0 {
			fmt.Fprintf(out, "  🔢 Track: %d\n", file.TrackNumber)
		} else {
			fmt.Fprintln(out, "  🔢 Track: -")
		}
		if file.TrackTitle != "" {
			fmt.Fprintf(out, "  🎙️ Track Title: %s\n", file.TrackTitle)
		}
		fmt.Fprintf(out, "  💿 Album: %s\n", valueOrDash(file.Album))
		if file.Error != "" {
			fmt.Fprintf(out, "  ⚠️ Error: %s\n", file.Error)
		}
		writeAdditionalMetadataFieldsVerbose(out, file.RawData)
	}
}

func scanMetadataJSON(
	inputDir string,
	useEmbedded bool,
	fieldMapping organizer.FieldMapping,
) (metadataJSONOutput, error) {
	info, err := os.Stat(inputDir)
	if err != nil {
		return metadataJSONOutput{}, fmt.Errorf(
			"error accessing metadata directory %s: %w",
			inputDir,
			err,
		)
	}
	if !info.IsDir() {
		return metadataJSONOutput{}, fmt.Errorf("%s is not a directory", inputDir)
	}

	output := metadataJSONOutput{
		Files: []metadataJSONFile{},
	}

	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !organizer.IsSupportedFile(filepath.Ext(path)) {
			return nil
		}

		output.Summary.FilesScanned++
		file := metadataJSONFile{
			Path:       path,
			SourceType: metadataSourceTypeForPath(path),
			Authors:    []string{},
			Series:     []string{},
		}

		provider := organizer.NewMetadataProvider(path, useEmbedded)
		metadata, err := provider.GetMetadata()
		if err != nil {
			file.Error = fmt.Sprintf("failed to extract metadata: %v", err)
			output.Summary.Errors++
			output.Files = append(output.Files, file)
			return nil
		}

		if !fieldMapping.IsEmpty() {
			metadata.ApplyFieldMapping(fieldMapping)
		}

		file.SourceType = metadata.SourceType
		file.Title = metadata.Title
		file.Authors = nonNilStrings(metadata.Authors)
		file.Series = nonNilStrings(metadata.Series)
		file.TrackNumber = metadata.TrackNumber
		file.TrackTitle = metadata.TrackTitle
		file.Album = metadata.Album
		file.RawData = metadata.RawData
		output.Files = append(output.Files, file)
		return nil
	})
	if err != nil {
		return metadataJSONOutput{}, fmt.Errorf(
			"error scanning metadata directory %s: %w",
			inputDir,
			err,
		)
	}

	return output, nil
}

func metadataInputDir(cmd *cobra.Command) string {
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
	return inputDir
}

func errMetadataDirRequired() error {
	return fmt.Errorf("--dir must be specified")
}

func syncMetadataFlagsToViper(cmd *cobra.Command, inputDir string) {
	viper.Set("dir", inputDir)

	useEmbedded, _ := cmd.Flags().GetBool("use-embedded-metadata")
	flat, _ := cmd.Flags().GetBool("flat")
	verbose, _ := cmd.Flags().GetBool("verbose")

	viper.Set("use-embedded-metadata", useEmbedded)
	viper.Set("flat", flat)
	viper.Set("verbose", verbose)

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
}

func metadataUseEmbedded(cmd *cobra.Command) bool {
	useEmbedded, _ := cmd.Flags().GetBool("use-embedded-metadata")
	flat, _ := cmd.Flags().GetBool("flat")
	return useEmbedded || flat
}

func metadataFieldMapping(cmd *cobra.Command) organizer.FieldMapping {
	authorFields := []string{}
	if rawAuthorFields, _ := cmd.Flags().GetString("author-fields"); rawAuthorFields != "" {
		for _, field := range strings.Split(rawAuthorFields, ",") {
			trimmed := strings.TrimSpace(field)
			if trimmed != "" {
				authorFields = append(authorFields, trimmed)
			}
		}
	}

	titleField, _ := cmd.Flags().GetString("title-field")
	seriesField, _ := cmd.Flags().GetString("series-field")
	trackField, _ := cmd.Flags().GetString("track-field")

	return organizer.FieldMapping{
		TitleField:   titleField,
		SeriesField:  seriesField,
		AuthorFields: authorFields,
		TrackField:   trackField,
	}
}

func metadataSourceTypeForPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".epub":
		return "epub"
	case ".mp3", ".m4b", ".m4a", ".ogg", ".flac":
		return "audio"
	default:
		return "unknown"
	}
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func joinedOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

func valueOrDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func writeAdditionalMetadataFields(out io.Writer, rawData map[string]interface{}) {
	fields := additionalMetadataFields(rawData)
	if len(fields) == 0 {
		return
	}

	fmt.Fprintln(out, "  Additional Fields:")
	for _, field := range fields {
		fmt.Fprintf(out, "    %s: %s\n", field.key, formatMetadataValue(field.value))
	}
}

func writeAdditionalMetadataFieldsVerbose(out io.Writer, rawData map[string]interface{}) {
	fields := additionalMetadataFields(rawData)
	if len(fields) == 0 {
		return
	}

	fmt.Fprintln(out, "  🔎 Additional Fields:")
	for _, field := range fields {
		fmt.Fprintf(out, "    %s: %s\n", field.key, formatMetadataValue(field.value))
	}
}

type metadataField struct {
	key   string
	value interface{}
}

func additionalMetadataFields(rawData map[string]interface{}) []metadataField {
	if len(rawData) == 0 {
		return nil
	}

	keys := make([]string, 0, len(rawData))
	for key, value := range rawData {
		if skipAdditionalMetadataField(key, value) {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fields := make([]metadataField, 0, len(keys))
	for _, key := range keys {
		fields = append(fields, metadataField{key: key, value: rawData[key]})
	}
	return fields
}

func skipAdditionalMetadataField(key string, value interface{}) bool {
	if value == nil {
		return true
	}
	if strings.HasPrefix(key, "_") {
		return true
	}
	if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
		return true
	}
	if isZeroMetadataNumber(value) {
		return true
	}
	if values, ok := value.([]string); ok && len(values) == 0 {
		return true
	}
	if values, ok := value.([]interface{}); ok && len(values) == 0 {
		return true
	}

	switch key {
	case "album", "artist", "authors", "series", "title", "track", "track_number", "track_title":
		return true
	case "discnumber":
		return true
	}
	return false
}

func isZeroMetadataNumber(value interface{}) bool {
	switch v := value.(type) {
	case int:
		return v == 0
	case int8:
		return v == 0
	case int16:
		return v == 0
	case int32:
		return v == 0
	case int64:
		return v == 0
	case uint:
		return v == 0
	case uint8:
		return v == 0
	case uint16:
		return v == 0
	case uint32:
		return v == 0
	case uint64:
		return v == 0
	case float32:
		return v == 0
	case float64:
		return v == 0
	default:
		return false
	}
}

func formatMetadataValue(value interface{}) string {
	switch v := value.(type) {
	case []string:
		return truncateMetadataValue(strings.Join(v, ", "))
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, fmt.Sprintf("%v", item))
		}
		return truncateMetadataValue(strings.Join(parts, ", "))
	default:
		return truncateMetadataValue(fmt.Sprintf("%v", value))
	}
}

func truncateMetadataValue(value string) string {
	const maxLen = 120
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen-3] + "..."
}
