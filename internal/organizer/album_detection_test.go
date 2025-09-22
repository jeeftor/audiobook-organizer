package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHasCommonPrefix(t *testing.T) {
	tests := []struct {
		name     string
		str1     string
		str2     string
		expected bool
	}{
		{
			name:     "Same strings",
			str1:     "Book Title",
			str2:     "Book Title",
			expected: false, // No separator found
		},
		{
			name:     "With dash separator",
			str1:     "Book Title - Track 01",
			str2:     "Book Title - Track 02",
			expected: true,
		},
		{
			name:     "With colon separator",
			str1:     "Book Title: Part 1",
			str2:     "Book Title: Part 2",
			expected: true,
		},
		{
			name:     "With comma separator",
			str1:     "Book Title, Chapter 1",
			str2:     "Book Title, Chapter 2",
			expected: true,
		},
		{
			name:     "Different titles",
			str1:     "Book Title One",
			str2:     "Book Title Two",
			expected: false,
		},
		{
			name:     "Short prefix",
			str1:     "A - Part 1",
			str2:     "A - Part 2",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasCommonPrefix(tt.str1, tt.str2)
			if result != tt.expected {
				t.Errorf("hasCommonPrefix(%q, %q) = %v, want %v", tt.str1, tt.str2, result, tt.expected)
			}
		})
	}
}

func TestHasTrackNumberPattern(t *testing.T) {
	tests := []struct {
		name     string
		str1     string
		str2     string
		expected bool
	}{
		{
			name:     "Track pattern",
			str1:     "Book Title Track 1",
			str2:     "Book Title Track 2",
			expected: true,
		},
		{
			name:     "Part pattern",
			str1:     "Book Title Part 1",
			str2:     "Book Title Part 2",
			expected: true,
		},
		{
			name:     "Chapter pattern",
			str1:     "Book Title Chapter 1",
			str2:     "Book Title Chapter 2",
			expected: true,
		},
		{
			name:     "Different titles with numbers",
			str1:     "Book 1",
			str2:     "Book 2",
			expected: true, // Similar without numbers
		},
		{
			name:     "Completely different",
			str1:     "Book One",
			str2:     "Another Book",
			expected: false,
		},
		{
			name:     "Case insensitive",
			str1:     "Book Title track 1",
			str2:     "Book Title TRACK 2",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasTrackNumberPattern(tt.str1, tt.str2)
			if result != tt.expected {
				t.Errorf("hasTrackNumberPattern(%q, %q) = %v, want %v", tt.str1, tt.str2, result, tt.expected)
			}
		})
	}
}

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Special characters",
			input:    "Title & Author @ Work",
			expected: "title and author at work",
		},
		{
			name:     "Plus signs",
			input:    "Book+Plus+Signs",
			expected: "bookplusplusplussigns",
		},
		{
			name:     "Multiple symbols",
			input:    "Book$$$Money%Percent",
			expected: "bookdollarmoneypercentpercent",
		},
		{
			name:     "Dots and underscores",
			input:    "Book.With.Dots_And_Underscores",
			expected: "book with dots and underscores",
		},
		{
			name:     "Slashes and backslashes",
			input:    "Book/With\\Slashes",
			expected: "bookwithslashes",
		},
		{
			name:     "Extra whitespace",
			input:    "  Book   With   Spaces  ",
			expected: "book with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeString(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "With numbers",
			input:    "Track 123",
			expected: "Track ",
		},
		{
			name:     "Mixed numbers and text",
			input:    "Chapter1Part2Volume3",
			expected: "ChapterPartVolume",
		},
		{
			name:     "No numbers",
			input:    "Book Title",
			expected: "Book Title",
		},
		{
			name:     "Only numbers",
			input:    "12345",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripNumbers(tt.input)
			if result != tt.expected {
				t.Errorf("stripNumbers(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStringSimilarity(t *testing.T) {
	tests := []struct {
		name      string
		s1        string
		s2        string
		minExpect float64
		maxExpect float64
	}{
		{
			name:      "Identical strings",
			s1:        "Book Title",
			s2:        "Book Title",
			minExpect: 1.0,
			maxExpect: 1.0,
		},
		{
			name:      "Similar strings",
			s1:        "Book Title",
			s2:        "Book Titles",
			minExpect: 0.8,
			maxExpect: 0.95,
		},
		{
			name:      "Different strings",
			s1:        "Book Title",
			s2:        "Another Book",
			minExpect: 0.0,
			maxExpect: 0.5,
		},
		{
			name:      "Case insensitive",
			s1:        "BOOK TITLE",
			s2:        "book title",
			minExpect: 1.0,
			maxExpect: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringSimilarity(tt.s1, tt.s2)
			if result < tt.minExpect || result > tt.maxExpect {
				t.Errorf("stringSimilarity(%q, %q) = %v, want between %v and %v",
					tt.s1, tt.s2, result, tt.minExpect, tt.maxExpect)
			}
		})
	}
}

func TestCreateAlbumKey(t *testing.T) {
	// Create a minimal organizer for testing
	org := &Organizer{
		config: OrganizerConfig{},
	}

	tests := []struct {
		name     string
		metadata Metadata
		expected string
	}{
		{
			name: "Basic metadata",
			metadata: Metadata{
				Title:   "Book Title",
				Authors: []string{"Author Name"},
			},
			expected: "author name|book title",
		},
		{
			name: "With series",
			metadata: Metadata{
				Title:   "Book Title",
				Authors: []string{"Author Name"},
				Series:  []string{"Series Name"},
			},
			expected: "author name|book title|series name",
		},
		{
			name: "Multiple authors",
			metadata: Metadata{
				Title:   "Book Title",
				Authors: []string{"Author One", "Author Two"},
			},
			expected: "author one,author two|book title",
		},
		{
			name: "Special characters",
			metadata: Metadata{
				Title:   "Book & Title+Special",
				Authors: []string{"Author@Name"},
			},
			expected: "authoratname|book and titleplusspecial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := org.createAlbumKey(tt.metadata)
			if result != tt.expected {
				t.Errorf("createAlbumKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCreateAlbumKeyWithSpecialCharacters(t *testing.T) {
	// Create a minimal organizer for testing
	org := &Organizer{
		config: OrganizerConfig{},
	}

	tests := []struct {
		name     string
		metadata Metadata
		expected string
	}{
		{
			name: "Special characters in title",
			metadata: Metadata{
				Title:   "Book & Title: Special Edition",
				Authors: []string{"Author Name"},
			},
			expected: "author name|book and title special edition",
		},
		{
			name: "Special characters in author",
			metadata: Metadata{
				Title:   "Book Title",
				Authors: []string{"Author @ Name"},
			},
			expected: "author at name|book title",
		},
		{
			name: "Multiple special characters",
			metadata: Metadata{
				Title:   "Book & Title: Special $$ Edition",
				Authors: []string{"Author @ Name"},
				Series:  []string{"Series #1"},
			},
			expected: "author at name|book and title special dollar dollar edition|series",
		},
		{
			name: "Dots and underscores",
			metadata: Metadata{
				Title:   "Book.With.Dots_And_Underscores",
				Authors: []string{"Author.Name"},
			},
			expected: "author name|book with dots and underscores",
		},
		{
			name: "Multiple authors with special characters",
			metadata: Metadata{
				Title:   "Book Title",
				Authors: []string{"Author & Co", "Second @ Author"},
			},
			expected: "author and co,second at author|book title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := org.createAlbumKey(tt.metadata)
			if result != tt.expected {
				t.Errorf("createAlbumKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCreateAlbumKeyWithSeriesVariations(t *testing.T) {
	// Create a minimal organizer for testing
	org := &Organizer{
		config: OrganizerConfig{},
	}

	tests := []struct {
		name     string
		metadata Metadata
		expected string
	}{
		{
			name: "Single series",
			metadata: Metadata{
				Title:   "Book Title",
				Authors: []string{"Author Name"},
				Series:  []string{"Series Name"},
			},
			expected: "author name|book title|series name",
		},
		{
			name: "Multiple series - should use first",
			metadata: Metadata{
				Title:   "Book Title",
				Authors: []string{"Author Name"},
				Series:  []string{"Primary Series", "Secondary Series"},
			},
			expected: "author name|book title|primary series",
		},
		{
			name: "Empty series",
			metadata: Metadata{
				Title:   "Book Title",
				Authors: []string{"Author Name"},
				Series:  []string{""},
			},
			expected: "author name|book title",
		},
		{
			name: "Series with same name as title",
			metadata: Metadata{
				Title:   "Series Name: Book Title",
				Authors: []string{"Author Name"},
				Series:  []string{"Series Name"},
			},
			expected: "author name|series name book title|series name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := org.createAlbumKey(tt.metadata)
			if result != tt.expected {
				t.Errorf("createAlbumKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestShouldProcessAsAlbum(t *testing.T) {
	// Create a test organizer
	org := &Organizer{
		config: OrganizerConfig{
			FieldMapping: FieldMapping{},
		},
	}

	// Save the original functions to restore later
	originalReadDir := readDirFunc
	originalNewAudioMetadataProvider := newAudioMetadataProviderFunc

	// Restore the original functions when the test completes
	defer func() {
		readDirFunc = originalReadDir
		newAudioMetadataProviderFunc = originalNewAudioMetadataProvider
	}()

	tests := []struct {
		name           string
		mockDirEntries []os.DirEntry
		mockMetadata   map[string]Metadata
		expected       bool
	}{
		{
			name: "Single audio file - not an album",
			mockDirEntries: []os.DirEntry{
				mockDirEntry{name: "audiobook.mp3", isDir: false},
				mockDirEntry{name: "cover.jpg", isDir: false},
			},
			mockMetadata: map[string]Metadata{
				"audiobook.mp3": {
					Title:   "Audiobook Title",
					Authors: []string{"Author Name"},
				},
			},
			expected: false,
		},
		{
			name: "Multiple audio files with consistent metadata - is an album",
			mockDirEntries: []os.DirEntry{
				mockDirEntry{name: "track01.mp3", isDir: false},
				mockDirEntry{name: "track02.mp3", isDir: false},
				mockDirEntry{name: "track03.mp3", isDir: false},
				mockDirEntry{name: "cover.jpg", isDir: false},
			},
			mockMetadata: map[string]Metadata{
				"track01.mp3": {
					Title:       "Album Title",
					Authors:     []string{"Album Artist"},
					TrackNumber: 1,
				},
				"track02.mp3": {
					Title:       "Album Title",
					Authors:     []string{"Album Artist"},
					TrackNumber: 2,
				},
				"track03.mp3": {
					Title:       "Album Title",
					Authors:     []string{"Album Artist"},
					TrackNumber: 3,
				},
			},
			expected: true,
		},
		{
			name: "Multiple audio files with inconsistent titles but sequential tracks - is an album",
			mockDirEntries: []os.DirEntry{
				mockDirEntry{name: "track01.mp3", isDir: false},
				mockDirEntry{name: "track02.mp3", isDir: false},
				mockDirEntry{name: "track03.mp3", isDir: false},
			},
			mockMetadata: map[string]Metadata{
				"track01.mp3": {
					Title:       "Book Title - Part 1",
					Authors:     []string{"Book Author"},
					TrackNumber: 1,
				},
				"track02.mp3": {
					Title:       "Book Title - Part 2",
					Authors:     []string{"Book Author"},
					TrackNumber: 2,
				},
				"track03.mp3": {
					Title:       "Book Title - Part 3",
					Authors:     []string{"Book Author"},
					TrackNumber: 3,
				},
			},
			expected: true,
		},
		{
			name: "Multiple audio files with inconsistent metadata - not an album",
			mockDirEntries: []os.DirEntry{
				mockDirEntry{name: "book1.mp3", isDir: false},
				mockDirEntry{name: "book2.mp3", isDir: false},
			},
			mockMetadata: map[string]Metadata{
				"book1.mp3": {
					Title:   "Book One",
					Authors: []string{"Author One"},
				},
				"book2.mp3": {
					Title:   "Book Two",
					Authors: []string{"Author Two"},
				},
			},
			expected: false,
		},
		{
			name: "Multiple audio files with same series but different titles - is an album",
			mockDirEntries: []os.DirEntry{
				mockDirEntry{name: "book1.mp3", isDir: false},
				mockDirEntry{name: "book2.mp3", isDir: false},
			},
			mockMetadata: map[string]Metadata{
				"book1.mp3": {
					Title:       "Book Series Volume 1",
					Authors:     []string{"Series Author"},
					Series:      []string{"Book Series"},
					TrackNumber: 1,
				},
				"book2.mp3": {
					Title:       "Book Series Volume 2",
					Authors:     []string{"Series Author"},
					Series:      []string{"Book Series"},
					TrackNumber: 2,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock os.ReadDir to return our test entries
			readDirFunc = func(dirPath string) ([]os.DirEntry, error) {
				return tt.mockDirEntries, nil
			}

			// Mock NewAudioMetadataProvider to return our test metadata
			newAudioMetadataProviderFunc = func(filePath string) MetadataProvider {
				fileName := filepath.Base(filePath)
				metadata, exists := tt.mockMetadata[fileName]

				// Create a simple mock that implements MetadataProvider
				mockProvider := &mockAudioProvider{}
				if !exists {
					mockProvider.metadata = Metadata{}
					mockProvider.err = fmt.Errorf("no metadata for %s", fileName)
				} else {
					mockProvider.metadata = metadata
					mockProvider.err = nil
				}

				return mockProvider
			}

			// Test the function
			result := org.shouldProcessAsAlbum("test/dir")
			if result != tt.expected {
				t.Errorf("shouldProcessAsAlbum() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGroupFilesByAlbum(t *testing.T) {
	// Create a test organizer
	org := &Organizer{
		config: OrganizerConfig{
			FieldMapping: FieldMapping{},
		},
	}

	// Save the original functions to restore later
	originalNewAudioMetadataProvider := newAudioMetadataProviderFunc

	// Restore the original functions when the test completes
	defer func() {
		newAudioMetadataProviderFunc = originalNewAudioMetadataProvider
	}()

	tests := []struct {
		name           string
		dirEntries     []os.DirEntry
		mockMetadata   map[string]Metadata
		expectedGroups int
		expectedFiles  map[string]int // albumKey -> number of files
	}{
		{
			name: "Single album with multiple tracks",
			dirEntries: []os.DirEntry{
				mockDirEntry{name: "track01.mp3", isDir: false},
				mockDirEntry{name: "track02.mp3", isDir: false},
				mockDirEntry{name: "track03.mp3", isDir: false},
				mockDirEntry{name: "cover.jpg", isDir: false},
			},
			mockMetadata: map[string]Metadata{
				"track01.mp3": {
					Title:       "Album Title",
					Authors:     []string{"Album Artist"},
					TrackNumber: 1,
				},
				"track02.mp3": {
					Title:       "Album Title",
					Authors:     []string{"Album Artist"},
					TrackNumber: 2,
				},
				"track03.mp3": {
					Title:       "Album Title",
					Authors:     []string{"Album Artist"},
					TrackNumber: 3,
				},
			},
			expectedGroups: 1,
			expectedFiles: map[string]int{
				"album artist|album title": 3,
			},
		},
		{
			name: "Multiple albums in one directory",
			dirEntries: []os.DirEntry{
				mockDirEntry{name: "album1_track1.mp3", isDir: false},
				mockDirEntry{name: "album1_track2.mp3", isDir: false},
				mockDirEntry{name: "album2_track1.mp3", isDir: false},
				mockDirEntry{name: "album2_track2.mp3", isDir: false},
			},
			mockMetadata: map[string]Metadata{
				"album1_track1.mp3": {
					Title:       "Album One",
					Authors:     []string{"Artist One"},
					TrackNumber: 1,
				},
				"album1_track2.mp3": {
					Title:       "Album One",
					Authors:     []string{"Artist One"},
					TrackNumber: 2,
				},
				"album2_track1.mp3": {
					Title:       "Album Two",
					Authors:     []string{"Artist Two"},
					TrackNumber: 1,
				},
				"album2_track2.mp3": {
					Title:       "Album Two",
					Authors:     []string{"Artist Two"},
					TrackNumber: 2,
				},
			},
			expectedGroups: 2,
			expectedFiles: map[string]int{
				"artist one|album one": 2,
				"artist two|album two": 2,
			},
		},
		{
			name: "Album with series information",
			dirEntries: []os.DirEntry{
				mockDirEntry{name: "series_book1.mp3", isDir: false},
				mockDirEntry{name: "series_book2.mp3", isDir: false},
			},
			mockMetadata: map[string]Metadata{
				"series_book1.mp3": {
					Title:       "Book One",
					Authors:     []string{"Series Author"},
					Series:      []string{"Book Series"},
					TrackNumber: 1,
				},
				"series_book2.mp3": {
					Title:       "Book Two",
					Authors:     []string{"Series Author"},
					Series:      []string{"Book Series"},
					TrackNumber: 2,
				},
			},
			expectedGroups: 2, // Different titles create different groups
			expectedFiles: map[string]int{
				"series author|book one|book series":  1,
				"series author|book two|book series":  1,
			},
		},
		{
			name: "Special characters in metadata",
			dirEntries: []os.DirEntry{
				mockDirEntry{name: "special1.mp3", isDir: false},
				mockDirEntry{name: "special2.mp3", isDir: false},
			},
			mockMetadata: map[string]Metadata{
				"special1.mp3": {
					Title:       "Book & Title+Special",
					Authors:     []string{"Author@Name"},
					TrackNumber: 1,
				},
				"special2.mp3": {
					Title:       "Book & Title+Special",
					Authors:     []string{"Author@Name"},
					TrackNumber: 2,
				},
			},
			expectedGroups: 1,
			expectedFiles: map[string]int{
				"authoratname|book and titleplusspecial": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock NewAudioMetadataProvider to return our test metadata
			newAudioMetadataProviderFunc = func(filePath string) MetadataProvider {
				fileName := filepath.Base(filePath)
				metadata, exists := tt.mockMetadata[fileName]

				// Create a simple mock that implements MetadataProvider
				mockProvider := &mockAudioProvider{}
				if !exists {
					mockProvider.metadata = Metadata{}
					mockProvider.err = fmt.Errorf("no metadata for %s", fileName)
				} else {
					mockProvider.metadata = metadata
					mockProvider.err = nil
				}

				return mockProvider
			}

			// Test the function
			albumGroups, err := org.groupFilesByAlbum("test/dir", tt.dirEntries)
			if err != nil {
				t.Fatalf("groupFilesByAlbum() error = %v", err)
			}

			// Check number of groups
			if len(albumGroups) != tt.expectedGroups {
				t.Errorf("groupFilesByAlbum() returned %d groups, want %d", len(albumGroups), tt.expectedGroups)
			}

			// Check number of files in each group
			for key, group := range albumGroups {
				expectedCount, exists := tt.expectedFiles[key]
				if !exists {
					t.Errorf("Unexpected album key: %s", key)
					continue
				}
				if len(group.Files) != expectedCount {
					t.Errorf("Album group %s has %d files, want %d", key, len(group.Files), expectedCount)
				}
			}
		})
	}
}

func TestAlbumGroupSorting(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		trackNums map[string]int
		expected  []string
	}{
		{
			name:      "Sort by track number",
			files:     []string{"file3.mp3", "file1.mp3", "file2.mp3"},
			trackNums: map[string]int{"file1.mp3": 1, "file2.mp3": 2, "file3.mp3": 3},
			expected:  []string{"file1.mp3", "file2.mp3", "file3.mp3"},
		},
		{
			name:      "Some files without track numbers",
			files:     []string{"fileC.mp3", "fileA.mp3", "file1.mp3"},
			trackNums: map[string]int{"file1.mp3": 1, "fileA.mp3": 0, "fileC.mp3": 0},
			expected:  []string{"file1.mp3", "fileA.mp3", "fileC.mp3"}, // Track number first, then alphabetical
		},
		{
			name:      "No track numbers",
			files:     []string{"fileC.mp3", "fileA.mp3", "fileB.mp3"},
			trackNums: map[string]int{},
			expected:  []string{"fileA.mp3", "fileB.mp3", "fileC.mp3"}, // Alphabetical order
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create album group
			ag := NewAlbumGroup(Metadata{})

			// Add files with track numbers
			for _, file := range tt.files {
				trackNum := 0
				if num, exists := tt.trackNums[file]; exists {
					trackNum = num
				}
				ag.AddFile(file, trackNum)
			}

			// Sort files
			ag.SortFilesByTrackNumber()

			// Check if sorted correctly
			if len(ag.Files) != len(tt.expected) {
				t.Fatalf("SortFilesByTrackNumber() resulted in %d files, want %d", len(ag.Files), len(tt.expected))
			}

			for i, file := range ag.Files {
				if file != tt.expected[i] {
					t.Errorf("SortFilesByTrackNumber() at position %d: got %s, want %s", i, file, tt.expected[i])
				}
			}
		})
	}
}

// Mock types for testing
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return m.isDir }
func (m mockDirEntry) Type() os.FileMode          { return 0 }
func (m mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }

type mockMetadataProvider struct {
	metadata Metadata
	err      error
}

func (m *mockMetadataProvider) GetMetadata() (Metadata, error) {
	return m.metadata, m.err
}

// mockAudioMetadataProvider for testing AudioMetadataProvider
type mockAudioMetadataProvider struct {
	metadata Metadata
	err      error
}

func (m *mockAudioMetadataProvider) GetMetadata() (Metadata, error) {
	return m.metadata, m.err
}

// For testing, we'll create a simple mock that looks like AudioMetadataProvider
// but uses our test data instead of reading real files
type mockAudioProvider struct {
	metadata Metadata
	err      error
}

func (m *mockAudioProvider) GetMetadata() (Metadata, error) {
	return m.metadata, m.err
}

// testUnifiedMetadataProvider wraps UnifiedMetadataProvider but delegates to test data
type testUnifiedMetadataProvider struct {
	UnifiedMetadataProvider // Embed to satisfy the type requirement
	testData                *mockAudioProvider
}

func (t *testUnifiedMetadataProvider) GetMetadata() (Metadata, error) {
	if t.testData != nil {
		return t.testData.GetMetadata()
	}
	// Fallback to embedded behavior if no test data
	return t.UnifiedMetadataProvider.GetMetadata()
}

// Note: readDirFunc is declared in album_detection.go
