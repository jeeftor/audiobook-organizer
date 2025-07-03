//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// TestTrackNumberInFilenames tests the behavior of track number detection in filenames
func TestTrackNumberInFilenames(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected int
	}{
		{"single_digit", "track1.mp3", 1},
		{"two_digits", "track12.mp3", 12},
		{"with_underscore", "track_1.mp3", 1},
		{"with_dash", "track-1.mp3", 1},
		{"no_number", "track.mp3", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := organizer.ExtractTrackNumber(tt.filename)
			if got != tt.expected {
				t.Errorf("ExtractTrackNumber(%q) = %v, want %v", tt.filename, got, tt.expected)
			}
		})
	}
}

// TestNonFlatStructureWithMetadata verifies the behavior with non-flat directory structure and metadata
func TestNonFlatStructureWithMetadata(t *testing.T) {
	testFiles := []testFile{
		{
			Path: "book1/chapter1.mp3",
			Metadata: &organizer.Metadata{
				Title:   "Book 1",
				Authors: []string{"Author 1"},
				Series:  []string{"Series 1"},
			},
		},
		{
			Path: "book1/chapter2.mp3",
			Metadata: &organizer.Metadata{
				Title:   "Book 1",
				Authors: []string{"Author 1"},
				Series:  []string{"Series 1"},
			},
		},
	}

	env := setupTestEnvironment(t, testFiles)
	defer env.Cleanup()

	// TODO: Add test logic here
}

// TestFlatVsNonFlatStructure compares flat and non-flat directory structures
func TestFlatVsNonFlatStructure(t *testing.T) {
	tests := []struct {
		name string
		flat bool
	}{
		{"flat_mode", true},
		{"non_flat_mode", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Add test logic here
			t.Logf("Testing %s (flat=%v)", tt.name, tt.flat)
		})
	}
}
