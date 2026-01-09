package organizer

import (
	"testing"
)

func TestConvertToFirstLast(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "already first last",
			input: "Brandon Sanderson",
			want:  "Brandon Sanderson",
		},
		{
			name:  "last comma first",
			input: "Sanderson, Brandon",
			want:  "Brandon Sanderson",
		},
		{
			name:  "multi-word first name",
			input: "von Neumann, John",
			want:  "John von Neumann",
		},
		{
			name:  "single name",
			input: "Cicero",
			want:  "Cicero",
		},
		{
			name:  "multiple commas",
			input: "Last, First, Middle",
			want:  "Last, First, Middle", // Can't convert, return as-is
		},
		{
			name:  "whitespace handling",
			input: "Sanderson , Brandon",
			want:  "Brandon Sanderson",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToFirstLast(tt.input)
			if got != tt.want {
				t.Errorf("ConvertToFirstLast(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvertToLastFirst(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "first last",
			input: "Brandon Sanderson",
			want:  "Sanderson, Brandon",
		},
		{
			name:  "already last first",
			input: "Sanderson, Brandon",
			want:  "Sanderson, Brandon",
		},
		{
			name:  "multi-word first name",
			input: "Brandon von Sanderson",
			want:  "Sanderson, Brandon von",
		},
		{
			name:  "single name",
			input: "Cicero",
			want:  "Cicero",
		},
		{
			name:  "three part name",
			input: "Jean-Luc Picard",
			want:  "Picard, Jean-Luc",
		},
		{
			name:  "four part name",
			input: "Martin Luther King Jr",
			want:  "Jr, Martin Luther King",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToLastFirst(tt.input)
			if got != tt.want {
				t.Errorf("ConvertToLastFirst(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  AuthorFormat
	}{
		{
			name:  "first last detected",
			input: "Brandon Sanderson",
			want:  AuthorFormatFirstLast,
		},
		{
			name:  "last first detected",
			input: "Sanderson, Brandon",
			want:  AuthorFormatLastFirst,
		},
		{
			name:  "single name",
			input: "Cicero",
			want:  AuthorFormatFirstLast,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFormat(tt.input)
			if got != tt.want {
				t.Errorf("DetectFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAuthorFormatter_FormatAuthor(t *testing.T) {
	tests := []struct {
		name   string
		format AuthorFormat
		input  string
		want   string
	}{
		{
			name:   "first last - already correct",
			format: AuthorFormatFirstLast,
			input:  "Brandon Sanderson",
			want:   "Brandon Sanderson",
		},
		{
			name:   "first last - convert from last first",
			format: AuthorFormatFirstLast,
			input:  "Sanderson, Brandon",
			want:   "Brandon Sanderson",
		},
		{
			name:   "last first - already correct",
			format: AuthorFormatLastFirst,
			input:  "Sanderson, Brandon",
			want:   "Sanderson, Brandon",
		},
		{
			name:   "last first - convert from first last",
			format: AuthorFormatLastFirst,
			input:  "Brandon Sanderson",
			want:   "Sanderson, Brandon",
		},
		{
			name:   "preserve - no change",
			format: AuthorFormatPreserve,
			input:  "Brandon Sanderson",
			want:   "Brandon Sanderson",
		},
		{
			name:   "preserve - keeps comma format",
			format: AuthorFormatPreserve,
			input:  "Sanderson, Brandon",
			want:   "Sanderson, Brandon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewAuthorFormatter(tt.format)
			got := formatter.FormatAuthor(tt.input)
			if got != tt.want {
				t.Errorf("FormatAuthor(%q) with format %v = %q, want %q", tt.input, tt.format, got, tt.want)
			}
		})
	}
}

func TestNewAuthorFormatter(t *testing.T) {
	tests := []struct {
		name   string
		format AuthorFormat
	}{
		{
			name:   "first last formatter",
			format: AuthorFormatFirstLast,
		},
		{
			name:   "last first formatter",
			format: AuthorFormatLastFirst,
		},
		{
			name:   "preserve formatter",
			format: AuthorFormatPreserve,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewAuthorFormatter(tt.format)
			if formatter == nil {
				t.Fatal("NewAuthorFormatter() returned nil")
			}
			if formatter.format != tt.format {
				t.Errorf("NewAuthorFormatter() format = %v, want %v", formatter.format, tt.format)
			}
		})
	}
}
