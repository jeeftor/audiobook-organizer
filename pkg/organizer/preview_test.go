package organizer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

func TestGeneratePreview_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		inputDir  string
		outputDir string
		config    *OrganizerConfig
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "empty input directory",
			inputDir:  "",
			outputDir: "/output",
			config: &OrganizerConfig{
				FieldMapping: organizer.DefaultFieldMapping(),
			},
			wantErr: true,
			errMsg:  "input and output directories are required",
		},
		{
			name:      "empty output directory",
			inputDir:  "/input",
			outputDir: "",
			config: &OrganizerConfig{
				FieldMapping: organizer.DefaultFieldMapping(),
			},
			wantErr: true,
			errMsg:  "input and output directories are required",
		},
		{
			name:      "both empty",
			inputDir:  "",
			outputDir: "",
			config: &OrganizerConfig{
				FieldMapping: organizer.DefaultFieldMapping(),
			},
			wantErr: true,
			errMsg:  "input and output directories are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GeneratePreview(tt.inputDir, tt.outputDir, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePreview() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf(
						"GeneratePreview() error = %v, want error containing %s",
						err,
						tt.errMsg,
					)
				}
			}
		})
	}
}

func TestGeneratePreview_EmptyDirectory(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	config := &OrganizerConfig{
		Layout:       "author-series-title",
		FieldMapping: organizer.DefaultFieldMapping(),
	}

	result, err := GeneratePreview(inputDir, outputDir, config)
	if err != nil {
		t.Fatalf("GeneratePreview() unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("GeneratePreview() returned nil result")
	}

	if result.TotalFiles != 0 {
		t.Errorf("GeneratePreview() TotalFiles = %d, expected 0", result.TotalFiles)
	}

	if len(result.Moves) != 0 {
		t.Errorf("GeneratePreview() returned %d moves, expected 0", len(result.Moves))
	}

	if result.ConflictCount != 0 {
		t.Errorf("GeneratePreview() ConflictCount = %d, expected 0", result.ConflictCount)
	}
}

func TestCalculateTargetPaths(t *testing.T) {
	config := &OrganizerConfig{
		Layout:       "author-series-title",
		OutputDir:    "/output",
		FieldMapping: organizer.DefaultFieldMapping(),
	}

	metadata1 := organizer.NewMetadata()
	metadata1.Title = "Book One"
	metadata1.Authors = []string{"Author A"}
	metadata1.Series = []string{"Series X"}
	metadata1.SourcePath = "/input/book1/metadata.json"
	metadata1.SourceType = "json"

	metadata2 := organizer.NewMetadata()
	metadata2.Title = "Book Two"
	metadata2.Authors = []string{"Author B"}
	metadata2.Series = []string{"Series Y"}
	metadata2.SourcePath = "/input/book2/metadata.json"
	metadata2.SourceType = "json"

	audiobooks := []Metadata{metadata1, metadata2}

	moves, err := CalculateTargetPaths(audiobooks, config)
	if err != nil {
		t.Fatalf("CalculateTargetPaths() error: %v", err)
	}

	if len(moves) != 2 {
		t.Fatalf("CalculateTargetPaths() returned %d moves, expected 2", len(moves))
	}

	// Verify first move
	if moves[0].SourcePath != "/input/book1" {
		t.Errorf("Move 0 SourcePath = %s, expected /input/book1", moves[0].SourcePath)
	}

	// Verify target path contains author and series
	if !contains(moves[0].TargetPath, "Author A") {
		t.Errorf("Move 0 TargetPath missing author: %s", moves[0].TargetPath)
	}
	if !contains(moves[0].TargetPath, "Series X") {
		t.Errorf("Move 0 TargetPath missing series: %s", moves[0].TargetPath)
	}

	// Verify moves are not marked as conflicts initially
	if moves[0].IsConflict {
		t.Error("Move 0 should not be marked as conflict initially")
	}
	if moves[1].IsConflict {
		t.Error("Move 1 should not be marked as conflict initially")
	}
}

func TestDetectConflicts(t *testing.T) {
	// Create moves with intentional conflict
	moves := []PreviewMove{
		{
			SourcePath: "/input/book1",
			TargetPath: "/output/Author/Series/Book",
			IsConflict: false,
		},
		{
			SourcePath: "/input/book2",
			TargetPath: "/output/Author/Series/Book", // Same target = conflict
			IsConflict: false,
		},
		{
			SourcePath: "/input/book3",
			TargetPath: "/output/Other/Book",
			IsConflict: false,
		},
	}

	result := detectConflicts(moves)

	if result.TotalFiles != 3 {
		t.Errorf("TotalFiles = %d, expected 3", result.TotalFiles)
	}

	if result.ConflictCount != 2 {
		t.Errorf("ConflictCount = %d, expected 2 (both conflicting moves)", result.ConflictCount)
	}

	// First two moves should be marked as conflicts
	if !result.Moves[0].IsConflict {
		t.Error("Move 0 should be marked as conflict")
	}
	if !result.Moves[1].IsConflict {
		t.Error("Move 1 should be marked as conflict")
	}

	// Third move should not be conflict
	if result.Moves[2].IsConflict {
		t.Error("Move 2 should not be marked as conflict")
	}

	// Verify conflict reasons are set
	if result.Moves[0].ConflictReason == "" {
		t.Error("Move 0 conflict reason should be set")
	}
	if result.Moves[1].ConflictReason == "" {
		t.Error("Move 1 conflict reason should be set")
	}
}

func TestDetectConflicts_NoConflicts(t *testing.T) {
	// Create moves with no conflicts
	moves := []PreviewMove{
		{
			SourcePath: "/input/book1",
			TargetPath: "/output/Author1/Series1/Book1",
			IsConflict: false,
		},
		{
			SourcePath: "/input/book2",
			TargetPath: "/output/Author2/Series2/Book2",
			IsConflict: false,
		},
	}

	result := detectConflicts(moves)

	if result.ConflictCount != 0 {
		t.Errorf("ConflictCount = %d, expected 0", result.ConflictCount)
	}

	for i, move := range result.Moves {
		if move.IsConflict {
			t.Errorf("Move %d should not be marked as conflict", i)
		}
	}
}

func TestFilterPreviewByIndices(t *testing.T) {
	preview := &PreviewResult{
		Moves: []PreviewMove{
			{SourcePath: "/input/book1", TargetPath: "/output/book1", IsConflict: false},
			{SourcePath: "/input/book2", TargetPath: "/output/book2", IsConflict: true},
			{SourcePath: "/input/book3", TargetPath: "/output/book3", IsConflict: false},
			{SourcePath: "/input/book4", TargetPath: "/output/book4", IsConflict: true},
		},
		TotalFiles:    4,
		ConflictCount: 2,
	}

	tests := []struct {
		name              string
		selectedIndices   []int
		expectedMoves     int
		expectedConflicts int
	}{
		{
			name:              "select all",
			selectedIndices:   []int{0, 1, 2, 3},
			expectedMoves:     4,
			expectedConflicts: 2,
		},
		{
			name:              "select first two",
			selectedIndices:   []int{0, 1},
			expectedMoves:     2,
			expectedConflicts: 1,
		},
		{
			name:              "select non-conflicts",
			selectedIndices:   []int{0, 2},
			expectedMoves:     2,
			expectedConflicts: 0,
		},
		{
			name:              "select conflicts only",
			selectedIndices:   []int{1, 3},
			expectedMoves:     2,
			expectedConflicts: 2,
		},
		{
			name:              "empty selection returns original",
			selectedIndices:   []int{},
			expectedMoves:     4,
			expectedConflicts: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterPreviewByIndices(preview, tt.selectedIndices)

			if filtered.TotalFiles != tt.expectedMoves {
				t.Errorf("TotalFiles = %d, expected %d", filtered.TotalFiles, tt.expectedMoves)
			}

			if len(filtered.Moves) != tt.expectedMoves {
				t.Errorf("len(Moves) = %d, expected %d", len(filtered.Moves), tt.expectedMoves)
			}

			if filtered.ConflictCount != tt.expectedConflicts {
				t.Errorf(
					"ConflictCount = %d, expected %d",
					filtered.ConflictCount,
					tt.expectedConflicts,
				)
			}
		})
	}
}

func TestFormatPreviewMove(t *testing.T) {
	tests := []struct {
		name           string
		move           PreviewMove
		expectContains []string
	}{
		{
			name: "normal move",
			move: PreviewMove{
				SourcePath: "/input/book",
				TargetPath: "/output/Author/Book",
				IsConflict: false,
			},
			expectContains: []string{"✓", "/input/book", "→", "/output/Author/Book"},
		},
		{
			name: "conflict move",
			move: PreviewMove{
				SourcePath:     "/input/book1",
				TargetPath:     "/output/Author/Book",
				IsConflict:     true,
				ConflictReason: "Multiple books targeting same path",
			},
			expectContains: []string{
				"⚠️",
				"/input/book1",
				"→",
				"/output/Author/Book",
				"Conflict:",
				"Multiple books",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPreviewMove(tt.move)

			for _, expected := range tt.expectContains {
				if !contains(result, expected) {
					t.Errorf("FormatPreviewMove() result missing '%s':\n%s", expected, result)
				}
			}
		})
	}
}

func TestGeneratePreview_Integration(t *testing.T) {
	// Create temporary input and output directories
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create a simple metadata.json file
	bookDir := filepath.Join(inputDir, "test-book")
	os.MkdirAll(bookDir, 0o755)

	metadataContent := `{
		"title": "Test Book",
		"authors": ["Test Author"],
		"series": ["Test Series"]
	}`
	metadataPath := filepath.Join(bookDir, "metadata.json")
	os.WriteFile(metadataPath, []byte(metadataContent), 0o644)

	config := &OrganizerConfig{
		Layout:       "author-series-title",
		FieldMapping: organizer.DefaultFieldMapping(),
	}

	result, err := GeneratePreview(inputDir, outputDir, config)
	if err != nil {
		t.Fatalf("GeneratePreview() error: %v", err)
	}

	if result.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, expected 1", result.TotalFiles)
	}

	if len(result.Moves) != 1 {
		t.Fatalf("Expected 1 move, got %d", len(result.Moves))
	}

	move := result.Moves[0]

	// Verify source path
	if !contains(move.SourcePath, "test-book") {
		t.Errorf("SourcePath should contain 'test-book': %s", move.SourcePath)
	}

	// Verify target path contains expected components
	if !contains(move.TargetPath, "Test Author") {
		t.Errorf("TargetPath missing author: %s", move.TargetPath)
	}
	if !contains(move.TargetPath, "Test Series") {
		t.Errorf("TargetPath missing series: %s", move.TargetPath)
	}
	if !contains(move.TargetPath, "Test Book") {
		t.Errorf("TargetPath missing title: %s", move.TargetPath)
	}
}
