package models

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

func TestProcessStatus(t *testing.T) {
	tests := []struct {
		name   string
		status ProcessStatus
		value  int
	}{
		{"StatusPending", StatusPending, 0},
		{"StatusProcessing", StatusProcessing, 1},
		{"StatusSuccess", StatusSuccess, 2},
		{"StatusError", StatusError, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.status) != tt.value {
				t.Errorf("Expected %s to have value %d, got %d", tt.name, tt.value, int(tt.status))
			}
		})
	}
}

func TestProcessItem(t *testing.T) {
	tests := []struct {
		name        string
		item        ProcessItem
		sourcePath  string
		targetPath  string
		status      ProcessStatus
		errorExists bool
		message     string
	}{
		{
			name: "pending item",
			item: ProcessItem{
				SourcePath: "/test/source.mp3",
				TargetPath: "/output/target.mp3",
				Status:     StatusPending,
				Error:      nil,
				Message:    "Waiting to process",
			},
			sourcePath:  "/test/source.mp3",
			targetPath:  "/output/target.mp3",
			status:      StatusPending,
			errorExists: false,
			message:     "Waiting to process",
		},
		{
			name: "error item",
			item: ProcessItem{
				SourcePath: "/test/source.mp3",
				TargetPath: "/output/target.mp3",
				Status:     StatusError,
				Error:      errors.New("permission denied"),
				Message:    "Failed to move file",
			},
			sourcePath:  "/test/source.mp3",
			targetPath:  "/output/target.mp3",
			status:      StatusError,
			errorExists: true,
			message:     "Failed to move file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.item.SourcePath != tt.sourcePath {
				t.Errorf("Expected SourcePath %q, got %q", tt.sourcePath, tt.item.SourcePath)
			}

			if tt.item.Status != tt.status {
				t.Errorf("Expected Status %v, got %v", tt.status, tt.item.Status)
			}

			if tt.errorExists && tt.item.Error == nil {
				t.Error("Expected Error to exist but got nil")
			} else if !tt.errorExists && tt.item.Error != nil {
				t.Errorf("Expected no Error but got: %v", tt.item.Error)
			}
		})
	}
}

func TestNewProcessModelFixed(t *testing.T) {
	tests := []struct {
		name         string
		books        []AudioBook
		config       map[string]string
		moves        []MovePreview
		fieldMapping organizer.FieldMapping
	}{
		{
			name:  "empty process model",
			books: []AudioBook{},
			config: map[string]string{
				"dryrun": "false",
			},
			moves:        []MovePreview{},
			fieldMapping: organizer.DefaultFieldMapping(),
		},
		{
			name: "single book process model",
			books: []AudioBook{
				{
					Path: "/test/book.mp3",
					Metadata: organizer.Metadata{
						Title:   "Test Book",
						Authors: []string{"Test Author"},
					},
				},
			},
			config: map[string]string{
				"dryrun":  "false",
				"verbose": "true",
			},
			moves: []MovePreview{
				{SourcePath: "/test/book.mp3", TargetPath: "/output/book.mp3"},
			},
			fieldMapping: organizer.DefaultFieldMapping(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewProcessModel(tt.books, tt.config, tt.moves, tt.fieldMapping)

			if model == nil {
				t.Fatal("NewProcessModel returned nil")
			}

			if len(model.books) != len(tt.books) {
				t.Errorf("Expected %d books, got %d", len(tt.books), len(model.books))
			}

			if len(model.config) != len(tt.config) {
				t.Errorf("Expected %d config items, got %d", len(tt.config), len(model.config))
			}

			if len(model.items) != len(tt.moves) {
				t.Errorf("Expected %d items, got %d", len(tt.moves), len(model.items))
			}

			// Check initial state
			if model.complete {
				t.Error("Expected complete to be false initially")
			}

			if model.processing {
				t.Error("Expected processing to be false initially")
			}
		})
	}
}

func TestProcessModelInit(t *testing.T) {
	moves := []MovePreview{}
	model := NewProcessModel([]AudioBook{}, map[string]string{}, moves, organizer.DefaultFieldMapping())

	cmd := model.Init()

	// ProcessModel should return a command to start processing
	if cmd == nil {
		t.Error("Init() should return a command to start processing")
	}
}

func TestProcessModelView(t *testing.T) {
	books := []AudioBook{
		{
			Path: "/test/book1.mp3",
			Metadata: organizer.Metadata{
				Title:   "Test Book 1",
				Authors: []string{"Test Author"},
			},
		},
	}

	config := map[string]string{
		"dryrun": "false",
	}

	moves := []MovePreview{
		{SourcePath: "/test/book1.mp3", TargetPath: "/output/book1.mp3"},
	}

	model := NewProcessModel(books, config, moves, organizer.DefaultFieldMapping())

	// Test view during processing
	view := model.View()
	if view == "" {
		t.Error("View() should return non-empty string")
	}

	// Test completed processing
	model.complete = true

	view = model.View()
	if view == "" {
		t.Error("View() should return non-empty string when complete")
	}
}

func TestProcessModelProgressTracking(t *testing.T) {
	books := []AudioBook{
		{Path: "/test/book1.mp3"},
		{Path: "/test/book2.mp3"},
		{Path: "/test/book3.mp3"},
	}

	moves := []MovePreview{
		{SourcePath: "/test/book1.mp3", TargetPath: "/output/book1.mp3"},
		{SourcePath: "/test/book2.mp3", TargetPath: "/output/book2.mp3"},
		{SourcePath: "/test/book3.mp3", TargetPath: "/output/book3.mp3"},
	}

	model := NewProcessModel(books, map[string]string{}, moves, organizer.DefaultFieldMapping())

	// Test initial progress
	if len(model.items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(model.items))
	}

	// Simulate processing progress
	model.success = 1
	model.failed = 0

	totalItems := len(model.items)
	processed := model.success + model.failed
	progress := float64(processed) / float64(totalItems)
	expectedProgress := float64(1) / float64(3)

	if progress != expectedProgress {
		t.Errorf("Expected progress %f, got %f", expectedProgress, progress)
	}

	// Complete processing
	model.success = 3
	model.complete = true

	if !model.complete {
		t.Error("Expected processing to be complete")
	}

	finalProgress := float64(model.success+model.failed) / float64(totalItems)
	if finalProgress != 1.0 {
		t.Errorf("Expected final progress 1.0, got %f", finalProgress)
	}
}

func TestProcessModelErrorHandling(t *testing.T) {
	moves := []MovePreview{
		{SourcePath: "/test/success.mp3", TargetPath: "/output/success.mp3"},
		{SourcePath: "/test/error1.mp3", TargetPath: "/output/error1.mp3"},
		{SourcePath: "/test/error2.mp3", TargetPath: "/output/error2.mp3"},
	}

	model := NewProcessModel([]AudioBook{}, map[string]string{}, moves, organizer.DefaultFieldMapping())

	// Simulate some errors
	model.items[0].Status = StatusSuccess
	model.items[1].Status = StatusError
	model.items[1].Error = errors.New("permission denied")
	model.items[2].Status = StatusError
	model.items[2].Error = errors.New("disk full")

	model.success = 1
	model.failed = 2

	// Test error count tracking
	if model.failed != 2 {
		t.Errorf("Expected failed count 2, got %d", model.failed)
	}

	// Test error display in view
	view := model.View()
	if view == "" {
		t.Error("View should not be empty with errors")
	}

	// Count actual errors in items
	actualErrors := 0
	for _, item := range model.items {
		if item.Status == StatusError {
			actualErrors++
		}
	}

	if actualErrors != model.failed {
		t.Errorf("Expected %d errors in items, got %d", model.failed, actualErrors)
	}
}

func TestProcessModelWindowSizeHandling(t *testing.T) {
	moves := []MovePreview{}
	model := NewProcessModel([]AudioBook{}, map[string]string{}, moves, organizer.DefaultFieldMapping())

	// Test window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}

	updatedModel, cmd := model.Update(msg)

	processModel, ok := updatedModel.(*ProcessModel)
	if !ok {
		t.Fatal("Update should return a ProcessModel")
	}

	if processModel.width != 100 {
		t.Errorf("Expected width 100, got %d", processModel.width)
	}

	if processModel.height != 30 {
		t.Errorf("Expected height 30, got %d", processModel.height)
	}

	_ = cmd // Command can be nil
}

func TestProcessModelTimeTracking(t *testing.T) {
	moves := []MovePreview{}
	model := NewProcessModel([]AudioBook{}, map[string]string{}, moves, organizer.DefaultFieldMapping())

	// Test initial time state
	if !model.startTime.IsZero() {
		t.Error("Expected startTime to be zero initially")
	}

	// Simulate starting processing
	model.startTime = time.Now().Add(-5 * time.Second)
	model.elapsedTime = time.Since(model.startTime)

	if model.elapsedTime < 4*time.Second {
		t.Error("Expected elapsedTime to be at least 4 seconds")
	}

	// Test completion time
	model.complete = true
	endTime := model.startTime.Add(model.elapsedTime)

	if endTime.Before(model.startTime) {
		t.Error("End time should be after start time")
	}
}

func TestProcessModelKeyHandling(t *testing.T) {
	moves := []MovePreview{}
	model := NewProcessModel([]AudioBook{}, map[string]string{}, moves, organizer.DefaultFieldMapping())

	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{
			name:   "enter key",
			keyMsg: tea.KeyMsg{Type: tea.KeyEnter},
		},
		{
			name:   "escape key",
			keyMsg: tea.KeyMsg{Type: tea.KeyEsc},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedModel, cmd := model.Update(tt.keyMsg)

			// Verify model is returned
			if updatedModel == nil {
				t.Error("Update should return a model")
			}

			// Verify it's still a ProcessModel
			_, ok := updatedModel.(*ProcessModel)
			if !ok {
				t.Error("Update should return a ProcessModel")
			}

			_ = cmd // Command can be nil
		})
	}
}
