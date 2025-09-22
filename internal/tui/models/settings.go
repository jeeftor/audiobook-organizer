package models

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Setting represents a configurable setting
type Setting struct {
	Name        string
	Description string
	Options     []string
	Value       int
	Focused     bool
}

// FieldMappingSetting represents a field mapping option
type FieldMappingSetting struct {
	Name        string
	Description string
	Options     []string
	Value       int
	Focused     bool
}

// SettingsModel represents the settings screen
type SettingsModel struct {
	settings     []Setting
	cursor       int
	width        int
	height       int
	selectedBooks []AudioBook // Selected books from the book list
	showAdvanced bool         // Whether to show advanced field mapping options
	fieldMappings []FieldMappingSetting // Advanced field mapping settings
	fieldCursor   int          // Cursor for field mapping settings
	filterString string        // Filter string for searching settings
	filtering    bool          // Whether we're currently in filtering mode
}

// NewSettingsModel creates a new settings model
func NewSettingsModel(selectedBooks []AudioBook) *SettingsModel {
	settings := []Setting{
		{
			Name:        "Layout",
			Description: "How to organize the output directory structure",
			Options:     []string{"author-only", "author-title", "author-series-title"},
			Value:       2, // Default to author-series-title
			Focused:     false,
		},
		{
			Name:        "Use Embedded Metadata",
			Description: "Use metadata embedded in the audiobook files",
			Options:     []string{"No", "Yes"},
			Value:       1, // Default to Yes
			Focused:     false,
		},
		{
			Name:        "Flat Mode",
			Description: "Process each file individually (vs. directory-based)",
			Options:     []string{"No", "Yes"},
			Value:       0, // Default to No
			Focused:     false,
		},
		{
			Name:        "Dry Run",
			Description: "Preview changes without moving files",
			Options:     []string{"No", "Yes"},
			Value:       0, // Default to No
			Focused:     false,
		},
		{
			Name:        "Verbose",
			Description: "Show detailed output during processing",
			Options:     []string{"No", "Yes"},
			Value:       1, // Default to Yes
			Focused:     false,
		},
		{
			Name:        "Advanced Settings",
			Description: "Show advanced metadata field mapping options",
			Options:     []string{"No", "Yes"},
			Value:       0, // Default to No
			Focused:     false,
		},
	}

	// Initialize field mappings for advanced settings
	fieldMappings := []FieldMappingSetting{
		{
			Name:        "Title Field",
			Description: "Field to use as title",
			Options:     []string{"title", "album", "series", "track_title"},
			Value:       0, // Default to title
			Focused:     false,
		},
		{
			Name:        "Series Field",
			Description: "Field to use as series",
			Options:     []string{"series", "album", "title"},
			Value:       0, // Default to series
			Focused:     false,
		},
		{
			Name:        "Author Field",
			Description: "Field to use as author",
			Options:     []string{"authors", "artist", "album_artist", "composer"},
			Value:       0, // Default to authors
			Focused:     false,
		},
		{
			Name:        "Track Field",
			Description: "Field to use for track number",
			Options:     []string{"track", "track_number", "disc"},
			Value:       0, // Default to track
			Focused:     false,
		},
	}

	return &SettingsModel{
		settings:      settings,
		cursor:        0,
		selectedBooks: selectedBooks,
		showAdvanced:  false,
		fieldMappings: fieldMappings,
		fieldCursor:   0,
	}
}

// Init initializes the model
func (m *SettingsModel) Init() tea.Cmd {
	return nil
}

// formatFullMetadata formats all metadata fields for display
func formatFullMetadata(metadata *organizer.Metadata) string {
	var content strings.Builder

	// Define styles for field names and values
	fieldStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAFF"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Format primary metadata fields
	if metadata.Title != "" {
		content.WriteString(fmt.Sprintf("%s: %s\n",
			fieldStyle.Render("Title"),
			valueStyle.Render(metadata.Title)))
	}

	if len(metadata.Authors) > 0 {
		content.WriteString(fmt.Sprintf("%s: %s\n",
			fieldStyle.Render("Authors"),
			valueStyle.Render(strings.Join(metadata.Authors, ", "))))
	}

	// Use GetValidSeries() to get a string representation of the series
	if series := metadata.GetValidSeries(); series != "" {
		content.WriteString(fmt.Sprintf("%s: %s\n",
			fieldStyle.Render("Series"),
			valueStyle.Render(series)))
	}

	if metadata.TrackNumber != 0 {
		content.WriteString(fmt.Sprintf("%s: %s\n",
			fieldStyle.Render("Track Number"),
			valueStyle.Render(fmt.Sprintf("%d", metadata.TrackNumber))))
	}

	// Format additional metadata fields
	if metadata.Album != "" {
		content.WriteString(fmt.Sprintf("%s: %s\n",
			fieldStyle.Render("Album"),
			valueStyle.Render(metadata.Album)))
	}

	if metadata.TrackTitle != "" {
		content.WriteString(fmt.Sprintf("%s: %s\n",
			fieldStyle.Render("Track Title"),
			valueStyle.Render(metadata.TrackTitle)))
	}

	// Display raw metadata fields
	if len(metadata.RawData) > 0 {
		content.WriteString("\n" + fieldStyle.Render("Additional Metadata Fields:") + "\n")

		// Sort keys for consistent display
		keys := make([]string, 0, len(metadata.RawData))
		for k := range metadata.RawData {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Display each raw metadata field
		for _, key := range keys {
			// Skip fields we've already displayed
			if key == "title" || key == "album" || key == "series" ||
			   key == "track" || key == "track_number" || key == "track_title" {
				continue
			}

			value := metadata.RawData[key]
			content.WriteString(fmt.Sprintf("  %s: %s\n",
				fieldStyle.Render(key),
				valueStyle.Render(fmt.Sprintf("%v", value))))
		}
	}

	return content.String()
}

// formatFieldMapping formats the field mapping configuration for display
func formatFieldMapping(mapping organizer.FieldMapping) string {
	var content strings.Builder

	// Define styles for field names and values
	fieldStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFAAAA"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Format field mappings
	content.WriteString(fmt.Sprintf("%s: %s\n",
		fieldStyle.Render("Title Field"),
		valueStyle.Render(mapping.TitleField)))

	content.WriteString(fmt.Sprintf("%s: %s\n",
		fieldStyle.Render("Series Field"),
		valueStyle.Render(mapping.SeriesField)))

	content.WriteString(fmt.Sprintf("%s: %s\n",
		fieldStyle.Render("Author Fields"),
		valueStyle.Render(strings.Join(mapping.AuthorFields, ", "))))

	content.WriteString(fmt.Sprintf("%s: %s\n",
		fieldStyle.Render("Track Field"),
		valueStyle.Render(mapping.TrackField)))

	return content.String()
}

// GetSettings returns the current settings
func (m *SettingsModel) GetSettings() map[string]string {
	settings := make(map[string]string)
	for _, setting := range m.settings {
		settings[setting.Name] = setting.Options[setting.Value]
	}
	return settings
}

// applyFilter filters settings based on the current filter string
func (m *SettingsModel) applyFilter() {
	// If filter string is empty, do nothing
	if m.filterString == "" {
		return
	}

	// Find the first setting that matches the filter string
	for i, setting := range m.settings {
		if strings.Contains(strings.ToLower(setting.Name), strings.ToLower(m.filterString)) ||
		   strings.Contains(strings.ToLower(setting.Description), strings.ToLower(m.filterString)) {
			// Move cursor to this setting
			m.cursor = i
			return
		}
	}
}

// Update handles messages and user input
func (m *SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// Check if we're in advanced mode
		if m.showAdvanced && msg.String() != "esc" {
			// Handle advanced settings navigation
			switch msg.String() {
			case "up", "k":
				if m.fieldCursor > 0 {
					m.fieldCursor--
				}

			case "down", "j":
				if m.fieldCursor < len(m.fieldMappings)-1 {
					m.fieldCursor++
				}

			case "left", "h":
				// Decrease the value of the current field mapping
				if m.fieldMappings[m.fieldCursor].Value > 0 {
					m.fieldMappings[m.fieldCursor].Value--
				}

			case "right", "l":
				// Increase the value of the current field mapping
				if m.fieldMappings[m.fieldCursor].Value < len(m.fieldMappings[m.fieldCursor].Options)-1 {
					m.fieldMappings[m.fieldCursor].Value++
				}

			case "tab":
				// Focus/unfocus the current field mapping
				m.fieldMappings[m.fieldCursor].Focused = !m.fieldMappings[m.fieldCursor].Focused
			}
		} else if m.filtering {
			// Handle filtering mode
			switch msg.Type {
			case tea.KeyRunes:
				// Add character to filter string
				m.filterString += msg.String()
				// Apply filter and reset cursor
				m.applyFilter()

			case tea.KeyBackspace:
				// Remove character from filter string
				if len(m.filterString) > 0 {
					m.filterString = m.filterString[:len(m.filterString)-1]
					// Apply filter and reset cursor
					m.applyFilter()
				}

			case tea.KeyEsc, tea.KeyEnter:
				// Exit filtering mode
				m.filtering = false
				m.filterString = ""
			}
		} else {
			// Handle regular settings navigation
			switch msg.String() {
			case "/":
				// Enter filtering mode
				m.filtering = true
				m.filterString = ""
				return m, nil

			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}

			case "down", "j":
				if m.cursor < len(m.settings)-1 {
					m.cursor++
				}

			case "left", "h":
				// Decrease the value of the current setting
				if m.settings[m.cursor].Value > 0 {
					m.settings[m.cursor].Value--

					// If this is the Advanced Settings toggle, update showAdvanced
					if m.settings[m.cursor].Name == "Advanced Settings" {
						m.showAdvanced = m.settings[m.cursor].Value == 1
					}
				}

			case "right", "l":
				// Increase the value of the current setting
				if m.settings[m.cursor].Value < len(m.settings[m.cursor].Options)-1 {
					m.settings[m.cursor].Value++

					// If this is the Advanced Settings toggle, update showAdvanced
					if m.settings[m.cursor].Name == "Advanced Settings" {
						m.showAdvanced = m.settings[m.cursor].Value == 1
					}
				}

			case "tab":
				// Focus/unfocus the current setting
				m.settings[m.cursor].Focused = !m.settings[m.cursor].Focused
			}
		}

		// Handle ESC key to exit advanced mode
		if msg.String() == "esc" && m.showAdvanced {
			m.showAdvanced = false
			m.settings[5].Value = 0 // Reset Advanced Settings toggle
		}
	}

	return m, nil
}

// View renders the UI
func (m *SettingsModel) View() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)

	// Different header based on mode
	if m.showAdvanced {
		content.WriteString(headerStyle.Render("🔧 Advanced Metadata Field Mapping") + "\n\n")
		content.WriteString("Configure which fields to use for metadata:\n\n")
	} else {
		content.WriteString(headerStyle.Render("⚙️ Organization Settings") + "\n\n")
		content.WriteString("Configure how your audiobooks will be organized:\n\n")
	}

	// Show filter if active
	if m.filtering {
		content.WriteString(filterStyle.Render("Filter: " + m.filterString + "_") + "\n\n")
	}

	// If showing advanced settings, display field mappings
	if m.showAdvanced {
		content.WriteString("\n\nAdvanced Metadata Field Mappings:\n")

		for i, field := range m.fieldMappings {
			// Cursor indicator
			cursor := " "
			if i == m.fieldCursor {
				cursor = ">"
			}

			// Determine styles based on cursor position and focus
			var nameStyle, valueStyle lipgloss.Style

			if i == m.fieldCursor {
				nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFF00"))
				valueStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
			} else {
				nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
				valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAFFAA"))
			}

			// Field name and description
			content.WriteString(fmt.Sprintf("%s %s: %s\n",
				cursor,
				nameStyle.Render(field.Name),
				field.Description))

			// Field options
			optionsStr := ""
			for j, option := range field.Options {
				if j == field.Value {
					optionsStr += "[" + valueStyle.Render(option) + "] "
				} else {
					optionsStr += "[ " + option + " ] "
				}
			}
			content.WriteString("  " + optionsStr + "\n\n")
		}
	} else {
		// Regular settings
		for i, setting := range m.settings {
			// Determine styles based on cursor position and focus
			var nameStyle, valueStyle lipgloss.Style

			if i == m.cursor {
				nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFF00"))
				valueStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
			} else {
				nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
				valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAFFAA"))
			}

			// Cursor indicator
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}

			// Setting name and description
			content.WriteString(fmt.Sprintf("%s %s: %s\n",
				cursor,
				nameStyle.Render(setting.Name),
				setting.Description))

			// Setting options
			optionsStr := ""
			for j, option := range setting.Options {
				if j == setting.Value {
					optionsStr += "[" + valueStyle.Render(option) + "] "
				} else {
					optionsStr += "[ " + option + " ] "
				}
			}
			content.WriteString("  " + optionsStr + "\n\n")
		}
	}

	// Show preview of output paths and full metadata based on current settings
	if len(m.selectedBooks) > 0 {
		// Show different preview based on mode
		if m.showAdvanced {
			// First show the output path preview
			content.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render("Preview of Output Paths:") + "\n\n")

			// Get current layout setting
			layoutSetting := m.settings[0].Options[m.settings[0].Value]
			// Whether to use embedded metadata (affects display only)
			embeddedMetadataEnabled := m.settings[1].Value == 1 // Yes is index 1

			// Create field mapping from current settings
			fieldMapping := organizer.FieldMapping{
				TitleField:   m.fieldMappings[0].Options[m.fieldMappings[0].Value],
				SeriesField:  m.fieldMappings[1].Options[m.fieldMappings[1].Value],
				AuthorFields: []string{m.fieldMappings[2].Options[m.fieldMappings[2].Value]},
				TrackField:   m.fieldMappings[3].Options[m.fieldMappings[3].Value],
			}

			// Show preview for up to 3 books
			previewCount := 3
			if len(m.selectedBooks) < previewCount {
				previewCount = len(m.selectedBooks)
			}

			for i := 0; i < previewCount; i++ {
				book := m.selectedBooks[i]

				// Get filename for display
				filename := filepath.Base(book.Path)

				// Generate output path based on current settings and field mapping
				outputPath := GenerateOutputPathWithLayout(book, layoutSetting, embeddedMetadataEnabled)

				// Display book info and output path
				content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#AAFFAA")).Render(filename) + "\n")
				content.WriteString("  → " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAAAA")).Render(outputPath) + "\n")

				// Add metadata info
				authors := "Unknown"
				if len(book.Metadata.Authors) > 0 {
					authors = strings.Join(book.Metadata.Authors, ", ")
				}

				series := "None"
				if validSeries := book.Metadata.GetValidSeries(); validSeries != "" {
					series = validSeries
				}

				content.WriteString(fmt.Sprintf("  Author: %s | Series: %s\n\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAFF")).Render(authors),
					lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAAFF")).Render(series)))
			}

			// Then show the metadata and field mapping preview
			content.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render("Full Metadata Preview:") + "\n\n")

			// Show full metadata for the first selected book
			if len(m.selectedBooks) > 0 {
				book := m.selectedBooks[0]

				// Get filename for display
				filename := filepath.Base(book.Path)
				content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAFFAA")).Render(filename) + "\n\n")

				// Display all available metadata fields
				content.WriteString(formatFullMetadata(&book.Metadata) + "\n")

				// Show field mapping preview
				content.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFF00")).Render("Field Mapping Preview:") + "\n")

				// Display field mapping preview
				content.WriteString(formatFieldMapping(fieldMapping) + "\n")
			}
		} else {
			content.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render("Preview of Output Paths:") + "\n\n")

			// Get current layout setting
			layoutSetting := m.settings[0].Options[m.settings[0].Value]
			// Whether to use embedded metadata (affects display only)
			embeddedMetadataEnabled := m.settings[1].Value == 1 // Yes is index 1
			flatMode := m.settings[2].Value == 1 // Yes is index 1

			// Show preview for up to 3 books
			previewCount := 3
			if len(m.selectedBooks) < previewCount {
				previewCount = len(m.selectedBooks)
			}

			for i := 0; i < previewCount; i++ {
				book := m.selectedBooks[i]

				// Get filename for display
				filename := filepath.Base(book.Path)

				// Generate output path based on current settings
				outputPath := GenerateOutputPathWithLayout(book, layoutSetting, embeddedMetadataEnabled)

				// Display book info and output path
				content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#AAFFAA")).Render(filename) + "\n")
				content.WriteString("  → " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAAAA")).Render(outputPath) + "\n")

				// Add metadata info
				authors := "Unknown"
				if len(book.Metadata.Authors) > 0 {
					authors = strings.Join(book.Metadata.Authors, ", ")
				}

				series := "None"
				if validSeries := book.Metadata.GetValidSeries(); validSeries != "" {
					series = validSeries
				}

				content.WriteString(fmt.Sprintf("  Author: %s | Series: %s\n\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAFF")).Render(authors),
					lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAAFF")).Render(series)))
			}

			// Add note about flat mode if enabled
			if flatMode {
				content.WriteString(lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#FFFF00")).Render("Note: Flat mode is enabled - each file will be processed individually") + "\n\n")
			}
		}
	}

	// Footer with help text
	footerText := "\n↑/↓: Navigate • ←/→: Change value • Enter: Continue • q: Back"
	if m.showAdvanced {
		footerText = "\n↑/↓: Navigate • ←/→: Change value • Esc: Back to settings • Enter: Continue"
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render(footerText)

	content.WriteString(footer)

	return content.String()
}

// GetFieldMapping returns the current field mapping configuration
func (m *SettingsModel) GetFieldMapping() organizer.FieldMapping {
	// Create default field mapping
	defaultMapping := organizer.FieldMapping{
		TitleField:   "title",
		SeriesField:  "series",
		AuthorFields: []string{"authors", "artist", "album_artist"},
		TrackField:   "track",
	}

	// If advanced mode is enabled, use the custom field mappings
	if m.showAdvanced {
		return organizer.FieldMapping{
			TitleField:   m.fieldMappings[0].Options[m.fieldMappings[0].Value],
			SeriesField:  m.fieldMappings[1].Options[m.fieldMappings[1].Value],
			AuthorFields: []string{m.fieldMappings[2].Options[m.fieldMappings[2].Value]},
			TrackField:   m.fieldMappings[3].Options[m.fieldMappings[3].Value],
		}
	}

	return defaultMapping
}

// GetConfig returns the current configuration as a map
func (m *SettingsModel) GetConfig() map[string]string {
	config := make(map[string]string)

	// Add settings to config
	for _, setting := range m.settings {
		config[setting.Name] = setting.Options[setting.Value]
	}

	// Add field mappings if advanced mode is enabled
	if m.showAdvanced {
		// Add field mappings to config
		config["Title Field"] = m.fieldMappings[0].Options[m.fieldMappings[0].Value]
		config["Series Field"] = m.fieldMappings[1].Options[m.fieldMappings[1].Value]
		config["Author Field"] = m.fieldMappings[2].Options[m.fieldMappings[2].Value]
		config["Track Field"] = m.fieldMappings[3].Options[m.fieldMappings[3].Value]
	}

	return config
}

// GenerateOutputPathWithLayout creates a preview of the output path based on the selected layout
func GenerateOutputPathWithLayout(book AudioBook, layout string, useEmbeddedMetadata bool) string {
	// Get filename for fallback
	base := filepath.Base(book.Path)
	fileTitle := strings.TrimSuffix(base, filepath.Ext(base))

	// Get metadata values with fallbacks
	author := "Unknown"
	if len(book.Metadata.Authors) > 0 {
		author = book.Metadata.Authors[0]
	}

	// Prefer filename over generic series title
	title := fileTitle
	if useEmbeddedMetadata && book.Metadata.Title != "" {
		title = book.Metadata.Title
	}

	series := ""
	if validSeries := book.Metadata.GetValidSeries(); validSeries != "" {
		series = validSeries
	}

	// Generate path based on layout
	switch layout {
	case "author-only":
		return filepath.Join(author, title)
	case "author-title":
		return filepath.Join(author, title)
	case "author-series-title":
		if series != "" {
			return filepath.Join(author, series, title)
		}
		return filepath.Join(author, title)
	default:
		return filepath.Join(author, title)
	}
}
