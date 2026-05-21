package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePathsReportsExistingAndInvalidDirectories(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	root := t.TempDir()
	existingDir := filepath.Join(root, "input")
	filePath := filepath.Join(root, "book.mp3")
	missingDir := filepath.Join(root, "missing")

	if err := os.Mkdir(existingDir, 0o755); err != nil {
		t.Fatalf("create input dir: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("audio"), 0o644); err != nil {
		t.Fatalf("write file path: %v", err)
	}

	resp, err := service.ValidatePaths(context.Background(), PathValidationRequest{
		Paths: []PathValidationItem{
			{ID: "blank", Path: "  ", Kind: "existing-directory"},
			{ID: "source", Path: existingDir, Kind: "existing-directory"},
			{ID: "file", Path: filePath, Kind: "existing-directory"},
			{ID: "missing", Path: missingDir, Kind: "existing-directory"},
		},
	})
	if err != nil {
		t.Fatalf("ValidatePaths() error = %v", err)
	}

	assertPathValidation(t, resp, "blank", false, "Path is required.")
	assertPathValidation(t, resp, "source", true, "")
	assertPathValidation(t, resp, "file", false, "Path is not a directory:")
	assertPathValidation(t, resp, "missing", false, "Directory does not exist:")
}

func TestValidatePathsChecksOutputParentWithoutCreatingOutput(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	root := t.TempDir()
	existingOutput := filepath.Join(root, "output")
	newOutput := filepath.Join(root, "new-output")
	missingParentOutput := filepath.Join(root, "missing-parent", "output")
	fileOutput := filepath.Join(root, "output-file")

	if err := os.Mkdir(existingOutput, 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}
	if err := os.WriteFile(fileOutput, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write output file: %v", err)
	}

	resp, err := service.ValidatePaths(context.Background(), PathValidationRequest{
		Paths: []PathValidationItem{
			{ID: "existing", Path: existingOutput, Kind: "output-directory"},
			{ID: "new", Path: newOutput, Kind: "output-directory"},
			{ID: "missing-parent", Path: missingParentOutput, Kind: "output-directory"},
			{ID: "file-output", Path: fileOutput, Kind: "output-directory"},
		},
	})
	if err != nil {
		t.Fatalf("ValidatePaths() error = %v", err)
	}

	assertPathValidation(t, resp, "existing", true, "")
	assertPathValidation(t, resp, "new", true, "")
	assertPathValidation(
		t,
		resp,
		"missing-parent",
		false,
		"Output parent directory does not exist:",
	)
	assertPathValidation(t, resp, "file-output", false, "Output path is not a directory:")
	assertFileNotExists(t, newOutput)
}

func TestValidatePathsHonorsCanceledContext(t *testing.T) {
	service := NewService(DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.ValidatePaths(ctx, PathValidationRequest{
		Paths: []PathValidationItem{{ID: "source", Path: t.TempDir()}},
	})
	if err == nil {
		t.Fatal("ValidatePaths() error = nil, want canceled context error")
	}
}

func assertPathValidation(
	t *testing.T,
	resp *PathValidationResponse,
	id string,
	wantValid bool,
	wantErrorPrefix string,
) {
	t.Helper()

	for _, result := range resp.Results {
		if result.ID != id {
			continue
		}
		if result.Valid != wantValid {
			t.Fatalf("%s valid = %v, want %v", id, result.Valid, wantValid)
		}
		if wantErrorPrefix == "" && result.Error != "" {
			t.Fatalf("%s error = %q, want empty", id, result.Error)
		}
		if wantErrorPrefix != "" && !strings.HasPrefix(result.Error, wantErrorPrefix) {
			t.Fatalf("%s error = %q, want prefix %q", id, result.Error, wantErrorPrefix)
		}
		return
	}

	t.Fatalf("missing validation result for %q", id)
}
