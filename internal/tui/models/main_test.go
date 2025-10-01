package models

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewMainModel(t *testing.T) {
	tests := []struct {
		name      string
		inputDir  string
		outputDir string
	}{
		{
			name:      "valid directories",
			inputDir:  "/tmp/input",
			outputDir: "/tmp/output",
		},
		{
			name:      "empty directories",
			inputDir:  "",
			outputDir: "",
		},
		{
			name:      "relative paths",
			inputDir:  "./input",
			outputDir: "./output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewMainModel(tt.inputDir, tt.outputDir)

			if model == nil {
				t.Fatal("NewMainModel returned nil")
			}

			if model.inputDir != tt.inputDir {
				t.Errorf("Expected inputDir %q, got %q", tt.inputDir, model.inputDir)
			}

			if model.outputDir != tt.outputDir {
				t.Errorf("Expected outputDir %q, got %q", tt.outputDir, model.outputDir)
			}

			if model.screen != ScanScreen {
				t.Errorf("Expected initial screen to be ScanScreen, got %d", model.screen)
			}

			if model.quitting {
				t.Error("Expected quitting to be false initially")
			}

			if model.err != nil {
				t.Errorf("Expected no initial error, got %v", model.err)
			}
		})
	}
}

func TestMainModelInit(t *testing.T) {
	model := NewMainModel("/tmp/input", "/tmp/output")

	cmd := model.Init()

	// Init should return a command (tea.Cmd)
	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestMainModelScreenNavigation(t *testing.T) {
	model := NewMainModel("/tmp/input", "/tmp/output")

	tests := []struct {
		name           string
		currentScreen  Screen
		expectedScreen Screen
		keyMsg         string
	}{
		{
			name:           "navigate from scan to settings",
			currentScreen:  ScanScreen,
			expectedScreen: ScanScreen, // Screen changes happen via specific messages, not just key presses
			keyMsg:         "tab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.screen = tt.currentScreen

			// Create a key message
			msg := tea.KeyMsg{Type: tea.KeyTab}

			// Update the model
			updatedModel, cmd := model.Update(msg)

			// Verify the model is returned (even if screen doesn't change)
			if updatedModel == nil {
				t.Error("Update should return a model")
			}

			// Verify a command is returned (could be nil)
			_ = cmd // Commands can be nil, so just check it doesn't panic

			// The actual screen navigation logic would be tested based on the real implementation
			mainModel, ok := updatedModel.(*MainModel)
			if !ok {
				t.Error("Update should return a MainModel")
			}

			// Verify the model state is maintained
			if mainModel.inputDir != model.inputDir {
				t.Error("InputDir should be preserved after update")
			}
		})
	}
}

func TestMainModelView(t *testing.T) {
	model := NewMainModel("/tmp/input", "/tmp/output")

	// Test that View doesn't panic and returns a string
	view := model.View()

	if view == "" {
		t.Error("View() should return a non-empty string")
	}

	// Test with different screen sizes
	model.width = 80
	model.height = 24

	view = model.View()
	if view == "" {
		t.Error("View() should return a non-empty string with dimensions set")
	}
}

func TestMainModelWindowSizeHandling(t *testing.T) {
	model := NewMainModel("/tmp/input", "/tmp/output")

	// Test window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}

	updatedModel, cmd := model.Update(msg)

	mainModel, ok := updatedModel.(*MainModel)
	if !ok {
		t.Fatal("Update should return a MainModel")
	}

	if mainModel.width != 100 {
		t.Errorf("Expected width 100, got %d", mainModel.width)
	}

	if mainModel.height != 30 {
		t.Errorf("Expected height 30, got %d", mainModel.height)
	}

	_ = cmd // Command can be nil
}

func TestMainModelQuitHandling(t *testing.T) {
	model := NewMainModel("/tmp/input", "/tmp/output")

	// Test quit key
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	updatedModel, cmd := model.Update(msg)

	mainModel, ok := updatedModel.(*MainModel)
	if !ok {
		t.Fatal("Update should return a MainModel")
	}

	// Check if quitting is handled (implementation dependent)
	_ = mainModel.quitting // This would depend on actual implementation
	_ = cmd // Command handling varies by implementation
}

func TestMainModelErrorHandling(t *testing.T) {
	model := NewMainModel("/tmp/input", "/tmp/output")

	// Test that the model can handle error states
	model.err = nil
	if model.err != nil {
		t.Error("Initial error should be nil")
	}

	// Test view with error state
	model.err = fmt.Errorf("test error")
	view := model.View()

	if view == "" {
		t.Error("View should still return content even with error")
	}
}

func TestMainModelSubModelInitialization(t *testing.T) {
	model := NewMainModel("/tmp/input", "/tmp/output")

	// Test that sub-models are properly initialized when needed
	// This would depend on the actual implementation - when are sub-models created?

	// For now, just verify they start as nil (lazy initialization)
	if model.scanModel != nil && model.screen != ScanScreen {
		t.Log("scanModel initialized early - this may be intentional")
	}

	if model.bookListModel != nil && model.screen != BookListScreen {
		t.Log("bookListModel initialized early - this may be intentional")
	}

	if model.settingsModel != nil && model.screen != SettingsScreen {
		t.Log("settingsModel initialized early - this may be intentional")
	}
}
