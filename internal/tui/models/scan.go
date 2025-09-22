package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// ScanMsg is sent when a new audiobook is found
type ScanMsg struct {
	Book AudioBook
}

// ScanCompleteMsg is sent when scanning is complete
type ScanCompleteMsg struct {
	Books []AudioBook
}

// AudioBook represents an audiobook with its metadata
type AudioBook struct {
	Path        string
	Metadata    organizer.Metadata
	Selected    bool
	IsPartOfAlbum bool  // Indicates if this file is part of a multi-file album
	AlbumName   string   // Name of the album this file belongs to
	TrackNumber int      // Track number within the album
	TotalTracks int      // Total number of tracks in the album
}

// ScanModel represents the scanning screen
type ScanModel struct {
	inputDir    string
	scanning    bool
	complete    bool
	books       []AudioBook
	scannedDirs int
	scannedFiles int
	startTime   time.Time
	elapsedTime time.Duration
}

// NewScanModel creates a new scan model
func NewScanModel(inputDir string) *ScanModel {
	return &ScanModel{
		inputDir:  inputDir,
		scanning:  false,
		complete:  false,
		books:     []AudioBook{},
		startTime: time.Now(),
	}
}

// Init initializes the model and automatically starts scanning
func (m *ScanModel) Init() tea.Cmd {
	// Automatically start scanning without requiring user input
	return m.startScan()
}

// startScan begins the scanning process
func (m *ScanModel) startScan() tea.Cmd {
	m.scanning = true
	m.startTime = time.Now()

	return tea.Sequence(
		// Start a ticker to update the UI while scanning
		tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			m.elapsedTime = time.Since(m.startTime)
			return nil
		}),
		// Run the scan in a background process
		func() tea.Msg {
			// Perform the scan
			books := m.scanDirectory(m.inputDir)

			// Return completion message
			return ScanCompleteMsg{Books: books}
		},
	)
}

// scanDirectory scans a directory for audiobooks
func (m *ScanModel) scanDirectory(dir string) []AudioBook {
	var books []AudioBook

	// Use the organizer package to scan for audiobooks and extract metadata
	m.scannedDirs++

	// Define supported file extensions
	extensions := []string{".m4b", ".mp3", ".m4a", ".epub"}

	// Add debug logging to a file
	logFile, _ := os.OpenFile("scan_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer logFile.Close()

	logFile.WriteString(fmt.Sprintf("Scanning directory: %s\n", dir))

	// First pass: collect all audio files and their metadata
	type fileInfo struct {
		path     string
		metadata organizer.Metadata
		dir      string
	}

	fileInfos := []fileInfo{}

	// Map to track directories with multiple audio files (potential albums)
	dirFileCount := make(map[string]int)

	// Walk through the directory to collect all files
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logFile.WriteString(fmt.Sprintf("Error accessing path %s: %v\n", path, err))
			return nil
		}

		if info.IsDir() {
			m.scannedDirs++
			logFile.WriteString(fmt.Sprintf("Found directory: %s\n", path))
			return nil
		}

		m.scannedFiles++
		ext := strings.ToLower(filepath.Ext(path))
		logFile.WriteString(fmt.Sprintf("Checking file: %s (ext: %s)\n", path, ext))

		for _, validExt := range extensions {
			if ext == validExt {
				logFile.WriteString(fmt.Sprintf("Found supported file: %s\n", path))

				// Extract metadata using the organizer package's UnifiedMetadataProvider
				provider := organizer.NewMetadataProvider(path)
				metadata, err := provider.GetMetadata()
				if err != nil {
					logFile.WriteString(fmt.Sprintf("Error extracting metadata: %v\n", err))

					// If metadata extraction fails, create basic metadata from filename
					baseName := filepath.Base(path)
					metadata = organizer.Metadata{
						Title:   baseName,
						Authors: []string{"Unknown Author"},
					}
					logFile.WriteString(fmt.Sprintf("Using fallback metadata: Title=%q\n", baseName))
				} else {
					logFile.WriteString(fmt.Sprintf("Successfully extracted metadata: Title=%q, Authors=%v\n",
						metadata.Title, metadata.Authors))
				}

				// Store file info for later processing
				dirPath := filepath.Dir(path)
				fileInfos = append(fileInfos, fileInfo{
					path:     path,
					metadata: metadata,
					dir:      dirPath,
				})

				// Count files per directory for album detection
				dirFileCount[dirPath]++
				break
			}
		}

		return nil
	})

	// Second pass: group files into albums or process as individual files
	// Map to track which directories have been processed as albums
	processedDirs := make(map[string]bool)

	// Group files by directory for album detection
	dirFiles := make(map[string][]fileInfo)
	for _, fi := range fileInfos {
		dirFiles[fi.dir] = append(dirFiles[fi.dir], fi)
	}

	// Process each directory
	for dir, files := range dirFiles {
		// Skip if already processed
		if processedDirs[dir] {
			continue
		}

		// Check if this directory should be processed as an album
		isAlbum := len(files) > 1

		// If it's an album, check for consistent metadata
		if isAlbum {
			// Check for consistent title and author across files
			var albumTitle, albumArtist string
			consistentMetadata := true

			// Get first file's metadata as reference
			if len(files) > 0 {
				albumTitle = files[0].metadata.Title
				if len(files[0].metadata.Authors) > 0 {
					albumArtist = files[0].metadata.Authors[0]
				}

				// Check if other files have matching metadata
				for i := 1; i < len(files); i++ {
					currentTitle := files[i].metadata.Title
					var currentArtist string
					if len(files[i].metadata.Authors) > 0 {
						currentArtist = files[i].metadata.Authors[0]
					}

					// Check if title and artist match
					if currentTitle != albumTitle || (albumArtist != "" && currentArtist != "" && currentArtist != albumArtist) {
						// Check for track number patterns in title
						if !organizer.HasTrackNumberPattern(currentTitle, albumTitle) && !organizer.HasCommonPrefix(currentTitle, albumTitle) {
							consistentMetadata = false
							break
						}
					}
				}
			}

			// Process as album if metadata is consistent
			if consistentMetadata {
				logFile.WriteString(fmt.Sprintf("Processing directory as album: %s\n", dir))

				// Create album name from common metadata
				albumName := albumTitle
				if albumArtist != "" {
					albumName = albumArtist + " - " + albumName
				}

				// Sort files by track number if available
				organizer.SortFilesByTrackNumber(files)

				// Create AudioBook entries for each file in the album
				totalTracks := len(files)
				for i, file := range files {
					trackNumber := i + 1 // Default to position in sorted list

					// Use actual track number if available
					if file.metadata.TrackNumber > 0 {
						trackNumber = file.metadata.TrackNumber
					}

					book := AudioBook{
						Path:         file.path,
						Metadata:     file.metadata,
						Selected:     true,
						IsPartOfAlbum: true,
						AlbumName:    albumName,
						TrackNumber:  trackNumber,
						TotalTracks:  totalTracks,
					}
					books = append(books, book)
				}

				// Mark this directory as processed
				processedDirs[dir] = true
				continue
			}
		}

		// Process files individually if not an album or inconsistent metadata
		for _, file := range files {
			if !processedDirs[dir] {
				book := AudioBook{
					Path:         file.path,
					Metadata:     file.metadata,
					Selected:     true,
					IsPartOfAlbum: false,
				}
				books = append(books, book)
			}
		}
	}

	logFile.WriteString(fmt.Sprintf("Scan complete. Total books found: %d\n\n", len(books)))
	return books
}

// Update handles messages and user input
func (m *ScanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ScanMsg:
		// A new book was found
		m.books = append(m.books, msg.Book)
		return m, nil

	case ScanCompleteMsg:
		// Scanning is complete
		m.books = msg.Books
		m.complete = true
		m.scanning = false

		// If books were found, automatically proceed to book list after a short delay
		if len(m.books) > 0 {
			return m, tea.Sequence(
				tea.Tick(time.Millisecond*800, func(_ time.Time) tea.Msg {
					return ScanCompleteMsg{Books: m.books}
				}),
			)
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "enter" && !m.scanning && !m.complete {
			// Start scanning when Enter is pressed
			return m, m.startScan()
		} else if msg.String() == "r" && m.complete {
			// Restart scanning when 'r' is pressed after completion
			m.complete = false
			m.scanning = false
			m.books = []AudioBook{}
			m.scannedDirs = 0
			m.scannedFiles = 0
			return m, m.startScan()
		}
	}

	return m, nil
}

// View renders the UI
func (m *ScanModel) View() string {
	var content strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("📚 Audiobook Scanner")

	content.WriteString(title + "\n\n")

	if !m.scanning && !m.complete {
		// Initial state
		content.WriteString("Press Enter to start scanning directory:\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(m.inputDir))
		content.WriteString("\n\n")
		content.WriteString("This will search for audiobooks in the directory and its subdirectories.")
	} else if m.scanning {
		// Scanning state
		content.WriteString("🔍 Scanning for audiobooks...\n\n")
		content.WriteString(fmt.Sprintf("Directory: %s\n", m.inputDir))
		content.WriteString(fmt.Sprintf("Directories scanned: %d\n", m.scannedDirs))
		content.WriteString(fmt.Sprintf("Files checked: %d\n", m.scannedFiles))
		content.WriteString(fmt.Sprintf("Books found: %d\n", len(m.books)))
		content.WriteString(fmt.Sprintf("Elapsed time: %s\n", m.elapsedTime.Round(time.Second)))

		// Add a spinner
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinnerChar := spinner[int(time.Now().UnixNano()/100000000)%len(spinner)]
		content.WriteString("\n" + spinnerChar + " Scanning...")
	} else if m.complete {
		// Complete state
		content.WriteString("✅ Scan complete!\n\n")
		content.WriteString(fmt.Sprintf("Found %d audiobooks in %s\n\n", len(m.books), m.elapsedTime.Round(time.Second)))

		if len(m.books) > 0 {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("Press any key to continue to book selection..."))
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true).Render("No audiobooks found.") + "\n\n")
			content.WriteString("This could be because:\n")
			content.WriteString("1. The directory doesn't contain supported audiobook files (.m4b, .mp3, .m4a, .epub)\n")
			content.WriteString("2. The files don't have readable metadata\n\n")
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render("Press 'r' to scan again or 'q' to quit"))
		}
	}

	return content.String()
}

// IsComplete returns whether scanning is complete
func (m *ScanModel) IsComplete() bool {
	return m.complete
}

// GetBooks returns the list of found books
func (m *ScanModel) GetBooks() []AudioBook {
	return m.books
}
