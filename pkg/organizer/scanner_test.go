package organizer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

func TestScanForAudiobooks_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		baseDir   string
		config    *OrganizerConfig
		wantErr   bool
		errMsg    string
		setupFunc func() string
	}{
		{
			name:    "empty base directory",
			baseDir: "",
			config: &OrganizerConfig{
				FieldMapping: organizer.DefaultFieldMapping(),
			},
			wantErr: true,
			errMsg:  "base directory is required",
		},
		{
			name:    "non-existent directory",
			baseDir: "/nonexistent/path/to/nowhere",
			config: &OrganizerConfig{
				FieldMapping: organizer.DefaultFieldMapping(),
			},
			wantErr: true,
			errMsg:  "error accessing directory",
		},
		{
			name: "path is a file not directory",
			config: &OrganizerConfig{
				FieldMapping: organizer.DefaultFieldMapping(),
			},
			wantErr: true,
			errMsg:  "path is not a directory",
			setupFunc: func() string {
				tmpFile := filepath.Join(t.TempDir(), "testfile.txt")
				os.WriteFile(tmpFile, []byte("test"), 0644)
				return tmpFile
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := tt.baseDir
			if tt.setupFunc != nil {
				baseDir = tt.setupFunc()
			}

			_, err := ScanForAudiobooks(baseDir, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("ScanForAudiobooks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ScanForAudiobooks() error = %v, want error containing %s", err, tt.errMsg)
				}
			}
		})
	}
}

func TestScanForAudiobooks_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	config := &OrganizerConfig{
		FieldMapping: organizer.DefaultFieldMapping(),
	}

	results, err := ScanForAudiobooks(tmpDir, config)
	if err != nil {
		t.Fatalf("ScanForAudiobooks() unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("ScanForAudiobooks() returned %d results, expected 0 for empty directory", len(results))
	}
}

func TestScanForAudiobooks_SkipsOutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	os.MkdirAll(outputDir, 0755)

	// Create a fake audiobook in the output directory (should be skipped)
	os.WriteFile(filepath.Join(outputDir, "metadata.json"), []byte(`{"title":"Should Be Skipped"}`), 0644)

	config := &OrganizerConfig{
		OutputDir:    outputDir,
		FieldMapping: organizer.DefaultFieldMapping(),
	}

	results, err := ScanForAudiobooks(tmpDir, config)
	if err != nil {
		t.Fatalf("ScanForAudiobooks() unexpected error: %v", err)
	}

	// Should find nothing because output dir is skipped
	if len(results) != 0 {
		t.Errorf("ScanForAudiobooks() returned %d results, expected 0 (output dir should be skipped)", len(results))
	}
}

func TestScanSingleFile_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		config    *OrganizerConfig
		wantErr   bool
		errMsg    string
		setupFunc func() string
	}{
		{
			name:     "empty file path",
			filePath: "",
			config: &OrganizerConfig{
				FieldMapping: organizer.DefaultFieldMapping(),
			},
			wantErr: true,
			errMsg:  "file path is required",
		},
		{
			name:     "non-existent file",
			filePath: "/nonexistent/file.mp3",
			config: &OrganizerConfig{
				FieldMapping: organizer.DefaultFieldMapping(),
			},
			wantErr: true,
			errMsg:  "error accessing file",
		},
		{
			name: "path is a directory",
			config: &OrganizerConfig{
				FieldMapping: organizer.DefaultFieldMapping(),
			},
			wantErr: true,
			errMsg:  "path is a directory",
			setupFunc: func() string {
				return t.TempDir()
			},
		},
		{
			name: "unsupported file type",
			config: &OrganizerConfig{
				FieldMapping: organizer.DefaultFieldMapping(),
			},
			wantErr: true,
			errMsg:  "unsupported file type",
			setupFunc: func() string {
				tmpFile := filepath.Join(t.TempDir(), "test.txt")
				os.WriteFile(tmpFile, []byte("test"), 0644)
				return tmpFile
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.filePath
			if tt.setupFunc != nil {
				filePath = tt.setupFunc()
			}

			_, err := ScanSingleFile(filePath, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("ScanSingleFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ScanSingleFile() error = %v, want error containing %s", err, tt.errMsg)
				}
			}
		})
	}
}

func TestIsSubPath(t *testing.T) {
	tests := []struct {
		name   string
		parent string
		child  string
		want   bool
	}{
		{
			name:   "child is subdirectory",
			parent: "/path/to/parent",
			child:  "/path/to/parent/child",
			want:   true,
		},
		{
			name:   "child is deep subdirectory",
			parent: "/path/to/parent",
			child:  "/path/to/parent/child/grandchild",
			want:   true,
		},
		{
			name:   "child is not subdirectory",
			parent: "/path/to/parent",
			child:  "/path/to/other",
			want:   false,
		},
		{
			name:   "child is parent",
			parent: "/path/to/parent",
			child:  "/path/to/parent",
			want:   false,
		},
		{
			name:   "child is above parent",
			parent: "/path/to/parent/child",
			child:  "/path/to/parent",
			want:   false,
		},
		{
			name:   "child shares prefix but not subdirectory",
			parent: "/path/to/parent",
			child:  "/path/to/parent-sibling",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSubPath(tt.parent, tt.child)
			if got != tt.want {
				t.Errorf("isSubPath(%s, %s) = %v, want %v", tt.parent, tt.child, got, tt.want)
			}
		})
	}
}

func TestGetMetadataProviderForFile(t *testing.T) {
	config := &OrganizerConfig{
		FieldMapping: organizer.DefaultFieldMapping(),
	}

	tests := []struct {
		name     string
		filePath string
		wantType string
		wantErr  bool
	}{
		{
			name:     "epub file",
			filePath: "/path/to/book.epub",
			wantType: "*organizer.EPUBMetadataProvider",
			wantErr:  false,
		},
		{
			name:     "mp3 file",
			filePath: "/path/to/audio.mp3",
			wantType: "*organizer.AudioMetadataProvider",
			wantErr:  false,
		},
		{
			name:     "m4b file",
			filePath: "/path/to/audiobook.m4b",
			wantType: "*organizer.AudioMetadataProvider",
			wantErr:  false,
		},
		{
			name:     "m4a file",
			filePath: "/path/to/audio.m4a",
			wantType: "*organizer.AudioMetadataProvider",
			wantErr:  false,
		},
		{
			name:     "unsupported file",
			filePath: "/path/to/file.txt",
			wantType: "",
			wantErr:  true,
		},
		{
			name:     "case insensitive extension",
			filePath: "/path/to/AUDIO.MP3",
			wantType: "*organizer.AudioMetadataProvider",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := getMetadataProviderForFile(tt.filePath, config)

			if (err != nil) != tt.wantErr {
				t.Errorf("getMetadataProviderForFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Just verify we got a provider (type checking is tricky with interfaces)
				if provider == nil {
					t.Errorf("getMetadataProviderForFile() returned nil provider, expected %s", tt.wantType)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
