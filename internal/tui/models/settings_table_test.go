package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

func TestNewSettingsTableModel(t *testing.T) {
	tests := []struct {
		name          string
		selectedBooks []AudioBook
		showAdvanced  bool
	}{
		{
			name:          "empty book list",
			selectedBooks: []AudioBook{},
			showAdvanced:  false,
		},
		{
			name: "single book",
			selectedBooks: []AudioBook{
				{Metadata: organizer.Metadata{Title: "Test Book", Authors: []string{"Test Author"}}},
			},
			showAdvanced: false,
		},
		{
			name: "multiple books with advanced",
			selectedBooks: []AudioBook{
				{Metadata: organizer.Metadata{Title: "Book 1", Authors: []string{"Author 1"}}},
				{Metadata: organizer.Metadata{Title: "Book 2", Authors: []string{"Author 2"}}},
			},
			showAdvanced: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewSettingsTableModel(tt.selectedBooks, tt.showAdvanced)

			if model == nil {
				t.Fatal("NewSettingsTableModel returned nil")
			}

			if len(model.selectedBooks) != len(tt.selectedBooks) {
				t.Errorf("Expected %d selected books, got %d", len(tt.selectedBooks), len(model.selectedBooks))
			}

			if model.showAdvanced != tt.showAdvanced {
				t.Errorf("Expected showAdvanced=%v, got %v", tt.showAdvanced, model.showAdvanced)
			}

			if len(model.fieldMappings) == 0 {
				t.Error("Expected field mappings to be initialized")
			}

			if model.metadataWidget == nil {
				t.Error("Expected metadata widget to be initialized")
			}

			if model.pathPreviewWidget == nil {
				t.Error("Expected path preview widget to be initialized")
			}
		})
	}
}

func TestSettingsTableModelWithMode(t *testing.T) {
	books := []AudioBook{
		{Metadata: organizer.Metadata{Title: "Test Book"}},
	}

	tests := []struct {
		name     string
		scanMode string
	}{
		{"flat mode", "Flat"},
		{"embedded mode", "Embedded"},
		{"normal mode", "Normal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewSettingsTableModelWithMode(books, false, tt.scanMode)

			if model.scanMode != tt.scanMode {
				t.Errorf("Expected scanMode=%q, got %q", tt.scanMode, model.scanMode)
			}
		})
	}
}

func TestSettingsTableModelInit(t *testing.T) {
	model := NewSettingsTableModel([]AudioBook{}, false)

	cmd := model.Init()

	// Init should return a WindowSize command
	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestSettingsTableModelGetConfig(t *testing.T) {
	model := NewSettingsTableModel([]AudioBook{}, false)

	config := model.GetConfig()

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}

	// Should have some config values
	if len(config) == 0 {
		t.Error("Expected config to have values")
	}
}

func TestSettingsTableModelGetFieldMapping(t *testing.T) {
	model := NewSettingsTableModel([]AudioBook{}, false)

	mapping := model.GetFieldMapping()

	// Should have default field mapping values
	if mapping.TitleField == "" {
		t.Error("Expected TitleField to be set")
	}

	if mapping.SeriesField == "" {
		t.Error("Expected SeriesField to be set")
	}

	if len(mapping.AuthorFields) == 0 {
		t.Error("Expected AuthorFields to be set")
	}
}

func TestSettingsTableModelGetSelectedBooks(t *testing.T) {
	books := []AudioBook{
		{Metadata: organizer.Metadata{Title: "Book 1"}},
		{Metadata: organizer.Metadata{Title: "Book 2"}},
	}

	model := NewSettingsTableModel(books, false)

	selectedBooks := model.GetSelectedBooks()

	if len(selectedBooks) != len(books) {
		t.Errorf("Expected %d books, got %d", len(books), len(selectedBooks))
	}
}

func TestSettingsTableModelShouldAdvance(t *testing.T) {
	model := NewSettingsTableModel([]AudioBook{}, false)

	// Initially should be able to advance
	if !model.ShouldAdvance() {
		t.Error("Expected ShouldAdvance() to return true initially")
	}

	// When popup is showing, should not advance
	model.showPopup = true
	if model.ShouldAdvance() {
		t.Error("Expected ShouldAdvance() to return false when popup is showing")
	}

	// When popup was just closed, should not advance
	model.showPopup = false
	model.justClosedPopup = true
	if model.ShouldAdvance() {
		t.Error("Expected ShouldAdvance() to return false when popup was just closed")
	}
}

func TestSettingsTableModelView(t *testing.T) {
	model := NewSettingsTableModel([]AudioBook{}, false)

	// Set dimensions for view
	model.width = 120
	model.height = 40

	view := model.View()

	if view == "" {
		t.Error("View() should return non-empty string")
	}
}

func TestSettingsTableModelUpdate(t *testing.T) {
	model := NewSettingsTableModel([]AudioBook{}, false)
	model.width = 120
	model.height = 40

	// Test window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)

	settingsModel := updatedModel.(*SettingsTableModel)
	if settingsModel.width != 100 || settingsModel.height != 50 {
		t.Error("Window size should be updated")
	}
}

func TestSettingsTableModelFocusAreas(t *testing.T) {
	model := NewSettingsTableModel([]AudioBook{}, false)
	model.width = 120
	model.height = 40

	// Initially should be on table focus
	if model.focusArea != TableFocus {
		t.Error("Expected initial focus to be TableFocus")
	}

	// Test tab key to switch focus
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ := model.Update(msg)

	settingsModel := updatedModel.(*SettingsTableModel)
	if settingsModel.focusArea != MetadataFocus {
		t.Error("Expected focus to switch to MetadataFocus after tab")
	}

	// Tab again should go back to table
	updatedModel, _ = settingsModel.Update(msg)
	settingsModel = updatedModel.(*SettingsTableModel)
	if settingsModel.focusArea != TableFocus {
		t.Error("Expected focus to switch back to TableFocus after second tab")
	}
}
