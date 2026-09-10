package organizer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func safetyWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func safetyRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func safetyOrganizer(t *testing.T, dry bool) (*Organizer, string, string) {
	t.Helper()
	root := t.TempDir()
	source, output := filepath.Join(root, "input"), filepath.Join(root, "output")
	safetyWrite(
		t,
		filepath.Join(source, "book", "metadata.json"),
		`{"title":"Book","authors":["Author"]}`,
	)
	safetyWrite(t, filepath.Join(source, "book", "audio.mp3"), "original audio")
	org, err := NewOrganizer(&OrganizerConfig{BaseDir: source, OutputDir: output, DryRun: dry})
	if err != nil {
		t.Fatal(err)
	}
	return org, source, output
}

func TestSafetyOrganizeRejectsOccupiedDestination(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	target := filepath.Join(output, "Author", "Book", "audio.mp3")
	safetyWrite(t, target, "existing audiobook")
	if err := org.Execute(); err == nil {
		t.Fatal("collision must fail")
	}
	if got := safetyRead(t, target); got != "existing audiobook" {
		t.Fatalf("overwritten target: %q", got)
	}
	if got := safetyRead(t, filepath.Join(source, "book", "audio.mp3")); got != "original audio" {
		t.Fatal(got)
	}
	if _, err := os.Stat(filepath.Join(source, "book", "metadata.json")); err != nil {
		t.Fatal("metadata moved despite collision", err)
	}
}

func TestSafetyUndoDryRunPreservesFilesAndLog(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	if err := org.Execute(); err != nil {
		t.Fatal(err)
	}
	before := safetyRead(t, org.GetLogPath())
	undo, err := NewOrganizer(
		&OrganizerConfig{BaseDir: source, OutputDir: output, Undo: true, DryRun: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := undo.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := safetyRead(t, org.GetLogPath()); got != before {
		t.Fatal("undo preview changed log")
	}
	if got := safetyRead(t, filepath.Join(output, "Author", "Book", "audio.mp3")); got != "original audio" {
		t.Fatal(got)
	}
	if _, err := os.Stat(filepath.Join(source, "book", "audio.mp3")); !os.IsNotExist(err) {
		t.Fatal("undo preview restored file")
	}
}

func TestSafetyFailedUndoRetainsOnlyPendingMovesAndRetries(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	if err := org.Execute(); err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(source, "book", "audio.mp3")
	safetyWrite(t, original, "new original must survive")
	undo, err := NewOrganizer(&OrganizerConfig{BaseDir: source, OutputDir: output, Undo: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := undo.Execute(); err == nil {
		t.Fatal("blocked undo must fail")
	}
	if got := safetyRead(t, original); got != "new original must survive" {
		t.Fatal(got)
	}
	var entries []LogEntry
	if err := json.Unmarshal([]byte(safetyRead(t, org.GetLogPath())), &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || len(entries[0].Files) != 1 || entries[0].Files[0].From != "audio.mp3" {
		t.Fatalf("pending log: %+v", entries)
	}
	if err := os.Remove(original); err != nil {
		t.Fatal(err)
	}
	if err := undo.Execute(); err != nil {
		t.Fatal("retry failed", err)
	}
	if got := safetyRead(t, original); got != "original audio" {
		t.Fatal(got)
	}
	if _, err := os.Stat(org.GetLogPath()); !os.IsNotExist(err) {
		t.Fatal("completed undo retained log")
	}
}

func TestSafetyFailedMoveNeverAppearsInLogOrSummary(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	if err := os.MkdirAll(filepath.Join(output, "Author", "Book", "audio.mp3"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := org.Execute(); err == nil {
		t.Fatal("failed move reported success")
	}
	if len(org.GetSummary().Moves) != 0 {
		t.Fatal("failed book included in summary")
	}
	if _, err := os.Stat(org.GetLogPath()); err == nil {
		t.Fatal("failed moves logged")
	}
	if _, err := os.Stat(filepath.Join(source, "book", "metadata.json")); err != nil {
		t.Fatal(err)
	}
}

func TestSafetyPartialMoveLogsOnlyCompletedFiles(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	book := filepath.Join(source, "book")
	entries, err := os.ReadDir(book)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate the sidecar disappearing after directory enumeration.
	if err := os.Remove(filepath.Join(book, "metadata.json")); err != nil {
		t.Fatal(err)
	}
	files, err := org.processDirectoryFiles(entries, book, output, nil)
	if err == nil {
		t.Fatal("partial failure was swallowed")
	}
	if len(files) != 1 || files[0].From != "audio.mp3" {
		t.Fatalf("completed moves = %+v", files)
	}
	if err := org.updateLogAndCleanup(book, output, files); err != nil {
		t.Fatal(err)
	}
	var log []LogEntry
	if err := json.Unmarshal([]byte(safetyRead(t, org.GetLogPath())), &log); err != nil {
		t.Fatal(err)
	}
	if len(log) != 1 || len(log[0].Files) != 1 || log[0].Files[0].From != "audio.mp3" {
		t.Fatalf("inaccurate log: %+v", log)
	}
	if err := org.undoMoves(); err != nil {
		t.Fatal(err)
	}
	if safetyRead(t, filepath.Join(book, "audio.mp3")) != "original audio" {
		t.Fatal("partial undo lost audio")
	}
}

func TestSafetyCopyFallbackPreservesOccupiedTargetAndSupportsUndo(t *testing.T) {
	root := t.TempDir()
	source, target := filepath.Join(root, "source"), filepath.Join(root, "target")
	safetyWrite(t, source, "source bytes")
	safetyWrite(t, target, "target bytes")
	if err := copyNoReplace(source, target); err == nil {
		t.Fatal("copy overwrote destination")
	}
	if safetyRead(t, source) != "source bytes" || safetyRead(t, target) != "target bytes" {
		t.Fatal("collision changed data")
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := copyNoReplace(source, target); err != nil {
		t.Fatal(err)
	}
	if err := copyNoReplace(target, source); err != nil {
		t.Fatal("reverse copy failed", err)
	}
	if safetyRead(t, source) != "source bytes" {
		t.Fatal("copy round trip changed data")
	}
}

func TestSafetyDryRunNewOutputAndEmptyCleanup(t *testing.T) {
	org, source, output := safetyOrganizer(t, true)
	org.config.RemoveEmpty = true
	if err := os.Mkdir(filepath.Join(source, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- org.Execute() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("dry-run cleanup did not terminate")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("preview created output")
	}
	if _, err := os.Stat(filepath.Join(source, "empty")); err != nil {
		t.Fatal("preview removed empty dir", err)
	}
	if len(org.GetSummary().Moves) != 1 {
		t.Fatal("missing preview move")
	}
}

func TestSafetyMetadataComponentsAndSymlinkContainment(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	if got := org.SanitizePath("A/../../../escaped"); strings.ContainsAny(got, "/\\") {
		t.Fatalf("unsafe component %q", got)
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(output, "Author")); err != nil {
		t.Fatal(err)
	}
	if err := org.Execute(); err == nil {
		t.Fatal("destination symlink escaped output")
	}
	if _, err := os.Stat(filepath.Join(source, "book", "audio.mp3")); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatal("mutated outside output", err)
	}
}

func TestSafetyFlatSelectionAndLayouts(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/test-scenarios/single-file/single_book.mp3")
	if err != nil {
		t.Fatal(err)
	}
	input, output := t.TempDir(), t.TempDir()
	for _, name := range []string{"a.mp3", "b.mp3"} {
		safetyWrite(t, filepath.Join(input, name), string(fixture))
	}
	org, err := NewOrganizer(
		&OrganizerConfig{
			BaseDir:             input,
			OutputDir:           output,
			Flat:                true,
			UseEmbeddedMetadata: true,
			AllowedSourcePaths:  []string{filepath.Join(input, "a.mp3")},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := org.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(org.GetSummary().Moves) != 1 {
		t.Fatalf("moved unselected files: %+v", org.GetSummary())
	}
	if got := safetyRead(t, filepath.Join(input, "b.mp3")); got != string(fixture) {
		t.Fatal("unselected file changed")
	}
	metadata := Metadata{Authors: []string{"Author"}, Title: "Book", Series: []string{"Series #2"}}
	for layout, relative := range map[string]string{"author-only": "Author", "author-title": "Author/Book", "author-series": "Author/Series", "author-series-title": "Author/Series/Book", "author-series-title-number": "Author/Series/#2 - Book", "series-title": "Series/Book", "series-title-number": "Series/#2 - Book"} {
		t.Run(layout, func(t *testing.T) {
			cfg := &OrganizerConfig{BaseDir: input, OutputDir: output, Layout: layout, DryRun: true}
			one, err := NewOrganizer(cfg)
			if err != nil {
				t.Fatal(err)
			}
			got, err := one.calculateSingleFileTargetDirE(filepath.Join(input, "b.mp3"), metadata)
			if err != nil || got != filepath.Join(output, filepath.FromSlash(relative)) {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}
