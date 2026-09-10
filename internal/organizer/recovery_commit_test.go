package organizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecoveryChainedRenameUndoPreservesIntermediateOccupant(t *testing.T) {
	root := t.TempDir()
	original, first, second := filepath.Join(
		root,
		"audio.mp3",
	), filepath.Join(
		root,
		"First.mp3",
	), filepath.Join(
		root,
		"Second.mp3",
	)
	safetyWrite(t, original, "original audio")
	r := safetyRenamer(t, root, "{title}")
	for _, pair := range [][2]string{{original, first}, {first, second}} {
		if err := r.RenameFile(pair[0], pair[1]); err != nil {
			t.Fatal(err)
		}
		if err := r.SaveLog(); err != nil {
			t.Fatal(err)
		}
	}
	safetyWrite(t, first, "unrelated occupant")
	if err := r.UndoRenames(); err == nil {
		t.Fatal("expected blocked undo")
	}
	if safetyRead(t, first) != "unrelated occupant" || safetyRead(t, second) != "original audio" {
		t.Fatal("blocked undo moved unrelated contents")
	}
	if _, err := os.Stat(original); !os.IsNotExist(err) {
		t.Fatal("older dependency executed after failed restore")
	}
	if err := os.Remove(first); err != nil {
		t.Fatal(err)
	}
	if err := r.UndoRenames(); err != nil {
		t.Fatal(err)
	}
	if safetyRead(t, original) != "original audio" {
		t.Fatal("retry lost original")
	}
}

func TestRecoveryChainedOrganizeUndoPreservesIntermediateOccupant(t *testing.T) {
	o, source, output := safetyOrganizer(t, false)
	original := filepath.Join(source, "book")
	first, second := filepath.Join(output, "First"), filepath.Join(output, "Second")
	for _, pair := range [][2]string{{original, first}, {first, second}} {
		if err := o.executeMove(pair[0], pair[1], nil); err != nil {
			t.Fatal(err)
		}
	}
	blocker := filepath.Join(first, "audio.mp3")
	safetyWrite(t, blocker, "unrelated occupant")
	if err := o.undoMoves(); err == nil {
		t.Fatal("expected blocked undo")
	}
	if safetyRead(t, blocker) != "unrelated occupant" ||
		safetyRead(t, filepath.Join(second, "audio.mp3")) != "original audio" {
		t.Fatal("blocked undo moved unrelated contents")
	}
	if _, err := os.Stat(filepath.Join(original, "audio.mp3")); !os.IsNotExist(err) {
		t.Fatal("older dependency executed after failed restore")
	}
	if err := os.Remove(blocker); err != nil {
		t.Fatal(err)
	}
	if err := o.undoMoves(); err != nil {
		t.Fatal(err)
	}
	if safetyRead(t, filepath.Join(original, "audio.mp3")) != "original audio" {
		t.Fatal("retry lost original")
	}
}

func TestRecoveryLogFailureRollsBackOrganize(t *testing.T) {
	for _, single := range []bool{false, true} {
		t.Run(map[bool]string{false: "directory", true: "single"}[single], func(t *testing.T) {
			o, source, output := safetyOrganizer(t, false)
			// Simulate a late log publication failure after constructor validation.
			if err := os.Mkdir(o.GetLogPath(), 0o755); err != nil {
				t.Fatal(err)
			}
			original := filepath.Join(source, "book", "audio.mp3")
			target := filepath.Join(output, "Author", "Book", "audio.mp3")
			var err error
			if single {
				err = o.executeSingleFileMove(original, target, Metadata{})
			} else {
				err = o.Execute()
			}
			if err == nil {
				t.Fatal("expected log error")
			}
			if safetyRead(t, original) != "original audio" {
				t.Fatal("log failure stranded original")
			}
			if _, err := os.Stat(target); !os.IsNotExist(err) {
				t.Fatal("rolled-back destination remains")
			}
			if len(o.logEntries) != 0 || len(o.GetSummary().Moves) != 0 {
				t.Fatal("rolled-back operation still recorded")
			}
		})
	}
}

func TestRecoveryLogFailureRollsBackRename(t *testing.T) {
	root := t.TempDir()
	original, target := filepath.Join(root, "audio.mp3"), filepath.Join(root, "Book.mp3")
	safetyWrite(t, original, "original audio")
	r := safetyRenamer(t, root, "{title}")
	if err := os.Mkdir(r.GetLogPath(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := r.RenameFile(original, target); err == nil {
		t.Fatal("expected log error")
	}
	if safetyRead(t, original) != "original audio" {
		t.Fatal("log failure stranded original")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("rolled-back destination remains")
	}
	if len(r.logEntries) != 0 {
		t.Fatal("rolled-back rename still recorded")
	}
}

func TestRecoveryRollbackFailureReportsRemainingMove(t *testing.T) {
	o, source, output := safetyOrganizer(t, false)
	if err := os.Mkdir(o.GetLogPath(), 0o755); err != nil {
		t.Fatal(err)
	}
	from := filepath.Join(source, "book", "audio.mp3")
	to := filepath.Join(output, "audio.mp3")
	if err := moveNoReplace(from, to); err != nil {
		t.Fatal(err)
	}
	safetyWrite(t, from, "new occupant")
	err := o.updateLogAndCleanup(
		filepath.Dir(from),
		output,
		[]FilePair{{From: "audio.mp3", To: "audio.mp3"}},
	)
	if err == nil || !strings.Contains(err.Error(), "rollback") ||
		!strings.Contains(err.Error(), to) {
		t.Fatalf("missing recovery guidance: %v", err)
	}
	if safetyRead(t, from) != "new occupant" || safetyRead(t, to) != "original audio" ||
		len(o.logEntries) != 1 || len(o.logEntries[0].Files) != 1 {
		t.Fatal("failed rollback lost contents or pending recovery state")
	}
}
