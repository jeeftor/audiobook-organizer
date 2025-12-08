package models

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// RescanRequestMsg is sent when user requests a rescan with different mode
type RescanRequestMsg struct {
	FlatMode    bool
	UseEmbedded bool
}

// BookGroup represents a group of related audiobook files
type BookGroup struct {
	Key            string              // Unique key for the group (author-series-title)
	Title          string              // Display title
	Author         string              // Author name
	Series         string              // Series name (if any)
	Files          []AudioBook         // Files in this group
	Expanded       bool                // Whether the group is expanded
	Selected       bool                // Whether the group is selected for processing
	SampleMetadata *organizer.Metadata // Sample metadata from first file
}

// BookGroupModel represents the grouped book view with metadata preview
type BookGroupModel struct {
	groups         []BookGroup
	filteredGroups []BookGroup // Groups after filter applied
	allBooks       []AudioBook
	cursor         int
	scrollOffset   int
	width          int
	height         int
	showMetadata   bool   // Whether to show detailed metadata
	selectedGroup  int    // Currently selected group for metadata view
	scanMode       string // Current scan mode (for display)
	inputDir       string // Input directory path
	outputDir      string // Output directory path
	fallbackMsg    string // Message shown when mode fallback occurred
	filtering      bool   // Whether filter input is active
	filterText     string // Current filter text
}

// NewBookGroupModel creates a new book group model from scanned books
func NewBookGroupModel(books []AudioBook) *BookGroupModel {
	return NewBookGroupModelWithMode(books, "Normal", "", "", "")
}

// NewBookGroupModelWithMode creates a new book group model with scan mode info
func NewBookGroupModelWithMode(books []AudioBook, scanMode, inputDir, outputDir, fallbackMsg string) *BookGroupModel {
	m := &BookGroupModel{
		allBooks:     books,
		showMetadata: false,
		scanMode:     scanMode,
		inputDir:     inputDir,
		outputDir:    outputDir,
		fallbackMsg:  fallbackMsg,
	}

	// In Flat mode, each file is its own group
	if scanMode == "Flat" {
		m.groupBooksFlat()
	} else {
		m.groupBooks()
	}
	m.applyFilter() // Initialize filteredGroups
	return m
}

// groupBooksFlat groups books by identical metadata (Flat mode)
// In flat mode, files from different books may be in the same directory,
// so we group by metadata similarity (author + album/title) not by directory
func (m *BookGroupModel) groupBooksFlat() {
	groupMap := make(map[string]*BookGroup)

	for _, book := range m.allBooks {
		author := book.Metadata.GetFirstAuthor("Unknown Author")
		title := book.Metadata.Title
		if title == "" {
			title = filepath.Base(book.Path)
		}

		// Get album for grouping
		album := book.Metadata.Album
		if album == "" {
			if rawAlbum, ok := book.Metadata.RawData["album"].(string); ok {
				album = rawAlbum
			}
		}

		// In flat mode, group by author + album (or title if no album)
		// This groups files with IDENTICAL metadata together
		var key string
		var displayTitle string
		if album != "" {
			key = fmt.Sprintf("%s|%s", author, album)
			displayTitle = album
		} else {
			key = fmt.Sprintf("%s|%s", author, title)
			displayTitle = title
		}

		group, exists := groupMap[key]
		if !exists {
			group = &BookGroup{
				Key:            key,
				Title:          displayTitle,
				Author:         author,
				Series:         book.Metadata.GetValidSeries(),
				Files:          []AudioBook{},
				Expanded:       false,
				Selected:       true,
				SampleMetadata: &book.Metadata,
			}
			groupMap[key] = group
		}
		group.Files = append(group.Files, book)
	}

	// Convert map to slice
	m.groups = make([]BookGroup, 0, len(groupMap))
	for _, group := range groupMap {
		// Sort files within group by track number
		sort.Slice(group.Files, func(i, j int) bool {
			if group.Files[i].TrackNumber != group.Files[j].TrackNumber {
				return group.Files[i].TrackNumber < group.Files[j].TrackNumber
			}
			return group.Files[i].Path < group.Files[j].Path
		})
		m.groups = append(m.groups, *group)
	}

	// Sort by author, then title
	sort.Slice(m.groups, func(i, j int) bool {
		if m.groups[i].Author != m.groups[j].Author {
			return m.groups[i].Author < m.groups[j].Author
		}
		return m.groups[i].Title < m.groups[j].Title
	})
}

// applyFilter filters groups based on filterText
func (m *BookGroupModel) applyFilter() {
	if m.filterText == "" {
		m.filteredGroups = m.groups
		return
	}

	filter := strings.ToLower(m.filterText)
	m.filteredGroups = []BookGroup{}

	for _, g := range m.groups {
		// Match against author, title, or series
		if strings.Contains(strings.ToLower(g.Author), filter) ||
			strings.Contains(strings.ToLower(g.Title), filter) ||
			strings.Contains(strings.ToLower(g.Series), filter) {
			m.filteredGroups = append(m.filteredGroups, g)
		}
	}

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filteredGroups) {
		m.cursor = 0
		m.scrollOffset = 0
	}
}

// groupBooks groups books by DIRECTORY (Non-Flat/Embedded mode)
// In non-flat mode, each directory = one book, all files in same dir are grouped together
// Metadata is taken from the first file in each directory
func (m *BookGroupModel) groupBooks() {
	// Map to collect books by directory
	groupMap := make(map[string]*BookGroup)

	for _, book := range m.allBooks {
		// Group by directory - all files in same directory = one book
		dirPath := filepath.Dir(book.Path)

		group, exists := groupMap[dirPath]
		if !exists {
			// Use metadata from first file in directory
			author := book.Metadata.GetFirstAuthor("Unknown Author")
			series := book.Metadata.GetValidSeries()
			title := book.Metadata.Title

			// Get album for display title
			album := book.Metadata.Album
			if album == "" {
				if rawAlbum, ok := book.Metadata.RawData["album"].(string); ok {
					album = rawAlbum
				}
			}

			// Determine display title
			displayTitle := album
			if displayTitle == "" {
				displayTitle = series
			}
			if displayTitle == "" {
				displayTitle = title
			}
			if displayTitle == "" {
				displayTitle = filepath.Base(dirPath)
			}

			group = &BookGroup{
				Key:            dirPath,
				Title:          displayTitle,
				Author:         author,
				Series:         series,
				Files:          []AudioBook{},
				Expanded:       false,
				Selected:       true,
				SampleMetadata: &book.Metadata,
			}
			groupMap[dirPath] = group
		}

		group.Files = append(group.Files, book)
	}

	// Convert map to slice and sort
	m.groups = make([]BookGroup, 0, len(groupMap))
	for _, group := range groupMap {
		// Sort files within group by track number or filename
		sort.Slice(group.Files, func(i, j int) bool {
			if group.Files[i].TrackNumber != group.Files[j].TrackNumber {
				return group.Files[i].TrackNumber < group.Files[j].TrackNumber
			}
			return group.Files[i].Path < group.Files[j].Path
		})
		m.groups = append(m.groups, *group)
	}

	// Sort groups by author, then title
	sort.Slice(m.groups, func(i, j int) bool {
		if m.groups[i].Author != m.groups[j].Author {
			return m.groups[i].Author < m.groups[j].Author
		}
		return m.groups[i].Title < m.groups[j].Title
	})
}

// Init initializes the model
func (m *BookGroupModel) Init() tea.Cmd {
	// Request window size on init
	return tea.WindowSize()
}

// Update handles messages and user input
func (m *BookGroupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// Handle filter input mode
		if m.filtering {
			switch msg.String() {
			case "enter":
				// Exit filter mode, keep filter applied
				m.filtering = false
				return m, nil
			case "esc":
				// Cancel filter, clear text
				m.filtering = false
				m.filterText = ""
				m.applyFilter()
				return m, nil
			case "backspace":
				if len(m.filterText) > 0 {
					m.filterText = m.filterText[:len(m.filterText)-1]
					m.applyFilter()
				}
				return m, nil
			default:
				// Add character to filter (only printable chars)
				if len(msg.String()) == 1 {
					m.filterText += msg.String()
					m.applyFilter()
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "/": // Start filtering
			m.filtering = true
			m.filterText = ""
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}
		case "down", "j":
			if m.cursor < len(m.filteredGroups)-1 {
				m.cursor++
				m.adjustScroll()
			}
		case " ": // Toggle selection
			if m.cursor < len(m.filteredGroups) {
				// Find the actual group in m.groups and toggle it
				for i := range m.groups {
					if m.groups[i].Key == m.filteredGroups[m.cursor].Key {
						m.groups[i].Selected = !m.groups[i].Selected
						m.filteredGroups[m.cursor].Selected = m.groups[i].Selected
						break
					}
				}
			}
		case "tab", "e": // Expand/collapse
			if m.cursor < len(m.filteredGroups) {
				for i := range m.groups {
					if m.groups[i].Key == m.filteredGroups[m.cursor].Key {
						m.groups[i].Expanded = !m.groups[i].Expanded
						m.filteredGroups[m.cursor].Expanded = m.groups[i].Expanded
						break
					}
				}
			}
		case "m": // Toggle metadata view
			m.showMetadata = !m.showMetadata
			m.selectedGroup = m.cursor
		case "A": // Select all filtered groups
			for i := range m.filteredGroups {
				for j := range m.groups {
					if m.groups[j].Key == m.filteredGroups[i].Key {
						m.groups[j].Selected = true
						m.filteredGroups[i].Selected = true
						break
					}
				}
			}
		case "N": // Deselect all filtered groups
			for i := range m.filteredGroups {
				for j := range m.groups {
					if m.groups[j].Key == m.filteredGroups[i].Key {
						m.groups[j].Selected = false
						m.filteredGroups[i].Selected = false
						break
					}
				}
			}
		case "enter":
			// Proceed to next screen
			return m, nil
		case "q":
			if m.filterText != "" {
				// Clear filter first
				m.filterText = ""
				m.applyFilter()
				return m, nil
			}
			return m, tea.Quit

		// Rescan options
		case "f", "F": // Rescan with flat mode
			return m, func() tea.Msg {
				return RescanRequestMsg{FlatMode: true, UseEmbedded: true}
			}
		case "M": // Rescan with embedded metadata mode
			return m, func() tea.Msg {
				return RescanRequestMsg{FlatMode: false, UseEmbedded: true}
			}
		case "r": // Rescan with normal mode
			return m, func() tea.Msg {
				return RescanRequestMsg{FlatMode: false, UseEmbedded: false}
			}
		}
	}

	return m, nil
}

// adjustScroll adjusts the scroll offset to keep cursor visible
func (m *BookGroupModel) adjustScroll() {
	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}

	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+maxVisible {
		m.scrollOffset = m.cursor - maxVisible + 1
	}
}

// View renders the UI
func (m *BookGroupModel) View() string {
	var content strings.Builder

	// Header with current mode
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	content.WriteString(headerStyle.Render("📚 AUDIOBOOK GROUPS"))
	content.WriteString("  ")
	content.WriteString(modeStyle.Render(fmt.Sprintf("[Mode: %s]", m.scanMode)))
	content.WriteString("\n")

	// Show directories
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00AAFF"))
	if m.inputDir != "" {
		content.WriteString(dimStyle.Render("📂 Input:  ") + pathStyle.Render(m.inputDir) + "\n")
	}
	if m.outputDir != "" {
		content.WriteString(dimStyle.Render("📁 Output: ") + pathStyle.Render(m.outputDir) + "\n")
	}

	// Show fallback message if present
	if m.fallbackMsg != "" {
		warnStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Bold(true)
		content.WriteString(warnStyle.Render("⚠️  "+m.fallbackMsg) + "\n")
	}
	content.WriteString("\n")

	// Stats
	selectedCount := 0
	totalFiles := 0
	for _, g := range m.groups {
		if g.Selected {
			selectedCount++
			totalFiles += len(g.Files)
		}
	}

	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	content.WriteString(statsStyle.Render(fmt.Sprintf("Found %d book groups (%d files) • Selected: %d groups",
		len(m.groups), len(m.allBooks), selectedCount)))

	// Show filter info
	if m.filterText != "" {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00AAFF"))
		content.WriteString(filterStyle.Render(fmt.Sprintf(" • Showing: %d matching \"%s\"", len(m.filteredGroups), m.filterText)))
	}
	content.WriteString("\n")

	// Show filter input if active
	if m.filtering {
		filterInputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333366")).
			Padding(0, 1)
		content.WriteString(filterInputStyle.Render(fmt.Sprintf("🔍 Filter: %s_", m.filterText)) + "\n")
	}
	content.WriteString("\n")

	// Show metadata panel if enabled
	if m.showMetadata && m.selectedGroup < len(m.filteredGroups) {
		content.WriteString(m.renderMetadataPanel(m.filteredGroups[m.selectedGroup]))
		content.WriteString("\n")
	}

	// Calculate available space for group list
	// Reserve: header(1) + dirs(3) + stats(2) + footer(3) = 9 lines minimum
	reservedLines := 9
	if m.showMetadata {
		reservedLines += 15 // Reserve space for metadata panel
	}
	if m.filtering {
		reservedLines += 1 // Filter input line
	}

	maxVisible := m.height - reservedLines
	if maxVisible < 5 {
		maxVisible = 5
	}
	// Use reasonable default if height not set yet
	if m.height == 0 {
		maxVisible = 20
	}

	endIdx := m.scrollOffset + maxVisible
	if endIdx > len(m.filteredGroups) {
		endIdx = len(m.filteredGroups)
	}

	for i := m.scrollOffset; i < endIdx; i++ {
		group := m.filteredGroups[i]
		content.WriteString(m.renderGroup(group, i == m.cursor))
	}

	// Scroll indicator if needed
	if len(m.filteredGroups) > maxVisible && endIdx < len(m.filteredGroups) {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
		remaining := len(m.filteredGroups) - endIdx
		content.WriteString(dimStyle.Render(fmt.Sprintf("    ... %d more groups (scroll with ↑/↓)\n", remaining)))
	}

	content.WriteString("\n")

	// Scan mode options - show current mode highlighted with descriptions
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Italic(true)

	content.WriteString("🔄 Rescan: ")
	if m.scanMode == "Flat" {
		content.WriteString(activeStyle.Render("F=Flat"))
	} else {
		content.WriteString(helpStyle.Render("F=Flat"))
	}
	content.WriteString(descStyle.Render("(each file)"))
	content.WriteString(helpStyle.Render(" • "))
	if m.scanMode == "Embedded" {
		content.WriteString(activeStyle.Render("M=Embedded"))
	} else {
		content.WriteString(helpStyle.Render("M=Embedded"))
	}
	content.WriteString(descStyle.Render("(group by album)"))
	content.WriteString(helpStyle.Render(" • "))
	if m.scanMode == "Normal" {
		content.WriteString(activeStyle.Render("r=Normal"))
	} else {
		content.WriteString(helpStyle.Render("r=Normal"))
	}
	content.WriteString(descStyle.Render("(metadata.json)"))
	content.WriteString("\n")

	// Help text
	if m.filtering {
		content.WriteString(helpStyle.Render("Type to filter • Enter: Apply • Esc: Cancel"))
	} else {
		content.WriteString(helpStyle.Render("↑/↓: Navigate • Space: Toggle • /: Filter • Tab/e: Expand • m: Metadata • A/N: Select All • Enter: Settings"))
	}

	return content.String()
}

// renderGroup renders a single book group
func (m *BookGroupModel) renderGroup(group BookGroup, isCursor bool) string {
	var content strings.Builder

	// Selection indicator
	selectIcon := "☐"
	if group.Selected {
		selectIcon = "☑"
	}

	// Cursor indicator
	cursor := "  "
	if isCursor {
		cursor = "▶ "
	}

	// Expand indicator
	expandIcon := "▸"
	if group.Expanded {
		expandIcon = "▾"
	}

	// Style based on cursor position
	var titleStyle, infoStyle lipgloss.Style
	if isCursor {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
		infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAFFAA"))
	} else if group.Selected {
		titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	} else {
		titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	}

	// Main line: [select] [expand] Author - Title (N files)
	mainLine := fmt.Sprintf("%s%s %s %s - %s (%d files)",
		cursor, selectIcon, expandIcon, group.Author, group.Title, len(group.Files))
	content.WriteString(titleStyle.Render(mainLine) + "\n")

	// Series info if different from title
	if group.Series != "" && group.Series != group.Title {
		content.WriteString(infoStyle.Render(fmt.Sprintf("      Series: %s", group.Series)) + "\n")
	}

	// Expanded view - show files (first 3, middle, last 3)
	if group.Expanded {
		fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		totalFiles := len(group.Files)

		if totalFiles <= 7 {
			// Show all files if 7 or fewer
			for _, file := range group.Files {
				filename := filepath.Base(file.Path)
				trackInfo := ""
				if file.TrackNumber > 0 {
					trackInfo = fmt.Sprintf(" [Track %d]", file.TrackNumber)
				}
				content.WriteString(fileStyle.Render(fmt.Sprintf("      📄 %s%s", filename, trackInfo)) + "\n")
			}
		} else {
			// Show first 3
			for i := 0; i < 3; i++ {
				file := group.Files[i]
				filename := filepath.Base(file.Path)
				trackInfo := ""
				if file.TrackNumber > 0 {
					trackInfo = fmt.Sprintf(" [Track %d]", file.TrackNumber)
				}
				content.WriteString(fileStyle.Render(fmt.Sprintf("      📄 %s%s", filename, trackInfo)) + "\n")
			}

			// Show ellipsis with count
			content.WriteString(fileStyle.Render(fmt.Sprintf("      ... %d more files ...", totalFiles-6)) + "\n")

			// Show last 3
			for i := totalFiles - 3; i < totalFiles; i++ {
				file := group.Files[i]
				filename := filepath.Base(file.Path)
				trackInfo := ""
				if file.TrackNumber > 0 {
					trackInfo = fmt.Sprintf(" [Track %d]", file.TrackNumber)
				}
				content.WriteString(fileStyle.Render(fmt.Sprintf("      📄 %s%s", filename, trackInfo)) + "\n")
			}
		}
	}

	return content.String()
}

// renderMetadataPanel renders the metadata preview panel
func (m *BookGroupModel) renderMetadataPanel(group BookGroup) string {
	var content strings.Builder

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	var inner strings.Builder

	// Header
	inner.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFF00")).
		Render("🎵 Metadata Preview") + "\n")
	inner.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).
		Render("────────────────────────────────────────") + "\n")

	if group.SampleMetadata != nil {
		md := group.SampleMetadata

		// Show detected fields
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00AAFF"))
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)

		// Title
		inner.WriteString(labelStyle.Render("📖 Title: "))
		inner.WriteString(valueStyle.Render(md.Title))
		inner.WriteString(fieldStyle.Render(" ← field: 'title'") + "\n")

		// Authors
		inner.WriteString(labelStyle.Render("👥 Authors: "))
		if len(md.Authors) > 0 {
			inner.WriteString(valueStyle.Render(strings.Join(md.Authors, ", ")))
		} else {
			inner.WriteString(valueStyle.Render("Unknown"))
		}
		inner.WriteString(fieldStyle.Render(" ← field: 'authors'") + "\n")

		// Series
		inner.WriteString(labelStyle.Render("📚 Series: "))
		if len(md.Series) > 0 && md.Series[0] != "" {
			inner.WriteString(valueStyle.Render(strings.Join(md.Series, ", ")))
		} else {
			inner.WriteString(valueStyle.Render("(none)"))
		}
		inner.WriteString(fieldStyle.Render(" ← field: 'series'") + "\n")

		// Album (often same as series/title for audiobooks)
		if md.Album != "" {
			inner.WriteString(labelStyle.Render("💿 Album: "))
			inner.WriteString(valueStyle.Render(md.Album))
			inner.WriteString(fieldStyle.Render(" ← field: 'album'") + "\n")
		}

		// Track
		if md.TrackNumber > 0 {
			inner.WriteString(labelStyle.Render("🔢 Track: "))
			inner.WriteString(valueStyle.Render(fmt.Sprintf("%d", md.TrackNumber)))
			inner.WriteString(fieldStyle.Render(" ← field: 'track'") + "\n")
		}

		// Source type
		inner.WriteString(labelStyle.Render("📂 Source: "))
		inner.WriteString(valueStyle.Render(md.SourceType) + "\n")

		// Show raw data fields if available
		if len(md.RawData) > 0 {
			inner.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).
				Render("Available fields: "))

			keys := make([]string, 0, len(md.RawData))
			for k := range md.RawData {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// Show first few fields
			maxFields := 8
			if len(keys) > maxFields {
				inner.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).
					Render(strings.Join(keys[:maxFields], ", ") + "..."))
			} else {
				inner.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).
					Render(strings.Join(keys, ", ")))
			}
			inner.WriteString("\n")
		}
	}

	content.WriteString(boxStyle.Render(inner.String()))
	return content.String()
}

// GetSelectedBooks returns all books from selected groups
func (m *BookGroupModel) GetSelectedBooks() []AudioBook {
	var books []AudioBook
	for _, group := range m.groups {
		if group.Selected {
			books = append(books, group.Files...)
		}
	}
	return books
}

// GetGroups returns all book groups
func (m *BookGroupModel) GetGroups() []BookGroup {
	return m.groups
}

// GetTotalBookCount returns the total number of books available
func (m *BookGroupModel) GetTotalBookCount() int {
	return len(m.allBooks)
}
