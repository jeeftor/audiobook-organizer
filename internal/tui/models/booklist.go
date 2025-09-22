package models

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BookItem represents an item in the book list
type BookItem struct {
	book     AudioBook
	selected bool
}

// FilterState represents the current filter state
type FilterState struct {
	query     string
	filtering bool
}

// BookListModel represents the book list screen
type BookListModel struct {
	books       []AudioBook
	items       []list.Item
	list        list.Model
	filterState FilterState
	selected    map[int]bool
	width       int
	height      int
}

// Implement the list.Item interface for BookItem
func (i BookItem) Title() string {
	// Add a checkmark if selected
	prefix := ""
	if i.selected {
		prefix = "✓ "
	}

	// Add album indicator if part of a multi-file album
	if i.book.IsPartOfAlbum {
		prefix += "📀 "
	}

	// Prefer filename over metadata title for display
	base := filepath.Base(i.book.Path)
	fileTitle := strings.TrimSuffix(base, filepath.Ext(base))

	// Use filename as title instead of generic series title
	title := fileTitle

	// If filename is empty for some reason, fall back to metadata title
	if title == "" {
		title = i.book.Metadata.Title
	}

	// Add track number information if part of an album
	if i.book.IsPartOfAlbum && i.book.TrackNumber > 0 {
		title = fmt.Sprintf("%s [Track %d/%d]", title, i.book.TrackNumber, i.book.TotalTracks)
	}

	return prefix + title
}

func (i BookItem) Description() string {
	var desc strings.Builder

	// First line: Author and Series info
	if len(i.book.Metadata.Authors) > 0 {
		desc.WriteString("Author: " + strings.Join(i.book.Metadata.Authors, ", "))
	} else {
		desc.WriteString("Author: Unknown")
	}

	// Add series info if available
	if series := i.book.Metadata.GetValidSeries(); series != "" {
		if desc.Len() > 0 {
			desc.WriteString(" | ")
		}
		desc.WriteString("Series: " + series)
	}

	// Add album info if part of a multi-file album
	if i.book.IsPartOfAlbum {
		desc.WriteString("\nAlbum: " + i.book.AlbumName)
		if i.book.TrackNumber > 0 {
			desc.WriteString(fmt.Sprintf(" (Track %d of %d)", i.book.TrackNumber, i.book.TotalTracks))
		}
	}

	// Second line: Input path (shortened if too long)
	path := i.book.Path
	if len(path) > 60 {
		// Shorten the path by keeping the beginning and end
		dir := filepath.Dir(path)
		base := filepath.Base(path)
		if len(dir) > 40 {
			dir = dir[:20] + "..." + dir[len(dir)-20:]
		}
		path = filepath.Join(dir, base)
	}
	desc.WriteString("\nInput: " + path)

	// Third line: Output path preview
	outputPath := generateOutputPathPreview(i.book)
	desc.WriteString("\nOutput: " + outputPath)

	return desc.String()
}

func (i BookItem) FilterValue() string {
	// This is used for filtering - include filename, title, authors and series
	base := filepath.Base(i.book.Path)
	fileTitle := strings.TrimSuffix(base, filepath.Ext(base))
	authors := strings.Join(i.book.Metadata.Authors, " ")
	series := i.book.Metadata.GetValidSeries()
	metadataTitle := i.book.Metadata.Title

	// Combine all searchable fields
	return fmt.Sprintf("%s %s %s %s", fileTitle, metadataTitle, authors, series)
}

// CustomDelegate is a custom delegate for the book list
type CustomDelegate struct {
	Styles struct {
		NormalTitle, SelectedTitle                lipgloss.Style
		NormalDesc, SelectedDesc                  lipgloss.Style
		NormalItemStyle, SelectedItemStyle        lipgloss.Style
		DimmedDesc                                lipgloss.Style
	}
	ItemHeight     int
	ItemSpacing    int
	SelectedPrefix  string
	UnselectedPrefix string
}

// NewCustomDelegate creates a new custom delegate
func NewCustomDelegate() list.ItemDelegate {
	d := CustomDelegate{}

	// Initialize styles struct
	d.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#2BFFB5")).Bold(true)
	d.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7D56F4")).Bold(true)
	d.Styles.NormalDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	d.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD"))
	d.Styles.DimmedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	d.Styles.NormalItemStyle = lipgloss.NewStyle().PaddingLeft(2)
	d.Styles.SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Background(lipgloss.Color("#333333"))

	// Set other properties - make items very compact to show more books
	d.ItemHeight = 1
	d.ItemSpacing = 0
	d.SelectedPrefix = "✓ "
	d.UnselectedPrefix = "  "

	return d
}

// Render renders a list item
func (d CustomDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		titleStyle, descStyle lipgloss.Style
	)

	// Get the book item
	bookItem, ok := item.(BookItem)
	if !ok {
		return
	}

	// Always regenerate title and description to ensure they're up to date
	title := bookItem.Title()
	desc := bookItem.Description()

	// Set styles based on selection
	if index == m.Index() {
		titleStyle = d.Styles.SelectedTitle
		descStyle = d.Styles.SelectedDesc
	} else {
		titleStyle = d.Styles.NormalTitle
		descStyle = d.Styles.NormalDesc
	}

	// Split description into lines for better styling
	descLines := strings.Split(desc, "\n")
	styledDesc := ""

	// Apply styles to each line
	for i, line := range descLines {
		if i > 0 {
			styledDesc += "\n"
		}
		styledDesc += descStyle.Render(line)
	}

	// Apply styles to title
	styledTitle := titleStyle.Render(title)

	// Combine title and description
	str := styledTitle + "\n" + styledDesc

	// Apply item style
	if index == m.Index() {
		str = d.Styles.SelectedItemStyle.Render(str)
	} else {
		str = d.Styles.NormalItemStyle.Render(str)
	}

	// Write to output
	fmt.Fprintf(w, "%s\n%s", str, strings.Repeat("\n", d.ItemSpacing))
}

// Height returns the height of the delegate
func (d CustomDelegate) Height() int {
	return d.ItemHeight + d.ItemSpacing
}

// Spacing returns the spacing of the delegate
func (d CustomDelegate) Spacing() int {
	return d.ItemSpacing
}

// Update is called when the list is updated
func (d CustomDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// NewBookListModel creates a new book list model
func NewBookListModel(books []AudioBook) *BookListModel {
	// Convert books to list items
	items := make([]list.Item, len(books))
	for i, book := range books {
		items[i] = BookItem{book: book, selected: true} // Default to selected
	}

	// Create the list model with custom delegate
	l := list.New(items, NewCustomDelegate(), 20, 20) // Set height to show multiple books
	l.Title = "📚 Audiobooks Found"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(true)
	l.SetShowPagination(true)
	l.Styles.Title = l.Styles.Title.Background(lipgloss.Color("#7D56F4")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 1)

	// Initialize selected map with all books selected by default
	selected := make(map[int]bool)
	for i := range books {
		selected[i] = true
	}

	return &BookListModel{
		books:    books,
		items:    items,
		list:     l,
		selected: selected,
		filterState: FilterState{
			query:     "",
			filtering: false,
		},
	}
}

// Init initializes the model
func (m *BookListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and user input
func (m *BookListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 6) // Leave room for header and footer

	case tea.KeyMsg:
		switch msg.String() {
		case " ": // Space to toggle selection
			idx := m.list.Index()
			if idx >= 0 && idx < len(m.items) {
				m.selected[idx] = !m.selected[idx]

				// Update the item's selected state
				item := m.items[idx].(BookItem)
				item.selected = m.selected[idx]
				m.items[idx] = item
			}
			return m, nil

		case "a": // Select all
			for i := range m.items {
				m.selected[i] = true
				item := m.items[i].(BookItem)
				item.selected = true
				m.items[i] = item
			}
			return m, nil

		case "n": // Deselect all
			for i := range m.items {
				m.selected[i] = false
				item := m.items[i].(BookItem)
				item.selected = false
				m.items[i] = item
			}
			return m, nil

		case "/": // Start filtering
			m.filterState.filtering = true
			m.list.SetFilteringEnabled(true)
			// Send a slash key to start filtering
			return m, func() tea.Msg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}} }
		}
	}

	// Handle list updates
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel

	// Handle escape key to clear filter
	if msg, ok := msg.(tea.KeyMsg); ok && msg.Type == tea.KeyEsc && m.filterState.filtering {
		m.filterState.filtering = false
		m.filterState.query = ""
		m.list.ResetFilter()
	}

	// Update filter state
	if m.filterState.filtering {
		m.filterState.query = m.list.FilterValue()
		if m.filterState.query == "" {
			m.filterState.filtering = false
		}
	}

	return m, cmd
}

// View renders the UI
func (m *BookListModel) View() string {
	var content strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("📚 Audiobook Selection")

	content.WriteString(header + "\n\n")

	// Count selected books
	selectedCount := 0
	for _, selected := range m.selected {
		if selected {
			selectedCount++
		}
	}

	// Selection info with better styling
	selectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true).
		MarginBottom(1)

	selectionInfo := fmt.Sprintf("Selected: %d of %d books", selectedCount, len(m.items))
	content.WriteString(selectionStyle.Render(selectionInfo) + "\n")

	// Add book count if no books are found
	if len(m.items) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Render("No audiobooks found. Please go back and scan again.")
		content.WriteString(emptyMsg + "\n\n")
	}

	// List view
	content.WriteString(m.list.View())

	// Show filter status if filtering
	if m.filterState.filtering && m.filterState.query != "" {
		filterStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			MarginTop(1)

		filterText := fmt.Sprintf("Filtering by: %s", m.filterState.query)
		content.WriteString("\n" + filterStyle.Render(filterText))
	}

	// Footer with help text
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		MarginTop(1)

	footerText := "Space: toggle selection • a: select all • n: deselect all • /: filter • Esc: clear filter • Enter: continue"
	footer := "\n" + footerStyle.Render(footerText)

	content.WriteString(footer)

	return content.String()
}

// GetSelectedBooks returns the currently selected books
func (m *BookListModel) GetSelectedBooks() []AudioBook {
	selected := []AudioBook{}

	for i := range m.books {
		if m.selected[i] {
			selected = append(selected, m.books[i])
		}
	}

	return selected
}

// GetBooks returns all books in the list
func (m *BookListModel) GetBooks() []AudioBook {
	return m.books
}
