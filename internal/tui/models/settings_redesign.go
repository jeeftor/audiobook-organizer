package models

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	summaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Padding(0, 1)

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#5555AA")).
				Padding(0, 1)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFF00")).
				Background(lipgloss.Color("#333333"))

	normalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA"))
)

// FocusArea represents which area has focus
type FocusArea int

const (
	TableFocus FocusArea = iota
	MetadataFocus
)

// SettingsTableModel represents the redesigned settings screen with table layout
type SettingsTableModel struct {
	table            table.Model
	metadataViewport viewport.Model
	selectedBooks    []AudioBook
	width            int
	height           int
	showAdvanced     bool
	focusArea        FocusArea

	// Popup state
	showPopup        bool
	popupOptions     []string
	popupSelection   int
	popupSettingIdx  int
	justClosedPopup  bool

	// Metadata navigation
	metadataBookIndex int

	// Settings values
	settings      []Setting
	fieldMappings []FieldMappingSetting
}

// NewSettingsTableModel creates a new table-based settings model
func NewSettingsTableModel(selectedBooks []AudioBook, showAdvanced bool) *SettingsTableModel {
	// Create settings (same as before)
	settings := []Setting{
		{Name: "Layout", Description: "Directory structure", Options: []string{"author-only", "author-title", "author-series-title", "author-series-title-number"}, Value: 2},
		{Name: "Use Embedded Metadata", Description: "Use file metadata", Options: []string{"No", "Yes"}, Value: 1},
		{Name: "Flat Mode", Description: "Process files individually", Options: []string{"No", "Yes"}, Value: 0},
		{Name: "Dry Run", Description: "Preview without moving", Options: []string{"No", "Yes"}, Value: 0},
		{Name: "Verbose", Description: "Detailed output", Options: []string{"No", "Yes"}, Value: 1},
	}

	fieldMappings := []FieldMappingSetting{
		{Name: "Layout", Description: "Directory structure", Options: []string{"author-only", "author-title", "author-series-title", "author-series-title-number"}, Value: 2},
		{Name: "Flat Mode", Description: "Process individually", Options: []string{"No", "Yes"}, Value: 0},
		{Name: "Title Field", Description: "Field for title", Options: []string{"title", "album", "series", "track_title"}, Value: 0},
		{Name: "Series Field", Description: "Field for series", Options: []string{"series", "album", "title"}, Value: 0},
		{Name: "Author Fields", Description: "Author priority", Options: []string{"authors→artist→album_artist", "authors→narrators→artist", "artist→album_artist→composer", "authors only"}, Value: 0},
		{Name: "Track Field", Description: "Field for track", Options: []string{"track", "track_number", "disc"}, Value: 0},
	}

	// Combine ALL settings into one unified view
	// Always show all settings (basic + advanced field mappings)
	allSettings := []FieldMappingSetting{
		// Basic settings converted to FieldMappingSetting format
		{Name: "Use Embedded Metadata", Description: "Use file metadata", Options: []string{"No", "Yes"}, Value: 1},
		{Name: "Dry Run", Description: "Preview without moving", Options: []string{"No", "Yes"}, Value: 0},
		{Name: "Verbose", Description: "Detailed output", Options: []string{"No", "Yes"}, Value: 1},
		{Name: "───────────────────", Description: "separator", Options: []string{""}, Value: 0}, // Visual separator
		// Advanced field mapping settings
		{Name: "Layout", Description: "Directory structure", Options: []string{"author-only", "author-title", "author-series-title", "author-series-title-number"}, Value: 2},
		{Name: "Flat Mode", Description: "Process individually", Options: []string{"No", "Yes"}, Value: 0},
		{Name: "Title Field", Description: "Field for title", Options: []string{"title", "album", "series", "track_title"}, Value: 0},
		{Name: "Series Field", Description: "Field for series", Options: []string{"series", "album", "title"}, Value: 0},
		{Name: "Author Fields", Description: "Author priority", Options: []string{"authors→artist→album_artist", "authors→narrators→artist", "artist→album_artist→composer", "authors only"}, Value: 0},
		{Name: "Track Field", Description: "Field for track", Options: []string{"track", "track_number", "disc"}, Value: 0},
	}

	// Calculate max width needed for "Current" column
	maxCurrentWidth := 10
	for _, setting := range allSettings {
		for _, opt := range setting.Options {
			if len(opt) > maxCurrentWidth {
				maxCurrentWidth = len(opt)
			}
		}
	}
	// Add padding
	maxCurrentWidth += 2

	// Create table columns - 3 columns for better visibility
	columns := []table.Column{
		{Title: "Setting", Width: 22},
		{Title: "Current", Width: maxCurrentWidth},
		{Title: "Options", Width: 50},
	}

	// Create table rows with all options visible
	var rows []table.Row
	optionsColWidth := 50
	for _, setting := range allSettings {
		// Show available options in a compact format
		optionsStr := strings.Join(setting.Options, " | ")
		if len(optionsStr) > optionsColWidth {
			optionsStr = optionsStr[:optionsColWidth-3] + "..."
		}

		rows = append(rows, table.Row{
			setting.Name,
			setting.Options[setting.Value],
			optionsStr,
		})
	}

	// Store the unified settings
	fieldMappings = allSettings

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10), // Show all 10 rows
	)

	s := table.DefaultStyles()
	s.Header = tableHeaderStyle
	s.Selected = selectedRowStyle
	t.SetStyles(s)

	// Create viewport for metadata (will be resized on first WindowSizeMsg)
	metadataVp := viewport.New(100, 15)
	metadataVp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#AA7DFF")).
		Padding(0, 1)

	m := &SettingsTableModel{
		table:            t,
		metadataViewport: metadataVp,
		selectedBooks:    selectedBooks,
		settings:         settings,
		fieldMappings:    fieldMappings,
		showAdvanced:     showAdvanced,
		focusArea:        TableFocus,
	}

	// Initialize metadata content
	m.updateMetadata()

	return m
}

// Init initializes the model
func (m *SettingsTableModel) Init() tea.Cmd {
	// Request window size on init
	return tea.WindowSize()
}

// Update handles messages and user input
func (m *SettingsTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate available space more accurately
		// Count actual lines in our UI:
		// - Header: 3 lines (emoji title + newline + newline)
		// - Table: 12 lines (1 header + 10 rows + 1 blank line after border)
		// - Table borders/padding: 2 lines
		// - Metadata title bar: 1 line
		// - Output preview: 5 lines
		// - Footer: 2 lines (newline + text)
		usedLines := 3 + 12 + 2 + 1 + 5 + 2  // = 25

		// Give all remaining space to metadata viewport
		// The viewport's border/padding is handled by its own Style, not counted here
		metadataHeight := msg.Height - usedLines

		// Cap at a reasonable maximum to avoid too much empty space
		if metadataHeight > 30 {
			metadataHeight = 30
		}

		// Ensure reasonable minimum
		if metadataHeight < 10 {
			metadataHeight = 10
		}

		vpWidth := msg.Width - 4  // Account for viewport borders

		m.metadataViewport.Width = vpWidth
		m.metadataViewport.Height = metadataHeight

		// Update metadata content
		m.updateMetadata()

	case tea.KeyMsg:
		// Reset the justClosedPopup flag on any key press
		if m.justClosedPopup {
			m.justClosedPopup = false
		}

		keyStr := msg.String()

		// Handle c/n globally to advance to next screen - ALWAYS works regardless of popup
		// Let these pass through to main.go (don't consume them)
		if keyStr == "c" || keyStr == "n" {
			// Don't consume - let main.go handle advancing to next screen
			// Just continue processing normally (fall through)
		} else if m.showPopup {
			// Handle popup keys - consume ALL other keys when popup is showing
			switch keyStr {
			case "up", "k":
				if m.popupSelection > 0 {
					m.popupSelection--
				}
				return m, nil
			case "down", "j":
				if m.popupSelection < len(m.popupOptions)-1 {
					m.popupSelection++
				}
				return m, nil
			case "enter", " ":
				// Apply selection and close popup
				m.fieldMappings[m.popupSettingIdx].Value = m.popupSelection
				m.updateTableRow(m.popupSettingIdx)
				m.updateMetadata()
				m.showPopup = false
				m.justClosedPopup = true
				// IMPORTANT: Return here to consume the Enter key
				return m, nil
			case "esc", "q":
				// Cancel popup
				m.showPopup = false
				m.justClosedPopup = true
				return m, nil
			default:
				// Consume any other keys when popup is showing
				return m, nil
			}
		}

		// Normal key handling
		switch msg.String() {
		case "tab":
			// Toggle between table and metadata
			if m.focusArea == TableFocus {
				m.focusArea = MetadataFocus
				m.table.Blur()
			} else {
				m.focusArea = TableFocus
				m.table.Focus()
			}

		case "up", "k":
			// If metadata is focused, scroll it; otherwise navigate table
			if m.focusArea == MetadataFocus {
				m.metadataViewport.LineUp(1)
			} else {
				m.table, cmd = m.table.Update(msg)
			}

		case "down", "j":
			// If metadata is focused, scroll it; otherwise navigate table
			if m.focusArea == MetadataFocus {
				m.metadataViewport.LineDown(1)
			} else {
				m.table, cmd = m.table.Update(msg)
			}

		case "left", "h":
			if m.focusArea == TableFocus {
				// Change settings when table is focused
				cursor := m.table.Cursor()
				if cursor < len(m.fieldMappings) && m.fieldMappings[cursor].Name != "───────────────────" {
					// For simple toggles (2 options), cycle backward
					if len(m.fieldMappings[cursor].Options) == 2 {
						if m.fieldMappings[cursor].Value > 0 {
							m.fieldMappings[cursor].Value--
							m.updateTableRow(cursor)
							m.updateMetadata()
						}
					}
				}
			} else if m.focusArea == MetadataFocus {
				// Navigate to previous book in metadata view
				if m.metadataBookIndex > 0 {
					m.metadataBookIndex--
					m.updateMetadata()
				}
			}

		case "right", "l":
			if m.focusArea == TableFocus {
				// Change settings when table is focused
				cursor := m.table.Cursor()
				if cursor < len(m.fieldMappings) && m.fieldMappings[cursor].Name != "───────────────────" {
					// For simple toggles (2 options), cycle forward
					if len(m.fieldMappings[cursor].Options) == 2 {
						if m.fieldMappings[cursor].Value < 1 {
							m.fieldMappings[cursor].Value++
							m.updateTableRow(cursor)
							m.updateMetadata()
						}
					} else {
						// For complex options (3+), show popup
						m.showPopup = true
						m.popupOptions = m.fieldMappings[cursor].Options
						m.popupSelection = m.fieldMappings[cursor].Value
						m.popupSettingIdx = cursor
					}
				}
			} else if m.focusArea == MetadataFocus {
				// Navigate to next book in metadata view
				if m.metadataBookIndex < len(m.selectedBooks)-1 {
					m.metadataBookIndex++
					m.updateMetadata()
				}
			}

		case "enter", " ":
			// Show popup picker for complex settings when table is focused
			if m.focusArea == TableFocus {
				cursor := m.table.Cursor()
				if cursor < len(m.fieldMappings) && m.fieldMappings[cursor].Name != "───────────────────" {
					// Only show popup for settings with 3+ options
					if len(m.fieldMappings[cursor].Options) >= 3 {
						m.showPopup = true
						m.popupOptions = m.fieldMappings[cursor].Options
						m.popupSelection = m.fieldMappings[cursor].Value
						m.popupSettingIdx = cursor
						return m, nil
					}
				}
			}
			// Enter no longer advances to next screen - use c/n instead

		case "pgup":
			// Only works when metadata is focused
			if m.focusArea == MetadataFocus {
				m.metadataViewport.LineUp(5)
			}

		case "pgdown":
			// Only works when metadata is focused
			if m.focusArea == MetadataFocus {
				m.metadataViewport.LineDown(5)
			}

		default:
			// Let table handle other keys when focused
			if m.focusArea == TableFocus {
				m.table, cmd = m.table.Update(msg)
			}
		}
	}

	return m, cmd
}

// updateTableRow updates a single row in the table
func (m *SettingsTableModel) updateTableRow(index int) {
	rows := m.table.Rows()
	if index < len(rows) && index < len(m.fieldMappings) {
		rows[index][1] = m.fieldMappings[index].Options[m.fieldMappings[index].Value]
		m.table.SetRows(rows)
	}
}

// colorizeOutputPath applies different colors to path components based on layout
func (m *SettingsTableModel) colorizeOutputPath(path string, layout string) string {
	parts := strings.Split(path, string(filepath.Separator))

	// Color scheme for different components
	authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9500"))      // Orange for author
	seriesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF"))      // Cyan for series
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))       // Green for title
	fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))        // Gray for filename
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))   // Gray for /

	// Skip the output directory (first part) and work with the rest
	if len(parts) > 0 {
		parts = parts[1:]
	}

	var coloredParts []string

	switch layout {
	case "author-only":
		// author/filename
		if len(parts) >= 2 {
			coloredParts = []string{
				authorStyle.Render(parts[0]),
				fileStyle.Render(parts[1]),
			}
		}
	case "author-title":
		// author/title/filename
		if len(parts) >= 3 {
			coloredParts = []string{
				authorStyle.Render(parts[0]),
				titleStyle.Render(parts[1]),
				fileStyle.Render(parts[2]),
			}
		}
	case "author-series-title", "author-series-title-number":
		// author/series/title/filename
		if len(parts) >= 4 {
			coloredParts = []string{
				authorStyle.Render(parts[0]),
				seriesStyle.Render(parts[1]),
				titleStyle.Render(parts[2]),
				fileStyle.Render(parts[3]),
			}
		} else if len(parts) >= 3 {
			// No series, fallback to author/title/filename
			coloredParts = []string{
				authorStyle.Render(parts[0]),
				titleStyle.Render(parts[1]),
				fileStyle.Render(parts[2]),
			}
		}
	default:
		// Fallback: just color each part
		for _, part := range parts {
			coloredParts = append(coloredParts, titleStyle.Render(part))
		}
	}

	// Join with colored separator
	return strings.Join(coloredParts, separatorStyle.Render("/"))
}

// generateOutputPreview generates a simple preview of output paths (non-scrollable)
func (m *SettingsTableModel) generateOutputPreview() string {
	var content strings.Builder

	// Get layout setting (index 4 in unified list)
	layout := "author-series-title"
	if len(m.fieldMappings) > 4 {
		layout = m.fieldMappings[4].Options[m.fieldMappings[4].Value]
	}

	// Get field mapping configuration
	fieldMapping := m.GetFieldMapping()

	// Show preview for up to 3 books
	previewCount := 3
	if len(m.selectedBooks) < previewCount {
		previewCount = len(m.selectedBooks)
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))
	content.WriteString(titleStyle.Render("Output Path Preview:") + "\n")

	for i := 0; i < previewCount; i++ {
		book := m.selectedBooks[i]

		// Generate output path using universal function
		outputPath := GenerateOutputPath(book, layout, fieldMapping, "output")

		// Colorize and format path
		coloredPath := m.colorizeOutputPath(outputPath, layout)
		content.WriteString("  " + coloredPath + "\n")
	}

	if len(m.selectedBooks) > previewCount {
		moreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
		content.WriteString(moreStyle.Render(fmt.Sprintf("  ... and %d more", len(m.selectedBooks)-previewCount)) + "\n")
	}

	return content.String()
}

// updateMetadata updates the metadata viewport content
func (m *SettingsTableModel) updateMetadata() {
	var content strings.Builder

	if len(m.selectedBooks) == 0 {
		content.WriteString("No books selected")
		m.metadataViewport.SetContent(content.String())
		return
	}

	// Get field mapping to see what's being used
	fieldMapping := m.GetFieldMapping()

	// Get the book to display based on current index
	if m.metadataBookIndex >= len(m.selectedBooks) {
		m.metadataBookIndex = 0
	}
	book := m.selectedBooks[m.metadataBookIndex]

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFAAFF"))

	// Color styles matching the output path components
	authorLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9500"))  // Orange for author
	seriesLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF"))  // Cyan for series
	titleLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))   // Green for title
	defaultLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAFF")) // Default for other fields

	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	usedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	checkmark := usedStyle.Render("✓ ")

	content.WriteString(titleStyle.Render(fmt.Sprintf("Metadata Preview (Book %d/%d):", m.metadataBookIndex+1, len(m.selectedBooks))) + "\n\n")

	// Show key metadata fields with checkmarks for fields being used
	// Title field - use title color (green)
	titleCheck := ""
	if fieldMapping.TitleField == "title" {
		titleCheck = checkmark
	}
	content.WriteString(titleCheck + titleLabelStyle.Render("Title: ") + valueStyle.Render(book.Metadata.Title) + "\n")

	// Authors - use author color (orange)
	authorCheck := ""
	for _, af := range fieldMapping.AuthorFields {
		if af == "authors" {
			authorCheck = checkmark
			break
		}
	}
	if len(book.Metadata.Authors) > 0 {
		content.WriteString(authorCheck + authorLabelStyle.Render("Authors: ") + valueStyle.Render(strings.Join(book.Metadata.Authors, ", ")) + "\n")
	}

	// Series - use series color (cyan)
	seriesCheck := ""
	if fieldMapping.SeriesField == "series" {
		seriesCheck = checkmark
	}
	if series := book.Metadata.GetValidSeries(); series != "" {
		content.WriteString(seriesCheck + seriesLabelStyle.Render("Series: ") + valueStyle.Render(series) + "\n")
	}

	// Album - could be title or series, determine color based on mapping
	albumCheck := ""
	albumLabelStyle := defaultLabelStyle
	if fieldMapping.TitleField == "album" {
		albumCheck = checkmark
		albumLabelStyle = titleLabelStyle
	} else if fieldMapping.SeriesField == "album" {
		albumCheck = checkmark
		albumLabelStyle = seriesLabelStyle
	}
	if book.Metadata.Album != "" {
		content.WriteString(albumCheck + albumLabelStyle.Render("Album: ") + valueStyle.Render(book.Metadata.Album) + "\n")
	}

	// Track title - could be used as title field
	trackTitleCheck := ""
	trackTitleLabelStyle := defaultLabelStyle
	if fieldMapping.TitleField == "track_title" {
		trackTitleCheck = checkmark
		trackTitleLabelStyle = titleLabelStyle
	}
	if book.Metadata.TrackTitle != "" {
		content.WriteString(trackTitleCheck + trackTitleLabelStyle.Render("Track Title: ") + valueStyle.Render(book.Metadata.TrackTitle) + "\n")
	}

	// Track number
	trackCheck := ""
	if fieldMapping.TrackField == "track" || fieldMapping.TrackField == "track_number" {
		trackCheck = checkmark
	}
	if book.Metadata.TrackNumber != 0 {
		content.WriteString(trackCheck + defaultLabelStyle.Render("Track Number: ") + valueStyle.Render(fmt.Sprintf("%d", book.Metadata.TrackNumber)) + "\n")
	}

	// Show file path (shortened to last 3 components)
	pathParts := strings.Split(book.Path, string(filepath.Separator))
	displayPath := book.Path
	if len(pathParts) > 3 {
		displayPath = ".../" + strings.Join(pathParts[len(pathParts)-3:], "/")
	}
	content.WriteString("\n" + defaultLabelStyle.Render("Source: ") + valueStyle.Render(displayPath) + "\n")
	content.WriteString(defaultLabelStyle.Render("Source Type: ") + valueStyle.Render(book.Metadata.SourceType) + "\n")

	// Show raw metadata fields if available
	if len(book.Metadata.RawData) > 0 {
		content.WriteString("\n" + defaultLabelStyle.Render("Raw Metadata Fields:") + "\n")
		for key, val := range book.Metadata.RawData {
			// Check if this raw field is being used and determine color
			rawCheck := ""
			rawLabelStyle := defaultLabelStyle

			// Check if it's the title field
			if key == fieldMapping.TitleField {
				rawCheck = checkmark
				rawLabelStyle = titleLabelStyle
			}
			// Check if it's the series field
			if key == fieldMapping.SeriesField {
				rawCheck = checkmark
				rawLabelStyle = seriesLabelStyle
			}
			// Check if it's an author field
			for _, af := range fieldMapping.AuthorFields {
				if key == af {
					rawCheck = checkmark
					rawLabelStyle = authorLabelStyle
					break
				}
			}
			// Check if it's the track field
			if key == fieldMapping.TrackField {
				rawCheck = checkmark
				// Track field doesn't have a special color
			}

			content.WriteString(fmt.Sprintf("  %s%s: %v\n", rawCheck, rawLabelStyle.Render(key), val))
		}
	}

	m.metadataViewport.SetContent(content.String())
	// Reset scroll position to top when content changes
	m.metadataViewport.GotoTop()
}

// renderMetadataViewport renders the metadata viewport with scroll indicator
func (m *SettingsTableModel) renderMetadataViewport() string {
	vp := m.metadataViewport
	focused := m.focusArea == MetadataFocus

	// Calculate scroll position
	scrollIndicator := ""
	atTop := vp.YOffset <= 0
	atBottom := vp.AtBottom()

	if vp.TotalLineCount() > vp.Height {
		if atTop {
			scrollIndicator = " ▼"
		} else if atBottom {
			scrollIndicator = " ▲"
		} else {
			scrollPercent := float64(vp.YOffset) / float64(vp.TotalLineCount()-vp.Height)
			scrollIndicator = fmt.Sprintf(" ▲%d%%▼", int(scrollPercent*100))
		}
	}

	// Create title bar
	focusIndicator := ""
	if focused {
		focusIndicator = " ●"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	if !focused {
		titleStyle = titleStyle.
			Foreground(lipgloss.Color("#888888")).
			Background(lipgloss.Color("#333333"))
	}

	scrollStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Background(lipgloss.Color("#7D56F4"))

	if !focused {
		scrollStyle = scrollStyle.Background(lipgloss.Color("#333333"))
	}

	titleBar := titleStyle.Render("Metadata" + focusIndicator)
	if scrollIndicator != "" {
		titleBar += scrollStyle.Render(scrollIndicator)
	}

	return titleBar + "\n" + vp.View()
}

// getMetadataFieldValue retrieves the value of a metadata field from a book
func (m *SettingsTableModel) getMetadataFieldValue(fieldName string, book *AudioBook) string {
	// Special handling for comma-separated author fields (e.g., "authors,artist,album_artist")
	if strings.Contains(fieldName, ",") {
		fields := strings.Split(fieldName, ",")
		var values []string
		for _, field := range fields {
			field = strings.TrimSpace(field)
			val := m.getSingleMetadataFieldValue(field, book)
			if val != "" {
				values = append(values, fmt.Sprintf("%s=%s", field, val))
			}
		}
		if len(values) > 0 {
			return strings.Join(values, "; ")
		}
		return ""
	}

	return m.getSingleMetadataFieldValue(fieldName, book)
}

// getSingleMetadataFieldValue retrieves a single metadata field value
func (m *SettingsTableModel) getSingleMetadataFieldValue(fieldName string, book *AudioBook) string {
	// First check RawData
	if val, ok := book.Metadata.RawData[fieldName]; ok && val != nil {
		if strVal, ok := val.(string); ok {
			return strVal
		}
		return fmt.Sprintf("%v", val)
	}

	// Then check structured fields
	switch fieldName {
	case "title":
		return book.Metadata.Title
	case "album":
		return book.Metadata.Album
	case "series":
		return book.Metadata.GetValidSeries()
	case "track_title":
		return book.Metadata.TrackTitle
	case "track", "track_number":
		if book.Metadata.TrackNumber != 0 {
			return fmt.Sprintf("%d", book.Metadata.TrackNumber)
		}
	case "authors", "artist", "album_artist":
		if len(book.Metadata.Authors) > 0 {
			return strings.Join(book.Metadata.Authors, ", ")
		}
	}

	return ""
}

// renderPopup renders a centered popup selector
func (m *SettingsTableModel) renderPopup() string {
	var content strings.Builder

	// Get setting name
	settingName := m.fieldMappings[m.popupSettingIdx].Name

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	content.WriteString(titleStyle.Render("Select " + settingName) + "\n\n")

	// Options list
	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFF00")).
		Background(lipgloss.Color("#333333"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true)

	// Get current book's metadata for showing values
	var currentBook *AudioBook
	if m.metadataBookIndex < len(m.selectedBooks) {
		currentBook = &m.selectedBooks[m.metadataBookIndex]
	}

	for i, option := range m.popupOptions {
		var optionText string

		// Special handling for Layout field - show path preview
		if settingName == "Layout" && currentBook != nil {
			// Get field mapping
			fieldMapping := m.GetFieldMapping()

			// Generate a preview path for this layout using universal function
			previewPath := GenerateOutputPath(*currentBook, option, fieldMapping, "output")

			// Colorize the preview
			coloredPreview := m.colorizeOutputPath(previewPath, option)

			// Format with layout name and preview
			optionText = fmt.Sprintf("%-20s %s", option, valueStyle.Render(coloredPreview))
		} else {
			// Get the value for this field from metadata
			var fieldValue string
			if currentBook != nil {
				fieldValue = m.getMetadataFieldValue(option, currentBook)
			}

			// Format the option with value
			optionText = option
			if fieldValue != "" {
				// Truncate long values
				maxLen := 50
				if len(fieldValue) > maxLen {
					fieldValue = fieldValue[:maxLen-3] + "..."
				}
				optionText = fmt.Sprintf("%-15s %s", option, valueStyle.Render(fieldValue))
			}
		}

		if i == m.popupSelection {
			content.WriteString("  → " + selectedStyle.Render(optionText) + "\n")
		} else {
			content.WriteString("    " + normalStyle.Render(optionText) + "\n")
		}
	}

	// Footer
	content.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content.WriteString(helpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Cancel"))

	// Box the popup
	popupStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFFF00")).
		Padding(1, 2).
		Background(lipgloss.Color("#000000"))

	popup := popupStyle.Render(content.String())

	// Center the popup
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popup)
}

// View renders the UI
func (m *SettingsTableModel) View() string {
	// Header with debug info
	debugInfo := fmt.Sprintf(" [Terminal: %dx%d, Viewport: %d lines]", m.width, m.height, m.metadataViewport.Height)
	header := headerStyle.Render("⚙️ All Settings (Basic + Advanced)"+debugInfo) + "\n\n"

	// Table with border
	tableBorderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	tableView := tableBorderStyle.Render(m.table.View()) + "\n"

	// Metadata viewport with scroll indicator (no extra spacing)
	metadataPane := m.renderMetadataViewport() + "\n"

	// Output path preview (non-scrollable, very bottom)
	outputPreview := m.generateOutputPreview()

	// Footer - show different help based on focus
	helpText := ""
	if m.showPopup {
		helpText = "Selecting option..."
	} else if m.focusArea == MetadataFocus {
		helpText = "TAB: Back to settings • ←/→: Browse books • ↑/↓: Scroll • PgUp/PgDn: Scroll faster • c/n: Continue • q: Back"
	} else {
		helpText = "↑/↓: Navigate • ←/→ or Enter: Pick value • TAB: View metadata • c/n: Continue • q: Back"
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("\n" + helpText)

	baseView := header + tableView + metadataPane + outputPreview + footer

	// If popup is showing, overlay it
	if m.showPopup {
		return m.renderPopup()
	}

	return baseView
}

// ShouldAdvance returns true if Enter should advance to next screen
func (m *SettingsTableModel) ShouldAdvance() bool {
	// Don't advance if popup is showing or was just closed
	return !m.showPopup && !m.justClosedPopup
}

// GetConfig returns the current configuration
func (m *SettingsTableModel) GetConfig() map[string]string {
	config := make(map[string]string)

	// Convert unified fieldMappings to config map
	for _, fm := range m.fieldMappings {
		if fm.Name != "───────────────────" { // Skip separator
			config[fm.Name] = fm.Options[fm.Value]
		}
	}

	return config
}

// GetFieldMapping returns the field mapping configuration
func (m *SettingsTableModel) GetFieldMapping() organizer.FieldMapping {
	// Unified settings indices:
	// 0: Use Embedded Metadata
	// 1: Dry Run
	// 2: Verbose
	// 3: separator
	// 4: Layout
	// 5: Flat Mode
	// 6: Title Field
	// 7: Series Field
	// 8: Author Fields
	// 9: Track Field

	// Parse author fields
	var authorFields []string
	if len(m.fieldMappings) > 8 {
		authorFieldsOption := m.fieldMappings[8].Options[m.fieldMappings[8].Value]
		switch authorFieldsOption {
		case "authors→artist→album_artist":
			authorFields = []string{"authors", "artist", "album_artist"}
		case "authors→narrators→artist":
			authorFields = []string{"authors", "narrators", "artist"}
		case "artist→album_artist→composer":
			authorFields = []string{"artist", "album_artist", "composer"}
		case "authors only":
			authorFields = []string{"authors"}
		default:
			authorFields = []string{"authors", "artist", "album_artist"}
		}
	} else {
		authorFields = []string{"authors", "artist", "album_artist"}
	}

	return organizer.FieldMapping{
		TitleField:   m.fieldMappings[6].Options[m.fieldMappings[6].Value],
		SeriesField:  m.fieldMappings[7].Options[m.fieldMappings[7].Value],
		AuthorFields: authorFields,
		TrackField:   m.fieldMappings[9].Options[m.fieldMappings[9].Value],
	}
}
