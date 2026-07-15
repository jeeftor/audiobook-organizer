package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

func TestRenameUsesABSMetadataSourceForPreviewAndRun(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	inputDir := t.TempDir()
	filePath := filepath.Join(inputDir, "source.mp3")
	writeFile(t, filePath, "fake audio")

	absServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/api/libraries/lib-main/items" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{{
				"id":        "item-1",
				"libraryId": "lib-main",
				"path":      "/abs/input",
				"libraryFiles": []map[string]any{{
					"metadata": map[string]string{"path": "/abs/input/source.mp3"},
				}},
				"media": map[string]any{
					"metadata": map[string]any{
						"title": "ABS Rename Book", "authors": []map[string]string{{"name": "ABS Author"}},
					},
					"audioFiles": []map[string]any{{
						"metadata": map[string]string{
							"path": "/abs/input/source.mp3",
						}, "trackNumFromMeta": 4,
					}},
				},
			}},
			"total": 1,
		})
	}))
	defer absServer.Close()

	config := RenameConfigDTO{
		BaseDir: inputDir, Template: "{track} - {author} - {title}", AuthorFormat: "first-last",
		Recursive: true, PreservePath: true, MetadataSource: metadataSourceABS,
		ABS: ABSConfigDTO{
			URL:       absServer.URL,
			Token:     "test-token",
			LibraryID: "lib-main",
			PathMappings: []PathMappingDTO{{
				ABSPrefix: "/abs/input", LocalPrefix: inputDir,
			}},
		},
	}
	preview, err := service.PreviewRename(context.Background(), RenameRequest{Config: config})
	if err != nil {
		t.Fatalf("PreviewRename() error = %v", err)
	}
	if got, want := filepath.Base(preview.Candidates[0].ProposedPath), "04 - ABS Author - ABS Rename Book.mp3"; got != want {
		t.Fatalf("preview proposed filename = %q, want %q", got, want)
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("preview moved source file: %v", err)
	}

	run, err := service.RunRename(context.Background(), RenameRequest{Config: config})
	if err != nil {
		t.Fatalf("RunRename() error = %v", err)
	}
	if run.Summary.FilesRenamed != 1 {
		t.Fatalf("FilesRenamed = %d, want 1", run.Summary.FilesRenamed)
	}
	assertPathExists(t, filepath.Join(inputDir, "04 - ABS Author - ABS Rename Book.mp3"))
}

func TestPreviewRenameReturnsRealSummaryState(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	inputDir := createRenameFixture(t)

	resp, err := service.PreviewRename(context.Background(), RenameRequest{
		Config: RenameConfigDTO{
			BaseDir:      inputDir,
			Template:     "{author} - {title}",
			DryRun:       false,
			AuthorFormat: "first-last",
			Recursive:    true,
			PreservePath: true,
			FieldMapping: FieldMappingDTO{
				TitleField:   "title",
				SeriesField:  "series",
				AuthorFields: []string{"authors"},
			},
		},
	})
	if err != nil {
		t.Fatalf("PreviewRename() error = %v", err)
	}

	if got := len(resp.Candidates); got != 4 {
		t.Fatalf("candidate count = %d, want 4", got)
	}
	if resp.Summary.FilesScanned != 4 {
		t.Fatalf("FilesScanned = %d, want 4", resp.Summary.FilesScanned)
	}
	if resp.Summary.FilesSkipped != 2 {
		t.Fatalf("FilesSkipped = %d, want 2", resp.Summary.FilesSkipped)
	}
	if resp.Summary.ConflictsFound != 1 {
		t.Fatalf("ConflictsFound = %d, want 1", resp.Summary.ConflictsFound)
	}
	if len(resp.Summary.Errors) != 1 {
		t.Fatalf("Errors length = %d, want 1", len(resp.Summary.Errors))
	}

	conflict := findRenameCandidate(t, resp.Candidates, "02-conflict-b")
	if !conflict.IsConflict {
		t.Fatal("second duplicate candidate should be marked as conflict")
	}
	if got := filepath.Base(conflict.ProposedPath); got != "Conflict Author - Conflict Book (2).mp3" {
		t.Fatalf(
			"conflict proposed filename = %q, want %q",
			got,
			"Conflict Author - Conflict Book (2).mp3",
		)
	}

	noop := findRenameCandidate(t, resp.Candidates, "03-noop")
	if !noop.IsNoOp {
		t.Fatal("matching filename candidate should be marked as no-op")
	}

	broken := findRenameCandidate(t, resp.Candidates, "04-broken")
	if broken.Error == "" {
		t.Fatal("broken audio candidate should report an extraction error")
	}
}

func TestRunRenameAppliesCandidatesAndWritesLog(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	inputDir := createRenameFixture(t)

	resp, err := service.RunRename(context.Background(), RenameRequest{
		Config: RenameConfigDTO{
			BaseDir:      inputDir,
			Template:     "{author} - {title}",
			DryRun:       true,
			AuthorFormat: "first-last",
			Recursive:    true,
			PreservePath: true,
			FieldMapping: FieldMappingDTO{
				TitleField:   "title",
				SeriesField:  "series",
				AuthorFields: []string{"authors"},
			},
		},
	})
	if err != nil {
		t.Fatalf("RunRename() error = %v", err)
	}

	if got := len(resp.Candidates); got != 4 {
		t.Fatalf("candidate count = %d, want 4", got)
	}
	if resp.Summary.FilesScanned != 4 {
		t.Fatalf("FilesScanned = %d, want 4", resp.Summary.FilesScanned)
	}
	if resp.Summary.FilesRenamed != 2 {
		t.Fatalf("FilesRenamed = %d, want 2", resp.Summary.FilesRenamed)
	}
	if resp.Summary.FilesSkipped != 2 {
		t.Fatalf("FilesSkipped = %d, want 2", resp.Summary.FilesSkipped)
	}
	if resp.Summary.ConflictsFound != 1 {
		t.Fatalf("ConflictsFound = %d, want 1", resp.Summary.ConflictsFound)
	}
	if len(resp.Summary.Errors) != 1 {
		t.Fatalf("Errors length = %d, want 1", len(resp.Summary.Errors))
	}

	wantLogPath := filepath.Join(inputDir, ".abook-rename.log")
	if resp.LogPath != wantLogPath {
		t.Fatalf("LogPath = %q, want %q", resp.LogPath, wantLogPath)
	}
	assertPathExists(t, wantLogPath)
	assertPathExists(
		t,
		filepath.Join(inputDir, "01-conflict-a", "Conflict Author - Conflict Book.mp3"),
	)
	assertPathExists(
		t,
		filepath.Join(inputDir, "02-conflict-b", "Conflict Author - Conflict Book (2).mp3"),
	)
	assertPathExists(t, filepath.Join(inputDir, "03-noop", "Noop Author - Noop Book.mp3"))
	assertPathExists(t, filepath.Join(inputDir, "04-broken", "broken.mp3"))
	assertPathMissing(t, filepath.Join(inputDir, "01-conflict-a", "original-a.mp3"))
	assertPathMissing(t, filepath.Join(inputDir, "02-conflict-b", "original-b.mp3"))
}

func TestRunRenameHonorsAllowedCurrentPaths(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	inputDir := createRenameFixture(t)
	selectedPath := filepath.Join(inputDir, "01-conflict-a", "original-a.mp3")

	resp, err := service.RunRename(context.Background(), RenameRequest{
		Config: RenameConfigDTO{
			BaseDir:             inputDir,
			Template:            "{author} - {title}",
			DryRun:              true,
			AuthorFormat:        "first-last",
			Recursive:           true,
			PreservePath:        true,
			AllowedCurrentPaths: []string{selectedPath},
			FieldMapping: FieldMappingDTO{
				TitleField:   "title",
				SeriesField:  "series",
				AuthorFields: []string{"authors"},
			},
		},
	})
	if err != nil {
		t.Fatalf("RunRename() error = %v", err)
	}

	if got := len(resp.Candidates); got != 1 {
		t.Fatalf("candidate count = %d, want 1", got)
	}
	if resp.Summary.FilesScanned != 1 {
		t.Fatalf("FilesScanned = %d, want 1", resp.Summary.FilesScanned)
	}
	if resp.Summary.FilesRenamed != 1 {
		t.Fatalf("FilesRenamed = %d, want 1", resp.Summary.FilesRenamed)
	}
	if resp.Summary.FilesSkipped != 0 {
		t.Fatalf("FilesSkipped = %d, want 0", resp.Summary.FilesSkipped)
	}
	if resp.Summary.ConflictsFound != 0 {
		t.Fatalf("ConflictsFound = %d, want 0", resp.Summary.ConflictsFound)
	}
	if len(resp.Summary.Errors) != 0 {
		t.Fatalf("Errors length = %d, want 0", len(resp.Summary.Errors))
	}

	assertPathExists(
		t,
		filepath.Join(inputDir, "01-conflict-a", "Conflict Author - Conflict Book.mp3"),
	)
	assertPathExists(t, filepath.Join(inputDir, "02-conflict-b", "original-b.mp3"))
	assertPathExists(t, filepath.Join(inputDir, "03-noop", "Noop Author - Noop Book.mp3"))
	assertPathExists(t, filepath.Join(inputDir, "04-broken", "broken.mp3"))
	assertPathMissing(t, selectedPath)
	assertPathMissing(
		t,
		filepath.Join(inputDir, "02-conflict-b", "Conflict Author - Conflict Book.mp3"),
	)
}

func createRenameFixture(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	sourceAudio := filepath.Join(
		"..",
		"..",
		"testdata",
		"mp3flat",
		"charlesdexterward_01_lovecraft_64kb.mp3",
	)
	createRenameBook(t, root, "01-conflict-a", "original-a.mp3", sourceAudio, `{
		"title": "Conflict Book",
		"authors": ["Conflict Author"],
		"series": ["Rename Series #1"]
	}`)
	createRenameBook(t, root, "02-conflict-b", "original-b.mp3", sourceAudio, `{
		"title": "Conflict Book",
		"authors": ["Conflict Author"],
		"series": ["Rename Series #1"]
	}`)
	createRenameBook(t, root, "03-noop", "Noop Author - Noop Book.mp3", sourceAudio, `{
		"title": "Noop Book",
		"authors": ["Noop Author"]
	}`)

	brokenDir := filepath.Join(root, "04-broken")
	if err := os.MkdirAll(brokenDir, 0o755); err != nil {
		t.Fatalf("create broken dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(brokenDir, "broken.mp3"), []byte("not audio"), 0o644); err != nil {
		t.Fatalf("write broken audio: %v", err)
	}

	return root
}

func createRenameBook(t *testing.T, root, dirName, audioName, sourceAudio, metadata string) {
	t.Helper()
	bookDir := filepath.Join(root, dirName)
	if err := os.MkdirAll(bookDir, 0o755); err != nil {
		t.Fatalf("create book dir: %v", err)
	}
	copyFile(t, sourceAudio, filepath.Join(bookDir, audioName))
	if err := os.WriteFile(filepath.Join(bookDir, "metadata.json"), []byte(metadata), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
}

func copyFile(t *testing.T, source, target string) {
	t.Helper()
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read fixture audio: %v", err)
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		t.Fatalf("write fixture audio: %v", err)
	}
}

func findRenameCandidate(
	t *testing.T,
	candidates []organizer.RenameCandidate,
	pathPart string,
) organizer.RenameCandidate {
	t.Helper()
	for _, candidate := range candidates {
		if strings.Contains(candidate.CurrentPath, pathPart) {
			return candidate
		}
	}
	t.Fatalf("candidate containing %q not found", pathPart)
	return organizer.RenameCandidate{}
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path to exist: %s\nstat error: %v", path, err)
	}
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected path to be missing: %s\nstat error: %v", path, err)
	}
}
