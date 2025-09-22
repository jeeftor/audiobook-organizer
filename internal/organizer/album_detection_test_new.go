package organizer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldProcessAsAlbumWithFiles(t *testing.T) {
	// Create a test organizer
	org := &Organizer{
		config: OrganizerConfig{
			Verbose:             false,
			UseEmbeddedMetadata: true,
			FieldMapping:        DefaultFieldMapping(),
		},
	}

	tests := []struct {
		name     string
		testDir  string // Directory under testdata/
		expected bool
		reason   string
	}{
		{
			name:     "Single audio file should not be album",
			testDir:  "test-scenarios/single-file",
			expected: false,
			reason:   "Single files should not be treated as albums",
		},
		{
			name:     "Multi-part MP3 series should be album",
			testDir:  "mp3", // Contains multiple tracks from same series
			expected: true,
			reason:   "Multiple MP3 files with consistent metadata should be album",
		},
		{
			name:     "Mixed unrelated books should not be album",
			testDir:  "test-scenarios/mixed-unrelated",
			expected: false,
			reason:   "Files from different series/authors should not be grouped",
		},
		{
			name:     "Empty directory should not be album",
			testDir:  "test-scenarios/empty-dir",
			expected: false,
			reason:   "Empty directories have no audio files",
		},
		{
			name:     "Non-audio files only should not be album",
			testDir:  "test-scenarios/non-audio-only",
			expected: false,
			reason:   "Directories with no audio files should not be albums",
		},
		{
			name:     "Files without track numbers but same metadata should be album",
			testDir:  "test-scenarios/no-tracks",
			expected: true,
			reason:   "Files with consistent metadata should be album even without track numbers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build path to test directory
			testPath := filepath.Join("testdata", tt.testDir)

			// Call the actual function with real files
			result := org.shouldProcessAsAlbum(testPath)

			if result != tt.expected {
				t.Errorf("shouldProcessAsAlbum(%s) = %v, want %v\nReason: %s",
					tt.testDir, result, tt.expected, tt.reason)
			}
		})
	}
}

func TestGroupFilesByAlbumWithFiles(t *testing.T) {
	// Create a test organizer
	org := &Organizer{
		config: OrganizerConfig{
			Verbose:             false,
			UseEmbeddedMetadata: true,
			FieldMapping:        DefaultFieldMapping(),
		},
	}

	tests := []struct {
		name           string
		testDir        string
		expectedGroups int
		description    string
	}{
		{
			name:           "Multi-part series should create one album group",
			testDir:        "mp3",
			expectedGroups: 6, // Based on the different series in testdata/mp3/
			description:    "MP3 directory contains multiple series that should be grouped separately",
		},
		{
			name:           "Single files should create individual groups",
			testDir:        "test-scenarios/single-file",
			expectedGroups: 1,
			description:    "Single file should create one group",
		},
		{
			name:           "Mixed unrelated should create separate groups",
			testDir:        "test-scenarios/mixed-unrelated",
			expectedGroups: 2,
			description:    "Two unrelated books should create two groups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build path to test directory
			testPath := filepath.Join("testdata", tt.testDir)

			// Get directory entries
			entries, err := readDirFunc(testPath)
			if err != nil {
				t.Fatalf("Failed to read test directory %s: %v", testPath, err)
			}

			// Filter to audio files only
			var audioEntries []os.DirEntry
			for _, entry := range entries {
				if !entry.IsDir() && IsSupportedAudioFile(filepath.Ext(entry.Name())) {
					audioEntries = append(audioEntries, entry)
				}
			}

			// Group the files
			albumGroups, err := org.groupFilesByAlbum(testPath, audioEntries)
			if err != nil {
				t.Fatalf("groupFilesByAlbum failed: %v", err)
			}

			if len(albumGroups) != tt.expectedGroups {
				t.Errorf("groupFilesByAlbum() created %d groups, want %d\nDescription: %s",
					len(albumGroups), tt.expectedGroups, tt.description)
			}
		})
	}
}
