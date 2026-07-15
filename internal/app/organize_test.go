package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestOrganizeUsesABSMetadataSourceForPreviewAndRun(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	root := t.TempDir()
	inputDir := filepath.Join(root, "input")
	outputDir := filepath.Join(root, "output")
	bookDir := filepath.Join(inputDir, "unsorted")
	if err := os.MkdirAll(bookDir, 0o755); err != nil {
		t.Fatalf("create input fixture: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("create output fixture: %v", err)
	}
	writeFile(t, filepath.Join(bookDir, "audio.m4b"), "fake audio")

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
			"results": []map[string]any{
				{
					"id":        "item-1",
					"libraryId": "lib-main",
					"path":      "/abs/input/unsorted",
					"media": map[string]any{
						"metadata": map[string]any{
							"title":   "ABS Test Book",
							"authors": []map[string]string{{"name": "ABS Author"}},
						},
					},
				},
			},
			"total": 1,
		})
	}))
	defer absServer.Close()

	config := organizeTestConfig(inputDir, outputDir, false)
	config.MetadataSource = metadataSourceABS
	config.ABS = ABSConfigDTO{
		URL:       absServer.URL,
		Token:     "test-token",
		LibraryID: "lib-main",
		PathMappings: []PathMappingDTO{{
			ABSPrefix:   "/abs/input",
			LocalPrefix: inputDir,
		}},
	}

	preview, err := service.PreviewOrganize(context.Background(), OrganizeRequest{Config: config})
	if err != nil {
		t.Fatalf("PreviewOrganize() error = %v", err)
	}
	if got := len(preview.Summary.Moves); got != 1 {
		t.Fatalf("preview moves = %d, want 1", got)
	}
	assertFileExists(t, filepath.Join(bookDir, "audio.m4b"))
	assertFileNotExists(t, filepath.Join(outputDir, "ABS Author", "ABS Test Book", "audio.m4b"))

	run, err := service.RunOrganize(context.Background(), OrganizeRequest{Config: config})
	if err != nil {
		t.Fatalf("RunOrganize() error = %v", err)
	}
	if got := len(run.Summary.Moves); got != 1 {
		t.Fatalf("run moves = %d, want 1", got)
	}
	assertFileExists(t, filepath.Join(outputDir, "ABS Author", "ABS Test Book", "audio.m4b"))
}

func TestPreviewOrganizeForcesDryRunAndOmitsLogPath(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	inputDir, outputDir := createOrganizeFixture(t)

	resp, err := service.PreviewOrganize(context.Background(), OrganizeRequest{
		Config: organizeTestConfig(inputDir, outputDir, false),
	})
	if err != nil {
		t.Fatalf("PreviewOrganize() error = %v", err)
	}

	if resp.LogPath != "" {
		t.Fatalf("PreviewOrganize() LogPath = %q, want empty", resp.LogPath)
	}
	if got := len(resp.Summary.MetadataFound); got != 1 {
		t.Fatalf("MetadataFound length = %d, want 1", got)
	}
	if got := len(resp.Summary.Moves); got != 1 {
		t.Fatalf("Moves length = %d, want 1", got)
	}
	assertFileExists(t, filepath.Join(inputDir, "test_book", "audio.mp3"))
	assertFileNotExists(t, filepath.Join(outputDir, "App Author", "App Test Book", "audio.mp3"))
}

func TestPreviewOrganizeReportsMetadataMissingWithoutVerbose(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	inputDir, outputDir, missingDir := createOrganizeFixtureWithMissingMetadata(t)

	resp, err := service.PreviewOrganize(context.Background(), OrganizeRequest{
		Config: organizeTestConfig(inputDir, outputDir, false),
	})
	if err != nil {
		t.Fatalf("PreviewOrganize() error = %v", err)
	}

	assertStringSliceContains(t, resp.Summary.MetadataMissing, mustResolvePath(t, inputDir))
	assertStringSliceContains(t, resp.Summary.MetadataMissing, mustResolvePath(t, missingDir))
}

func TestPreviewOrganizeUsesCustomLayoutTemplate(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	inputDir, outputDir := createOrganizeFixture(t)
	config := organizeTestConfig(inputDir, outputDir, false)
	config.LayoutTemplate = "${author}/${series}/${series-count} - ${title} (${narrator})"

	resp, err := service.PreviewOrganize(context.Background(), OrganizeRequest{
		Config: config,
	})
	if err != nil {
		t.Fatalf("PreviewOrganize() error = %v", err)
	}

	if got := len(resp.Summary.Moves); got != 1 {
		t.Fatalf("Moves length = %d, want 1", got)
	}
	wantTarget := filepath.Join(
		outputDir,
		"App Author",
		"App Series",
		"1 - App Test Book (App Narrator)",
	)
	if resp.Summary.Moves[0].To != wantTarget {
		t.Fatalf("move target = %q, want %q", resp.Summary.Moves[0].To, wantTarget)
	}
	assertFileNotExists(t, filepath.Join(wantTarget, "audio.mp3"))
}

func TestRunOrganizeForcesRunAndReturnsLogPath(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	inputDir, outputDir := createOrganizeFixture(t)

	resp, err := service.RunOrganize(context.Background(), OrganizeRequest{
		Config: organizeTestConfig(inputDir, outputDir, true),
	})
	if err != nil {
		t.Fatalf("RunOrganize() error = %v", err)
	}

	resolvedOutputDir, err := filepath.EvalSymlinks(outputDir)
	if err != nil {
		t.Fatalf("resolve output dir: %v", err)
	}
	wantLogPath := filepath.Join(resolvedOutputDir, ".abook-org.log")
	if resp.LogPath != wantLogPath {
		t.Fatalf("RunOrganize() LogPath = %q, want %q", resp.LogPath, wantLogPath)
	}
	if got := len(resp.Summary.MetadataFound); got != 1 {
		t.Fatalf("MetadataFound length = %d, want 1", got)
	}
	if got := len(resp.Summary.Moves); got != 1 {
		t.Fatalf("Moves length = %d, want 1", got)
	}
	assertFileExists(t, filepath.Join(outputDir, "App Author", "App Test Book", "audio.mp3"))
}

func TestRunOrganizeHonorsAllowedSourcePaths(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	inputDir, outputDir := createOrganizeFixture(t)
	selectedDir := filepath.Join(inputDir, "selected_book")

	if err := os.MkdirAll(selectedDir, 0o755); err != nil {
		t.Fatalf("create selected book dir: %v", err)
	}
	writeFile(
		t,
		filepath.Join(selectedDir, "metadata.json"),
		`{"title":"Selected Book","authors":["Selected Author"]}`,
	)
	writeFile(t, filepath.Join(selectedDir, "selected.mp3"), "fake audio")

	config := organizeTestConfig(inputDir, outputDir, true)
	config.AllowedSourcePaths = []string{selectedDir}
	resp, err := service.RunOrganize(context.Background(), OrganizeRequest{
		Config: config,
	})
	if err != nil {
		t.Fatalf("RunOrganize() error = %v", err)
	}

	if got := len(resp.Summary.Moves); got != 1 {
		t.Fatalf("Moves length = %d, want 1", got)
	}
	assertFileExists(
		t,
		filepath.Join(outputDir, "Selected Author", "Selected Book", "selected.mp3"),
	)
	assertFileExists(t, filepath.Join(inputDir, "test_book", "audio.mp3"))
	assertFileNotExists(t, filepath.Join(outputDir, "App Author", "App Test Book", "audio.mp3"))
}

func organizeTestConfig(inputDir, outputDir string, dryRun bool) OrganizerConfigDTO {
	return OrganizerConfigDTO{
		BaseDir:   inputDir,
		OutputDir: outputDir,
		DryRun:    dryRun,
		Layout:    "author-title",
		FieldMapping: FieldMappingDTO{
			TitleField:   "title",
			SeriesField:  "series",
			AuthorFields: []string{"authors"},
		},
	}
}

func createOrganizeFixture(t *testing.T) (string, string) {
	t.Helper()

	root := t.TempDir()
	inputDir := filepath.Join(root, "input")
	outputDir := filepath.Join(root, "output")
	bookDir := filepath.Join(inputDir, "test_book")

	if err := os.MkdirAll(bookDir, 0o755); err != nil {
		t.Fatalf("create book dir: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}
	writeFile(
		t,
		filepath.Join(bookDir, "metadata.json"),
		`{"title":"App Test Book","authors":["App Author"],"series":["App Series #1"],"narrator":"App Narrator"}`,
	)
	writeFile(t, filepath.Join(bookDir, "audio.mp3"), "fake audio")

	return inputDir, outputDir
}

func createOrganizeFixtureWithMissingMetadata(t *testing.T) (string, string, string) {
	t.Helper()

	inputDir, outputDir := createOrganizeFixture(t)
	missingDir := filepath.Join(inputDir, "missing_metadata")
	if err := os.MkdirAll(missingDir, 0o755); err != nil {
		t.Fatalf("create missing metadata dir: %v", err)
	}
	writeFile(t, filepath.Join(missingDir, "orphan.mp3"), "fake audio")

	return inputDir, outputDir, missingDir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %s\nstat error: %v", path, err)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected file not to exist: %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func assertStringSliceContains(t *testing.T, values []string, want string) {
	t.Helper()
	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("expected %q in %v", want, values)
}

func mustResolvePath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("resolve %s: %v", path, err)
	}
	return resolved
}
