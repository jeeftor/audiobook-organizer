//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReviewCLIRegressions(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "audiobook-organizer")
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	audio, err := os.ReadFile(
		filepath.Join(root, "testdata/test-scenarios/single-file/single_book.mp3"),
	)
	if err != nil {
		t.Fatal(err)
	}
	put := func(t *testing.T, path string, data []byte) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	read := func(t *testing.T, path string) []byte {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	exists := func(path string) bool { _, err := os.Stat(path); return err == nil }
	run := func(t *testing.T, stdin string, wantError bool, args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binary, args...)
		cmd.Dir = t.TempDir()
		cmd.Stdin = strings.NewReader(stdin)
		for _, v := range os.Environ() {
			if !strings.HasPrefix(v, "AO_") && !strings.HasPrefix(v, "AUDIOBOOK_ORGANIZER_") {
				cmd.Env = append(cmd.Env, v)
			}
		}
		output, err := cmd.CombinedOutput()
		if ctx.Err() != nil {
			t.Fatalf("command hung: %v\n%s", args, output)
		}
		if (err != nil) != wantError {
			t.Fatalf("command %v: %v\n%s", args, err, output)
		}
		return string(output)
	}
	book := func(t *testing.T) (string, string, []string) {
		t.Helper()
		base := t.TempDir()
		input, output := filepath.Join(base, "in"), filepath.Join(base, "out")
		put(t,
			filepath.Join(input, "book", "metadata.json"),
			[]byte(`{"title":"Book","authors":["Author"]}`),
		)
		put(t, filepath.Join(input, "book", "audio.mp3"), []byte("book audio"))
		return input, output, []string{"--input", input, "--output", output}
	}
	t.Run("occupied_destination_and_failed_move_log", func(t *testing.T) {
		for _, directory := range []bool{false, true} {
			input, output, args := book(t)
			target := filepath.Join(output, "Author", "Book", "audio.mp3")
			if directory {
				if err := os.MkdirAll(target, 0o755); err != nil {
					t.Fatal(err)
				}
			} else {
				put(t, target, []byte("existing"))
			}
			run(t, "", true, args...)
			if string(read(t, filepath.Join(input, "book", "audio.mp3"))) != "book audio" ||
				!exists(filepath.Join(input, "book", "metadata.json")) {
				t.Fatal("collision changed source")
			}
			if !directory && string(read(t, target)) != "existing" {
				t.Fatal("overwritten destination")
			}
			if exists(filepath.Join(output, ".abook-org.log")) {
				t.Fatal("failed operation logged")
			}
		}
	})
	t.Run("dry_undo_and_retry_failed_undo", func(t *testing.T) {
		input, output, args := book(t)
		run(t, "", false, args...)
		logPath := filepath.Join(output, ".abook-org.log")
		before := read(t, logPath)
		run(t, "", false, append(args, "--undo", "--dry-run")...)
		original := filepath.Join(input, "book", "audio.mp3")
		if exists(original) || !bytes.Equal(before, read(t, logPath)) {
			t.Fatal("dry undo changed state")
		}
		put(t, original, []byte("new file"))
		run(t, "", true, append(args, "--undo")...)
		if string(read(t, original)) != "new file" || !exists(logPath) {
			t.Fatal("failed undo lost data or log")
		}
		if err := os.Remove(original); err != nil {
			t.Fatal(err)
		}
		run(t, "", false, append(args, "--undo")...)
		if string(read(t, original)) != "book audio" || exists(logPath) {
			t.Fatal("undo retry did not finish")
		}
	})
	t.Run("organize_history_survives_separate_runs", func(t *testing.T) {
		input, output, args := book(t)
		run(t, "", false, args...)
		put(
			t,
			filepath.Join(input, "second", "metadata.json"),
			[]byte(`{"title":"Second","authors":["Author"]}`),
		)
		put(t, filepath.Join(input, "second", "audio.mp3"), []byte("second audio"))
		run(t, "", false, args...)
		run(t, "", false, append(args, "--undo")...)
		if string(read(t, filepath.Join(input, "book", "audio.mp3"))) != "book audio" ||
			string(read(t, filepath.Join(input, "second", "audio.mp3"))) != "second audio" ||
			exists(filepath.Join(output, ".abook-org.log")) {
			t.Fatal("undo did not restore both runs")
		}
	})
	t.Run("rename_history_and_partial_undo_retry", func(t *testing.T) {
		input, _, _ := book(t)
		dir := filepath.Join(input, "book")
		put(t, filepath.Join(dir, "audio.mp3"), audio)
		args := []string{"rename", "--dir", dir}
		run(t, "", false, append(args, "--template", "First - {title}")...)
		run(t, "", false, append(args, "--template", "Second - {title}")...)
		logPath := filepath.Join(dir, ".abook-rename.log")
		before := read(t, logPath)
		run(t, "", false, append(args, "--undo", "--dry-run")...)
		if !bytes.Equal(before, read(t, logPath)) {
			t.Fatal("dry rename undo changed log")
		}
		original := filepath.Join(dir, "audio.mp3")
		put(t, original, []byte("blocker"))
		run(t, "", true, append(args, "--undo")...)
		if string(read(t, original)) != "blocker" ||
			!bytes.Equal(read(t, filepath.Join(dir, "First - Book.mp3")), audio) {
			t.Fatal("partial undo lost content")
		}
		var pending []json.RawMessage
		if err := json.Unmarshal(read(t, logPath), &pending); err != nil || len(pending) != 1 {
			t.Fatalf("expected one pending rename: %s", read(t, logPath))
		}
		if err := os.Remove(original); err != nil {
			t.Fatal(err)
		}
		run(t, "", false, append(args, "--undo")...)
		if !bytes.Equal(read(t, original), audio) || exists(logPath) {
			t.Fatal("rename undo retry did not restore original")
		}
	})
	t.Run("invalid_recovery_log_blocks_new_moves", func(t *testing.T) {
		for _, rename := range []bool{false, true} {
			input, output, args := book(t)
			original := filepath.Join(input, "book", "audio.mp3")
			put(t, original, audio)
			logPath := filepath.Join(output, ".abook-org.log")
			if rename {
				args = []string{"rename", "--dir", filepath.Dir(original), "--template", "{title}"}
				logPath = filepath.Join(filepath.Dir(original), ".abook-rename.log")
			}
			put(t, logPath, []byte("invalid log"))
			run(t, "", true, args...)
			if !bytes.Equal(read(t, original), audio) || string(read(t, logPath)) != "invalid log" {
				t.Fatal("invalid recovery log allowed mutation")
			}
		}
	})
	t.Run("dry_run_new_output_and_empty_cleanup", func(t *testing.T) {
		input, output, args := book(t)
		empty := filepath.Join(input, "empty")
		if err := os.Mkdir(empty, 0o755); err != nil {
			t.Fatal(err)
		}
		run(t, "", false, append(args, "--dry-run", "--remove-empty")...)
		if exists(output) || !exists(empty) {
			t.Fatal("dry run changed directories")
		}
	})
	t.Run("declined_cleanup_terminates", func(t *testing.T) {
		input := t.TempDir()
		empty := filepath.Join(input, "empty")
		if err := os.Mkdir(empty, 0o755); err != nil {
			t.Fatal(err)
		}
		run(t, "n\n", false, "--input", input, "--remove-empty", "--prompt")
		if !exists(empty) {
			t.Fatal("declined directory removed")
		}
	})
	t.Run("metadata_cannot_escape_output", func(t *testing.T) {
		input, output, args := book(t)
		put(t,
			filepath.Join(input, "book", "metadata.json"),
			[]byte(`{"title":"Book","authors":["A/../../../escaped"]}`),
		)
		run(t, "", false, args...)
		var entries []struct {
			TargetPath string `json:"target_path"`
		}
		if err := json.Unmarshal(read(t, filepath.Join(output, ".abook-org.log")), &entries); err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatal("missing move")
		}
		resolved, err := filepath.EvalSymlinks(output)
		if err != nil {
			t.Fatal(err)
		}
		rel, err := filepath.Rel(resolved, entries[0].TargetPath)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Fatalf("escaped: %s", entries[0].TargetPath)
		}
	})
	t.Run("rename_preserves_noop_target", func(t *testing.T) {
		input, _, _ := book(t)
		dir := filepath.Join(input, "book")
		put(t, filepath.Join(dir, "audio.mp3"), audio)
		existing := append(append([]byte(nil), audio...), []byte("different contents")...)
		put(t, filepath.Join(dir, "Book.mp3"), existing)
		run(t, "", false, "rename", "--dir", dir, "--template", "{title}")
		if !bytes.Equal(read(t, filepath.Join(dir, "Book.mp3")), existing) ||
			!bytes.Equal(read(t, filepath.Join(dir, "Book (2).mp3")), audio) {
			t.Fatal("rename lost content")
		}
	})
	t.Run("strict_missing_year", func(t *testing.T) {
		input, _, _ := book(t)
		dir := filepath.Join(input, "book")
		put(t, filepath.Join(dir, "audio.mp3"), audio)
		run(t, "", true, "rename", "--dir", dir, "--template", "{year} - {title}", "--strict")
		if !exists(filepath.Join(dir, "audio.mp3")) ||
			exists(filepath.Join(dir, ".abook-rename.log")) {
			t.Fatal("strict validation mutated files")
		}
	})
	t.Run("sqlite_discovery_from_committed_abs_database", func(t *testing.T) {
		database := filepath.Join(
			root,
			"test",
			"abs",
			"baseline-config",
			"plain",
			"config",
			"absdatabase.sqlite",
		)
		before := read(t, database)
		output := run(
			t,
			"",
			false,
			"abs",
			"test-paths",
			"--abs-sqlite",
			database,
			"--dir",
			"/audiobooks",
		)
		if !strings.Contains(output, "Path discovery SUCCESS!") ||
			!bytes.Equal(before, read(t, database)) {
			t.Fatal("SQLite discovery failed or changed the committed database")
		}
	})
}
