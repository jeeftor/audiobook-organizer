// internal/organizer/logging.go
package organizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// printGroupedMoveSummary displays moves grouped by destination book with pretty formatting
func (o *Organizer) printGroupedMoveSummary() {
	// Group moves by Author/Series (first 2 path components after output dir)
	// This handles cases where bad titles create multiple subdirectories
	bookMoves := make(map[string][]MoveSummary)

	for _, move := range o.summary.Moves {
		// Get relative path from output dir
		relPath := move.To
		if o.config.OutputDir != "" {
			if rel, err := filepath.Rel(o.config.OutputDir, move.To); err == nil {
				relPath = rel
			}
		}

		// Extract Author/Series (first 2 components)
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		var bookKey string
		if len(parts) >= 2 {
			bookKey = filepath.Join(parts[0], parts[1]) // Author/Series
		} else {
			bookKey = filepath.Dir(move.To) // Fallback
		}

		bookMoves[bookKey] = append(bookMoves[bookKey], move)
	}

	// Sort book keys
	var bookKeys []string
	for key := range bookMoves {
		bookKeys = append(bookKeys, key)
	}
	sort.Strings(bookKeys)

	// Fancy header with lipgloss
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#5F00D7")).
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#AF87FF")).
		Padding(0, 2).
		MarginTop(1).
		MarginBottom(1).
		Align(lipgloss.Center)

	// Book section header - colorful box
	bookHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#0087AF")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D7FF")).
		Padding(0, 1).
		MarginTop(1)

	// Color styles for path components
	authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")) // Green
	seriesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")) // Cyan
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))  // Yellow
	fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))   // White
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))    // Dark gray

	// Ellipsis styling
	ellipsisStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Italic(true).
		Faint(true)

	// Print fancy header
	fmt.Println(headerStyle.Render("✨ 📚  GROUPED MOVE SUMMARY  📚 ✨"))

	for _, bookKey := range bookKeys {
		moves := bookMoves[bookKey]

		// Sort files within each book by destination path
		sort.Slice(moves, func(i, j int) bool {
			return moves[i].To < moves[j].To
		})

		// Book header section
		fmt.Printf("\n%s\n",
			bookHeaderStyle.Render(fmt.Sprintf("📖 %s (%d files)", bookKey, len(moves))))

		// List files with their destination paths, color-coded
		maxDisplay := 5
		displayCount := len(moves)
		if displayCount > maxDisplay {
			displayCount = maxDisplay
		}

		for i := 0; i < displayCount; i++ {
			relPath := moves[i].To
			if o.config.OutputDir != "" {
				if rel, err := filepath.Rel(o.config.OutputDir, moves[i].To); err == nil {
					relPath = rel
				}
			}

			// Split path and color each component
			parts := strings.Split(filepath.ToSlash(relPath), "/")
			var coloredParts []string

			for idx, part := range parts {
				switch {
				case idx == 0:
					// Author - green
					coloredParts = append(coloredParts, authorStyle.Render(part))
				case idx == 1:
					// Series - cyan
					coloredParts = append(coloredParts, seriesStyle.Render(part))
				case idx == len(parts)-1:
					// Filename - white
					coloredParts = append(coloredParts, fileStyle.Render(part))
				default:
					// Title/subdirs - yellow
					coloredParts = append(coloredParts, titleStyle.Render(part))
				}
			}

			fmt.Printf("  %s\n", strings.Join(coloredParts, sepStyle.Render("/")))
		}

		if len(moves) > maxDisplay {
			remaining := len(moves) - maxDisplay
			fmt.Printf("  %s\n",
				ellipsisStyle.Render(fmt.Sprintf("... %d more files ...", remaining)))
		}
	}

	// Fancy footer with totals
	footerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#00875F")).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#00D787")).
		Padding(0, 2).
		MarginTop(1).
		Align(lipgloss.Center)

	totalText := fmt.Sprintf("📊 TOTAL: %d Books • %d Files",
		len(bookKeys),
		len(o.summary.Moves))
	fmt.Println(footerStyle.Render(totalText))
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

	// Display grouped summary if there are moves
	if len(o.summary.Moves) > 0 {
		o.printGroupedMoveSummary()
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
