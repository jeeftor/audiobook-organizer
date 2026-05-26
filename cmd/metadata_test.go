package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	"github.com/spf13/cobra"
)

func TestMetadataCmd_HasJSONFlag(t *testing.T) {
	if metadataCmd == nil {
		t.Fatal("metadataCmd is nil")
	}

	if metadataCmd.Use != "metadata" {
		t.Errorf("metadataCmd.Use = %q, want %q", metadataCmd.Use, "metadata")
	}

	if flag := metadataCmd.Flags().Lookup("json"); flag == nil {
		t.Fatal("metadataCmd missing json flag")
	}
	if flag := metadataCmd.Flags().Lookup("pretty"); flag == nil {
		t.Fatal("metadataCmd missing pretty flag")
	}
}

func TestMetadataTuiCmd_Registered(t *testing.T) {
	if metadataTuiCmd == nil {
		t.Fatal("metadataTuiCmd is nil")
	}

	if metadataTuiCmd.Use != "metadata-tui" {
		t.Errorf("metadataTuiCmd.Use = %q, want %q", metadataTuiCmd.Use, "metadata-tui")
	}

	found := false
	for _, command := range rootCmd.Commands() {
		if command.Use == "metadata-tui" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("metadata-tui command is not registered")
	}
}

func TestShouldPrintStartupBanner_MetadataJSON(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "metadata suppresses banner",
			args: []string{"metadata", "--dir", "testdata/mp3flat"},
			want: false,
		},
		{
			name: "metadata json true suppresses banner",
			args: []string{"metadata", "--json=true"},
			want: false,
		},
		{
			name: "metadata json false suppresses banner",
			args: []string{"metadata", "--json=false"},
			want: false,
		},
		{
			name: "metadata json zero suppresses banner",
			args: []string{"metadata", "--json=0"},
			want: false,
		},
		{
			name: "metadata tui prints banner",
			args: []string{"metadata-tui", "--dir", "testdata/mp3flat"},
			want: true,
		},
		{
			name: "term diagnostics suppress banner",
			args: []string{"term"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldPrintStartupBanner(tt.args); got != tt.want {
				t.Errorf("shouldPrintStartupBanner(%v) = %t, want %t", tt.args, got, tt.want)
			}
		})
	}
}

func TestScanMetadataJSON_WithMP3FlatFixture(t *testing.T) {
	output, err := scanMetadataJSON(
		filepath.Join("..", "testdata", "mp3flat"),
		true,
		organizer.FieldMapping{},
	)
	if err != nil {
		t.Fatalf("scanMetadataJSON() error = %v", err)
	}

	if output.Summary.FilesScanned == 0 {
		t.Fatal("scanMetadataJSON() scanned no files")
	}
	if len(output.Files) != output.Summary.FilesScanned {
		t.Fatalf(
			"scanMetadataJSON() files = %d, want %d",
			len(output.Files),
			output.Summary.FilesScanned,
		)
	}

	first := output.Files[0]
	if first.Path == "" {
		t.Fatal("first metadata result missing path")
	}
	if first.SourceType == "" {
		t.Fatal("first metadata result missing source_type")
	}
	if first.Authors == nil {
		t.Fatal("first metadata result authors should be an empty array or populated array")
	}
	if first.Series == nil {
		t.Fatal("first metadata result series should be an empty array or populated array")
	}
}

func TestRunMetadataJSON_WritesParseableJSON(t *testing.T) {
	tmpDir := t.TempDir()
	fixturePath := filepath.Join(
		"..",
		"testdata",
		"mp3flat",
		"charlesdexterward_01_lovecraft_64kb.mp3",
	)
	fixtureBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read MP3 fixture %s: %v", fixturePath, err)
	}
	audioPath := filepath.Join(tmpDir, "book.mp3")
	if err := os.WriteFile(audioPath, fixtureBytes, 0o644); err != nil {
		t.Fatalf("failed to write test audio file: %v", err)
	}
	metadataPath := filepath.Join(tmpDir, "metadata.json")
	metadataContent := `{
		"title": "The JSON Book",
		"authors": ["Example Author"],
		"series": ["Example Series"],
		"track_number": 7,
		"publishedYear": 2024,
		"narrator": "Example Narrator"
	}`
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0o644); err != nil {
		t.Fatalf("failed to write metadata.json: %v", err)
	}

	cmd := newMetadataJSONTestCommand(t)
	cmd.Flags().Set("dir", tmpDir)
	cmd.Flags().Set("json", "true")

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runMetadataJSON(cmd, tmpDir); err != nil {
		t.Fatalf("runMetadataJSON() error = %v", err)
	}

	var output metadataJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("metadata JSON did not parse: %v\n%s", err, buf.String())
	}

	if output.Summary.FilesScanned != 1 {
		t.Fatalf("files_scanned = %d, want 1", output.Summary.FilesScanned)
	}
	if len(output.Files) != 1 {
		t.Fatalf("files length = %d, want 1", len(output.Files))
	}

	file := output.Files[0]
	if file.Path != audioPath {
		t.Errorf("path = %q, want %q", file.Path, audioPath)
	}
	if file.SourceType != "json" {
		t.Errorf("source_type = %q, want %q", file.SourceType, "json")
	}
	if file.Title != "The JSON Book" {
		t.Errorf("title = %q, want %q", file.Title, "The JSON Book")
	}
	if len(file.Authors) != 1 || file.Authors[0] != "Example Author" {
		t.Errorf("authors = %v, want [Example Author]", file.Authors)
	}
	if len(file.Series) != 1 || file.Series[0] != "Example Series" {
		t.Errorf("series = %v, want [Example Series]", file.Series)
	}
	if file.TrackNumber != 1 {
		t.Errorf("track_number = %d, want 1", file.TrackNumber)
	}
	if got := file.RawData["publishedYear"]; got != float64(2024) {
		t.Errorf("raw_data[publishedYear] = %v, want 2024", got)
	}
	if got := file.RawData["narrator"]; got != "Example Narrator" {
		t.Errorf("raw_data[narrator] = %v, want Example Narrator", got)
	}
}

func TestRunMetadataText_WritesTerminalOutput(t *testing.T) {
	tmpDir := t.TempDir()
	fixturePath := filepath.Join(
		"..",
		"testdata",
		"mp3flat",
		"charlesdexterward_01_lovecraft_64kb.mp3",
	)
	fixtureBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read MP3 fixture %s: %v", fixturePath, err)
	}
	audioPath := filepath.Join(tmpDir, "book.mp3")
	if err := os.WriteFile(audioPath, fixtureBytes, 0o644); err != nil {
		t.Fatalf("failed to write test audio file: %v", err)
	}
	metadataPath := filepath.Join(tmpDir, "metadata.json")
	metadataContent := `{
		"title": "The Text Book",
		"authors": ["Example Author"],
		"series": ["Example Series"],
		"publishedYear": 2024,
		"narrator": "Example Narrator"
	}`
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0o644); err != nil {
		t.Fatalf("failed to write metadata.json: %v", err)
	}

	cmd := newMetadataJSONTestCommand(t)
	cmd.Flags().Set("dir", tmpDir)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runMetadataText(cmd, tmpDir); err != nil {
		t.Fatalf("runMetadataText() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Metadata scan: "+tmpDir) {
		t.Fatalf("terminal output missing scan header:\n%s", output)
	}
	if !strings.Contains(output, "Files scanned: 1") {
		t.Fatalf("terminal output missing file count:\n%s", output)
	}
	if !strings.Contains(output, "Path: "+audioPath) {
		t.Fatalf("terminal output missing file path:\n%s", output)
	}
	if !strings.Contains(output, "Source: json") {
		t.Fatalf("terminal output missing source type:\n%s", output)
	}
	if !strings.Contains(output, "Additional Fields:") {
		t.Fatalf("terminal output missing additional fields section:\n%s", output)
	}
	if !strings.Contains(output, "narrator: Example Narrator") {
		t.Fatalf("terminal output missing additional narrator field:\n%s", output)
	}
	if !strings.Contains(output, "publishedYear: 2024") {
		t.Fatalf("terminal output missing additional publishedYear field:\n%s", output)
	}
}

func TestRunMetadataTextVerbose_WritesVisualTerminalOutput(t *testing.T) {
	tmpDir := t.TempDir()
	fixturePath := filepath.Join(
		"..",
		"testdata",
		"mp3flat",
		"charlesdexterward_01_lovecraft_64kb.mp3",
	)
	fixtureBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read MP3 fixture %s: %v", fixturePath, err)
	}
	audioPath := filepath.Join(tmpDir, "book.mp3")
	if err := os.WriteFile(audioPath, fixtureBytes, 0o644); err != nil {
		t.Fatalf("failed to write test audio file: %v", err)
	}
	metadataPath := filepath.Join(tmpDir, "metadata.json")
	metadataContent := `{
		"title": "The Visual Book",
		"authors": ["Example Author"],
		"series": ["Example Series"],
		"publishedYear": 2024,
		"narrator": "Example Narrator"
	}`
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0o644); err != nil {
		t.Fatalf("failed to write metadata.json: %v", err)
	}

	cmd := newMetadataJSONTestCommand(t)
	cmd.Flags().Set("dir", tmpDir)
	cmd.Flags().Set("verbose", "true")

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runMetadataText(cmd, tmpDir); err != nil {
		t.Fatalf("runMetadataText() error = %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"🎧 Metadata scan",
		"📁 Directory: " + tmpDir,
		"📄 Files scanned: 1",
		"📄 " + audioPath,
		"JSON Metadata",
		"Hybrid Mode:",
		"Title: The Visual Book",
		"Authors: Example Author",
		"Series: Example Series",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("verbose terminal output missing %q:\n%s", want, output)
		}
	}
}

func TestRunMetadataTextPretty_UsesFormatterBackedOutput(t *testing.T) {
	tmpDir := t.TempDir()
	fixturePath := filepath.Join(
		"..",
		"testdata",
		"mp3flat",
		"charlesdexterward_01_lovecraft_64kb.mp3",
	)
	fixtureBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read MP3 fixture %s: %v", fixturePath, err)
	}
	audioPath := filepath.Join(tmpDir, "book.mp3")
	if err := os.WriteFile(audioPath, fixtureBytes, 0o644); err != nil {
		t.Fatalf("failed to write test audio file: %v", err)
	}
	metadataPath := filepath.Join(tmpDir, "metadata.json")
	metadataContent := `{
		"title": "The Pretty Book",
		"authors": ["Example Author"],
		"series": ["Example Series"]
	}`
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0o644); err != nil {
		t.Fatalf("failed to write metadata.json: %v", err)
	}

	cmd := newMetadataJSONTestCommand(t)
	cmd.Flags().Set("dir", tmpDir)
	cmd.Flags().Set("pretty", "true")

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runMetadataText(cmd, tmpDir); err != nil {
		t.Fatalf("runMetadataText() error = %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"🎧 Metadata scan",
		"📄 " + audioPath,
		"JSON Metadata",
		"Hybrid Mode:",
		"Title: The Pretty Book",
		"Authors: Example Author",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("pretty terminal output missing %q:\n%s", want, output)
		}
	}
}

func TestScanMetadataJSON_ReportsExtractionErrors(t *testing.T) {
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "broken.mp3")
	if err := os.WriteFile(audioPath, []byte("not real audio"), 0o644); err != nil {
		t.Fatalf("failed to write test audio file: %v", err)
	}

	output, err := scanMetadataJSON(tmpDir, true, organizer.FieldMapping{})
	if err != nil {
		t.Fatalf("scanMetadataJSON() error = %v", err)
	}

	if output.Summary.FilesScanned != 1 {
		t.Fatalf("files_scanned = %d, want 1", output.Summary.FilesScanned)
	}
	if output.Summary.Errors != 1 {
		t.Fatalf("errors = %d, want 1", output.Summary.Errors)
	}
	if len(output.Files) != 1 {
		t.Fatalf("files length = %d, want 1", len(output.Files))
	}
	if output.Files[0].Path != audioPath {
		t.Errorf("path = %q, want %q", output.Files[0].Path, audioPath)
	}
	if output.Files[0].Error == "" {
		t.Fatal("expected extraction error in JSON output")
	}
}

func TestAdditionalMetadataFields_FiltersEmptyCoreAndZeroValues(t *testing.T) {
	fields := additionalMetadataFields(map[string]interface{}{
		"title":       "Core Title",
		"authors":     []string{"Core Author"},
		"_source":     "internal marker",
		"empty":       "",
		"empty_list":  []string{},
		"year":        0,
		"disc":        0,
		"track_total": 0,
		"genre":       "speech",
		"identifier":  "id-123",
	})

	got := make([]string, 0, len(fields))
	for _, field := range fields {
		got = append(got, field.key)
	}

	want := []string{"genre", "identifier"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("additional fields = %v, want %v", got, want)
	}
}

func newMetadataJSONTestCommand(t *testing.T) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "metadata-test"}
	cmd.Flags().StringP("dir", "d", "", "Directory to scan for audiobooks")
	cmd.Flags().String("input", "", "Alias for --dir")
	cmd.Flags().Bool("use-embedded-metadata", false, "Force use of embedded metadata")
	cmd.Flags().Bool("flat", false, "Flat mode")
	cmd.Flags().Bool("json", false, "Write metadata scan results as JSON")
	cmd.Flags().Bool("pretty", false, "Write formatter-backed pretty metadata output")
	cmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	cmd.Flags().String("title-field", "", "Field to use for title")
	cmd.Flags().String("series-field", "", "Field to use for series")
	cmd.Flags().String("author-fields", "", "Comma-separated fields for authors")
	cmd.Flags().String("track-field", "", "Field to use for track number")
	return cmd
}
