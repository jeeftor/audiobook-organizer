// internal/organizer/plan_script.go
package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

// PlannedMove represents a single file move operation
type PlannedMove struct {
	SourcePath string
	TargetPath string
	Metadata   *Metadata // Optional metadata for documentation
}

// PlanScriptWriter handles writing dry-run plans to executable shell scripts
type PlanScriptWriter struct {
	scriptPath string
	moves      []PlannedMove
	baseDir    string
	outputDir  string
}

// NewPlanScriptWriter creates a new plan script writer
func NewPlanScriptWriter(scriptPath, baseDir, outputDir string) *PlanScriptWriter {
	return &PlanScriptWriter{
		scriptPath: scriptPath,
		moves:      []PlannedMove{},
		baseDir:    baseDir,
		outputDir:  outputDir,
	}
}

// AddMove adds a planned move to the script
func (p *PlanScriptWriter) AddMove(source, target string, metadata *Metadata) {
	p.moves = append(p.moves, PlannedMove{
		SourcePath: source,
		TargetPath: target,
		Metadata:   metadata,
	})
}

// WriteScript writes the plan to an executable shell script
func (p *PlanScriptWriter) WriteScript() error {
	if p.scriptPath == "" {
		return nil
	}

	file, err := os.Create(p.scriptPath)
	if err != nil {
		return fmt.Errorf("failed to create plan script: %w", err)
	}
	defer file.Close()

	// Write shebang and header
	fmt.Fprintf(file, "#!/bin/bash\n")
	fmt.Fprintf(file, "#\n")
	fmt.Fprintf(file, "# Audiobook Organizer - Move Plan Script\n")
	fmt.Fprintf(file, "# Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "# Source Directory: %s\n", p.baseDir)
	if p.outputDir != "" {
		fmt.Fprintf(file, "# Output Directory: %s\n", p.outputDir)
	}
	fmt.Fprintf(file, "# Total Moves: %d\n", len(p.moves))
	fmt.Fprintf(file, "#\n")
	fmt.Fprintf(file, "# Review this script before running!\n")
	fmt.Fprintf(file, "# Run with: bash %s\n", filepath.Base(p.scriptPath))
	fmt.Fprintf(file, "#\n\n")

	// Add safety options
	fmt.Fprintf(file, "set -e  # Exit on error\n")
	fmt.Fprintf(file, "set -u  # Exit on undefined variable\n\n")

	// Add dry-run option
	fmt.Fprintf(file, "# Set DRY_RUN=1 to preview without moving\n")
	fmt.Fprintf(file, "DRY_RUN=${DRY_RUN:-0}\n\n")

	// Add move function
	fmt.Fprintf(file, "move_file() {\n")
	fmt.Fprintf(file, "    local src=\"$1\"\n")
	fmt.Fprintf(file, "    local dst=\"$2\"\n")
	fmt.Fprintf(file, "    local dst_dir\n")
	fmt.Fprintf(file, "    dst_dir=$(dirname \"$dst\")\n")
	fmt.Fprintf(file, "    \n")
	fmt.Fprintf(file, "    if [[ $DRY_RUN -eq 1 ]]; then\n")
	fmt.Fprintf(file, "        echo \"[DRY-RUN] Would move: $src -> $dst\"\n")
	fmt.Fprintf(file, "    else\n")
	fmt.Fprintf(file, "        mkdir -p \"$dst_dir\"\n")
	fmt.Fprintf(file, "        mv \"$src\" \"$dst\"\n")
	fmt.Fprintf(file, "        echo \"Moved: $src -> $dst\"\n")
	fmt.Fprintf(file, "    fi\n")
	fmt.Fprintf(file, "}\n\n")

	// Group moves by book/metadata for better documentation
	fmt.Fprintf(file, "# ============================================\n")
	fmt.Fprintf(file, "# FILE MOVES\n")
	fmt.Fprintf(file, "# ============================================\n\n")

	// Track current book for grouping
	var currentBook string
	for _, move := range p.moves {
		// Add book header comment if metadata available and book changed
		if move.Metadata != nil {
			author := move.Metadata.GetFirstAuthor("Unknown Author")
			bookID := fmt.Sprintf("%s - %s", author, move.Metadata.Title)
			if bookID != currentBook {
				currentBook = bookID
				fmt.Fprintf(file, "\n# --------------------------------------------\n")
				fmt.Fprintf(file, "# Book: %s\n", move.Metadata.Title)
				fmt.Fprintf(file, "# Author: %s\n", author)
				if series := move.Metadata.GetValidSeries(); series != "" {
					fmt.Fprintf(file, "# Series: %s", series)
					if move.Metadata.TrackNumber > 0 {
						fmt.Fprintf(file, " #%d", move.Metadata.TrackNumber)
					}
					fmt.Fprintf(file, "\n")
				}
				fmt.Fprintf(file, "# --------------------------------------------\n")
			}
		}

		// Write the move command
		fmt.Fprintf(file, "move_file %s %s\n",
			shellQuote(move.SourcePath),
			shellQuote(move.TargetPath))
	}

	// Add summary at the end
	fmt.Fprintf(file, "\n# ============================================\n")
	fmt.Fprintf(file, "# SUMMARY\n")
	fmt.Fprintf(file, "# ============================================\n")
	fmt.Fprintf(file, "echo \"\"\n")
	fmt.Fprintf(file, "echo \"Plan complete: %d files processed\"\n", len(p.moves))

	// Make the script executable
	if err := os.Chmod(p.scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	color.Green("📝 Plan script written to: %s", p.scriptPath)
	color.Cyan("   Review and run with: bash %s", p.scriptPath)
	color.Cyan("   Preview first with:  DRY_RUN=1 bash %s", p.scriptPath)

	return nil
}

// shellQuote properly quotes a string for shell usage
func shellQuote(s string) string {
	// If the string contains no special characters, just quote it
	if !strings.ContainsAny(s, " \t\n'\"\\$`!") {
		return "\"" + s + "\""
	}

	// Use single quotes and escape any single quotes in the string
	escaped := strings.ReplaceAll(s, "'", "'\"'\"'")
	return "'" + escaped + "'"
}

// MoveCount returns the number of planned moves
func (p *PlanScriptWriter) MoveCount() int {
	return len(p.moves)
}

// PlanFileWriter handles writing dry-run plans to human-readable text files
type PlanFileWriter struct {
	filePath  string
	moves     []PlannedMove
	baseDir   string
	outputDir string
}

// NewPlanFileWriter creates a new plan file writer
func NewPlanFileWriter(filePath, baseDir, outputDir string) *PlanFileWriter {
	return &PlanFileWriter{
		filePath:  filePath,
		moves:     []PlannedMove{},
		baseDir:   baseDir,
		outputDir: outputDir,
	}
}

// AddMove adds a planned move to the file
func (p *PlanFileWriter) AddMove(source, target string, metadata *Metadata) {
	p.moves = append(p.moves, PlannedMove{
		SourcePath: source,
		TargetPath: target,
		Metadata:   metadata,
	})
}

// WriteFile writes the plan to a human-readable text file
func (p *PlanFileWriter) WriteFile() error {
	if p.filePath == "" {
		return nil
	}

	file, err := os.Create(p.filePath)
	if err != nil {
		return fmt.Errorf("failed to create plan file: %w", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "================================================================================\n")
	fmt.Fprintf(file, "                    AUDIOBOOK ORGANIZER - MOVE PLAN\n")
	fmt.Fprintf(file, "================================================================================\n\n")
	fmt.Fprintf(file, "Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "Source Directory: %s\n", p.baseDir)
	if p.outputDir != "" {
		fmt.Fprintf(file, "Output Directory: %s\n", p.outputDir)
	}
	fmt.Fprintf(file, "Total Files: %d\n\n", len(p.moves))

	fmt.Fprintf(file, "================================================================================\n")
	fmt.Fprintf(file, "                              PLANNED MOVES\n")
	fmt.Fprintf(file, "================================================================================\n\n")

	// Group moves by book for better readability
	var currentBook string
	for i, move := range p.moves {
		// Add book header if metadata available and book changed
		if move.Metadata != nil {
			author := move.Metadata.GetFirstAuthor("Unknown Author")
			bookID := fmt.Sprintf("%s - %s", author, move.Metadata.Title)
			if bookID != currentBook {
				currentBook = bookID
				if i > 0 {
					fmt.Fprintf(file, "\n")
				}
				fmt.Fprintf(file, "--------------------------------------------------------------------------------\n")
				fmt.Fprintf(file, "BOOK: %s\n", move.Metadata.Title)
				fmt.Fprintf(file, "AUTHOR: %s\n", author)
				if series := move.Metadata.GetValidSeries(); series != "" {
					fmt.Fprintf(file, "SERIES: %s", series)
					if move.Metadata.TrackNumber > 0 {
						fmt.Fprintf(file, " (#%d)", move.Metadata.TrackNumber)
					}
					fmt.Fprintf(file, "\n")
				}
				fmt.Fprintf(file, "--------------------------------------------------------------------------------\n")
			}
		}

		// Write the move
		fmt.Fprintf(file, "  FROM: %s\n", move.SourcePath)
		fmt.Fprintf(file, "    TO: %s\n", move.TargetPath)
	}

	// Write summary
	fmt.Fprintf(file, "\n================================================================================\n")
	fmt.Fprintf(file, "                                SUMMARY\n")
	fmt.Fprintf(file, "================================================================================\n\n")
	fmt.Fprintf(file, "Total files to be moved: %d\n", len(p.moves))
	fmt.Fprintf(file, "\nThis is a DRY-RUN plan. No files have been moved.\n")
	fmt.Fprintf(file, "To execute these moves, run the organizer without --plan-file.\n")

	color.Green("📄 Plan file written to: %s", p.filePath)

	return nil
}

// MoveCount returns the number of planned moves
func (p *PlanFileWriter) MoveCount() int {
	return len(p.moves)
}
