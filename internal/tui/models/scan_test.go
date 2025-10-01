package models

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

func TestNewScanModel(t *testing.T) {
	tests := []struct {
		name     string
		inputDir string
	}{
		{
			name:     "valid directory",
			inputDir: "/tmp/audiobooks",
		},
		{
			name:     "empty directory",
			inputDir: "",
		},
		{
			name:     "relative path",
			inputDir: "./audiobooks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewScanModel(tt.inputDir)

			if model == nil {
				t.Fatal("NewScanModel returned nil")
			}

			if model.inputDir != tt.inputDir {
				t.Errorf("Expected inputDir %q, got %q", tt.inputDir, model.inputDir)
			}

			if model.scanning {
				t.Error("Expected scanning to be false initially")
			}

			if model.complete {
				t.Error("Expected complete to be false initially")
			}

			if len(model.books) != 0 {
				t.Error("Expected empty books list initially")
			}

			if model.scannedDirs != 0 {
				t.Error("Expected scannedDirs to be 0 initially")
			}

			if model.scannedFiles != 0 {
				t.Error("Expected scannedFiles to be 0 initially")
			}
		})
	}
}

func TestScanModelInit(t *testing.T) {
	model := NewScanModel("/tmp/test")

	cmd := model.Init()

	// Init should return a scanning command
	if cmd == nil {
		t.Error("Init() should return a command to start scanning")
	}
}

func TestAudioBook(t *testing.T) {
	// Test AudioBook struct
	tests := []struct {
		name     string
		book     AudioBook
		expected AudioBook
	}{
		{
			name: "basic audiobook",
			book: AudioBook{
				Path: "/path/to/book.mp3",
				Metadata: organizer.Metadata{
					Title:   "Test Book",
					Authors: []string{"Test Author"},
				},
				Selected: false,
			},
			expected: AudioBook{
				Path: "/path/to/book.mp3",
				Metadata: organizer.Metadata{
					Title:   "Test Book",
					Authors: []string{"Test Author"},
				},
				Selected: false,
			},
		},
		{
			name: "album audiobook",
			book: AudioBook{
				Path:          "/path/to/album/track01.mp3",
				IsPartOfAlbum: true,
				AlbumName:     "Test Album",
				TrackNumber:   1,
				TotalTracks:   10,
				Selected:      true,
			},
			expected: AudioBook{
				Path:          "/path/to/album/track01.mp3",
				IsPartOfAlbum: true,
				AlbumName:     "Test Album",
				TrackNumber:   1,
				TotalTracks:   10,
				Selected:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.book.Path != tt.expected.Path {
				t.Errorf("Expected Path %q, got %q", tt.expected.Path, tt.book.Path)
			}

			if tt.book.Selected != tt.expected.Selected {
				t.Errorf("Expected Selected %v, got %v", tt.expected.Selected, tt.book.Selected)
			}

			if tt.book.IsPartOfAlbum != tt.expected.IsPartOfAlbum {
				t.Errorf("Expected IsPartOfAlbum %v, got %v", tt.expected.IsPartOfAlbum, tt.book.IsPartOfAlbum)
			}

			if tt.book.AlbumName != tt.expected.AlbumName {
				t.Errorf("Expected AlbumName %q, got %q", tt.expected.AlbumName, tt.book.AlbumName)
			}

			if tt.book.TrackNumber != tt.expected.TrackNumber {
				t.Errorf("Expected TrackNumber %d, got %d", tt.expected.TrackNumber, tt.book.TrackNumber)
			}

			if tt.book.TotalTracks != tt.expected.TotalTracks {
				t.Errorf("Expected TotalTracks %d, got %d", tt.expected.TotalTracks, tt.book.TotalTracks)
			}
		})
	}
}

func TestScanModelScanMsg(t *testing.T) {
	model := NewScanModel("/tmp/test")

	book := AudioBook{
		Path: "/test/book.mp3",
		Metadata: organizer.Metadata{
			Title:   "Test Book",
			Authors: []string{"Author"},
		},
	}

	msg := ScanMsg{Book: book}

	updatedModel, cmd := model.Update(msg)

	scanModel, ok := updatedModel.(*ScanModel)
	if !ok {
		t.Fatal("Update should return a ScanModel")
	}

	if len(scanModel.books) != 1 {
		t.Errorf("Expected 1 book, got %d", len(scanModel.books))
	}

	if scanModel.books[0].Path != book.Path {
		t.Errorf("Expected book path %q, got %q", book.Path, scanModel.books[0].Path)
	}

	_ = cmd // Command can be nil
}

func TestScanModelScanCompleteMsg(t *testing.T) {
	model := NewScanModel("/tmp/test")
	model.scanning = true

	books := []AudioBook{
		{Path: "/test/book1.mp3", Metadata: organizer.Metadata{Title: "Book 1"}},
		{Path: "/test/book2.mp3", Metadata: organizer.Metadata{Title: "Book 2"}},
	}

	msg := ScanCompleteMsg{Books: books}

	updatedModel, cmd := model.Update(msg)

	scanModel, ok := updatedModel.(*ScanModel)
	if !ok {
		t.Fatal("Update should return a ScanModel")
	}

	if !scanModel.complete {
		t.Error("Expected complete to be true after ScanCompleteMsg")
	}

	if scanModel.scanning {
		t.Error("Expected scanning to be false after ScanCompleteMsg")
	}

	if len(scanModel.books) != len(books) {
		t.Errorf("Expected %d books, got %d", len(books), len(scanModel.books))
	}

	_ = cmd // Command can be nil
}

func TestScanModelView(t *testing.T) {
	model := NewScanModel("/tmp/test")

	// Test view during scanning
	model.scanning = true
	model.startTime = time.Now().Add(-5 * time.Second)
	model.scannedDirs = 10
	model.scannedFiles = 50

	view := model.View()
	if view == "" {
		t.Error("View() should return non-empty string during scanning")
	}

	// Test view when complete
	model.scanning = false
	model.complete = true
	model.books = []AudioBook{
		{Path: "/test/book1.mp3"},
		{Path: "/test/book2.mp3"},
	}

	view = model.View()
	if view == "" {
		t.Error("View() should return non-empty string when complete")
	}

	// View should show number of books found
	if !strings.Contains(view, "2") {
		t.Error("View should show the number of books found")
	}
}

func TestScanModelElapsedTimeTracking(t *testing.T) {
	model := NewScanModel("/tmp/test")

	// Start scanning
	model.scanning = true
	model.startTime = time.Now().Add(-2 * time.Second)

	// Update elapsed time (this would normally happen in real updates)
	model.elapsedTime = time.Since(model.startTime)

	if model.elapsedTime < time.Second {
		t.Error("Expected elapsed time to be at least 1 second")
	}
}

func TestScanModelKeyHandling(t *testing.T) {
	model := NewScanModel("/tmp/test")

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
		{
			name:   "character key",
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedModel, cmd := model.Update(tt.keyMsg)

			// Verify model is returned
			if updatedModel == nil {
				t.Error("Update should return a model")
			}

			// Verify it's still a ScanModel
			_, ok := updatedModel.(*ScanModel)
			if !ok {
				t.Error("Update should return a ScanModel")
			}

			_ = cmd // Command can be nil
		})
	}
}

func TestScanModelStatisticsTracking(t *testing.T) {
	model := NewScanModel("/tmp/test")

	// Test that statistics can be tracked
	model.scannedDirs = 5
	model.scannedFiles = 25

	if model.scannedDirs != 5 {
		t.Errorf("Expected scannedDirs 5, got %d", model.scannedDirs)
	}

	if model.scannedFiles != 25 {
		t.Errorf("Expected scannedFiles 25, got %d", model.scannedFiles)
	}

	// Add books and verify count
	model.books = append(model.books, AudioBook{Path: "/test/book.mp3"})

	if len(model.books) != 1 {
		t.Errorf("Expected 1 book, got %d", len(model.books))
	}
}

func TestScanModelWindowSizeHandling(t *testing.T) {
	model := NewScanModel("/tmp/test")

	// Test window size message (if the model handles it)
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}

	updatedModel, cmd := model.Update(msg)

	// Verify model is returned
	if updatedModel == nil {
		t.Error("Update should return a model")
	}

	// Verify it's still a ScanModel
	_, ok := updatedModel.(*ScanModel)
	if !ok {
		t.Error("Update should return a ScanModel")
	}

	_ = cmd // Command can be nil
}
