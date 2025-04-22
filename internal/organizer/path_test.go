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
			want:         "Book: Title? (Part 1) <Test> |Series| *Special*",
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
			got := cleanSeriesName(tt.input)
			if got != tt.want {
				t.Errorf("cleanSeriesName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizePathWithProblematicMetadata(t *testing.T) {
	// Test cases for problematic metadata
	tests := []struct {
		name     string
		input    string
		expected string
		replace  string
	}{
		// Invalid characters
		{"colons", "The Book: With Colons?", "The Book_ With Colons_", ""},
		{"slashes", "Series/With/Slashes", "Series_With_Slashes", ""},
		{"invalid_chars", "Author*With|Invalid<Characters>", "Author_With_Invalid_Characters_", ""},

		// Symbols
		{"symbols", "Book & Symbols % $ # @ !", "Book & Symbols % $ # @ !", ""},
		{"trademark", "Seriesu2122 u00a92025", "Seriesu2122 u00a92025", ""},

		// Non-ASCII
		{"accents", "Cafu00e9 au lait", "Cafu00e9 au lait", ""},
		{"unicode", "Ru00e9sumu00e9 of u00c5ngstru00f6m", "Ru00e9sumu00e9 of u00c5ngstru00f6m", ""},

		// Spaces
		{"spaces", "  Book  With  Many    Spaces  ", "Book  With  Many    Spaces", ""},
		{"spaces_replaced", "Book With Spaces", "Book_With_Spaces", "_"},
		{"spaces_dot", "Book With Spaces", "Book.With.Spaces", "."},

		// Quotes and backslashes
		{"quotes", "Book \"Quoted\" Title", "Book _Quoted_ Title", ""},
		{"backslashes", "Series\\With\\Backslashes", "Series\\With\\Backslashes", ""},

		// Dots
		{"dots", "Book.With.Dots", "Book.With.Dots", ""},

		// Emoji
		{"emoji", "Book With Emoji ud83dudcdaud83dudd0d", "Book With Emoji ud83dudcdaud83dudd0d", ""},

		// HTML
		{"html", "Book With HTML <b>Tags</b>", "Book With HTML _b_Tags__b_", ""},
		{"script", "Author With <script>alert(\"XSS\")</script>", "Author With _script_alert(_XSS_)__script_", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create organizer with test config
			config := &OrganizerConfig{
				ReplaceSpace: tt.replace,
			}
			org := &Organizer{config: *config}

			// Test sanitization
			result := org.SanitizePath(tt.input)

			// Check if result matches expected output
			if result != tt.expected {
				t.Errorf("SanitizePath(%q) with replace_space=%q = %q; want %q",
					tt.input, tt.replace, result, tt.expected)
			}
		})
	}
}
