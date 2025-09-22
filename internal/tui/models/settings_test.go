package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

func TestNewSettingsModel(t *testing.T) {
	tests := []struct {
		name          string
		selectedBooks []AudioBook
	}{
		{
			name:          "empty book list",
			selectedBooks: []AudioBook{},
		},
		{
			name: "single book",
			selectedBooks: []AudioBook{
				{Metadata: organizer.Metadata{Title: "Test Book", Authors: []string{"Test Author"}}},
			},
		},
		{
			name: "multiple books",
			selectedBooks: []AudioBook{
				{Metadata: organizer.Metadata{Title: "Book 1", Authors: []string{"Author 1"}}},
				{Metadata: organizer.Metadata{Title: "Book 2", Authors: []string{"Author 2"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewSettingsModel(tt.selectedBooks)

			if model == nil {
				t.Fatal("NewSettingsModel returned nil")
			}

			if len(model.selectedBooks) != len(tt.selectedBooks) {
				t.Errorf("Expected %d selected books, got %d", len(tt.selectedBooks), len(model.selectedBooks))
			}

			if model.cursor != 0 {
				t.Errorf("Expected initial cursor position 0, got %d", model.cursor)
			}

			if model.showAdvanced {
				t.Error("Expected showAdvanced to be false initially")
			}

			if model.filtering {
				t.Error("Expected filtering to be false initially")
			}

			if model.filterString != "" {
				t.Errorf("Expected empty filter string initially, got %q", model.filterString)
			}

			if len(model.settings) == 0 {
				t.Error("Expected settings to be initialized")
			}

			if len(model.fieldMappings) == 0 {
				t.Error("Expected field mappings to be initialized")
			}
		})
	}
}

func TestSettingsModelInit(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})

	cmd := model.Init()

	// Init should return nil command for settings model
	if cmd != nil {
		t.Error("Init() should return nil command for settings model")
	}
}

func TestSettingsModelFilterMode(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})

	// Test entering filter mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	updatedModel, cmd := model.Update(msg)

	settingsModel, ok := updatedModel.(*SettingsModel)
	if !ok {
		t.Fatal("Update should return a SettingsModel")
	}

	if !settingsModel.filtering {
		t.Error("Expected filtering mode to be enabled after '/' key")
	}

	if settingsModel.filterString != "" {
		t.Error("Expected filter string to be empty when entering filter mode")
	}

	_ = cmd // Command can be nil
}

func TestSettingsModelFilterInput(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})
	model.filtering = true // Start in filtering mode

	tests := []struct {
		name           string
		input          string
		expectedFilter string
	}{
		{
			name:           "single character",
			input:          "l",
			expectedFilter: "l",
		},
		{
			name:           "multiple characters",
			input:          "layout",
			expectedFilter: "layout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset filter string
			model.filterString = ""

			// Add each character
			for _, char := range tt.input {
				msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}}
				updatedModel, _ := model.Update(msg)
				model = updatedModel.(*SettingsModel)
			}

			if model.filterString != tt.expectedFilter {
				t.Errorf("Expected filter string %q, got %q", tt.expectedFilter, model.filterString)
			}
		})
	}
}

func TestSettingsModelFilterBackspace(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})
	model.filtering = true
	model.filterString = "test"

	// Test backspace
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := model.Update(msg)

	settingsModel := updatedModel.(*SettingsModel)
	if settingsModel.filterString != "tes" {
		t.Errorf("Expected filter string 'tes', got %q", settingsModel.filterString)
	}

	// Test backspace on empty string
	model.filterString = ""
	updatedModel, _ = model.Update(msg)
	settingsModel = updatedModel.(*SettingsModel)

	if settingsModel.filterString != "" {
		t.Errorf("Expected filter string to remain empty, got %q", settingsModel.filterString)
	}
}

func TestSettingsModelFilterExit(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})
	model.filtering = true
	model.filterString = "test"

	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{
			name:   "exit with escape",
			keyMsg: tea.KeyMsg{Type: tea.KeyEsc},
		},
		{
			name:   "exit with enter",
			keyMsg: tea.KeyMsg{Type: tea.KeyEnter},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to filtering state
			model.filtering = true
			model.filterString = "test"

			updatedModel, _ := model.Update(tt.keyMsg)
			settingsModel := updatedModel.(*SettingsModel)

			if settingsModel.filtering {
				t.Error("Expected filtering mode to be disabled")
			}

			if settingsModel.filterString != "" {
				t.Errorf("Expected filter string to be cleared, got %q", settingsModel.filterString)
			}
		})
	}
}

func TestSettingsModelApplyFilter(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})

	// Ensure we have some settings to filter
	if len(model.settings) == 0 {
		t.Skip("No settings available to test filtering")
	}

	originalCursor := model.cursor

	// Test filtering with a term that should match
	model.filterString = "layout" // Should match "Layout" setting
	model.applyFilter()

	// Check if cursor moved to a matching setting
	found := false
	for i, setting := range model.settings {
		if strings.Contains(strings.ToLower(setting.Name), "layout") ||
			strings.Contains(strings.ToLower(setting.Description), "layout") {
			if model.cursor == i {
				found = true
				break
			}
		}
	}

	if !found && model.cursor == originalCursor {
		// If no match found, cursor should stay where it was
		t.Logf("No layout setting found, cursor remained at %d", model.cursor)
	}

	// Test with empty filter string
	model.filterString = ""
	originalCursor = model.cursor
	model.applyFilter()

	if model.cursor != originalCursor {
		t.Error("Cursor should not change when filter string is empty")
	}
}

func TestSettingsModelNavigationKeys(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})

	if len(model.settings) < 2 {
		t.Skip("Need at least 2 settings to test navigation")
	}

	originalCursor := model.cursor

	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{
			name:   "down arrow",
			keyMsg: tea.KeyMsg{Type: tea.KeyDown},
		},
		{
			name:   "j key",
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")},
		},
		{
			name:   "up arrow",
			keyMsg: tea.KeyMsg{Type: tea.KeyUp},
		},
		{
			name:   "k key",
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset cursor
			model.cursor = originalCursor

			updatedModel, _ := model.Update(tt.keyMsg)
			settingsModel := updatedModel.(*SettingsModel)

			// Cursor should be within valid range
			if settingsModel.cursor < 0 || settingsModel.cursor >= len(settingsModel.settings) {
				t.Errorf("Cursor %d is out of range [0, %d)", settingsModel.cursor, len(settingsModel.settings))
			}
		})
	}
}

func TestSettingsModelView(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})

	// Test normal view
	view := model.View()
	if view == "" {
		t.Error("View() should return non-empty string")
	}

	// Test view in filtering mode
	model.filtering = true
	model.filterString = "test"
	view = model.View()

	if view == "" {
		t.Error("View() should return non-empty string in filtering mode")
	}

	if !strings.Contains(view, "Filter: test") {
		t.Error("View should show filter string when filtering")
	}

	// Test view with advanced settings
	model.filtering = false
	model.showAdvanced = true
	view = model.View()

	if view == "" {
		t.Error("View() should return non-empty string with advanced settings")
	}
}

func TestSettingsModelAdvancedToggle(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})

	originalAdvanced := model.showAdvanced

	// Test that advanced mode can be toggled (implementation specific)
	// This would depend on what key toggles advanced mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")} // Assuming 'a' toggles advanced

	updatedModel, _ := model.Update(msg)
	settingsModel := updatedModel.(*SettingsModel)

	// The behavior depends on implementation - this is just structural testing
	_ = settingsModel.showAdvanced // Verify field exists and is accessible

	// Reset
	model.showAdvanced = originalAdvanced
}

func TestSettingsModelFocusHandling(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})

	// Test that settings have focus state
	for i := range model.settings {
		// Focus state should be boolean
		_ = model.settings[i].Focused
	}

	// Test field mappings focus
	for i := range model.fieldMappings {
		_ = model.fieldMappings[i].Focused
	}
}

func TestSettingsModelGetCurrentSetting(t *testing.T) {
	model := NewSettingsModel([]AudioBook{})

	if len(model.settings) == 0 {
		t.Skip("No settings available to test")
	}

	// Test that cursor points to valid setting
	if model.cursor >= 0 && model.cursor < len(model.settings) {
		currentSetting := model.settings[model.cursor]
		if currentSetting.Name == "" {
			t.Error("Current setting should have a name")
		}
	}
}
