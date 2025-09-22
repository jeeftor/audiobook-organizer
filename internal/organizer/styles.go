package organizer

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

// Styles defines all the Lipgloss styles used throughout the application.
// This centralizes styling to make it consistent and easy to modify.
var Styles = struct {
	// Text styles
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Normal      lipgloss.Style
	Bold        lipgloss.Style
	Faint       lipgloss.Style

	// Color styles
	Success     lipgloss.Style
	Error       lipgloss.Style
	Warning     lipgloss.Style
	Info        lipgloss.Style

	// UI element styles
	Prompt      lipgloss.Style
	Path        lipgloss.Style
	Highlight   lipgloss.Style

	// Icon styles
	IconSuccess lipgloss.Style
	IconError   lipgloss.Style
	IconWarning lipgloss.Style
	IconInfo    lipgloss.Style
	IconPrompt  lipgloss.Style
}{
	// Text styles
	Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")),
	Subtitle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD")),
	Normal:      lipgloss.NewStyle(),
	Bold:        lipgloss.NewStyle().Bold(true),
	Faint:       lipgloss.NewStyle().Faint(true),

	// Color styles - matching fatih/color defaults as closely as possible
	Success:     lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")), // Green
	Error:       lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")), // Red
	Warning:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")), // Yellow
	Info:        lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")), // Cyan

	// UI element styles
	Prompt:      lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")),
	Path:        lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")), // Yellow
	Highlight:   lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")), // Cyan

	// Icon styles - for emoji and symbols
	IconSuccess: lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")), // Green
	IconError:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")), // Red
	IconWarning: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")), // Yellow
	IconInfo:    lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")), // Cyan
	IconPrompt:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")), // White
}

// Helper functions to render text with styles

// RenderTitle renders text with the Title style
func RenderTitle(text string) string {
	return Styles.Title.Render(text)
}

// RenderSuccess renders text with the Success style
func RenderSuccess(text string) string {
	return Styles.Success.Render(text)
}

// RenderError renders text with the Error style
func RenderError(text string) string {
	return Styles.Error.Render(text)
}

// RenderWarning renders text with the Warning style
func RenderWarning(text string) string {
	return Styles.Warning.Render(text)
}

// RenderInfo renders text with the Info style
func RenderInfo(text string) string {
	return Styles.Info.Render(text)
}

// RenderPrompt renders text with the Prompt style
func RenderPrompt(text string) string {
	return Styles.Prompt.Render(text)
}

// RenderPath renders text with the Path style
func RenderPath(text string) string {
	return Styles.Path.Render(text)
}

// RenderHighlight renders text with the Highlight style
func RenderHighlight(text string) string {
	return Styles.Highlight.Render(text)
}

// Icon rendering functions

// RenderSuccessIcon renders an icon with the IconSuccess style
func RenderSuccessIcon(icon string) string {
	return Styles.IconSuccess.Render(icon)
}

// RenderErrorIcon renders an icon with the IconError style
func RenderErrorIcon(icon string) string {
	return Styles.IconError.Render(icon)
}

// RenderWarningIcon renders an icon with the IconWarning style
func RenderWarningIcon(icon string) string {
	return Styles.IconWarning.Render(icon)
}

// RenderInfoIcon renders an icon with the IconInfo style
func RenderInfoIcon(icon string) string {
	return Styles.IconInfo.Render(icon)
}

// RenderPromptIcon renders an icon with the IconPrompt style
func RenderPromptIcon(icon string) string {
	return Styles.IconPrompt.Render(icon)
}

// Format helpers that mimic fatih/color functions for easier migration

// PrintInfo prints text with Info style (cyan)
func PrintInfo(format string, a ...interface{}) {
	printStyled(Styles.Info, format, a...)
}

// PrintSuccess prints text with Success style (green)
func PrintSuccess(format string, a ...interface{}) {
	printStyled(Styles.Success, format, a...)
}

// PrintError prints text with Error style (red)
func PrintError(format string, a ...interface{}) {
	printStyled(Styles.Error, format, a...)
}

// PrintWarning prints text with Warning style (yellow)
func PrintWarning(format string, a ...interface{}) {
	printStyled(Styles.Warning, format, a...)
}

// Helper function to print styled text
func printStyled(style lipgloss.Style, format string, a ...interface{}) {
	if len(a) == 0 {
		fmt.Println(style.Render(format))
	} else {
		fmt.Println(style.Render(fmt.Sprintf(format, a...)))
	}
}

// Metadata-specific styling functions

// IconColor applies styling to icons in metadata display
func IconColor(text string) string {
	return Styles.IconInfo.Render(text)
}

// AuthorColor applies styling to author names
func AuthorColor(text string) string {
	return Styles.Highlight.Render(text)
}

// SeriesColor applies styling to series names
func SeriesColor(text string) string {
	return Styles.Warning.Render(text)
}

// TitleColor applies styling to book titles
func TitleColor(text string) string {
	return Styles.Bold.Render(text)
}

// TrackNumberColor applies styling to track numbers
func TrackNumberColor(text string) string {
	return Styles.Info.Render(text)
}

// FieldNameColor applies styling to field names in metadata
func FieldNameColor(text string) string {
	return Styles.Faint.Render(text)
}

// FilenameColor applies styling to filenames
func FilenameColor(text string) string {
	return Styles.Path.Render(text)
}

// TargetPathColor applies styling to target paths
func TargetPathColor(text string) string {
	return Styles.Highlight.Render(text)
}
