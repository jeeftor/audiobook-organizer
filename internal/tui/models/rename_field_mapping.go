package models

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// rescanCompleteMsg is sent when rescan completes
type rescanCompleteMsg struct {
	candidates []organizer.RenameCandidate
}

// rescanErrorMsg is sent when rescan fails
type rescanErrorMsg struct {
	err error
}

// ShowCommandMsg is sent when we want to show the command to run
type ShowCommandMsg struct {
	Command   string
	DryRunCmd string
	OutCmd    string
	Config    *organizer.RenamerConfig
}

// Widget focus states
type widgetFocus int

const (
	focusFieldMappings widgetFocus = iota
	focusSampleFiles
	focusRenamePreview
)

// RenameFieldMappingModel handles field mapping configuration
type RenameFieldMappingModel struct {
	candidates []organizer.RenameCandidate
	config     *organizer.RenamerConfig
	table      table.Model
	width      int
	height     int
	inputDir   string

	// Popup state for selecting field values
	showPopup           bool
	popupOptions        []string
	popupSelection      int
	popupSettingIdx     int
	previewFieldMapping organizer.FieldMapping // Temporary mapping for preview while navigating

	// Field mapping settings
	settings []FieldMappingSetting

	// Sample metadata from first file
	sampleMetadata *organizer.Metadata

	// Metadata mode: 0=json priority, 1=embedded only, 2=flat mode
	metadataMode int

	// Scanning state
	scanning bool

	// Widget focus and navigation
	focusedWidget widgetFocus
	sampleCursor  int
	previewCursor int

	// Metadata preview - show detailed metadata
	showMetadataPreview bool
	metadataBookIndex   int // Which sample to show detailed metadata for

	// Template selection
	showTemplatePopup bool
	templateOptions   []string
	templateSelection int
	selectedTemplate  string

	// Template builder
	availableFields   []string
	templateSlots     [4]int // -1 means empty, otherwise index into availableFields
	templateCursor    int    // Which field we're currently on (0-3)
	separators        []string
	selectedSeparator int
}

// NewRenameFieldMappingModel creates a new field mapping model
func NewRenameFieldMappingModel(candidates []organizer.RenameCandidate, config *organizer.RenamerConfig) *RenameFieldMappingModel {
	// Get sample metadata from first candidate
	var sampleMetadata *organizer.Metadata
	if len(candidates) > 0 {
		sampleMetadata = &candidates[0].Metadata
	}

	// Detect available fields from sample metadata
	availableFields := detectAvailableFields(candidates)

	// Determine initial metadata mode based on actual metadata.json presence
	metadataMode := 0 // Default: json priority
	if config.UseEmbeddedMetadata {
		metadataMode = 1 // Embedded only
	} else if len(candidates) > 0 {
		// Check if any candidate has metadata.json source
		hasMetadataJson := false
		for _, candidate := range candidates {
			if candidate.Metadata.SourceType == "json" {
				hasMetadataJson = true
				break
			}
		}
		// If no metadata.json found, default to embedded mode
		if !hasMetadataJson {
			metadataMode = 1
			config.UseEmbeddedMetadata = true
		}
	}

	// Remove metadata source from settings - it's now controlled by 'm' key
	settings := []FieldMappingSetting{
		{
			Name:        "Title Field",
			Description: "Field to use for title",
			Options:     []string{"title", "album", "series"},
			Value:       0,
		},
		{
			Name:        "Series Field",
			Description: "Field to use for series",
			Options:     []string{"series", "album", "title"},
			Value:       0,
		},
		{
			Name:        "Author Fields",
			Description: "Author field priority",
			Options:     []string{"authors→artist→album_artist", "authors→narrators→artist", "artist→album_artist", "authors only"},
			Value:       0,
		},
		{
			Name:        "Track Field",
			Description: "Field to use for track number",
			Options:     getFieldOptions(availableFields, []string{"track", "track_number", "trck", "trk"}),
			Value:       0,
		},
		{
			Name:        "Disc Field",
			Description: "Field to use for disc number",
			Options:     getFieldOptions(availableFields, []string{"disc", "discnumber", "disk", "tpos"}),
			Value:       0,
		},
	}

	// Create table columns
	columns := []table.Column{
		{Title: "Setting", Width: 20},
		{Title: "Current Value", Width: 30},
		{Title: "Available Options", Width: 40},
	}

	// Create table rows
	var rows []table.Row
	for _, setting := range settings {
		optionsStr := strings.Join(setting.Options, " | ")
		if len(optionsStr) > 40 {
			optionsStr = optionsStr[:37] + "..."
		}

		rows = append(rows, table.Row{
			setting.Name,
			setting.Options[setting.Value],
			optionsStr,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(4), // Always show all 4 settings rows
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	// Available fields for template building
	templateFields := []string{"author", "series", "track", "title"}
	templateSeparators := []string{" - ", "/", "_", " ", "."}

	// Start with empty template slots
	emptySlots := [4]int{-1, -1, -1, -1}

	// Apply default field mappings if not already set
	if config.FieldMapping.TitleField == "" {
		config.FieldMapping.TitleField = "title"
	}
	if config.FieldMapping.SeriesField == "" {
		config.FieldMapping.SeriesField = "series"
	}
	if len(config.FieldMapping.AuthorFields) == 0 {
		config.FieldMapping.AuthorFields = []string{"authors", "artist", "album_artist"}
	}
	if config.FieldMapping.TrackField == "" {
		config.FieldMapping.TrackField = "track"
	}

	// Populate template slots with actual metadata
	templateSlots := emptySlots
	if sampleMetadata != nil {
		if sampleMetadata.Title != "" {
			templateSlots[0] = 3 // title
		}
		if len(sampleMetadata.Series) > 0 {
			templateSlots[1] = 1 // series
		}
		if sampleMetadata.TrackNumber > 0 {
			templateSlots[2] = 2 // track
		}
		if len(sampleMetadata.Authors) > 0 {
			templateSlots[3] = 0 // author
		}
	}

	return &RenameFieldMappingModel{
		candidates:          candidates,
		config:              config,
		table:               t,
		settings:            settings,
		sampleMetadata:      sampleMetadata,
		metadataMode:        metadataMode,
		inputDir:            config.BaseDir,
		focusedWidget:       focusFieldMappings,
		showMetadataPreview: true,
		width:               80, // Default, will be updated by WindowSizeMsg
		height:              24, // Default, will be updated by WindowSizeMsg
		availableFields:     templateFields,
		templateSlots:       emptySlots,
		templateCursor:      0,
		separators:          templateSeparators,
		selectedSeparator:   0, // Default to " - "
	}
}

// detectAvailableFields analyzes candidates to find available metadata fields
func detectAvailableFields(candidates []organizer.RenameCandidate) map[string]bool {
	fields := make(map[string]bool)

	for _, candidate := range candidates {
		m := candidate.Metadata

		if m.Title != "" {
			fields["title"] = true
		}
		if m.Album != "" {
			fields["album"] = true
		}
		if len(m.Series) > 0 {
			fields["series"] = true
		}
		if m.TrackNumber > 0 {
			fields["track"] = true
		}

		// Check RawData for additional fields
		for key := range m.RawData {
			fields[key] = true
		}
	}

	return fields
}

// getFieldOptions returns field options, prioritizing available ones
func getFieldOptions(available map[string]bool, defaults []string) []string {
	var options []string

	// Add defaults that are available
	for _, field := range defaults {
		if available[field] {
			options = append(options, field)
		}
	}

	// If no defaults available, add them anyway
	if len(options) == 0 {
		options = defaults
	}

	return options
}

// Init initializes the model
func (m *RenameFieldMappingModel) Init() tea.Cmd {
	return tea.Batch(
		tea.WindowSize(), // Request window size on init
	)
}

// renderDetailedMetadataContent renders the detailed metadata preview content (like GUI mode)
func (m *RenameFieldMappingModel) renderDetailedMetadataContent(sampleIndices []int, bookIndex int) string {
	if bookIndex >= len(sampleIndices) {
		bookIndex = 0
	}

	idx := sampleIndices[bookIndex]
	candidate := m.candidates[idx]
	metadata := candidate.Metadata

	var content strings.Builder

	// Color styles for different field types
	titleLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	authorLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	seriesLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	defaultLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAFF"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))
	// Show the actual sample number (1-based index in the full list)
	sampleNum := bookIndex + 1
	content.WriteString(titleStyle.Render(fmt.Sprintf("Metadata Preview (#%d):", sampleNum)) + "\n\n")

	// File info - truncate filename if too long
	filename := filepath.Base(candidate.CurrentPath)
	if len(filename) > 45 {
		filename = filename[:42] + "..."
	}
	content.WriteString(defaultLabelStyle.Render("File: ") + valueStyle.Render(filename) + "\n")
	content.WriteString(defaultLabelStyle.Render("Source Type: ") + valueStyle.Render(metadata.SourceType) + "\n\n")

	// Check if we're in hybrid mode (metadata.json + embedded)
	isHybridMode := false
	if metadata.SourceType == "json" {
		if _, ok := metadata.RawData["_embedded_source"].(string); ok {
			isHybridMode = true
		}
	}

	// Raw metadata fields with inline indicators (sorted alphabetically)
	rawLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	jsonFieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))     // Gold for JSON fields
	embeddedFieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00CED1")) // Turquoise for embedded

	// Show header with hybrid mode indicator if applicable
	if isHybridMode {
		content.WriteString(titleStyle.Render("Raw Metadata Fields:") + " ")
		content.WriteString(jsonFieldStyle.Render("📁 metadata.json") + " | ")
		content.WriteString(embeddedFieldStyle.Render("🎵 Embedded") + "\n")
	} else {
		content.WriteString(titleStyle.Render("Raw Metadata Fields:") + "\n")
	}

	// Use preview mapping if popup is active, otherwise use config mapping
	fieldMapping := m.config.FieldMapping
	if m.showPopup {
		fieldMapping = m.previewFieldMapping
	}

	// Collect and sort keys, including synthetic fields
	var keys []string
	excludedFields := map[string]bool{
		"chapters":    true, // Large array of chapter data
		"description": true, // Long HTML description
		"tags":        true, // Usually not needed for renaming
	}

	for key, val := range metadata.RawData {
		// Skip nil values and empty strings (but keep 0, false, etc.)
		if val == nil {
			continue
		}
		if strVal, ok := val.(string); ok && strVal == "" {
			continue
		}
		// Skip excluded fields
		if excludedFields[key] {
			continue
		}
		keys = append(keys, key)
	}

	// Add series if it exists in metadata but not in RawData
	if len(metadata.Series) > 0 {
		hasSeriesInRaw := false
		for key := range metadata.RawData {
			if key == "series" {
				hasSeriesInRaw = true
				break
			}
		}
		if !hasSeriesInRaw {
			keys = append(keys, "series")
		}
	}

	sort.Strings(keys)

	// Iterate through sorted keys
	for _, key := range keys {
		// Skip the internal _embedded_source marker
		if key == "_embedded_source" {
			continue
		}

		var val interface{}

		// Get value from RawData or synthetic field
		if key == "series" && len(metadata.Series) > 0 {
			// Use series from metadata.Series array if not in RawData
			if rawVal, ok := metadata.RawData[key]; ok && rawVal != nil && rawVal != "" {
				val = rawVal
			} else {
				val = metadata.Series[0]
			}
		} else {
			val = metadata.RawData[key]
		}

		// Determine if this field came from embedded or JSON (in hybrid mode)
		isEmbeddedField := false
		sourceIndicator := ""
		if isHybridMode {
			// File-level fields that come from embedded audio
			embeddedFields := map[string]bool{
				"track": true, "trck": true, "trk": true, "track_number": true, "track_total": true,
				"disc": true, "disk": true, "discnumber": true, "tpos": true, "disc_total": true,
			}
			if embeddedFields[key] {
				isEmbeddedField = true
				sourceIndicator = " " + embeddedFieldStyle.Render("🎵")
			} else {
				sourceIndicator = " " + jsonFieldStyle.Render("📁")
			}
		}

		// Determine if this field is used and add field mapping indicator
		fieldIndicator := ""
		if key == fieldMapping.TitleField {
			fieldIndicator = " " + titleLabelStyle.Render("<- TITLE")
		} else if key == fieldMapping.SeriesField {
			fieldIndicator = " " + seriesLabelStyle.Render("<- SERIES")
		} else if key == fieldMapping.TrackField {
			fieldIndicator = " " + defaultLabelStyle.Render("<- TRACK")
		} else if key == fieldMapping.DiscField {
			fieldIndicator = " " + defaultLabelStyle.Render("<- DISC")
		} else {
			// Check if it's in author fields
			for _, af := range fieldMapping.AuthorFields {
				if key == af {
					fieldIndicator = " " + authorLabelStyle.Render("<- AUTHOR")
					break
				}
			}
		}

		// Choose field label style based on source
		fieldLabelStyle := rawLabelStyle
		if isHybridMode {
			if isEmbeddedField {
				fieldLabelStyle = embeddedFieldStyle
			} else {
				fieldLabelStyle = jsonFieldStyle
			}
		}

		content.WriteString(fmt.Sprintf("  %s: %v%s%s\n", fieldLabelStyle.Render(key), val, sourceIndicator, fieldIndicator))
	}

	return content.String()
}

// renderDetailedMetadata renders the detailed metadata preview widget (like GUI mode)
func (m *RenameFieldMappingModel) renderDetailedMetadata(sampleIndices []int) string {
	content := m.renderDetailedMetadataContent(sampleIndices, m.metadataBookIndex)

	// Render in a box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Width(m.width - 4)

	return box.Render(content)
}

// getSampleIndices returns indices of samples to display, mixing from different directories
func (m *RenameFieldMappingModel) getSampleIndices(count int) []int {
	if len(m.candidates) == 0 {
		return []int{}
	}

	// Group candidates by directory
	dirGroups := make(map[string][]int)
	for i, candidate := range m.candidates {
		dir := filepath.Dir(candidate.CurrentPath)
		dirGroups[dir] = append(dirGroups[dir], i)
	}

	// If only one directory, just return first N indices
	if len(dirGroups) == 1 {
		indices := make([]int, 0, count)
		for i := 0; i < count && i < len(m.candidates); i++ {
			indices = append(indices, i)
		}
		return indices
	}

	// Mix samples from different directories
	indices := make([]int, 0, count)
	dirs := make([]string, 0, len(dirGroups))
	for dir := range dirGroups {
		dirs = append(dirs, dir)
	}

	// Round-robin through directories
	dirIdx := 0
	dirPositions := make(map[string]int)

	for len(indices) < count {
		dir := dirs[dirIdx]
		pos := dirPositions[dir]

		if pos < len(dirGroups[dir]) {
			indices = append(indices, dirGroups[dir][pos])
			dirPositions[dir]++
		}

		dirIdx = (dirIdx + 1) % len(dirs)

		// Check if we've exhausted all directories
		allExhausted := true
		for _, d := range dirs {
			if dirPositions[d] < len(dirGroups[d]) {
				allExhausted = false
				break
			}
		}
		if allExhausted {
			break
		}
	}

	return indices
}

// rescanFiles rescans files with the current metadata mode
func (m *RenameFieldMappingModel) rescanFiles() tea.Msg {
	// Create a new renamer with updated config
	renamer, err := organizer.NewRenamer(m.config)
	if err != nil {
		return rescanErrorMsg{err: err}
	}

	// Scan files
	candidates, err := renamer.ScanFiles()
	if err != nil {
		return rescanErrorMsg{err: err}
	}

	return rescanCompleteMsg{candidates: candidates}
}

// Update handles messages
func (m *RenameFieldMappingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle popup
	if m.showPopup {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.popupSelection > 0 {
					m.popupSelection--
					// Update preview mapping for live feedback
					m.updatePreviewMapping()
				}
				return m, nil

			case "down", "j":
				if m.popupSelection < len(m.popupOptions)-1 {
					m.popupSelection++
					// Update preview mapping for live feedback
					m.updatePreviewMapping()
				}
				return m, nil

			case "left", "h":
				// Navigate to previous metadata sample (even in popup)
				if m.metadataBookIndex > 0 {
					m.metadataBookIndex--
				}
				return m, nil

			case "right", "l":
				// Navigate to next metadata sample (even in popup)
				sampleIndices := m.getSampleIndices(10)
				if m.metadataBookIndex < len(sampleIndices)-1 {
					m.metadataBookIndex++
				}
				return m, nil

			case "enter":
				// Apply selection - strip sample data from option
				selectedOption := m.popupOptions[m.popupSelection]
				// Strip metadata value after colon
				if idx := strings.Index(selectedOption, ": "); idx != -1 {
					selectedOption = selectedOption[:idx]
				}
				// Also handle old format for backwards compatibility
				if idx := strings.Index(selectedOption, "  (e.g.,"); idx != -1 {
					selectedOption = selectedOption[:idx]
				}

				// Update the setting with clean option
				m.settings[m.popupSettingIdx].Value = m.popupSelection
				// Also update the actual option text to be clean
				m.settings[m.popupSettingIdx].Options[m.popupSelection] = selectedOption

				m.config.FieldMapping = m.buildFieldMapping()
				m.showPopup = false
				return m, nil

			case "esc", "q":
				m.showPopup = false
				return m, nil
			}
		}
		return m, nil
	}

	// Handle main screen
	switch msg := msg.(type) {
	case rescanCompleteMsg:
		// Update candidates with rescanned data
		m.candidates = msg.candidates
		if len(msg.candidates) > 0 {
			m.sampleMetadata = &msg.candidates[0].Metadata
		}
		m.scanning = false
		return m, nil

	case rescanErrorMsg:
		// Handle rescan error
		m.scanning = false
		// Could show error to user, for now just continue
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle template popup navigation
		if m.showTemplatePopup {
			switch msg.String() {
			case "up", "k":
				// Navigate up through available fields
				if m.templateCursor > 0 {
					m.templateCursor--
				}
				return m, nil

			case "down", "j":
				// Navigate down through available fields
				if m.templateCursor < len(m.availableFields)-1 {
					m.templateCursor++
				}
				return m, nil

			case "1", "2", "3", "4":
				// Assign current field to this position
				pos, _ := strconv.Atoi(msg.String())
				pos-- // Convert to 0-indexed

				// Remove this field from any existing position
				for i := range m.templateSlots {
					if m.templateSlots[i] == m.templateCursor {
						m.templateSlots[i] = -1
					}
				}

				// Shift existing fields if needed
				if m.templateSlots[pos] != -1 {
					// Position is occupied, shift everything down
					for i := 3; i > pos; i-- {
						m.templateSlots[i] = m.templateSlots[i-1]
					}
				}

				// Assign current field to position
				m.templateSlots[pos] = m.templateCursor
				return m, nil

			case "s", "S":
				// Cycle separator
				m.selectedSeparator = (m.selectedSeparator + 1) % len(m.separators)
				return m, nil

			case "enter":
				// Apply template
				m.selectedTemplate = m.buildTemplatePreview()
				m.showTemplatePopup = false
				return m, nil

			case "esc", "q":
				m.showTemplatePopup = false
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q", "esc":
			// Go back to scan
			return m, nil

		case "left", "h":
			// Navigate to previous metadata sample
			if m.metadataBookIndex > 0 {
				m.metadataBookIndex--
			}
			return m, nil

		case "right", "l":
			// Navigate to next metadata sample
			sampleIndices := m.getSampleIndices(10)
			if m.metadataBookIndex < len(sampleIndices)-1 {
				m.metadataBookIndex++
			}
			return m, nil

		case "p":
			// Show template selection popup
			m.showTemplatePopup = true
			return m, nil

		case "s":
			// Quick assign: Series field
			if !m.showPopup && !m.showTemplatePopup {
				m.showPopup = true
				m.popupSettingIdx = 1 // Series Field
				m.popupOptions = m.getOptionsWithMetadata("Series Field", m.settings[1].Options)
				m.popupSelection = m.settings[1].Value
				// Initialize preview mapping
				m.updatePreviewMapping()
			}
			return m, nil

		case "a":
			// Quick assign: Author Fields
			if !m.showPopup && !m.showTemplatePopup {
				m.showPopup = true
				m.popupSettingIdx = 2 // Author Fields
				m.popupOptions = m.getOptionsWithMetadata("Author Fields", m.settings[2].Options)
				m.popupSelection = m.settings[2].Value
				// Initialize preview mapping
				m.updatePreviewMapping()
			}
			return m, nil

		case "t":
			// Quick assign: Title/Name field
			if !m.showPopup && !m.showTemplatePopup {
				m.showPopup = true
				m.popupSettingIdx = 0 // Title Field
				m.popupOptions = m.getOptionsWithMetadata("Title Field", m.settings[0].Options)
				m.popupSelection = m.settings[0].Value
				// Initialize preview mapping
				m.updatePreviewMapping()
			}
			return m, nil

		case "o":
			// Quick assign: Track field
			if !m.showPopup && !m.showTemplatePopup {
				m.showPopup = true
				m.popupSettingIdx = 3 // Track Field
				m.popupOptions = m.getOptionsWithMetadata("Track Field", m.settings[3].Options)
				m.popupSelection = m.settings[3].Value
				// Initialize preview mapping
				m.updatePreviewMapping()
			}
			return m, nil

		case "tab":
			// Cycle through widgets
			m.focusedWidget = (m.focusedWidget + 1) % 3
			return m, nil

		case "enter", " ":
			// Show popup for current row (only if field mappings focused)
			if m.focusedWidget == focusFieldMappings {
				idx := m.table.Cursor()
				if idx < len(m.settings) {
					m.popupSettingIdx = idx
					// Populate options with actual metadata values
					m.popupOptions = m.getOptionsWithMetadata(m.settings[idx].Name, m.settings[idx].Options)
					m.popupSelection = m.settings[idx].Value
					m.showPopup = true
				}
			}
			return m, nil

		case "m":
			// Cycle through metadata modes: json priority -> embedded -> flat
			m.metadataMode = (m.metadataMode + 1) % 3
			// Update config immediately
			switch m.metadataMode {
			case 0: // JSON priority
				m.config.UseEmbeddedMetadata = false
			case 1: // Embedded only
				m.config.UseEmbeddedMetadata = true
			case 2: // Flat mode
				m.config.UseEmbeddedMetadata = true
				// TODO: Set flat mode flag when implemented
			}
			// Trigger rescan with new metadata mode
			m.scanning = true
			return m, m.rescanFiles

		case "c", "n":
			// Generate and show command instead of going to template screen
			m.applyFieldMapping()

			// Build command strings (multiple variants)
			normalCmd := m.buildCommandString(false, "")
			dryRunCmd := m.buildCommandString(true, "")
			outCmd := m.buildCommandString(false, "./organized")

			return m, func() tea.Msg {
				return ShowCommandMsg{
					Command:   normalCmd,
					DryRunCmd: dryRunCmd,
					OutCmd:    outCmd,
					Config:    m.config,
				}
			}
		}
	}

	// Update table only if field mappings focused
	if m.focusedWidget == focusFieldMappings {
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

// updateTableRow updates a specific row in the table
func (m *RenameFieldMappingModel) updateTableRow(idx int) {
	if idx >= len(m.settings) {
		return
	}

	setting := m.settings[idx]
	optionsStr := strings.Join(setting.Options, " | ")
	if len(optionsStr) > 40 {
		optionsStr = optionsStr[:37] + "..."
	}

	rows := m.table.Rows()
	rows[idx] = table.Row{
		setting.Name,
		setting.Options[setting.Value],
		optionsStr,
	}
	m.table.SetRows(rows)
}

// buildCommandString builds the CLI command string with all options
func (m *RenameFieldMappingModel) buildCommandString(forceDryRun bool, outDir string) string {
	var parts []string
	parts = append(parts, "audiobook-organizer rename")

	// Add base directory
	parts = append(parts, fmt.Sprintf("--dir \"%s\"", m.config.BaseDir))

	// Add output directory if specified
	if outDir != "" {
		parts = append(parts, fmt.Sprintf("--out \"%s\"", outDir))
	}

	// Add template
	template := m.selectedTemplate
	if template == "" {
		template = "{author} - {series} - {track} - {title}"
	}
	parts = append(parts, fmt.Sprintf("--template \"%s\"", template))

	// Add field mappings (these flags exist in cmd/metadata.go and cmd/rename.go)
	if m.config.FieldMapping.TitleField != "" && m.config.FieldMapping.TitleField != "title" {
		parts = append(parts, fmt.Sprintf("--title-field %s", m.config.FieldMapping.TitleField))
	}
	if m.config.FieldMapping.SeriesField != "" && m.config.FieldMapping.SeriesField != "series" {
		parts = append(parts, fmt.Sprintf("--series-field %s", m.config.FieldMapping.SeriesField))
	}
	if m.config.FieldMapping.TrackField != "" && m.config.FieldMapping.TrackField != "track" {
		parts = append(parts, fmt.Sprintf("--track-field %s", m.config.FieldMapping.TrackField))
	}
	if m.config.FieldMapping.DiscField != "" && m.config.FieldMapping.DiscField != "disc" {
		parts = append(parts, fmt.Sprintf("--disc-field %s", m.config.FieldMapping.DiscField))
	}
	if len(m.config.FieldMapping.AuthorFields) > 0 {
		parts = append(parts, fmt.Sprintf("--author-fields %s", strings.Join(m.config.FieldMapping.AuthorFields, ",")))
	}

	// Add author format if configured (default is first-last)
	if m.config.AuthorFormat != 0 {
		var formatStr string
		switch m.config.AuthorFormat {
		case 0: // AuthorFormatFirstLast
			formatStr = "first-last"
		case 1: // AuthorFormatLastFirst
			formatStr = "last-first"
		case 2: // AuthorFormatPreserve
			formatStr = "preserve"
		}
		if formatStr != "" && formatStr != "first-last" {
			parts = append(parts, fmt.Sprintf("--author-format %s", formatStr))
		}
	}

	// Add metadata mode
	if m.config.UseEmbeddedMetadata {
		parts = append(parts, "--use-embedded-metadata")
	}

	// Add other flags
	if m.config.Recursive {
		parts = append(parts, "--recursive")
	}
	if m.config.DryRun || forceDryRun {
		parts = append(parts, "--dry-run")
	}
	if m.config.PreservePath {
		parts = append(parts, "--preserve-path")
	}
	if m.config.StrictMode {
		parts = append(parts, "--strict")
	}
	if m.config.PromptEnabled {
		parts = append(parts, "--prompt")
	}

	return strings.Join(parts, " ")
}

// applyFieldMapping applies the selected field mapping to the config
func (m *RenameFieldMappingModel) applyFieldMapping() {
	for _, setting := range m.settings {
		selectedOption := setting.Options[setting.Value]

		switch setting.Name {
		case "Title Field":
			m.config.FieldMapping.TitleField = selectedOption

		case "Series Field":
			m.config.FieldMapping.SeriesField = selectedOption

		case "Author Fields":
			// Parse the author fields option
			m.config.FieldMapping.AuthorFields = parseAuthorFields(selectedOption)

		case "Track Field":
			m.config.FieldMapping.TrackField = selectedOption

		case "Disc Field":
			m.config.FieldMapping.DiscField = selectedOption
		}
	}
}

// buildFieldMapping builds the field mapping from the settings
func (m *RenameFieldMappingModel) buildFieldMapping() organizer.FieldMapping {
	fieldMapping := organizer.FieldMapping{}

	for _, setting := range m.settings {
		selectedOption := setting.Options[setting.Value]

		switch setting.Name {
		case "Title Field":
			fieldMapping.TitleField = selectedOption

		case "Series Field":
			fieldMapping.SeriesField = selectedOption

		case "Author Fields":
			// Parse the author fields option
			fieldMapping.AuthorFields = parseAuthorFields(selectedOption)

		case "Track Field":
			fieldMapping.TrackField = selectedOption

		case "Disc Field":
			fieldMapping.DiscField = selectedOption
		}
	}

	return fieldMapping
}

// metadataBookIndexTableRow returns the metadata book index for a given table row
func (m *RenameFieldMappingModel) metadataBookIndexTableRow(idx int) int {
	if idx >= len(m.settings) {
		return 0
	}

	switch m.settings[idx].Name {
	case "Title Field":
		return 0
	case "Series Field":
		return 1
	case "Author Fields":
		return 2
	case "Track Field":
		return 3
	default:
		return 0
	}
}

// getOptionsWithMetadata enriches options with actual metadata values from files
func (m *RenameFieldMappingModel) getOptionsWithMetadata(settingName string, baseOptions []string) []string {
	// Collect unique values from candidates
	valueMap := make(map[string][]string) // field -> sample values

	// Use the currently visible sample indices (matching what's shown in metadata preview)
	sampleIndices := m.getSampleIndices(3)
	if len(sampleIndices) > 3 {
		sampleIndices = sampleIndices[:3]
	}

	for _, idx := range sampleIndices {
		if idx >= len(m.candidates) {
			continue
		}
		meta := m.candidates[idx].Metadata

		// Collect values based on setting type
		switch settingName {
		case "Title Field":
			if meta.Title != "" {
				valueMap["title"] = append(valueMap["title"], meta.Title)
			}
			if meta.Album != "" {
				valueMap["album"] = append(valueMap["album"], meta.Album)
			}
			if len(meta.Series) > 0 {
				valueMap["series"] = append(valueMap["series"], meta.Series[0])
			}

		case "Series Field":
			if len(meta.Series) > 0 {
				valueMap["series"] = append(valueMap["series"], meta.Series[0])
			}
			if meta.Album != "" {
				valueMap["album"] = append(valueMap["album"], meta.Album)
			}
			if meta.Title != "" {
				valueMap["title"] = append(valueMap["title"], meta.Title)
			}

		case "Track Field":
			if meta.TrackNumber > 0 {
				valueMap["track"] = append(valueMap["track"], fmt.Sprintf("%d", meta.TrackNumber))
			}

		case "Author Fields":
			if len(meta.Authors) > 0 {
				valueMap["authors"] = append(valueMap["authors"], meta.Authors[0])
			}
			if val, ok := meta.RawData["artist"]; ok {
				if str, ok := val.(string); ok && str != "" {
					valueMap["artist"] = append(valueMap["artist"], str)
				}
			}
			if val, ok := meta.RawData["album_artist"]; ok {
				if str, ok := val.(string); ok && str != "" {
					valueMap["album_artist"] = append(valueMap["album_artist"], str)
				}
			}
		}
	}

	// Build enriched options
	enrichedOptions := make([]string, 0, len(baseOptions))
	for _, opt := range baseOptions {
		enriched := opt

		// For author fields, show all components
		if settingName == "Author Fields" {
			// Parse the option to show sample values
			fields := parseAuthorFields(opt)
			var samples []string
			for _, field := range fields {
				if vals, ok := valueMap[field]; ok && len(vals) > 0 {
					samples = append(samples, vals[0])
				}
			}
			if len(samples) > 0 {
				enriched = fmt.Sprintf("%s  (e.g., %s)", opt, strings.Join(samples, " → "))
			}
		} else {
			// For single fields, show value directly with colon format
			if vals, ok := valueMap[opt]; ok && len(vals) > 0 {
				sample := vals[0]
				// Don't truncate - let the display handle it
				enriched = fmt.Sprintf("%s: %s", opt, sample)
			}
		}

		enrichedOptions = append(enrichedOptions, enriched)
	}

	return enrichedOptions
}

// updatePreviewMapping updates the preview field mapping based on current popup selection
func (m *RenameFieldMappingModel) updatePreviewMapping() {
	// Start with current config mapping
	m.previewFieldMapping = m.config.FieldMapping

	// Get the selected option (strip metadata if present)
	selectedOption := m.popupOptions[m.popupSelection]
	if idx := strings.Index(selectedOption, ": "); idx != -1 {
		selectedOption = selectedOption[:idx]
	}
	if idx := strings.Index(selectedOption, "  (e.g.,"); idx != -1 {
		selectedOption = selectedOption[:idx]
	}

	// Update the appropriate field based on which popup is active
	switch m.popupSettingIdx {
	case 0: // Title Field
		m.previewFieldMapping.TitleField = selectedOption
	case 1: // Series Field
		m.previewFieldMapping.SeriesField = selectedOption
	case 2: // Author Fields
		m.previewFieldMapping.AuthorFields = parseAuthorFields(selectedOption)
	case 3: // Track Field
		m.previewFieldMapping.TrackField = selectedOption
	}
}

// renderFieldMappingSummary renders a compact summary of current field mappings
func (m *RenameFieldMappingModel) renderFieldMappingSummary() string {
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00AAFF"))
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

	var sb strings.Builder
	sb.WriteString(labelStyle.Render("⚙️  Field Mappings:") + " ")

	// Show current mappings inline
	mappings := []string{
		fmt.Sprintf("%s=%s", keyStyle.Render("t:Title"), valueStyle.Render(m.settings[0].Options[m.settings[0].Value])),
		fmt.Sprintf("%s=%s", keyStyle.Render("s:Series"), valueStyle.Render(m.settings[1].Options[m.settings[1].Value])),
		fmt.Sprintf("%s=%s", keyStyle.Render("a:Author"), valueStyle.Render(m.settings[2].Options[m.settings[2].Value])),
		fmt.Sprintf("%s=%s", keyStyle.Render("o:Track"), valueStyle.Render(m.settings[3].Options[m.settings[3].Value])),
	}

	sb.WriteString(strings.Join(mappings, " | "))
	return sb.String()
}

// parseAuthorFields converts the display string to actual field list
func parseAuthorFields(option string) []string {
	// Strip any sample data after colon
	if idx := strings.Index(option, ": "); idx != -1 {
		option = option[:idx]
	}
	// Also handle old format for backwards compatibility
	if idx := strings.Index(option, "  (e.g.,"); idx != -1 {
		option = option[:idx]
	}

	switch option {
	case "authors→artist→album_artist":
		return []string{"authors", "artist", "album_artist"}
	case "authors→narrators→artist":
		return []string{"authors", "narrators", "artist"}
	case "artist→album_artist":
		return []string{"artist", "album_artist"}
	case "authors only":
		return []string{"authors"}
	default:
		return []string{"authors", "artist", "album_artist"}
	}
}

// View renders the model
func (m *RenameFieldMappingModel) View() string {
	var sb strings.Builder

	// Title and metadata mode - PROMINENT
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Background(lipgloss.Color("#333333"))
	modeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFF00")).Background(lipgloss.Color("#333333"))

	var modeText, modeIcon string
	switch m.metadataMode {
	case 0:
		modeIcon = "📁"
		modeText = "metadata.json (priority)"
	case 1:
		modeIcon = "🎵"
		modeText = "Embedded metadata only"
	case 2:
		modeIcon = "📂"
		modeText = "Flat mode (embedded)"
	default:
		modeIcon = "📁"
		modeText = "metadata.json (priority)"
	}

	// Check if any metadata.json files were found
	jsonFilesFound := 0
	embeddedFilesFound := 0
	for _, candidate := range m.candidates {
		if candidate.Metadata.SourceType == "json" {
			jsonFilesFound++
		} else if candidate.Metadata.SourceType == "audio" {
			embeddedFilesFound++
		}
	}

	// Build status indicator
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	var statusText string
	if m.metadataMode == 0 {
		// JSON priority mode - show what was found
		if jsonFilesFound > 0 && embeddedFilesFound > 0 {
			statusText = fmt.Sprintf(" (%d JSON, %d embedded)", jsonFilesFound, embeddedFilesFound)
		} else if jsonFilesFound > 0 {
			statusText = fmt.Sprintf(" (%d JSON files)", jsonFilesFound)
		} else {
			statusText = " (⚠️  No metadata.json found - using embedded)"
		}
	} else if m.metadataMode == 1 {
		statusText = fmt.Sprintf(" (%d files)", embeddedFilesFound)
	}

	// Prominent mode indicator with debug resolution
	debugInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(fmt.Sprintf(" [%dx%d]", m.width, m.height))
	modeHeader := titleStyle.Render(" Metadata Mode: ") + modeStyle.Render(modeIcon+" "+modeText+" ") + statusStyle.Render(statusText) + debugInfo
	sb.WriteString(modeHeader + "\n\n")

	// Show popup if active
	if m.showPopup {
		// Check if we can show 3-column layout with metadata + selection
		if m.width > 200 && m.showMetadataPreview {
			// 3-column integrated layout: metadata above, selection below
			return m.render3ColumnSelectionView()
		}

		// Fallback to single-column popup overlay
		numSamples := 5
		if m.height > 40 {
			numSamples = 10
		} else if m.height > 30 {
			numSamples = 7
		}
		if numSamples > len(m.candidates) {
			numSamples = len(m.candidates)
		}
		sampleIndices := m.getSampleIndices(numSamples)

		if m.showMetadataPreview && len(sampleIndices) > 0 {
			metadataPreview := m.renderDetailedMetadata(sampleIndices)
			sb.WriteString(metadataPreview + "\n\n")
		}

		sb.WriteString(m.renderPopupWithColumns(false))
		return sb.String()
	}

	// Show template popup if active
	if m.showTemplatePopup {
		sb.WriteString(m.renderTemplatePopup())
		return sb.String()
	}

	// Determine how many samples to show
	numSamples := 5
	if m.height > 40 {
		numSamples = 10
	} else if m.height > 30 {
		numSamples = 7
	}
	if numSamples > len(m.candidates) {
		numSamples = len(m.candidates)
	}

	sampleIndices := m.getSampleIndices(numSamples)

	// Metadata Preview - now with 3-column support
	if m.showMetadataPreview && len(sampleIndices) > 0 {
		// Check if we can fit 3 columns (width > 200) or 2 columns (width > 140)
		if m.width > 200 && len(sampleIndices) >= 3 {
			// Three column layout
			meta1 := m.renderDetailedMetadataContent(sampleIndices, m.metadataBookIndex)
			meta2Idx := (m.metadataBookIndex + 1) % len(sampleIndices)
			meta3Idx := (m.metadataBookIndex + 2) % len(sampleIndices)
			meta2 := m.renderDetailedMetadataContent(sampleIndices, meta2Idx)
			meta3 := m.renderDetailedMetadataContent(sampleIndices, meta3Idx)

			boxWidth := (m.width - 10) / 3
			box := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444")).
				Width(boxWidth)

			sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
				box.Render(meta1),
				box.Render(meta2),
				box.Render(meta3),
			) + "\n\n")
		} else if m.width > 140 && len(sampleIndices) >= 2 {
			// Two column layout
			meta1 := m.renderDetailedMetadataContent(sampleIndices, m.metadataBookIndex)
			meta2Idx := (m.metadataBookIndex + 1) % len(sampleIndices)
			meta2 := m.renderDetailedMetadataContent(sampleIndices, meta2Idx)

			boxWidth := (m.width - 6) / 2
			box := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444")).
				Width(boxWidth)

			sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
				box.Render(meta1),
				box.Render(meta2),
			) + "\n\n")
		} else {
			// Single column
			sb.WriteString(m.renderDetailedMetadata(sampleIndices) + "\n\n")
		}
	}

	// Field Mappings Summary (compact inline display)
	sb.WriteString(m.renderFieldMappingSummary() + "\n\n")

	// Sample Files list
	focusIndicator := ""
	if m.focusedWidget == focusSampleFiles {
		focusIndicator = " ▶"
	}
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00AAFF")).Render("📋 Sample Files"+focusIndicator) + "\n")

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))

	// Determine which samples are currently visible in metadata preview
	visibleStart := m.metadataBookIndex
	visibleEnd := m.metadataBookIndex
	if m.width > 200 {
		visibleEnd = m.metadataBookIndex + 2 // 3 columns
	} else if m.width > 140 {
		visibleEnd = m.metadataBookIndex + 1 // 2 columns
	}
	if visibleEnd >= len(sampleIndices) {
		visibleEnd = len(sampleIndices) - 1
	}

	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)

	for i, idx := range sampleIndices {
		if idx >= len(m.candidates) {
			break
		}

		// Show actual filename
		filename := filepath.Base(m.candidates[idx].CurrentPath)

		// Highlight if this sample is currently visible in metadata preview
		style := valueStyle
		if i >= visibleStart && i <= visibleEnd {
			style = highlightStyle
		}

		// Add sample number prefix
		sb.WriteString(labelStyle.Render(fmt.Sprintf("%2d. ", i+1)) + style.Render(filename) + "\n")
	}
	sb.WriteString("\n")

	// Rename Preview
	focusIndicator = ""
	if m.focusedWidget == focusRenamePreview {
		focusIndicator = " ▶"
	}

	// Show current template
	templateDisplay := m.selectedTemplate
	if templateDisplay == "" {
		templateDisplay = "{author} - {series} - {track} - {title}"
	}
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00AAFF")).Render("👁️  Rename Preview"+focusIndicator) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(" ["+templateDisplay+"]") + "\n")

	// Color styles matching field types
	authorColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")) // Orange
	seriesColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")) // Cyan
	titleColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))  // Green
	trackColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAFF"))  // Blue

	highlightNumberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	normalNumberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	for i, idx := range sampleIndices {
		if idx >= len(m.candidates) {
			break
		}
		candidate := m.candidates[idx]
		meta := candidate.Metadata

		// Build color-coded output using the selected template
		coloredOutput := m.renderTemplateWithColors(meta, authorColor, seriesColor, titleColor, trackColor)

		// Highlight number if this sample is currently visible in metadata preview
		numberStyle := normalNumberStyle
		if i >= visibleStart && i <= visibleEnd {
			numberStyle = highlightNumberStyle
		}

		// Show: number. filename → proposed_name
		sb.WriteString(
			numberStyle.Render(fmt.Sprintf("%2d. ", i+1)) +
				coloredOutput + "\n")
	}

	// Controls
	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
		Render("t: Title | s: Series | a: Author | o: Track | p: Template | m: Mode | ←→: Samples | c: Continue | Q: Back"))

	return sb.String()
}

// renderTemplatePopup renders the template builder popup
func (m *RenameFieldMappingModel) renderTemplatePopup() string {
	var sb strings.Builder

	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(70)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Background(lipgloss.Color("#333333"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00AAFF"))

	// Title
	sb.WriteString(titleStyle.Render(" Build Output Template ") + "\n\n")

	// Available fields to select from
	sb.WriteString(labelStyle.Render("Available Fields (↑↓ to navigate, 1-4 to assign position):") + "\n")
	for i, fieldName := range m.availableFields {
		cursor := "  "
		style := normalStyle
		if i == m.templateCursor {
			cursor = "→ "
			style = selectedStyle
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(fieldName)))
	}

	sb.WriteString("\n")

	// Current template slots
	sb.WriteString(labelStyle.Render("Template Slots:") + "\n")
	for i := 0; i < 4; i++ {
		slotContent := "<empty>"
		if m.templateSlots[i] != -1 {
			slotContent = m.availableFields[m.templateSlots[i]]
		}
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, selectedStyle.Render(slotContent)))
	}

	sb.WriteString("\n")

	// Separator selection
	sb.WriteString(labelStyle.Render("Separator (press s to cycle):") + "\n")
	for i, sep := range m.separators {
		cursor := "  "
		style := normalStyle
		if i == m.selectedSeparator {
			cursor = "→ "
			style = selectedStyle
		}
		displaySep := sep
		if sep == " " {
			displaySep = "<space>"
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(displaySep)))
	}

	sb.WriteString("\n")

	// Preview
	preview := m.buildTemplatePreview()
	if preview == "" {
		preview = "<empty template>"
	}
	sb.WriteString(labelStyle.Render("Preview: ") + selectedStyle.Render(preview) + "\n")

	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
		Render("↑↓: Navigate • 1-4: Assign Position • s: Separator • Enter: Apply • Esc: Cancel"))

	return "\n\n" + popupStyle.Render(sb.String())
}

// buildTemplatePreview builds a preview of the current template
func (m *RenameFieldMappingModel) buildTemplatePreview() string {
	var parts []string
	for _, fieldIdx := range m.templateSlots {
		if fieldIdx != -1 {
			parts = append(parts, "{"+m.availableFields[fieldIdx]+"}")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, m.separators[m.selectedSeparator])
}

// renderTemplateWithColors renders a template with color-coded fields
func (m *RenameFieldMappingModel) renderTemplateWithColors(meta organizer.Metadata, authorColor, seriesColor, titleColor, trackColor lipgloss.Style) string {
	template := m.selectedTemplate
	if template == "" {
		// Default template
		template = "{author} - {series} - {track} - {title}"
	}

	// Get field mapping to use correct source fields
	fieldMapping := m.config.FieldMapping

	// Replace placeholders with colored values using mapped fields
	output := template

	// Author - use mapped author fields
	authorValue := ""
	for _, field := range fieldMapping.AuthorFields {
		if field == "authors" && len(meta.Authors) > 0 {
			authorValue = meta.Authors[0]
			break
		} else if val, ok := meta.RawData[field]; ok {
			if str, ok := val.(string); ok && str != "" {
				authorValue = str
				break
			}
		}
	}
	if authorValue != "" {
		output = strings.ReplaceAll(output, "{author}", authorColor.Render(authorValue))
	} else {
		output = strings.ReplaceAll(output, "{author}", "")
	}

	// Series - use mapped series field
	seriesValue := ""
	if fieldMapping.SeriesField == "series" {
		seriesValue = meta.GetValidSeries()
	} else if val, ok := meta.RawData[fieldMapping.SeriesField]; ok {
		if str, ok := val.(string); ok {
			seriesValue = str
		}
	}
	if seriesValue != "" {
		output = strings.ReplaceAll(output, "{series}", seriesColor.Render(seriesValue))
	} else {
		output = strings.ReplaceAll(output, "{series}", "")
	}

	// Track - use mapped track field
	trackValue := ""
	if fieldMapping.TrackField == "track" && meta.TrackNumber > 0 {
		trackValue = fmt.Sprintf("%02d", meta.TrackNumber)
	} else if val, ok := meta.RawData[fieldMapping.TrackField]; ok {
		if num, ok := val.(int); ok {
			trackValue = fmt.Sprintf("%02d", num)
		} else if str, ok := val.(string); ok {
			trackValue = str
		}
	}
	if trackValue != "" {
		output = strings.ReplaceAll(output, "{track}", trackColor.Render(trackValue))
	} else {
		output = strings.ReplaceAll(output, "{track}", "")
	}

	// Title - use mapped title field
	titleValue := ""
	if fieldMapping.TitleField == "title" {
		titleValue = meta.Title
	} else if val, ok := meta.RawData[fieldMapping.TitleField]; ok {
		if str, ok := val.(string); ok {
			titleValue = str
		}
	}
	if titleValue != "" {
		output = strings.ReplaceAll(output, "{title}", titleColor.Render(titleValue))
	} else {
		output = strings.ReplaceAll(output, "{title}", "")
	}

	// Clean up multiple separators and leading/trailing separators
	output = strings.TrimSpace(output)
	// Remove patterns like " - - " or " -  - "
	for strings.Contains(output, "  ") {
		output = strings.ReplaceAll(output, "  ", " ")
	}
	// Remove leading/trailing separators
	output = strings.Trim(output, " -_/.")

	return output
}

// renderPopupWithColumns renders the option selection popup with optional multi-column layout
func (m *RenameFieldMappingModel) renderPopupWithColumns(useMultiColumn bool) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))

	sb.WriteString(titleStyle.Render(fmt.Sprintf("Select %s:", m.settings[m.popupSettingIdx].Name)) + "\n\n")

	numOptions := len(m.popupOptions)

	if useMultiColumn && numOptions > 2 {
		// 3-column layout for wide screens with metadata preview
		numColumns := 3
		rowsPerColumn := (numOptions + numColumns - 1) / numColumns

		// Calculate column width based on longest option
		maxLen := 0
		for _, opt := range m.popupOptions {
			if len(opt) > maxLen {
				maxLen = len(opt)
			}
		}
		colWidth := maxLen + 4 // cursor + spacing

		for row := 0; row < rowsPerColumn; row++ {
			for col := 0; col < numColumns; col++ {
				idx := col*rowsPerColumn + row
				if idx < numOptions {
					cursor := "  "
					style := normalStyle
					if idx == m.popupSelection {
						cursor = "→ "
						style = selectedStyle
					}

					option := m.popupOptions[idx]
					cellText := fmt.Sprintf("%s%-*s", cursor, colWidth-2, option)
					sb.WriteString(style.Render(cellText))
				} else if col < numColumns-1 {
					sb.WriteString(strings.Repeat(" ", colWidth))
				}
			}
			sb.WriteString("\n")
		}
	} else {
		// Simple single-column layout
		for i, option := range m.popupOptions {
			cursor := "  "
			style := normalStyle
			if i == m.popupSelection {
				cursor = "→ "
				style = selectedStyle
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(option)))
		}
	}

	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
		Render("↑↓: Navigate • ←→: Samples • Enter: Select • Esc: Cancel"))

	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	return popupStyle.Render(sb.String())
}

// render3ColumnSelectionView renders 3 metadata columns with selection lists below each
func (m *RenameFieldMappingModel) render3ColumnSelectionView() string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00AAFF"))
	sb.WriteString(titleStyle.Render(fmt.Sprintf("Select %s:", m.settings[m.popupSettingIdx].Name)) + "\n\n")

	// Get the same sample indices that are currently visible in the main metadata preview
	numSamples := 3
	if m.width > 200 {
		numSamples = 3
	}
	sampleIndices := m.getSampleIndices(numSamples)

	// Make sure we have exactly 3 samples for the 3-column layout
	if len(sampleIndices) < 3 {
		// Not enough samples, fall back to single column
		return m.renderPopupWithColumns(false)
	}

	// Use only the first 3 samples (matching what's visible in metadata preview)
	sampleIndices = sampleIndices[:3]

	// Render 3 metadata columns
	boxWidth := (m.width - 10) / 3

	var metadataColumns []string
	for colIdx, fileIdx := range sampleIndices {
		// Render metadata for this specific file
		metaContent := m.renderMetadataForFile(fileIdx, colIdx+1, boxWidth)
		boxStyle := lipgloss.NewStyle().
			Width(boxWidth).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(0, 1)
		metadataColumns = append(metadataColumns, boxStyle.Render(metaContent))
	}

	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, metadataColumns...) + "\n\n")

	// Render selection lists under each column - each showing values from that file's metadata
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

	var selectionColumns []string
	for col := 0; col < 3; col++ {
		var colContent strings.Builder
		fileIdx := sampleIndices[col]

		// Get file-specific metadata for this column
		if fileIdx < len(m.candidates) {
			meta := m.candidates[fileIdx].Metadata

			for i, option := range m.popupOptions {
				cursor := "  "
				style := normalStyle
				if i == m.popupSelection {
					cursor = "→ "
					style = selectedStyle
				}

				// Extract the field name (strip metadata value if present)
				fieldName := option
				if idx := strings.Index(option, ": "); idx != -1 {
					fieldName = option[:idx]
				}

				// Get the actual value from this file's metadata for this field
				var displayValue string
				switch m.settings[m.popupSettingIdx].Name {
				case "Title Field":
					if fieldName == "title" {
						displayValue = meta.Title
					} else if fieldName == "album" {
						displayValue = meta.Album
					} else if fieldName == "series" && len(meta.Series) > 0 {
						displayValue = meta.Series[0]
					} else if val, ok := meta.RawData[fieldName]; ok {
						displayValue = fmt.Sprintf("%v", val)
					}
				case "Series Field":
					if fieldName == "series" && len(meta.Series) > 0 {
						displayValue = meta.Series[0]
					} else if fieldName == "album" {
						displayValue = meta.Album
					} else if fieldName == "title" {
						displayValue = meta.Title
					} else if val, ok := meta.RawData[fieldName]; ok {
						displayValue = fmt.Sprintf("%v", val)
					}
				case "Track Field":
					if fieldName == "track" && meta.TrackNumber > 0 {
						displayValue = fmt.Sprintf("%d", meta.TrackNumber)
					} else if val, ok := meta.RawData[fieldName]; ok {
						displayValue = fmt.Sprintf("%v", val)
					}
				case "Author Fields":
					// For author fields, just show the field name
					displayValue = fieldName
				}

				// Display format
				displayText := option
				if displayValue != "" {
					// Truncate if too long for column
					maxLen := boxWidth - 10
					if len(displayValue) > maxLen {
						displayValue = displayValue[:maxLen-3] + "..."
					}
					displayText = fmt.Sprintf("%s: %s", fieldName, displayValue)
				}

				colContent.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(displayText)))
			}
		}

		boxStyle := lipgloss.NewStyle().
			Width(boxWidth).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)
		selectionColumns = append(selectionColumns, boxStyle.Render(colContent.String()))
	}

	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, selectionColumns...) + "\n\n")

	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
		Render("↑↓: Navigate • ←→: Samples • Enter: Select • Esc: Cancel"))

	return sb.String()
}
