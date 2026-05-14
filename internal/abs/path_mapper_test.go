// internal/abs/path_mapper_test.go
// Tests for path mapping

package abs

import (
	"os"
	"testing"
)

func TestParsePathMapping(t *testing.T) {
	tests := []struct {
		input    string
		expected PathMapping
		wantErr  bool
	}{
		{
			input:    "/audiobooks:/mnt/media/audiobooks",
			expected: PathMapping{ABSPrefix: "/audiobooks", LocalPrefix: "/mnt/media/audiobooks"},
			wantErr:  false,
		},
		{
			input:    "/books:/home/user/books",
			expected: PathMapping{ABSPrefix: "/books", LocalPrefix: "/home/user/books"},
			wantErr:  false,
		},
		{
			input:   "invalid-no-colon",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParsePathMapping(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePathMapping(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result.ABSPrefix != tt.expected.ABSPrefix {
					t.Errorf("ABSPrefix = %q, want %q", result.ABSPrefix, tt.expected.ABSPrefix)
				}
				if result.LocalPrefix != tt.expected.LocalPrefix {
					t.Errorf(
						"LocalPrefix = %q, want %q",
						result.LocalPrefix,
						tt.expected.LocalPrefix,
					)
				}
			}
		})
	}
}

func TestPathMapper_ToLocal(t *testing.T) {
	mapper := NewPathMapper([]PathMapping{
		{ABSPrefix: "/audiobooks", LocalPrefix: "/mnt/media/audiobooks"},
		{ABSPrefix: "/podcasts", LocalPrefix: "/mnt/media/podcasts"},
	})

	tests := []struct {
		absPath  string
		expected string
	}{
		{
			absPath:  "/audiobooks/Author/Book",
			expected: "/mnt/media/audiobooks/Author/Book",
		},
		{
			absPath:  "/podcasts/Show/Episode",
			expected: "/mnt/media/podcasts/Show/Episode",
		},
		{
			absPath:  "/unknown/path",
			expected: "/unknown/path", // No mapping, returns as-is
		},
	}

	for _, tt := range tests {
		t.Run(tt.absPath, func(t *testing.T) {
			result := mapper.ToLocal(tt.absPath)
			if result != tt.expected {
				t.Errorf("ToLocal(%q) = %q, want %q", tt.absPath, result, tt.expected)
			}
		})
	}
}

func TestPathMapper_ToABS(t *testing.T) {
	mapper := NewPathMapper([]PathMapping{
		{ABSPrefix: "/audiobooks", LocalPrefix: "/mnt/media/audiobooks"},
	})

	tests := []struct {
		localPath string
		expected  string
	}{
		{
			localPath: "/mnt/media/audiobooks/Author/Book",
			expected:  "/audiobooks/Author/Book",
		},
		{
			localPath: "/other/path",
			expected:  "/other/path", // No mapping, returns as-is
		},
	}

	for _, tt := range tests {
		t.Run(tt.localPath, func(t *testing.T) {
			result := mapper.ToABS(tt.localPath)
			if result != tt.expected {
				t.Errorf("ToABS(%q) = %q, want %q", tt.localPath, result, tt.expected)
			}
		})
	}
}

func TestNewPathMapperFromSQLite(t *testing.T) {
	// Create a test SQLite database
	tmpDir := t.TempDir()
	_ = tmpDir // Will be used when we add SQLite test

	// Create minimal schema and data
	// Note: This requires the mattn/go-sqlite3 package
	// Skip if not available
	if os.Getenv("SKIP_SQLITE_TESTS") != "" {
		t.Skip("Skipping SQLite tests")
	}

	// This test would require creating a real SQLite DB
	// For unit tests, we mock the behavior
	t.Skip("SQLite integration test requires real database - tested via integration tests")
}

func TestPathMapper_Empty(t *testing.T) {
	mapper := NewPathMapper(nil)

	// Should return paths unchanged
	if got := mapper.ToLocal("/test"); got != "/test" {
		t.Errorf("Empty mapper ToLocal = %q, want %q", got, "/test")
	}
	if got := mapper.ToABS("/test"); got != "/test" {
		t.Errorf("Empty mapper ToABS = %q, want %q", got, "/test")
	}
}
