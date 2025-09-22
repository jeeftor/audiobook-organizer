// internal/organizer/logging.go
package organizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Configuration for forcing dark mode
var (
	ForceDarkMode = false // Set this to true to force dark background everywhere
)

// SetForceDarkMode enables/disables forced dark mode
func SetForceDarkMode(enabled bool) {
	ForceDarkMode = enabled

	// When dark mode is enabled, we need to create dark mode styles
	// This is handled by the styles.go file
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

// Style definitions that work both normally and in forced dark mode
var (
	// Path component styles - these will be updated by updateStylesForDarkMode if needed
	AuthorStyle      = lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	SeriesStyle      = lipgloss.NewStyle().Background(lipgloss.Color("#0000FF")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	TitleStyle       = lipgloss.NewStyle().Background(lipgloss.Color("#00FF00")).Foreground(lipgloss.Color("#000000")).Bold(true)
	TrackNumberStyle = lipgloss.NewStyle().Background(lipgloss.Color("#FF00FF")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	FilenameStyle    = lipgloss.NewStyle().Background(lipgloss.Color("#FFFF00")).Foreground(lipgloss.Color("#000000")).Bold(true)

	// Metadata display styles
	IconStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	FieldNameStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	NormalTextStyle = lipgloss.NewStyle() // Default terminal color
	TargetPathStyle = lipgloss.NewStyle().Background(lipgloss.Color("#00FFFF")).Foreground(lipgloss.Color("#000000")).Bold(true)

	// File type styles
	AudioFileStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	EpubFileStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF"))
	M4bFileStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	M4aFileStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	GenericFileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF"))
)

// Print functions that respect the ForceDarkMode setting
func PrintBase(format string, a ...interface{}) {
	if len(a) == 0 {
		fmt.Println(format)
	} else {
		fmt.Println(fmt.Sprintf(format, a...))
	}
}

func PrintRed(format string, a ...interface{}) {
	printStyled(Styles.Error, format, a...)
}

func PrintGreen(format string, a ...interface{}) {
	printStyled(Styles.Success, format, a...)
}

func PrintYellow(format string, a ...interface{}) {
	printStyled(Styles.Warning, format, a...)
}

func PrintBlue(format string, a ...interface{}) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF"))
	printStyled(style, format, a...)
}

func PrintCyan(format string, a ...interface{}) {
	printStyled(Styles.Info, format, a...)
}

func PrintMagenta(format string, a ...interface{}) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF"))
	printStyled(style, format, a...)
}
