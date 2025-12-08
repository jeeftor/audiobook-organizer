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

// Setting represents a configurable setting
type Setting struct {
	Name        string
	Description string
	Options     []string
	Value       int
	Focused     bool
}

// FieldMappingSetting represents a field mapping configuration
type FieldMappingSetting struct {
	Name        string
	Description string
	Options     []string
	Value       int
	Focused     bool
}

// SettingsTableModel represents the redesigned settings screen with table layout
type SettingsTableModel struct {
	table             table.Model
	metadataViewport  viewport.Model
	metadataWidget    *MetadataWidget
	pathPreviewWidget *PathPreviewWidget
	selectedBooks     []AudioBook
	width             int
	height            int
	showAdvanced      bool
	focusArea         FocusArea

	// Popup state
	showPopup       bool
	popupOptions    []string
	popupSelection  int
	popupSettingIdx int
	justClosedPopup bool

	// Settings values
	settings      []Setting
	fieldMappings []FieldMappingSetting

	// Scan mode (set during scan, not changeable)
	scanMode string
}

// NewSettingsTableModel creates a new table-based settings model
func NewSettingsTableModel(selectedBooks []AudioBook, showAdvanced bool) *SettingsTableModel {
	return NewSettingsTableModelWithMode(selectedBooks, showAdvanced, "Embedded")
}

// NewSettingsTableModelWithMode creates a new table-based settings model with scan mode
func NewSettingsTableModelWithMode(selectedBooks []AudioBook, showAdvanced bool, scanMode string) *SettingsTableModel {
	// Create settings (same as before)
	settings := []Setting{
		{Name: "Layout", Description: "Directory structure", Options: []string{"author-only", "author-title", "author-series-title", "author-series-title-number"}, Value: 2},
		{Name: "Dry Run", Description: "Preview without moving", Options: []string{"No", "Yes"}, Value: 0},
		{Name: "Verbose", Description: "Detailed output", Options: []string{"No", "Yes"}, Value: 1},
	}

	fieldMappings := []FieldMappingSetting{
		{Name: "Layout", Description: "Directory structure", Options: []string{"author-only", "author-title", "author-series-title", "author-series-title-number"}, Value: 2},
		{Name: "Title Field", Description: "Field for title", Options: []string{"title", "album", "series", "track_title"}, Value: 0},
		{Name: "Series Field", Description: "Field for series", Options: []string{"series", "album", "title"}, Value: 0},
		{Name: "Author Fields", Description: "Author priority", Options: []string{"authors→artist→album_artist", "authors→narrators→artist", "artist→album_artist→composer", "authors only"}, Value: 0},
		{Name: "Track Field", Description: "Field for track", Options: []string{"track", "track_number", "disc"}, Value: 0},
	}

	// Combine ALL settings into one unified view
	// Always show all settings (basic + advanced field mappings)
	// Note: Scan mode (Flat/Embedded/Normal) is set during scan and shown in header, not changeable here
	allSettings := []FieldMappingSetting{
		// Basic settings converted to FieldMappingSetting format
		{Name: "Dry Run", Description: "Preview without moving", Options: []string{"No", "Yes"}, Value: 0},
		{Name: "Verbose", Description: "Detailed output", Options: []string{"No", "Yes"}, Value: 1},
		{Name: "───────────────────", Description: "separator", Options: []string{""}, Value: 0}, // Visual separator
		// Layout and structure settings
		{Name: "Layout", Description: "Directory structure", Options: []string{"author-only", "author-title", "author-series-title", "author-series-title-number"}, Value: 2},
		{Name: "───────────────────", Description: "separator", Options: []string{""}, Value: 0}, // Visual separator
		// Rename options
		{Name: "Add Track Numbers", Description: "Prefix with track #", Options: []string{"No", "Yes"}, Value: 0},
		{Name: "Rename Files", Description: "Rename using pattern", Options: []string{"No", "Yes"}, Value: 0},
		{Name: "Rename Pattern", Description: "Pattern for rename", Options: []string{"{track} - {title}", "{title}", "{track}. {title}", "{author} - {title}"}, Value: 0},
		{Name: "───────────────────", Description: "separator", Options: []string{""}, Value: 0}, // Visual separator
		// Field mapping settings
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
		table.WithHeight(15), // Show all rows including new rename options
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
		scanMode:         scanMode,
	}

	// Initialize metadata widget
	m.metadataWidget = NewMetadataWidget(selectedBooks, m.GetFieldMapping())

	// Initialize path preview widget
	m.pathPreviewWidget = NewPathPreviewWidget(selectedBooks, m.GetFieldMapping())

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
		usedLines := 3 + 12 + 2 + 1 + 5 + 2 // = 25

		// Give remaining space to metadata viewport, but cap to avoid whitespace
		// The viewport's border/padding is handled by its own Style, not counted here
		metadataHeight := msg.Height - usedLines

		// Cap at a reasonable maximum - 22 lines is enough for ~20 metadata fields
		if metadataHeight > 22 {
			metadataHeight = 22
		}

		// Ensure reasonable minimum
		if metadataHeight < 8 {
			metadataHeight = 8
		}

		vpWidth := msg.Width - 4 // Account for viewport borders

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

		// Handle c globally to advance to next screen - ALWAYS works regardless of popup
		// Let it pass through to main.go (don't consume it)
		if keyStr == "c" {
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
				m.metadataWidget.PrevBook()
				m.updateMetadata()
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
				m.metadataWidget.NextBook()
				m.updateMetadata()
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

		case "p":
			// Navigate to previous file (works from any focus area)
			m.metadataWidget.PrevBook()
			m.updateMetadata()

		case "n":
			// Navigate to next file (works from any focus area)
			m.metadataWidget.NextBook()
			m.updateMetadata()

		case "P":
			// Navigate to previous book group (works from any focus area)
			m.metadataWidget.PrevBookGroup()
			m.updateMetadata()

		case "N":
			// Navigate to next book group (works from any focus area)
			m.metadataWidget.NextBookGroup()
			m.updateMetadata()

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
	authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9500"))    // Orange for author
	seriesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF"))    // Cyan for series
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))     // Green for title
	fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))      // Gray for filename
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")) // Gray for /

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

// generateOutputPreview generates a simple preview of output paths using the widget
func (m *SettingsTableModel) generateOutputPreview() string {
	// Update widget settings from current field mappings
	m.updatePathPreviewWidget()

	// Use widget to render compact preview
	return m.pathPreviewWidget.RenderCompactPreview(3)
}

// updatePathPreviewWidget syncs widget settings with current field mappings
func (m *SettingsTableModel) updatePathPreviewWidget() {
	// Get layout setting (index 3 in unified list)
	layout := "author-series-title"
	if len(m.fieldMappings) > 3 {
		layout = m.fieldMappings[3].Options[m.fieldMappings[3].Value]
	}

	// Get Add Track Numbers setting (index 5)
	addTrackNumbers := false
	if len(m.fieldMappings) > 5 {
		addTrackNumbers = m.fieldMappings[5].Value == 1
	}

	// Get Rename Files setting (index 6)
	renameFiles := false
	if len(m.fieldMappings) > 6 {
		renameFiles = m.fieldMappings[6].Value == 1
	}

	// Get Rename Pattern (index 7)
	renamePattern := "{track} - {title}"
	if len(m.fieldMappings) > 7 {
		renamePattern = m.fieldMappings[7].Options[m.fieldMappings[7].Value]
	}

	// Update widget
	m.pathPreviewWidget.SetLayout(layout)
	m.pathPreviewWidget.SetFieldMapping(m.GetFieldMapping())
	m.pathPreviewWidget.SetAddTrackNumbers(addTrackNumbers)
	m.pathPreviewWidget.SetRenameFiles(renameFiles)
	m.pathPreviewWidget.SetRenamePattern(renamePattern)
}

// updateMetadata updates the metadata viewport content using the widget
func (m *SettingsTableModel) updateMetadata() {
	// Update widget's field mapping to reflect current settings
	m.metadataWidget.SetFieldMapping(m.GetFieldMapping())

	// Sync path preview widget's current index with metadata widget
	m.pathPreviewWidget.SetCurrentIndex(m.metadataWidget.CurrentIndex())

	// Update path preview widget settings
	m.updatePathPreviewWidget()

	// Get content from widget and set it in viewport
	m.metadataViewport.SetContent(m.metadataWidget.View())

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

// generateRenamePreview generates a preview of what the renamed file would look like
func (m *SettingsTableModel) generateRenamePreview(pattern string, book *AudioBook) string {
	// Get track number from metadata
	trackNum := book.TrackNumber
	if trackNum == 0 {
		trackNum = book.Metadata.TrackNumber
	}
	if trackNum == 0 {
		// Try raw metadata
		if rawTrack, ok := book.Metadata.RawData["track"].(float64); ok {
			trackNum = int(rawTrack)
		} else if rawTrack, ok := book.Metadata.RawData["track_number"].(float64); ok {
			trackNum = int(rawTrack)
		}
	}
	if trackNum == 0 {
		trackNum = 1 // Default for preview
	}

	// Get title
	title := book.Metadata.Title
	if title == "" {
		base := filepath.Base(book.Path)
		title = strings.TrimSuffix(base, filepath.Ext(base))
	}

	// Get author
	author := book.Metadata.GetFirstAuthor("Unknown")

	// Get extension from original file
	ext := filepath.Ext(book.Path)

	// Apply pattern
	result := pattern
	result = strings.ReplaceAll(result, "{track}", fmt.Sprintf("%02d", trackNum))
	result = strings.ReplaceAll(result, "{title}", title)
	result = strings.ReplaceAll(result, "{author}", author)

	return result + ext
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

	content.WriteString(titleStyle.Render("Select "+settingName) + "\n\n")

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
	currentBook := m.metadataWidget.CurrentBook()

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
			optionText = fmt.Sprintf("%-25s %s", option, coloredPreview)
		} else if settingName == "Rename Pattern" && currentBook != nil {
			// Show preview of renamed file for each pattern
			previewName := m.generateRenamePreview(option, currentBook)
			optionText = fmt.Sprintf("%-20s → %s", option, valueStyle.Render(previewName))
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
	// Header with debug info and scan mode
	debugInfo := fmt.Sprintf(" [Terminal: %dx%d, Viewport: %d lines]", m.width, m.height, m.metadataViewport.Height)

	// Scan mode styling
	modeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	modeDescStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Italic(true)

	modeInfo := modeStyle.Render(fmt.Sprintf(" [Scan Mode: %s]", m.scanMode))
	switch m.scanMode {
	case "Flat":
		modeInfo += " " + modeDescStyle.Render("(files grouped by metadata)")
	case "Embedded":
		modeInfo += " " + modeDescStyle.Render("(files grouped by directory)")
	case "Normal":
		modeInfo += " " + modeDescStyle.Render("(using metadata.json)")
	}

	header := headerStyle.Render("⚙️ All Settings (Basic + Advanced)"+debugInfo) + modeInfo + "\n\n"

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
		helpText = "TAB: Back to settings • p/n: Prev/Next file • P/N: Prev/Next book • ↑/↓: Scroll • c: Continue • q: Back"
	} else {
		helpText = "↑/↓: Navigate • ←/→ or Enter: Pick value • p/n: Files • P/N: Books • TAB: Metadata • c: Continue • q: Back"
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

// GetSelectedBooks returns the selected books
func (m *SettingsTableModel) GetSelectedBooks() []AudioBook {
	return m.selectedBooks
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
	// Unified settings indices (after adding rename options):
	// 0: Use Embedded Metadata
	// 1: Dry Run
	// 2: Verbose
	// 3: separator
	// 4: Layout
	// 5: Flat Mode
	// 6: separator
	// 7: Add Track Numbers
	// 8: Rename Files
	// 9: Rename Pattern
	// 10: separator
	// 11: Title Field
	// 12: Series Field
	// 13: Author Fields
	// 14: Track Field

	// Find settings by name for robustness
	var titleField, seriesField, trackField string
	var authorFields []string

	for _, fm := range m.fieldMappings {
		switch fm.Name {
		case "Title Field":
			titleField = fm.Options[fm.Value]
		case "Series Field":
			seriesField = fm.Options[fm.Value]
		case "Track Field":
			trackField = fm.Options[fm.Value]
		case "Author Fields":
			authorFieldsOption := fm.Options[fm.Value]
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
		}
	}

	// Defaults if not found
	if titleField == "" {
		titleField = "title"
	}
	if seriesField == "" {
		seriesField = "series"
	}
	if trackField == "" {
		trackField = "track"
	}
	if len(authorFields) == 0 {
		authorFields = []string{"authors", "artist", "album_artist"}
	}

	return organizer.FieldMapping{
		TitleField:   titleField,
		SeriesField:  seriesField,
		AuthorFields: authorFields,
		TrackField:   trackField,
	}
}
