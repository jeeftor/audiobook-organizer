package organizer

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// PreviewMove represents a single file/directory move operation
type PreviewMove struct {
	SourcePath      string `json:"source_path"`
	TargetPath      string `json:"target_path"`
	IsConflict      bool   `json:"is_conflict"`
	ConflictReason  string `json:"conflict_reason,omitempty"`
	Metadata        Metadata `json:"metadata,omitempty"`
}

// PreviewResult contains the complete preview of all operations
type PreviewResult struct {
	Moves          []PreviewMove `json:"moves"`
	TotalFiles     int           `json:"total_files"`
	ConflictCount  int           `json:"conflict_count"`
}

// GeneratePreview scans directories and generates a preview of file operations
// that would be performed during organization using the actual organizer logic.
//
// Parameters:
//   - inputDir: Source directory containing audiobooks
//   - outputDir: Target directory for organized audiobooks
//   - config: Organizer configuration
//
// Returns:
//   - PreviewResult: Complete preview with all moves and conflict information
//   - error: Any error encountered during preview generation
func GeneratePreview(inputDir, outputDir string, config *OrganizerConfig) (*PreviewResult, error) {
	if inputDir == "" || outputDir == "" {
		return nil, fmt.Errorf("input and output directories are required")
	}

	// Configure for preview mode (dry-run)
	previewConfig := *config
	previewConfig.BaseDir = inputDir
	previewConfig.OutputDir = outputDir
	previewConfig.DryRun = true
	previewConfig.Verbose = false // Disable verbose output for GUI

	// Create organizer using the actual organizer implementation
	org, err := NewOrganizer(&previewConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating organizer: %w", err)
	}

	// Execute in dry-run mode to get the actual moves that would happen
	err = org.Execute()
	if err != nil {
		return nil, fmt.Errorf("error generating preview: %w", err)
	}

	// Get the summary from the organizer
	summary := org.GetSummary()

	// Convert the summary moves to PreviewMoves
	var moves []PreviewMove
	for _, move := range summary.Moves {
		moves = append(moves, PreviewMove{
			SourcePath: move.From,
			TargetPath: move.To,
			IsConflict: false, // Will be calculated below
		})
	}

	// Detect conflicts
	result := detectConflicts(moves)

	return result, nil
}

// CalculateTargetPaths calculates the target path for each audiobook based on configuration
//
// Parameters:
//   - audiobooks: List of audiobooks with metadata
//   - config: Organizer configuration (determines layout pattern)
//
// Returns:
//   - []PreviewMove: List of moves with source and target paths
//   - error: Any error encountered during path calculation
func CalculateTargetPaths(audiobooks []Metadata, config *OrganizerConfig) ([]PreviewMove, error) {
	var moves []PreviewMove

	// Create layout calculator with sanitizer
	sanitizer := CreateSanitizerFunc(config)
	layoutCalc := organizer.NewLayoutCalculator(config, sanitizer)

	for _, metadata := range audiobooks {
		// Calculate target directory path using layout calculator
		targetDirPath := layoutCalc.CalculateTargetPath(metadata)

		// Determine source and target paths based on mode
		var sourcePath string
		var targetPath string

		if config.Flat && metadata.SourceType == "audio" {
			// Flat mode with audio files: each file is moved individually
			sourcePath = metadata.SourcePath
			filename := filepath.Base(metadata.SourcePath)
			targetPath = filepath.Join(targetDirPath, filename)
		} else if metadata.SourceType == "json" || metadata.SourceType == "epub" {
			// Metadata files: move the entire directory
			sourcePath = filepath.Dir(metadata.SourcePath)
			targetPath = targetDirPath
		} else {
			// Non-flat mode audio or other: treat as directory
			sourcePath = filepath.Dir(metadata.SourcePath)
			targetPath = targetDirPath
		}

		move := PreviewMove{
			SourcePath: sourcePath,
			TargetPath: targetPath,
			IsConflict: false,
			Metadata:   metadata,
		}

		moves = append(moves, move)
	}

	return moves, nil
}

// detectConflicts identifies when multiple audiobooks would be moved to the same location
func detectConflicts(moves []PreviewMove) *PreviewResult {
	result := &PreviewResult{
		Moves:      moves,
		TotalFiles: len(moves),
	}

	// Build map of target paths to detect duplicates
	targetMap := make(map[string][]int)
	for i, move := range moves {
		cleanTarget := filepath.Clean(move.TargetPath)
		targetMap[cleanTarget] = append(targetMap[cleanTarget], i)
	}

	// Mark conflicts
	for targetPath, indices := range targetMap {
		if len(indices) > 1 {
			// Multiple sources targeting the same destination
			for _, idx := range indices {
				result.Moves[idx].IsConflict = true
				result.Moves[idx].ConflictReason = fmt.Sprintf("Multiple books (%d) would be moved to: %s", len(indices), targetPath)
				result.ConflictCount++
			}
		}
	}

	return result
}

// FilterPreviewByIndices filters preview moves to only include selected indices
// This is useful when the GUI allows users to select which books to organize
//
// Parameters:
//   - preview: Complete preview result
//   - selectedIndices: List of indices to include (0-based)
//
// Returns:
//   - *PreviewResult: Filtered preview with only selected moves
func FilterPreviewByIndices(preview *PreviewResult, selectedIndices []int) *PreviewResult {
	if len(selectedIndices) == 0 {
		return preview
	}

	// Create set of selected indices for fast lookup
	selectedSet := make(map[int]bool)
	for _, idx := range selectedIndices {
		selectedSet[idx] = true
	}

	// Filter moves
	filtered := &PreviewResult{
		Moves: make([]PreviewMove, 0, len(selectedIndices)),
	}

	for i, move := range preview.Moves {
		if selectedSet[i] {
			filtered.Moves = append(filtered.Moves, move)
			if move.IsConflict {
				filtered.ConflictCount++
			}
		}
	}

	filtered.TotalFiles = len(filtered.Moves)

	return filtered
}

// FormatPreviewMove returns a human-readable string representation of a move
func FormatPreviewMove(move PreviewMove) string {
	var b strings.Builder

	if move.IsConflict {
		b.WriteString("⚠️  ")
	} else {
		b.WriteString("✓  ")
	}

	b.WriteString(move.SourcePath)
	b.WriteString(" → ")
	b.WriteString(move.TargetPath)

	if move.IsConflict {
		b.WriteString("\n   Conflict: ")
		b.WriteString(move.ConflictReason)
	}

	return b.String()
}
