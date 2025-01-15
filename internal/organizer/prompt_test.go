package organizer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPromptConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     bool
		wantMove bool
	}{
		{
			name:     "confirm_with_y",
			input:    "y\n",
			want:     true,
			wantMove: true,
		},
		{
			name:     "confirm_with_Y",
			input:    "Y\n",
			want:     true,
			wantMove: true,
		},
		{
			name:     "confirm_with_yes",
			input:    "yes\n",
			want:     true,
			wantMove: true,
		},
		{
			name:     "confirm_with_YES",
			input:    "YES\n",
			want:     true,
			wantMove: true,
		},
		{
			name:     "deny_with_n",
			input:    "n\n",
			want:     false,
			wantMove: false,
		},
		{
			name:     "deny_with_empty",
			input:    "\n",
			want:     false,
			wantMove: false,
		},
		{
			name:     "deny_with_invalid",
			input:    "invalid\n",
			want:     false,
			wantMove: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			sourceDir := filepath.Join(tempDir, "source")
			if err := os.MkdirAll(sourceDir, 0755); err != nil {
				t.Fatal(err)
			}

			// Create a temporary file for input
			inputFile, err := os.CreateTemp("", "input")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(inputFile.Name())
			defer inputFile.Close()

			// Write the test input
			if _, err := inputFile.WriteString(tt.input); err != nil {
				t.Fatal(err)
			}
			if _, err := inputFile.Seek(0, 0); err != nil {
				t.Fatal(err)
			}

			// Replace stdin
			oldStdin := os.Stdin
			os.Stdin = inputFile
			defer func() { os.Stdin = oldStdin }()

			metadata := Metadata{
				Authors: []string{"Test Author"},
				Title:   "Test Book",
			}

			org := New(
				tempDir,
				"",    // outputDir
				"",    // replaceSpace
				false, // verbose
				false, // dryRun
				true,  // prompt
				false, // undo
			)

			got := org.PromptForConfirmation(metadata, sourceDir, filepath.Join(tempDir, "Test Author/Test Book"))
			if got != tt.want {
				t.Errorf("PromptForConfirmation() = %v, want %v", got, tt.want)
			}
		})
	}
}
