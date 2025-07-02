// internal/organizer/logging.go
package organizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

func (o *Organizer) saveLog() error {
	logPath := o.GetLogPath()
	data, err := json.MarshalIndent(o.logEntries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(logPath, data, 0644)
}

func (o *Organizer) undoMoves() error {
	logPath := o.GetLogPath()
	data, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("no log file found at %s", logPath)
	}

	var entries []LogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("error parsing log: %v", err)
	}

	for _, entry := range entries {
		color.Yellow("↩️  Restoring files from %s to %s", entry.TargetPath, entry.SourcePath)
		if err := os.MkdirAll(entry.SourcePath, 0755); err != nil {
			color.Red("❌ Error creating source directory: %v", err)
			continue
		}

		for _, file := range entry.Files {
			oldPath := filepath.Join(entry.TargetPath, file)
			newPath := filepath.Join(entry.SourcePath, file)
			if o.config.Verbose {
				color.Blue("📦 Moving %s to %s", oldPath, newPath)
			}
			if err := os.Rename(oldPath, newPath); err != nil {
				color.Red("❌ Error moving %s: %v", oldPath, err)
			}
		}
	}

	if err := os.Remove(logPath); err != nil {
		color.Yellow("⚠️  Warning: couldn't remove log file: %v", err)
	}

	return nil
}

func (o *Organizer) printSummary(startTime time.Time) {
	duration := time.Since(startTime)

	fmt.Println("\n📊 Summary Report")
	color.White("⏱️  Duration: %v", duration.Round(time.Millisecond))

	color.Green("\n📚 Metadata files found: %d", len(o.summary.MetadataFound))
	if len(o.summary.MetadataFound) > 0 {
		fmt.Println("\n📖 Valid Audiobooks Found:")
		for _, path := range o.summary.MetadataFound {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var metadata Metadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				continue
			}
			if len(metadata.Authors) > 0 && metadata.Title != "" {
				color.Green("  📚 %s by %s", metadata.Title, strings.Join(metadata.Authors, ", "))
				if len(metadata.Series) > 0 && metadata.Series[0] != "" {
					cleanedSeries := cleanSeriesName(metadata.Series[0])
					color.Green("     📖 Series: %s", cleanedSeries)
				}
			}
		}
	}

	if len(o.summary.MetadataMissing) > 0 {
		color.Yellow("\n⚠️  Directories without metadata: %d", len(o.summary.MetadataMissing))
		if o.config.Verbose {
			for _, path := range o.summary.MetadataMissing {
				fmt.Printf("  - %s\n", path)
			}
		}
	}

	color.Cyan("\n🔄 Moves planned/executed: %d", len(o.summary.Moves))
	for _, move := range o.summary.Moves {
		fmt.Printf("  From: %s\n  To: %s\n\n", move.From, move.To)
	}

	// Print information about removed empty directories
	if o.config.RemoveEmpty && len(o.summary.EmptyDirsRemoved) > 0 {
		color.Yellow("\n🗑️  Empty directories removed: %d", len(o.summary.EmptyDirsRemoved))
		if o.config.Verbose {
			for _, path := range o.summary.EmptyDirsRemoved {
				fmt.Printf("  - %s\n", path)
			}
		}
	}

	if o.config.DryRun {
		color.Yellow("\n🔍 This was a dry run - no files were actually moved or directories removed")
	} else {
		color.Green("\n✅ Organization complete!")
	}
}
