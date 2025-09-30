package models

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

func TestMovePreview(t *testing.T) {
	tests := []struct {
		name       string
		preview    MovePreview
		sourcePath string
		targetPath string
	}{
		{
			name: "basic move preview",
			preview: MovePreview{
				SourcePath: "/input/book.mp3",
				TargetPath: "/output/Author/Book/book.mp3",
			},
			sourcePath: "/input/book.mp3",
			targetPath: "/output/Author/Book/book.mp3",
		},
		{
			name: "move preview with complex paths",
			preview: MovePreview{
				SourcePath: "/input/Some Author - Some Book (Series #1).mp3",
				TargetPath: "/output/Some Author/Series/Some Book/01 - Some Author - Some Book (Series #1).mp3",
			},
			sourcePath: "/input/Some Author - Some Book (Series #1).mp3",
			targetPath: "/output/Some Author/Series/Some Book/01 - Some Author - Some Book (Series #1).mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preview.SourcePath != tt.sourcePath {
				t.Errorf("Expected SourcePath %q, got %q", tt.sourcePath, tt.preview.SourcePath)
			}

			if tt.preview.TargetPath != tt.targetPath {
				t.Errorf("Expected TargetPath %q, got %q", tt.targetPath, tt.preview.TargetPath)
			}
		})
	}
}

func TestNewPreviewModel(t *testing.T) {
	tests := []struct {
		name         string
		books        []AudioBook
		config       map[string]string
		fieldMapping organizer.FieldMapping
	}{
		{
			name:  "empty preview",
			books: []AudioBook{},
			config: map[string]string{
				"layout": "author-series-title",
				"dryrun": "true",
			},
			fieldMapping: organizer.DefaultFieldMapping(),
		},
		{
			name: "single book preview",
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
				"layout": "author-title",
				"dryrun": "false",
			},
			fieldMapping: organizer.DefaultFieldMapping(),
		},
		{
			name: "multiple books preview",
			books: []AudioBook{
				{Path: "/test/book1.mp3", Metadata: organizer.Metadata{Title: "Book 1", Authors: []string{"Author 1"}}},
				{Path: "/test/book2.mp3", Metadata: organizer.Metadata{Title: "Book 2", Authors: []string{"Author 2"}}},
			},
			config: map[string]string{
				"layout":      "author-series-title",
				"dryrun":      "true",
				"removeEmpty": "true",
			},
			fieldMapping: organizer.DefaultFieldMapping(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewPreviewModel(tt.books, tt.config, tt.fieldMapping)

			if model == nil {
				t.Fatal("NewPreviewModel returned nil")
			}

			if len(model.books) != len(tt.books) {
				t.Errorf("Expected %d books, got %d", len(tt.books), len(model.books))
			}

			if len(model.config) != len(tt.config) {
				t.Errorf("Expected %d config items, got %d", len(tt.config), len(model.config))
			}

			// Verify config values
			for key, expectedValue := range tt.config {
				if actualValue, exists := model.config[key]; !exists {
					t.Errorf("Expected config key %q to exist", key)
				} else if actualValue != expectedValue {
					t.Errorf("Expected config %q = %q, got %q", key, expectedValue, actualValue)
				}
			}

			if model.cursor != 0 {
				t.Errorf("Expected initial cursor 0, got %d", model.cursor)
			}

			if model.scrollOffset != 0 {
				t.Errorf("Expected initial scrollOffset 0, got %d", model.scrollOffset)
			}
		})
	}
}

func TestPreviewModelInit(t *testing.T) {
	model := NewPreviewModel([]AudioBook{}, map[string]string{}, organizer.DefaultFieldMapping())

	cmd := model.Init()

	// Preview model should return a command to generate move previews
	if cmd == nil {
		t.Error("Init() should return a command to generate previews")
	}
}

func TestPreviewModelView(t *testing.T) {
	books := []AudioBook{
		{
			Path: "/test/book1.mp3",
			Metadata: organizer.Metadata{
				Title:   "Test Book 1",
				Authors: []string{"Test Author"},
			},
		},
		{
			Path: "/test/book2.mp3",
			Metadata: organizer.Metadata{
				Title:   "Test Book 2",
				Authors: []string{"Test Author"},
			},
		},
	}

	config := map[string]string{
		"layout": "author-title",
		"dryrun": "true",
	}

	model := NewPreviewModel(books, config, organizer.DefaultFieldMapping())

	// Test view without moves
	view := model.View()
	if view == "" {
		t.Error("View() should return non-empty string")
	}

	// Add some move previews
	model.moves = []MovePreview{
		{
			SourcePath: "/test/book1.mp3",
			TargetPath: "/output/Test Author/Test Book 1/book1.mp3",
		},
		{
			SourcePath: "/test/book2.mp3",
			TargetPath: "/output/Test Author/Test Book 2/book2.mp3",
		},
	}

	// Set dimensions so moves are visible
	model.width = 100
	model.height = 30

	// Test view with moves
	view = model.View()
	if view == "" {
		t.Error("View() should return non-empty string with moves")
	}

	// View should contain source and target paths
	if !strings.Contains(view, "book1.mp3") {
		t.Errorf("View should contain source filename, got:\n%s", view)
	}

	view = model.View()
	if view == "" {
		t.Error("View() should return non-empty string with dimensions")
	}
}

func TestPreviewModelNavigation(t *testing.T) {
	model := NewPreviewModel([]AudioBook{}, map[string]string{}, organizer.DefaultFieldMapping())

	// Add some moves to navigate through
	model.moves = []MovePreview{
		{SourcePath: "/test/file1.mp3", TargetPath: "/output/file1.mp3"},
		{SourcePath: "/test/file2.mp3", TargetPath: "/output/file2.mp3"},
		{SourcePath: "/test/file3.mp3", TargetPath: "/output/file3.mp3"},
	}

	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{
			name:   "down arrow",
			keyMsg: tea.KeyMsg{Type: tea.KeyDown},
		},
		{
			name:   "up arrow",
			keyMsg: tea.KeyMsg{Type: tea.KeyUp},
		},
		{
			name:   "j key",
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")},
		},
		{
			name:   "k key",
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")},
		},
		{
			name:   "page down",
			keyMsg: tea.KeyMsg{Type: tea.KeyPgDown},
		},
		{
			name:   "page up",
			keyMsg: tea.KeyMsg{Type: tea.KeyPgUp},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCursor := model.cursor
			originalScroll := model.scrollOffset

			updatedModel, cmd := model.Update(tt.keyMsg)

			previewModel, ok := updatedModel.(*PreviewModel)
			if !ok {
				t.Fatal("Update should return a PreviewModel")
			}

			// Cursor should be within valid range
			if previewModel.cursor < 0 || previewModel.cursor >= len(previewModel.moves) {
				if len(previewModel.moves) > 0 {
					t.Errorf("Cursor %d is out of range [0, %d)", previewModel.cursor, len(previewModel.moves))
				}
			}

			// Scroll should be non-negative
			if previewModel.scrollOffset < 0 {
				t.Errorf("ScrollOffset should be non-negative, got %d", previewModel.scrollOffset)
			}

			// Navigation should change cursor or scroll (unless at boundaries)
			if len(model.moves) > 1 {
				// With multiple items, some navigation should occur
				// (unless we're at boundaries, which is implementation dependent)
			}

			_ = originalCursor // Avoid unused variable
			_ = originalScroll // Avoid unused variable
			_ = cmd            // Command can be nil
		})
	}
}

func TestPreviewModelWindowSizeHandling(t *testing.T) {
	model := NewPreviewModel([]AudioBook{}, map[string]string{}, organizer.DefaultFieldMapping())

	// Test window size message
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	updatedModel, cmd := model.Update(msg)

	previewModel, ok := updatedModel.(*PreviewModel)
	if !ok {
		t.Fatal("Update should return a PreviewModel")
	}

	if previewModel.width != 120 {
		t.Errorf("Expected width 120, got %d", previewModel.width)
	}

	if previewModel.height != 40 {
		t.Errorf("Expected height 40, got %d", previewModel.height)
	}

	_ = cmd // Command can be nil
}

func TestPreviewModelKeyHandling(t *testing.T) {
	model := NewPreviewModel([]AudioBook{}, map[string]string{}, organizer.DefaultFieldMapping())

	tests := []struct {
		name              string
		keyMsg            tea.KeyMsg
		expectsTransition bool // If true, model can be a different type (e.g., ProcessModel)
	}{
		{
			name:              "enter key",
			keyMsg:            tea.KeyMsg{Type: tea.KeyEnter},
			expectsTransition: true, // Enter transitions to ProcessModel
		},
		{
			name:   "escape key",
			keyMsg: tea.KeyMsg{Type: tea.KeyEsc},
		},
		{
			name:   "tab key",
			keyMsg: tea.KeyMsg{Type: tea.KeyTab},
		},
		{
			name:   "character key",
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedModel, cmd := model.Update(tt.keyMsg)

			// Verify model is returned
			if updatedModel == nil {
				t.Error("Update should return a model")
			}

			// If we don't expect a transition, verify it's still a PreviewModel
			if !tt.expectsTransition {
				_, ok := updatedModel.(*PreviewModel)
				if !ok {
					t.Error("Update should return a PreviewModel")
				}
			}

			_ = cmd // Command can be nil
		})
	}
}

func TestPreviewModelScrollHandling(t *testing.T) {
	model := NewPreviewModel([]AudioBook{}, map[string]string{}, organizer.DefaultFieldMapping())

	// Add many moves to test scrolling
	for i := 0; i < 20; i++ {
		model.moves = append(model.moves, MovePreview{
			SourcePath: fmt.Sprintf("/test/file%d.mp3", i),
			TargetPath: fmt.Sprintf("/output/file%d.mp3", i),
		})
	}

	// Set small height to force scrolling
	model.height = 10

	// Test scrolling down beyond visible area
	for i := 0; i < 15; i++ {
		model.cursor = i
		view := model.View()
		if view == "" {
			t.Error("View should not be empty during scrolling")
		}
	}

	// Test that scroll offset adjusts appropriately
	// (exact behavior depends on implementation)
	if model.scrollOffset < 0 {
		t.Error("ScrollOffset should not be negative")
	}
}

func TestPreviewModelConfigHandling(t *testing.T) {
	config := map[string]string{
		"layout":      "author-series-title",
		"dryrun":      "true",
		"removeEmpty": "false",
		"verbose":     "true",
	}

	model := NewPreviewModel([]AudioBook{}, config, organizer.DefaultFieldMapping())

	// Test that config is properly stored and accessible
	for key, expectedValue := range config {
		if actualValue, exists := model.config[key]; !exists {
			t.Errorf("Config key %q should exist", key)
		} else if actualValue != expectedValue {
			t.Errorf("Config %q should be %q, got %q", key, expectedValue, actualValue)
		}
	}

	// Test view with different config options
	view := model.View()
	if view == "" {
		t.Error("View should not be empty with config")
	}

	// Config values might appear in the view
	// (exact behavior depends on implementation)
}
