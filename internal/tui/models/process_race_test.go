package models

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

func TestReviewProcessConcurrentView(t *testing.T) {
	root := t.TempDir()
	input, output := filepath.Join(root, "input"), filepath.Join(root, "output")
	if err := os.MkdirAll(input, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("../../../testdata/test-scenarios/single-file/single_book.mp3")
	if err != nil {
		t.Fatal(err)
	}
	var moves []MovePreview
	for i := 0; i < 20; i++ {
		source := filepath.Join(input, fmt.Sprintf("%d.mp3", i))
		if err := os.WriteFile(source, fixture, 0o644); err != nil {
			t.Fatal(err)
		}
		moves = append(
			moves,
			MovePreview{
				SourcePath: source,
				TargetPath: filepath.Join(output, filepath.Base(source)),
			},
		)
	}
	model := NewProcessModel(
		nil,
		map[string]string{"Input Directory": input, "Output Directory": output},
		moves,
		organizer.DefaultFieldMapping(),
	)
	model.height, model.width = 40, 120
	work := model.Init()
	done := make(chan tea.Msg, 1)
	go func() { done <- work() }()
	for {
		select {
		case msg := <-done:
			model.Update(msg)
			if !model.complete || model.processing || model.failed != 0 ||
				model.success != len(moves) {
				t.Fatalf("incorrect completion state: %+v", model)
			}
			for _, move := range moves {
				if _, err := os.Stat(move.SourcePath); !os.IsNotExist(err) {
					t.Fatalf("source not moved: %s", move.SourcePath)
				}
			}
			return
		default:
			_ = model.View()
		}
	}
}
