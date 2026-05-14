// cmd/abs.go
// Audiobookshelf integration commands

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jeeftor/audiobook-organizer/internal/abs"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Note: viper imported for accessing global --verbose flag

var (
	absURL          string
	absToken        string
	absLibraryID    string
	absSQLite       string
	absPathMaps     []string
	absAllLibraries bool
	absShowAll      bool
	absCheckFiles   bool
	absHeaderFile   string
	absHeaders      []string
)

// absCmd is the parent command for all ABS operations
var absCmd = &cobra.Command{
	Use:   "abs",
	Short: "Audiobookshelf integration",
	Long: `Audiobookshelf (ABS) integration commands.

Two modes supported:

1. API-Only Mode (--abs-url, --abs-token, --abs-path-map):
   Works with any ABS instance. You provide the path mapping manually.
   Example: --abs-path-map="/audiobooks:/mnt/media/audiobooks"

2. SQLite+API Mode (--abs-sqlite, --abs-url, --abs-token):
   Auto-discovers path mapping from ABS database.
   SQLite is read-only; all operations use API.

For Docker ABS:
  docker cp abs_container:/config/abs.sqlite /tmp/abs.sqlite
  audiobook-organizer abs scan --abs-sqlite=/tmp/abs.sqlite ...
`,
}

// absScanCmd scans audiobooks using ABS as metadata source
var absScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan audiobooks using ABS metadata",
	Example: `  # API-Only mode (manual path mapping)
  audiobook-organizer abs scan \
    --abs-url=http://localhost:13378 \
    --abs-token=eyJhbG... \
    --abs-path-map="/audiobooks:/mnt/media/audiobooks" \
    --dir=/mnt/media/audiobooks \
    --out=/mnt/organized

  # SQLite+API mode (auto path discovery)
  audiobook-organizer abs scan \
    --abs-sqlite=/var/lib/audiobookshelf/config/abs.sqlite \
    --abs-url=http://localhost:13378 \
    --abs-token=eyJhbG... \
    --dir=/mnt/media/audiobooks \
    --out=/mnt/organized

  # Trigger library scan after organization
  audiobook-organizer abs scan-trigger \
    --abs-url=http://localhost:13378 \
    --abs-token=eyJhbG... \
    --library=main`,
	RunE: runABSScan,
}

// absTestPathsCmd tests path discovery
var absTestPathsCmd = &cobra.Command{
	Use:   "test-paths",
	Short: "Test path mapping discovery",
	RunE:  runABSTestPaths,
}

// absScanTriggerCmd triggers a library scan
var absScanTriggerCmd = &cobra.Command{
	Use:   "scan-trigger",
	Short: "Trigger ABS library scan",
	RunE:  runABSScanTrigger,
}

// absWebSocketCmd tests WebSocket connection
var absWebSocketCmd = &cobra.Command{
	Use:   "websocket-test",
	Short: "Test WebSocket connection and scan events",
	Long:  "Connects to ABS WebSocket and listens for library scan events",
	RunE:  runABSWebSocketTest,
}

func init() {
	rootCmd.AddCommand(absCmd)
	absCmd.AddCommand(absScanCmd)
	absCmd.AddCommand(absTestPathsCmd)
	absCmd.AddCommand(absScanTriggerCmd)
	absCmd.AddCommand(absWebSocketCmd)

	// Common flags
	absCmd.PersistentFlags().
		StringVar(&absURL, "abs-url", "", "ABS API base URL (e.g., http://localhost:13378)")
	absCmd.PersistentFlags().StringVar(&absToken, "abs-token", "", "ABS API token")
	absCmd.PersistentFlags().StringVar(&absLibraryID, "abs-library", "main", "ABS library ID")

	// SQLite mode flag
	absCmd.PersistentFlags().
		StringVar(&absSQLite, "abs-sqlite", "", "Path to abs.sqlite (enables auto path discovery)")

	// API-only mode flag
	absCmd.PersistentFlags().
		StringSliceVar(&absPathMaps, "abs-path-map", nil, "Path mappings 'abs:local' (e.g., '/audiobooks:/mnt/media/audiobooks')")
	absScanCmd.Flags().
		BoolVar(&absAllLibraries, "abs-all-libraries", false, "Scan all libraries instead of just one (auto-detects which library each book belongs to)")
	absScanCmd.Flags().BoolVar(&absShowAll, "all", false, "Show all items (no limit)")
	absScanCmd.Flags().
		BoolVar(&absCheckFiles, "check-files", false, "Verify files exist on disk (slower)")

	// Header flags (for Cloudflare/proxy auth)
	absCmd.PersistentFlags().
		StringVar(&absHeaderFile, "header-file", "", "File with custom headers (KEY=VALUE format, one per line)")
	absCmd.PersistentFlags().
		StringSliceVar(&absHeaders, "header", nil, "Custom header (KEY=VALUE, can be used multiple times)")

	// Bind to viper for config file support
	viper.BindPFlag("abs.url", absCmd.PersistentFlags().Lookup("abs-url"))
	viper.BindPFlag("abs.token", absCmd.PersistentFlags().Lookup("abs-token"))
	viper.BindPFlag("abs.library", absCmd.PersistentFlags().Lookup("abs-library"))
	viper.BindPFlag("abs.sqlite", absCmd.PersistentFlags().Lookup("abs-sqlite"))
}

func runABSScan(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")

	// Validate inputs
	if absURL == "" {
		return fmt.Errorf("--abs-url is required (e.g., http://localhost:13378)")
	}
	if absToken == "" {
		return fmt.Errorf("--abs-token is required (get from ABS: Settings > Users > API Token)")
	}

	// Discovery mode: no --dir provided, just show library info
	if inputDir == "" {
		return runDiscoveryMode(absURL, absToken, verbose)
	}

	// Auto-detect library if not specified, or resolve by name
	if absLibraryID == "" || absLibraryID == "main" {
		selectedLib, err := selectLibrary(absURL, absToken, "")
		if err != nil {
			return err
		}
		absLibraryID = selectedLib
	} else {
		// User provided a library - try to resolve it (could be UUID or name)
		resolvedLib, err := selectLibrary(absURL, absToken, absLibraryID)
		if err != nil {
			return err
		}
		absLibraryID = resolvedLib
	}

	if verbose {
		fmt.Printf("📡 Using API: %s\n", absURL)
		fmt.Printf("📚 Library: %s\n", absLibraryID)
		fmt.Printf("📁 Input dir: %s\n", inputDir)
	}

	// Determine mode and create provider
	var provider *abs.MetadataProvider
	var err error

	if absAllLibraries {
		// Scan ALL libraries mode
		if len(absPathMaps) == 0 {
			return fmt.Errorf(
				"--abs-all-libraries requires --abs-path-map (need path mappings to match books to libraries)",
			)
		}
		fmt.Printf("Using ALL LIBRARIES mode (scanning all ABS libraries)\n")
		var mappings []abs.PathMapping
		for _, m := range absPathMaps {
			mapping, err := abs.ParsePathMapping(m)
			if err != nil {
				return err
			}
			mappings = append(mappings, mapping)
		}
		provider = abs.NewMetadataProviderAllLibraries(absURL, absToken, mappings)
	} else if absSQLite != "" {
		// SQLite+API mode: auto path discovery
		fmt.Printf("Using SQLite+API mode (auto path discovery from %s)\n", absSQLite)
		provider, err = abs.NewMetadataProviderWithSQLite(absURL, absToken, absLibraryID, absSQLite, inputDir)
		if err != nil {
			return fmt.Errorf("path discovery failed: %w\n\nHint: Use --abs-path-map for manual mode", err)
		}
	} else if len(absPathMaps) > 0 {
		// API-only mode: manual path mapping
		fmt.Printf("Using API-only mode (manual path mapping)\n")
		var mappings []abs.PathMapping
		for _, m := range absPathMaps {
			mapping, err := abs.ParsePathMapping(m)
			if err != nil {
				return err
			}
			mappings = append(mappings, mapping)
		}
		provider = abs.NewMetadataProvider(absURL, absToken, absLibraryID, mappings)
	} else {
		// No mode specified - try auto path mapping from library info
		fmt.Println("No path mapping specified. Attempting to auto-detect from library info...")
		mappings, err := autoDetectPathMappings(absURL, absToken, absLibraryID, inputDir)
		if err != nil {
			return fmt.Errorf("auto-detection failed: %w\n\nPlease provide --abs-path-map manually", err)
		}
		fmt.Printf("Auto-detected path mapping: ABS:%s -> Local:%s\n", mappings[0].ABSPrefix, mappings[0].LocalPrefix)
		provider = abs.NewMetadataProvider(absURL, absToken, absLibraryID, mappings)
	}

	// Show path mappings
	mappings := provider.GetPathMappings()
	if len(mappings) > 0 {
		fmt.Println("Path mappings:")
		for _, m := range mappings {
			fmt.Printf("  ABS: %s -> Local: %s\n", m.ABSPrefix, m.LocalPrefix)
		}
	}

	// Load all items from ABS
	if verbose {
		fmt.Println("📡 Fetching library items from ABS API...")
	} else {
		fmt.Println("Loading library items from ABS...")
	}
	if err := provider.LoadAllItems(); err != nil {
		return fmt.Errorf("loading items: %w", err)
	}

	// Get all items as metadata
	items, err := provider.GetAllItems()
	if err != nil {
		return fmt.Errorf("getting metadata: %w", err)
	}

	fmt.Printf("Found %d items in ABS library\n", len(items))

	// If all-libraries mode, show breakdown
	if absAllLibraries {
		byLib := provider.FindItemsByLibrary()
		fmt.Println("\nItems by library:")
		for libID, libItems := range byLib {
			fmt.Printf("  %s: %d items\n", libID, len(libItems))
		}
	}

	// Print items preview
	limit := 5
	if absShowAll {
		limit = len(items)
		fmt.Printf("\n📚 Showing all %d items:\n", len(items))
	} else {
		fmt.Printf("\n📚 Preview (first %d of %d items):\n", min(5, len(items)), len(items))
	}

	for i, item := range items {
		if i >= limit {
			break
		}
		author := "Unknown"
		if len(item.Authors) > 0 {
			author = item.Authors[0]
		}
		series := ""
		if len(item.Series) > 0 {
			series = fmt.Sprintf(" [%s]", item.Series[0])
		}

		// Get local path after mapping
		localPath := item.SourcePath

		// Check file existence if requested
		status := ""
		if absCheckFiles && localPath != "" {
			if _, err := os.Stat(localPath); err == nil {
				status = " ✅"
			} else {
				status = " ❌ MISSING"
			}
		}

		// Calculate target organization path
		targetPath := calculateTargetPath(localPath, author, item.Title, item.Series)
		needsMove := localPath != targetPath && localPath != "" && targetPath != ""

		// Compact output with emoji status indicators (no extra newlines within entry)
		if needsMove {
			organizer.PrintYellow("  %d. 🚚 %s - %s%s", i+1, author, item.Title, series)
			fmt.Printf("     C: %s\n", localPath)
			fmt.Printf("     T: %s\n", targetPath)
		} else if status == " ❌ MISSING" {
			organizer.PrintRed("  %d. ✗ %s - %s%s", i+1, author, item.Title, series)
			fmt.Printf("     C: %s\n", localPath)
		} else if absCheckFiles {
			// Check-files mode but no issues - compact with checkmark
			organizer.PrintGreen("  %d. ✓ %s - %s%s", i+1, author, item.Title, series)
		} else {
			// Normal compact mode
			organizer.PrintCyan("  %d. %s - %s%s", i+1, author, item.Title, series)
		}
		// Add blank line between books
		fmt.Println()
	}

	// Note about organization
	fmt.Println("\nNote: Full organization integration coming in next update.")
	fmt.Println("For now, use this command to verify ABS connectivity and metadata.")

	// If output dir specified and not dry run, trigger library scan at end
	if outputDir != "" && !dryRun {
		fmt.Println("\nTo trigger ABS library scan after organizing:")
		fmt.Printf(
			"  audiobook-organizer abs scan-trigger --abs-url=%s --abs-token=*** --library=%s\n",
			absURL,
			absLibraryID,
		)
	}

	return nil
}

func runABSTestPaths(cmd *cobra.Command, args []string) error {
	if absSQLite == "" {
		return fmt.Errorf("--abs-sqlite is required for path testing")
	}
	if inputDir == "" {
		return fmt.Errorf("--dir is required (path to test against)")
	}

	fmt.Printf("Testing path discovery from: %s\n", absSQLite)
	fmt.Printf("User input path: %s\n", inputDir)

	// Try to discover paths
	mapper, err := abs.NewPathMapperFromSQLite(absSQLite, inputDir)
	if err != nil {
		fmt.Printf("\nPath discovery FAILED: %v\n\n", err)

		// Show available libraries for debugging
		fmt.Println("Available ABS library folders:")
		folders, err := abs.ListLibraries(absSQLite)
		if err != nil {
			return fmt.Errorf("listing libraries: %w", err)
		}
		for _, f := range folders {
			fmt.Printf("  - %s (ABS path: %s)\n", f.FullPath, f.Path)
		}

		fmt.Printf("\nYour path '%s' must be under one of the above paths.\n", inputDir)
		return fmt.Errorf("path discovery failed")
	}

	fmt.Println("\nPath discovery SUCCESS!")
	fmt.Println("Mappings found:")
	for _, m := range mapper.Mappings {
		fmt.Printf("  ABS: %s -> Local: %s\n", m.ABSPrefix, m.LocalPrefix)
	}

	// Test a conversion
	testPath := "/audiobooks/Author/Book"
	localPath := mapper.ToLocal(testPath)
	fmt.Printf("\nExample conversion:\n")
	fmt.Printf("  ABS path:  %s\n", testPath)
	fmt.Printf("  Local path: %s\n", localPath)

	return nil
}

func runABSScanTrigger(cmd *cobra.Command, args []string) error {
	if absURL == "" {
		return fmt.Errorf("--abs-url is required")
	}
	if absToken == "" {
		return fmt.Errorf("--abs-token is required")
	}

	client := createABSClient(absURL, absToken)

	fmt.Printf("Triggering library scan for: %s\n", absLibraryID)
	if err := client.ScanLibrary(absLibraryID); err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	fmt.Println("Library scan triggered successfully!")
	fmt.Println("ABS will detect any moved/renamed files and update its database.")
	return nil
}

func runABSWebSocketTest(cmd *cobra.Command, args []string) error {
	if absURL == "" {
		return fmt.Errorf("--abs-url is required")
	}
	if absToken == "" {
		return fmt.Errorf("--abs-token is required")
	}

	fmt.Printf("Connecting to ABS WebSocket at %s...\n", absURL)

	wsClient := abs.NewWebSocketClient(absURL, absToken, absHeaderFile, absHeaders)
	if err := wsClient.Connect(); err != nil {
		return fmt.Errorf("websocket connection failed: %w", err)
	}
	defer wsClient.Close()

	fmt.Println("✓ WebSocket connected!")
	fmt.Println("Listening for scan events...")
	fmt.Println("(Trigger a scan manually from ABS web UI to see events)")
	fmt.Println("")

	// Set up event handlers
	wsClient.OnScanStart(func(scan abs.LibraryScan) {
		fmt.Printf("🟡 Scan START: Library '%s' (ID: %s)\n", scan.Name, scan.ID)
	})

	wsClient.OnScanComplete(func(results abs.LibraryScanResults) {
		fmt.Printf("🟢 Scan COMPLETE: %d added, %d updated, %d missing\n",
			results.Added, results.Updated, results.Missing)
	})

	// Keep connection alive for 60 seconds
	time.Sleep(60 * time.Second)

	fmt.Println("\nWebSocket test complete.")
	return nil
}

// selectLibrary fetches libraries and helps user select one
// Supports UUID or name matching (if unique)
func selectLibrary(apiURL, token string, preferredLib string) (string, error) {
	client := createABSClient(apiURL, token)

	libraries, err := client.GetLibraries()
	if err != nil {
		return "", fmt.Errorf("fetching libraries: %w", err)
	}

	if len(libraries) == 0 {
		return "", fmt.Errorf("no libraries found in ABS")
	}

	// If user provided a library, try to match it
	if preferredLib != "" && preferredLib != "main" {
		// First try exact ID match
		for _, lib := range libraries {
			if lib.ID == preferredLib {
				return lib.ID, nil
			}
		}

		// Try name match (case-insensitive)
		var nameMatches []abs.Library
		for _, lib := range libraries {
			if strings.EqualFold(lib.Name, preferredLib) {
				nameMatches = append(nameMatches, lib)
			}
		}

		if len(nameMatches) == 1 {
			fmt.Printf(
				"Matched library by name: %s (ID: %s)\n",
				nameMatches[0].Name,
				nameMatches[0].ID,
			)
			return nameMatches[0].ID, nil
		}

		if len(nameMatches) > 1 {
			return "", fmt.Errorf(
				"multiple libraries match name '%s' - use UUID instead",
				preferredLib,
			)
		}

		return "", fmt.Errorf("no library found with ID or name '%s'", preferredLib)
	}

	if len(libraries) == 1 {
		fmt.Printf("Auto-selected library: %s (only one available)\n", libraries[0].Name)
		return libraries[0].ID, nil
	}

	// Multiple libraries - show selection prompt
	fmt.Println("\nMultiple libraries found. Please select one:")
	for i, lib := range libraries {
		fmt.Printf("  %d. %s (ID: %s, Type: %s)\n", i+1, lib.Name, lib.ID, lib.MediaType)
		if len(lib.Folders) > 0 {
			fmt.Printf("     Folders: %s\n", lib.Folders[0].FullPath)
		}
	}

	return "", fmt.Errorf(
		"\nPlease specify a library with --abs-library=<id> or --abs-library=<name>\nAvailable: %s",
		formatLibraryIDs(libraries),
	)
}

// autoDetectPathMappings attempts to auto-detect path mapping from library info
func autoDetectPathMappings(
	apiURL, token, libraryID, userInputDir string,
) ([]abs.PathMapping, error) {
	client := createABSClient(apiURL, token)

	lib, err := client.GetLibrary(libraryID)
	if err != nil {
		return nil, fmt.Errorf("fetching library: %w", err)
	}

	if len(lib.Folders) == 0 {
		return nil, fmt.Errorf("library has no folders defined")
	}

	// Try to match user's input dir to library folder
	for _, folder := range lib.Folders {
		if folder.FullPath != "" && strings.HasPrefix(userInputDir, folder.FullPath) {
			return []abs.PathMapping{{
				ABSPrefix:   folder.Path,
				LocalPrefix: folder.FullPath,
			}}, nil
		}
	}

	// If no match, use first folder and assume user knows what they're doing
	// This might fail but gives them a starting point
	return []abs.PathMapping{{
		ABSPrefix:   lib.Folders[0].Path,
		LocalPrefix: userInputDir,
	}}, nil
}

func formatLibraryIDs(libraries []abs.Library) string {
	var ids []string
	for _, lib := range libraries {
		ids = append(ids, lib.ID)
	}
	return strings.Join(ids, ", ")
}

// createABSClient creates a new ABS client with optional custom headers
func createABSClient(apiURL, token string) *abs.Client {
	client := abs.NewClient(apiURL, token)

	// Load headers from file
	if absHeaderFile != "" {
		if err := client.LoadHeadersFromFile(absHeaderFile); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Warning: failed to load headers from %s: %v\n",
				absHeaderFile,
				err,
			)
		} else if viper.GetBool("verbose") {
			fmt.Printf("Loaded custom headers from %s\n", absHeaderFile)
		}
	}

	// Parse inline headers (--header flags)
	for _, h := range absHeaders {
		parts := strings.SplitN(h, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			client.SetHeader(key, value)
			if viper.GetBool("verbose") {
				fmt.Printf("Set header: %s\n", key)
			}
		}
	}

	return client
}

// runDiscoveryMode shows library info without requiring --dir
func runDiscoveryMode(apiURL, token string, verbose bool) error {
	client := createABSClient(apiURL, token)

	if verbose {
		fmt.Printf("📡 Connecting to ABS at %s\n", apiURL)
	}

	fmt.Println("🔍 ABS Discovery Mode (no --dir specified)")
	fmt.Println()

	// Fetch and show libraries
	libraries, err := client.GetLibraries()
	if err != nil {
		return fmt.Errorf("fetching libraries: %w", err)
	}

	fmt.Printf("Found %d librar%s:\n", len(libraries), plural(len(libraries), "y", "ies"))
	for i, lib := range libraries {
		fmt.Printf("\n  %d. %s (ID: %s)\n", i+1, lib.Name, lib.ID)
		fmt.Printf("     Type: %s\n", lib.MediaType)
		if len(lib.Folders) > 0 {
			for _, f := range lib.Folders {
				fmt.Printf("     Folder: %s (ABS: %s)\n", f.FullPath, f.Path)
			}
		}

		// Show item count for this library
		items, err := client.GetAllLibraryItems(lib.ID)
		if err == nil {
			fmt.Printf("     Items: %d\n", len(items))
		}
	}

	fmt.Println()
	fmt.Println("💡 To organize books, run with:")
	fmt.Println()
	if len(libraries) == 1 {
		fmt.Printf("   --dir=<path>  (library will be auto-selected)\n")
	} else {
		fmt.Printf("   --dir=<path> --abs-library=<id>\n")
		fmt.Println()
		fmt.Println("   Available library IDs:")
		for _, lib := range libraries {
			fmt.Printf("     --abs-library=%s  (%s)\n", lib.ID, lib.Name)
		}
	}
	fmt.Println()

	return nil
}

func plural(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// calculateTargetPath determines where a book SHOULD be based on metadata
// Sanitizes path components to ensure valid filesystem paths
func calculateTargetPath(currentPath, author, title string, series []string) string {
	if currentPath == "" || author == "" || title == "" {
		return ""
	}

	// Extract base directory (e.g., /data/IRC/books)
	baseDir := filepath.Dir(filepath.Dir(currentPath))

	// Sanitize path components (remove/replace invalid chars)
	safeAuthor := sanitizePathComponent(author)
	safeTitle := sanitizePathComponent(title)

	// Build target path: Author/Series/Title or Author/Title
	targetDir := filepath.Join(baseDir, safeAuthor)
	if len(series) > 0 && series[0] != "" {
		safeSeries := sanitizePathComponent(series[0])
		targetDir = filepath.Join(targetDir, safeSeries)
	}
	targetDir = filepath.Join(targetDir, safeTitle)

	return targetDir
}

// sanitizePathComponent removes/replaces characters that are invalid in file paths
func sanitizePathComponent(s string) string {
	// Invalid characters for file paths on most systems
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*", "/", "\\"}

	result := s
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Trim leading/trailing spaces and dots
	result = strings.TrimSpace(result)
	result = strings.Trim(result, ".")

	return result
}
