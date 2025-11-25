package organizer

import (
	"runtime"
	"testing"
)

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		replaceSpace string
		os           string
		want         string
	}{
		// Windows-specific tests
		{
			name:         "windows_invalid_chars",
			input:        "Book: Title? (Part 1) <Test> |Series| *Special*",
			replaceSpace: "",
			os:           "windows",
			want:         "Book_ Title_ (Part 1) _Test_ _Series_ _Special_",
		},
		{
			name:         "windows_with_space_replacement",
			input:        `Test: Book \ Series / Part`,
			replaceSpace: ".",
			os:           "windows",
			want:         "Test_.Book._.Series._.Part",
		},
		{
			name:         "windows_file_extension",
			input:        "Test.mp3",
			replaceSpace: "",
			os:           "windows",
			want:         "Test.mp3",
		},
		// Unix-specific tests
		{
			name:         "unix_invalid_chars",
			input:        "Book: Title? (Part 1) <Test> |Series| *Special*",
			replaceSpace: "",
			os:           "linux",
			want:         "Book_ Title_ (Part 1) _Test_ _Series_ _Special",
		},
		{
			name:         "unix_with_forward_slash",
			input:        "Test/Book/Series",
			replaceSpace: "",
			os:           "linux",
			want:         "Test_Book_Series",
		},
		{
			name:         "unix_with_space_replacement",
			input:        "Test Book Series",
			replaceSpace: ".",
			os:           "linux",
			want:         "Test.Book.Series",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that don't match current OS unless we're running all tests
			if tt.os != runtime.GOOS && tt.os != "" {
				t.Skipf("Skipping %s test on %s", tt.os, runtime.GOOS)
			}

			config := &OrganizerConfig{
				BaseDir:      "",
				OutputDir:    "",
				ReplaceSpace: tt.replaceSpace,
				Verbose:      false,
				DryRun:       false,
				Undo:         false,
				Prompt:       false,
				RemoveEmpty:  false,
			}
			org := NewOrganizer(config)

			got := org.SanitizePath(tt.input)
			if got != tt.want {
				t.Errorf("sanitizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanSeriesName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "series_with_number",
			input: "Test Series #1",
			want:  "Test Series",
		},
		{
			name:  "series_with_complex_number",
			input: "Test Series Part 1 #12",
			want:  "Test Series Part 1",
		},
		{
			name:  "series_without_number",
			input: "Test Series",
			want:  "Test Series",
		},
		{
			name:  "multiple_hash_symbols",
			input: "Test #Series Part 1 #12",
			want:  "Test #Series Part 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanSeriesName(tt.input)
			if got != tt.want {
				t.Errorf("CleanSeriesName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizePathWithProblematicMetadata(t *testing.T) {
	config := &OrganizerConfig{}
	org := NewOrganizer(config)
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "colons",
			input: "The Book: With Colons?",
			want:  org.SanitizePath("The Book: With Colons?"),
		},
		{
			name:  "slashes",
			input: "Series/With/Slashes",
			want:  org.SanitizePath("Series/With/Slashes"),
		},
		{
			name:  "invalid_chars",
			input: "Author*With|Invalid<Characters>",
			want:  org.SanitizePath("Author*With|Invalid<Characters>"),
		},
		{
			name:  "quotes",
			input: "Book \"Quoted\" Title",
			want:  org.SanitizePath("Book \"Quoted\" Title"),
		},
		{
			name:  "html",
			input: "Book With HTML <b>Tags</b>",
			want:  org.SanitizePath("Book With HTML <b>Tags</b>"),
		},
		{
			name:  "script",
			input: "Author With <script>alert(\"XSS\")</script>",
			want:  org.SanitizePath("Author With <script>alert(\"XSS\")</script>"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := org.SanitizePath(tt.input)
			if got != tt.want {
				t.Errorf("SanitizePath(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}
