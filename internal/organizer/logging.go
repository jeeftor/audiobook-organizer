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

// Configuration for forcing dark mode
var (
	ForceDarkMode = false // Set this to true to force dark background everywhere
)

// Dark mode color functions - black background with appropriate foreground colors
var (
	darkModeBase    = color.New(color.BgBlack, color.FgWhite)
	darkModeSuccess = color.New(color.BgBlack, color.FgGreen)
	darkModeWarning = color.New(color.BgBlack, color.FgYellow)
	darkModeError   = color.New(color.BgBlack, color.FgRed)
	darkModeInfo    = color.New(color.BgBlack, color.FgCyan)
	darkModeBlue    = color.New(color.BgBlack, color.FgBlue)
)

// SetForceDarkMode enables/disables forced dark mode
func SetForceDarkMode(enabled bool) {
	ForceDarkMode = enabled

	if enabled {
		// Override all the print functions to use black background
		updateColorFunctionsForDarkMode()
	}
}

// updateColorFunctionsForDarkMode overrides color functions to use black background
func updateColorFunctionsForDarkMode() {
	// Update path component colors for dark mode
	AuthorColor = color.New(color.BgBlack, color.FgHiRed, color.Bold).SprintFunc()
	SeriesColor = color.New(color.BgBlack, color.FgHiBlue, color.Bold).SprintFunc()
	TitleColor = color.New(color.BgBlack, color.FgHiGreen, color.Bold).SprintFunc()
	TrackNumberColor = color.New(color.BgBlack, color.FgHiMagenta, color.Bold).SprintFunc()
	FilenameColor = color.New(color.BgBlack, color.FgHiYellow, color.Bold).SprintFunc()

	// Update display colors
	IconColor = color.New(color.BgBlack, color.FgCyan).SprintFunc()
	FieldNameColor = color.New(color.BgBlack, color.FgWhite, color.Bold).SprintFunc()
	NormalTextColor = color.New(color.BgBlack, color.FgWhite).SprintFunc()
	TargetPathColor = color.New(color.BgBlack, color.FgHiCyan, color.Bold).SprintFunc()

	// Update file type colors
	AudioFileColor = color.New(color.BgBlack, color.FgGreen).SprintFunc()
	EpubFileColor = color.New(color.BgBlack, color.FgBlue).SprintFunc()
	M4bFileColor = color.New(color.BgBlack, color.FgCyan).SprintFunc()
	M4aFileColor = color.New(color.BgBlack, color.FgYellow).SprintFunc()
	GenericFileColor = color.New(color.BgBlack, color.FgMagenta).SprintFunc()
}

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
		PrintYellow("↩️  Restoring files from %s to %s", entry.TargetPath, entry.SourcePath)
		if err := os.MkdirAll(entry.SourcePath, 0755); err != nil {
			PrintRed("❌ Error creating source directory: %v", err)
			continue
		}

		for _, file := range entry.Files {
			oldPath := filepath.Join(entry.TargetPath, file)
			newPath := filepath.Join(entry.SourcePath, file)
			if o.config.Verbose {
				PrintBlue("📦 Moving %s to %s", oldPath, newPath)
			}
			if err := os.Rename(oldPath, newPath); err != nil {
				PrintRed("❌ Error moving %s: %v", oldPath, err)
			}
		}
	}

	if err := os.Remove(logPath); err != nil {
		PrintYellow("⚠️  Warning: couldn't remove log file: %v", err)
	}

	return nil
}

func (o *Organizer) printSummary(startTime time.Time) {
	duration := time.Since(startTime)

	PrintBase("\n📊 Summary Report")
	PrintBase("⏱️  Duration: %v", duration.Round(time.Millisecond))

	PrintGreen("\n📚 Metadata files found: %d", len(o.summary.MetadataFound))
	if len(o.summary.MetadataFound) > 0 {
		PrintBase("\n📖 Valid Audiobooks Found:")
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
				PrintGreen("  📚 %s by %s", metadata.Title, strings.Join(metadata.Authors, ", "))
				if len(metadata.Series) > 0 && metadata.Series[0] != "" {
					cleanedSeries := CleanSeriesName(metadata.Series[0])
					PrintGreen("     📖 Series: %s", cleanedSeries)
				}
			}
		}
	}

	if len(o.summary.MetadataMissing) > 0 {
		PrintYellow("\n⚠️  Directories without metadata: %d", len(o.summary.MetadataMissing))
		if o.config.Verbose {
			for _, path := range o.summary.MetadataMissing {
				PrintBase("  - %s", path)
			}
		}
	}

	PrintCyan("\n🔄 Moves planned/executed: %d", len(o.summary.Moves))
	for _, move := range o.summary.Moves {
		PrintBase("  From: %s", move.From)
		PrintBase("  To: %s\n", move.To)
	}

	// Print information about removed empty directories
	if o.config.RemoveEmpty && len(o.summary.EmptyDirsRemoved) > 0 {
		PrintYellow("\n🗑️  Empty directories removed: %d", len(o.summary.EmptyDirsRemoved))
		if o.config.Verbose {
			for _, path := range o.summary.EmptyDirsRemoved {
				PrintBase("  - %s", path)
			}
		}
	}

	if o.config.DryRun {
		PrintYellow("\n🔍 This was a dry run - no files were actually moved or directories removed")
	} else {
		PrintGreen("\n✅ Organization complete!")
	}
}

// Color scheme that works both normally and in forced dark mode
var (
	// Path component colors - these will be updated by updateColorFunctionsForDarkMode if needed
	AuthorColor      = color.New(color.BgRed, color.FgWhite, color.Bold).SprintFunc()
	SeriesColor      = color.New(color.BgBlue, color.FgWhite, color.Bold).SprintFunc()
	TitleColor       = color.New(color.BgGreen, color.FgBlack, color.Bold).SprintFunc()
	TrackNumberColor = color.New(color.BgMagenta, color.FgWhite, color.Bold).SprintFunc()
	FilenameColor    = color.New(color.BgYellow, color.FgBlack, color.Bold).SprintFunc()

	// Metadata display colors
	IconColor       = color.New(color.FgCyan).SprintFunc()
	FieldNameColor  = color.New(color.FgWhite, color.Bold).SprintFunc()
	NormalTextColor = color.New().SprintFunc() // Default terminal color
	TargetPathColor = color.New(color.BgCyan, color.FgBlack, color.Bold).SprintFunc()

	// File type colors
	AudioFileColor   = color.New(color.FgGreen).SprintFunc()
	EpubFileColor    = color.New(color.FgBlue).SprintFunc()
	M4bFileColor     = color.New(color.FgCyan).SprintFunc()
	M4aFileColor     = color.New(color.FgYellow).SprintFunc()
	GenericFileColor = color.New(color.FgMagenta).SprintFunc()
)

// Print functions that respect the ForceDarkMode setting
func PrintBase(format string, a ...interface{}) {
	if ForceDarkMode {
		darkModeBase.Printf(format+"\n", a...)
	} else {
		fmt.Printf(format+"\n", a...)
	}
}

func PrintRed(format string, a ...interface{}) {
	if ForceDarkMode {
		darkModeError.Printf(format+"\n", a...)
	} else {
		color.New(color.FgRed).Printf(format+"\n", a...)
	}
}

func PrintGreen(format string, a ...interface{}) {
	if ForceDarkMode {
		darkModeSuccess.Printf(format+"\n", a...)
	} else {
		color.New(color.FgGreen).Printf(format+"\n", a...)
	}
}

func PrintYellow(format string, a ...interface{}) {
	if ForceDarkMode {
		darkModeWarning.Printf(format+"\n", a...)
	} else {
		color.New(color.FgYellow).Printf(format+"\n", a...)
	}
}

func PrintBlue(format string, a ...interface{}) {
	if ForceDarkMode {
		darkModeBlue.Printf(format+"\n", a...)
	} else {
		color.New(color.FgBlue).Printf(format+"\n", a...)
	}
}

func PrintCyan(format string, a ...interface{}) {
	if ForceDarkMode {
		darkModeInfo.Printf(format+"\n", a...)
	} else {
		color.New(color.FgCyan).Printf(format+"\n", a...)
	}
}

func PrintMagenta(format string, a ...interface{}) {
	if ForceDarkMode {
		color.New(color.BgBlack, color.FgMagenta).Printf(format+"\n", a...)
	} else {
		color.New(color.FgMagenta).Printf(format+"\n", a...)
	}
}
