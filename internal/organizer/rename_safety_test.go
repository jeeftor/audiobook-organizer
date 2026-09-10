package organizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type safetyMetadataResolver struct{}

func (safetyMetadataResolver) MetadataForPath(path string) (Metadata, error) {
	return NewMetadataProvider(filepath.Dir(path), false).GetMetadata()
}

func safetyRenamer(t *testing.T, root, template string) *Renamer {
	t.Helper()
	r, err := NewRenamer(
		&RenamerConfig{
			BaseDir:          root,
			Template:         template,
			PreservePath:     true,
			Recursive:        true,
			MetadataResolver: safetyMetadataResolver{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRenameSafetyReservesExistingAndGeneratedTargets(t *testing.T) {
	root := t.TempDir()
	safetyWrite(t, filepath.Join(root, "metadata.json"), `{"title":"Book","authors":["Author"]}`)
	for name, content := range map[string]string{"Book.mp3": "keep", "Book (2).mp3": "also keep", "a.mp3": "first", "b.mp3": "second"} {
		safetyWrite(t, filepath.Join(root, name), content)
	}
	r := safetyRenamer(t, root, "{title}")
	preview, err := r.ScanFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range preview {
		if c.CurrentPath != c.ProposedPath && c.ProposedPath == filepath.Join(root, "Book.mp3") {
			t.Fatal("reserved no-op destination reused")
		}
	}
	if err := r.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := safetyRead(t, filepath.Join(root, "Book.mp3")); got != "keep" {
		t.Fatalf("overwritten: %q", got)
	}
	seen := map[string]bool{}
	for _, c := range preview {
		seen[safetyRead(t, c.ProposedPath)] = true
	}
	for _, content := range []string{"keep", "also keep", "first", "second"} {
		if !seen[content] {
			t.Errorf("lost %q", content)
		}
	}
}

func TestRenameSafetyLateCollisionDoesNotOverwrite(t *testing.T) {
	root := t.TempDir()
	a, b := filepath.Join(root, "a.mp3"), filepath.Join(root, "b.mp3")
	safetyWrite(t, a, "source")
	safetyWrite(t, b, "destination")
	r := safetyRenamer(t, root, "{title}")
	if err := r.RenameFile(a, b); err == nil {
		t.Fatal("expected collision error")
	}
	if safetyRead(t, a) != "source" || safetyRead(t, b) != "destination" {
		t.Fatal("collision changed files")
	}
	if len(r.logEntries) != 0 {
		t.Fatal("failed rename logged")
	}
}

func TestRenameSafetySelectionPreservesPreview(t *testing.T) {
	root := t.TempDir()
	safetyWrite(t, filepath.Join(root, "metadata.json"), `{"title":"Book"}`)
	safetyWrite(t, filepath.Join(root, "a.mp3"), "a")
	safetyWrite(t, filepath.Join(root, "b.mp3"), "b")
	r := safetyRenamer(t, root, "{title}")
	preview, err := r.ScanFiles()
	if err != nil {
		t.Fatal(err)
	}
	selected := preview[1]
	r.config.AllowedCurrentPaths = []string{selected.CurrentPath}
	filtered, err := r.ScanFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].ProposedPath != selected.ProposedPath {
		t.Fatalf("selection changed plan: %+v vs %+v", filtered, selected)
	}
	if err := r.Execute(); err != nil {
		t.Fatal(err)
	}
	if safetyRead(t, selected.ProposedPath) != "b" || safetyRead(t, preview[0].CurrentPath) != "a" {
		t.Fatal("selection did not match preview")
	}
}

func TestRenameSafetySeparateDirectoriesDoNotConflict(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"a", "b"} {
		safetyWrite(t, filepath.Join(root, dir, "metadata.json"), `{"title":"Book"}`)
		safetyWrite(t, filepath.Join(root, dir, "original.mp3"), dir)
	}
	r := safetyRenamer(t, root, "{title}")
	preview, err := r.ScanFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range preview {
		if c.IsConflict || filepath.Base(c.ProposedPath) != "Book.mp3" {
			t.Fatalf("false conflict: %+v", c)
		}
	}
	if err := r.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"a", "b"} {
		if safetyRead(t, filepath.Join(root, dir, "Book.mp3")) != dir {
			t.Fatal("wrong rename")
		}
	}
}

func TestRenameSafetyStrictMissingFieldAbortsBeforeMutation(t *testing.T) {
	root := t.TempDir()
	safetyWrite(t, filepath.Join(root, "metadata.json"), `{"title":"Book"}`)
	safetyWrite(t, filepath.Join(root, "original.mp3"), "original")
	r := safetyRenamer(t, root, "{year} - {title}")
	r.config.StrictMode = true
	preview, err := r.ScanFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(preview) != 1 || !strings.Contains(preview[0].Error, "year") {
		t.Fatalf("missing strict preview error: %+v", preview)
	}
	if err := r.Execute(); err == nil {
		t.Fatal("strict execution succeeded with missing year")
	}
	if safetyRead(t, filepath.Join(root, "original.mp3")) != "original" {
		t.Fatal("strict failure mutated source")
	}
	if _, err := os.Stat(r.GetLogPath()); !os.IsNotExist(err) {
		t.Fatal("strict failure created log")
	}
	for _, template := range []string{"{year|Unknown} - {title}", "{year - }{title}"} {
		r = safetyRenamer(t, root, template)
		r.config.StrictMode = true
		if _, err := r.GenerateNewPath(filepath.Join(root, "original.mp3"), Metadata{Title: "Book"}); err != nil {
			t.Fatalf("optional/fallback field rejected: %v", err)
		}
	}
}

func TestRenameSafetyStrictValidFileWaitsForWholePlanValidation(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"a-valid", "b-missing"} {
		metadata := `{"title":"Book","year":2020}`
		if directory == "b-missing" {
			metadata = `{"title":"Book"}`
		}
		safetyWrite(t, filepath.Join(root, directory, "metadata.json"), metadata)
		safetyWrite(t, filepath.Join(root, directory, "original.mp3"), directory)
	}
	r := safetyRenamer(t, root, "{year} - {title}")
	r.config.StrictMode = true
	if err := r.Execute(); err == nil {
		t.Fatal("strict run ignored an invalid candidate")
	}
	for _, directory := range []string{"a-valid", "b-missing"} {
		if safetyRead(t, filepath.Join(root, directory, "original.mp3")) != directory {
			t.Fatal("strict failure changed a source")
		}
	}
	if r.GetSummary().FilesRenamed != 0 {
		t.Fatal("strict validation allowed a partial run")
	}
}
