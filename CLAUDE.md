# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an audiobook organizer tool that organizes audiobook files based on metadata from either `metadata.json` files or embedded metadata in EPUB, MP3, and M4B files. The tool creates structured directory layouts and supports various organization patterns.

**Three modes available:**
1. **GUI (Desktop App)** - Wails v2-based desktop application with React frontend (`audiobook-organizer-gui`)
2. **TUI (Interactive Terminal)** - Bubbletea-based interactive terminal interface (`audiobook-organizer tui`)
3. **CLI (Command Line)** - Scriptable automation interface (`audiobook-organizer --dir=/path`)

The project consists of two separate binaries:
- `audiobook-organizer` - CLI/TUI modes (main Go binary)
- `audiobook-organizer-gui` - Desktop GUI application (Wails app in `audiobook-organizer-gui/` directory)

## Architecture

The project follows a standard Go CLI architecture with separation of concerns:

### Entry Point
- **`main.go`**: Entry point that delegates to the cmd package

### Commands (`cmd/`)
- **`root.go`**: Main CLI command with:
  - Cobra flag definitions with dual aliases (--dir/--input, --out/--output)
  - Viper configuration handling with multiple sources (flags, env vars, config file)
  - Custom environment variable handling supporting both `AO_*` and `AUDIOBOOK_ORGANIZER_*` prefixes
  - PreRun hooks for flag aliasing and validation
- **`gui.go`**: TUI command that launches the interactive interface
- **`version.go`**: Version command implementation

### Core Business Logic (`internal/organizer/`)
- **`organizer.go`**: Main organizer struct and execution logic, including:
  - `OrganizerConfig`: Configuration struct
  - `FileOps`: File system operations handler with dry-run support
  - `LayoutCalculator`: Path calculation based on layout configuration
- **`organize.go`**: File organization and moving logic
- **`metadata_providers.go`**: Metadata extraction from various file formats (EPUB, MP3, M4B, JSON)
- **`path.go`**: Path generation and sanitization utilities
- **`types.go`**: Core data structures:
  - `Metadata`: Main metadata struct with field mapping support
  - `FieldMapping`: Configuration for mapping metadata fields from different sources
  - `MetadataProvider`: Interface for metadata extraction
  - `LogEntry`, `Summary`, `MoveSummary`: Support types
- **`logging.go`**: Operation logging for undo functionality
- **`prompt.go`**: Interactive user prompts
- **`album_detection.go`**: Logic for detecting multi-file albums/audiobooks
- **`album_handler.go`**: Handles processing of multi-file albums
- **`metadata_formatter.go`**: Formatting metadata for display

### TUI Interface (`internal/tui/`)
- **`app.go`**: TUI application entry point
- **`models/`**: Bubbletea models for different screens:
  - `main.go`: Main coordinator model managing screen transitions
  - `scan.go`: Directory scanning screen
  - `booklist.go`: Book selection/review screen
  - `preview.go`: Preview changes screen
  - `settings.go`: Settings configuration screen
  - `process.go`: File processing/organization screen
  - `output_path.go`: Output path configuration

### GUI Interface (`audiobook-organizer-gui/`)
- **Wails v2 desktop application** with Go backend and React + TypeScript frontend
- **Backend (`audiobook-organizer-gui/`)**: Go code that wraps core organizer functionality
  - Exposes methods to frontend via Wails bindings
  - Handles file system operations and metadata extraction
  - Manages organization and rename workflows
  - **Undo system** (`undo.go`): Reverses organization operations using `.abook-org.log`
- **Frontend (`audiobook-organizer-gui/frontend/`)**: React + TypeScript UI with Tailwind CSS
  - **Three-pane layout** (resizable panels):
    - **Left**: Grouped file list with album grouping, expandable groups, multi-select with checkboxes
    - **Center**: Metadata editor with editable fields, field mapping configuration, live preview
    - **Right**: Options panel with scan mode, layout templates, rename template builder
  - **View modes**:
    - **Editing view**: Three-pane layout for metadata review and configuration
    - **Preview view**: Full-screen execution preview with before/after paths, bash commands, copy/move option
    - **Results view**: Success/failure screen with undo button, file counts, error details
  - **Key features**:
    - Auto-scan on folder open
    - Selection preservation during rescans and field mapping changes
    - Grouped/flat list toggle with album-based grouping
    - Color-coded path previews (author=orange, series=cyan, title=green)
    - Rename template builder with field buttons and live preview
    - Batch preview showing all selected files with INPUT → OUTPUT paths
    - Copy-to-clipboard for bash commands
    - Finder-style zebra striping on lists
    - Indeterminate checkboxes for partial group selection
  - **Current limitations**:
    - `OrganizeFiles` function disabled due to selection filtering bug
    - Needs proper file filtering before organization (only selected files should be processed)
- **Building GUI**: Requires Wails CLI (`wails build`) and platform-specific dependencies
- **GitHub Actions**: Automated builds for macOS, Linux, Windows in `.github/workflows/build-gui.yml`

### Testing Structure
- Unit tests are co-located with source files using `_test.go` suffix
- Integration tests are in `internal/organizer/integration/`
- Test utilities in `internal/organizer/integration/test_utils.go`
- Test programs in `cmd/tests/` for manual testing scenarios

## Development Commands

### Building
```bash
# Development build with version info embedded via LDFLAGS
make dev

# Production build using goreleaser (creates multi-platform binaries)
make build

# Clean all build artifacts
make clean
```

### Testing
```bash
# Run unit tests (default, fast)
make test
# or
make test-unit

# Run integration tests (slower, tagged with 'integration')
make test-integration

# Run all tests (both unit and integration)
make test-all

# Run tests with coverage report
make coverage

# Generate HTML coverage report
make coverage-html

# Run tests for a specific package
go test ./internal/organizer/

# Run a specific test
go test -run TestSpecificTest ./internal/organizer/

# Run tests in verbose mode
go test -v ./...

# Run only short/fast tests
go test -short ./...
```

### Dependencies
```bash
# Install/update dependencies
go mod tidy

# Download dependencies
go mod download

# Ensure gotestsum is installed (used by Makefile)
make ensure-gotestsum
```

### Running the Application
```bash
# CLI mode (organize audiobooks)
./bin/audiobook-organizer --dir=/path/to/books --out=/path/to/output

# TUI mode (interactive interface) - supports same flag aliases as CLI
./bin/audiobook-organizer gui --dir=/path/to/books --out=/path/to/output
# or
./bin/audiobook-organizer gui --input=/path/to/books --output=/path/to/output

# Dry run mode (preview changes)
./bin/audiobook-organizer --dir=/path/to/books --dry-run

# Version information
./bin/audiobook-organizer version

# GUI mode (desktop application)
make gui-dev                    # Development mode with hot reload
make gui-build                  # Production build
cd audiobook-organizer-gui && wails dev -appargs "--dir=../books --out=../output"
```

## Key Configuration Patterns

### Configuration Priority
The application uses Viper for configuration with the following priority (highest to lowest):
1. Command line flags
2. Environment variables (with flexible prefixing)
3. Configuration file (`~/.audiobook-organizer.yaml` or specified with `--config`)
4. Default values

### Flag Aliases
The root command supports dual aliases for input/output directories to accommodate different naming preferences:
- **Input directory**: `--dir` and `--input` are completely interchangeable
- **Output directory**: `--out` and `--output` are completely interchangeable
- **Important**: The aliasing logic is handled in `handleInputAliases()` in cmd/root.go:166

### Environment Variable Handling
All flags can be set via environment variables with flexible prefixing. The `envAliases` map (cmd/root.go:48) defines all valid environment variable names:
- `AO_*` (short form) - e.g., `AO_DIR`, `AO_VERBOSE`
- `AUDIOBOOK_ORGANIZER_*` (long form) - e.g., `AUDIOBOOK_ORGANIZER_DIR`, `AUDIOBOOK_ORGANIZER_VERBOSE`

### Field Mapping System
The tool supports flexible field mapping to handle different metadata structures (especially for MP3 files):
- **Author Fields**: Comma-separated list to try (e.g., `authors,narrators,album_artist,artist`)
- **Title Field**: Single field to use as title (e.g., `album`, `title`, `track_title`)
- **Series Field**: Field to use for series (e.g., `series`, `album`)
- **Track Field**: Field for track numbers (e.g., `track`, `track_number`)
- **Disc Field**: Field for disc numbers (e.g., `disc`, `discnumber`, `disk`, `tpos`)

Field mapping is implemented in `types.go` with the `FieldMapping` struct and `ApplyFieldMapping()` method.

## Directory Layout Options

The tool supports seven layout patterns via the `--layout` flag:
- `author-series-title`: Full hierarchy Author/Series/Title/ (default)
- `author-series-title-number`: Author/Series/#1 - Title/ (includes series number)
- `author-series`: Author/Series/ (series-focused, no title level)
- `author-title`: Author/Title/ (skips series level)
- `author-only`: Author/ (flattens all books to author directory)
- `series-title`: Series/Title/ (series-first organization, no author directory)
- `series-title-number`: Series/#1 - Title/ (series-first with numbering, no author directory)

Layout logic is handled by the `LayoutCalculator` struct in organizer.go:82.

**See also:** [docs/LAYOUTS.md](docs/LAYOUTS.md) for detailed layout comparison and examples.

## Metadata Sources and Processing

### Metadata Sources
1. **metadata.json files**: Primary source, created by audiobookshelf
2. **Embedded metadata**: Extracted from EPUB, MP3, and M4B files when `--use-embedded-metadata` is enabled
3. **Flat processing**: Use `--flat` to process files in flat directory structures (auto-enables embedded metadata)

### Metadata Architecture
- **MetadataProvider interface** (types.go:322): Abstraction for different metadata sources
- **Metadata struct** (types.go:54): Core metadata container with:
  - Standard fields (Title, Authors, Series, TrackNumber)
  - RawData map for flexible field mapping
  - FieldMapping configuration
  - Source tracking (SourceType, SourcePath)
- **Metadata extraction**: Implemented in `metadata_providers.go` for different file formats

### Processing Modes
- **Flat mode** (`--flat`): Each file processed independently, grouped only by identical metadata
- **Non-flat mode** (default): All files in a directory treated as one book
- **Album detection**: Multi-file audiobooks are detected and grouped (album_detection.go)

## Testing Considerations

### Test Organization
- **Unit tests**: Co-located with source files using `_test.go` suffix
- **Integration tests**: Located in `internal/organizer/integration/` directory
- **Test utilities**: Shared test helpers in `internal/organizer/integration/test_utils.go`
- **Manual test programs**: Located in `cmd/tests/` for specific scenarios

### Testing Tools
- **gotestsum**: Used by Makefile for better test output formatting (auto-installed via `make ensure-gotestsum`)
- **testify**: Used for assertions (`github.com/stretchr/testify`)
- Use table-driven tests where appropriate

### Test Execution
- Integration tests are tagged with `integration` build tag
- Use `make test-unit` for fast unit tests only
- Use `make test-integration` for slower integration tests
- Use `make test-all` to run everything
- Mock external dependencies (file system operations) for unit tests
- Integration tests should use temporary directories

## Build Process

The project uses a two-tier build system:

### Development Builds (Makefile)
- **Target**: `make dev`
- **Purpose**: Fast local development builds
- **Output**: `bin/audiobook-organizer`
- **LDFLAGS**: Embeds version information:
  - `cmd.buildVersion` - from `git describe --tags`
  - `cmd.buildCommit` - from `git rev-parse --short HEAD`
  - `cmd.buildTime` - current timestamp

### Production Builds (GoReleaser)
- **Target**: `make build` (snapshot) or `make release` (tagged release)
- **Purpose**: Multi-platform distribution builds
- **Config**: `.goreleaser.yaml`
- **Platforms**: Linux, Windows, macOS (amd64, arm64, arm)
- **Outputs**:
  - Binary archives (tar.gz, zip)
  - Package formats (deb, rpm, apk)
  - Homebrew formula
  - SBOMs (Software Bill of Materials)
- **LDFLAGS**: Similar version injection as Makefile

### Version Information
Version info is displayed via `audiobook-organizer version` command and is injected at build time into variables in the `cmd` package.

## Key Implementation Patterns

### Dry-Run Support
The `FileOps` struct (organizer.go:42) provides dry-run support for all file system operations:
- `CreateDirIfNotExists()`: Respects dry-run mode
- File operations are logged but not executed when dry-run is enabled
- This pattern should be followed for any new file system operations

### Undo Functionality
Operations are logged to `.abook-org.log` file in the output directory:
- `LogEntry` struct (types.go:303) tracks source/target paths and files
- Logging is handled in `logging.go`
- The `--undo` flag reverses operations based on the log file

### Path Sanitization
Path generation must handle special characters and spaces:
- Implemented in `path.go`
- `CleanSeriesName()` removes series numbers from names
- Space replacement is configurable via `--replace_space` flag

### TUI Architecture (Bubbletea)
The TUI uses the Elm Architecture pattern via Bubbletea:
- **Models**: Each screen is a separate model in `internal/tui/models/`
- **Messages**: Define state transitions between screens
- **Update**: Handle messages and return new state
- **View**: Render the current state
- **Screen flow**: Scan → BookList → Preview → Settings → Process

### Constants
Important constants are defined in organizer.go:16:
- `LogFileName = ".abook-org.log"`
- `MetadataFileName = "metadata.json"`
- `InvalidSeriesValue = "__INVALID_SERIES__"`

## Documentation Structure

The project has comprehensive documentation split across multiple files:

### User-Facing Documentation (`docs/`)
- **[README.md](README.md)**: Main entry point with user journey structure, equal GUI/TUI/CLI treatment
- **[docs/INSTALLATION.md](docs/INSTALLATION.md)**: Platform-specific installation for GUI and CLI/TUI binaries
- **[docs/GUI.md](docs/GUI.md)**: Comprehensive GUI guide with 4-screen workflow, field mapping, template builder
- **[docs/TUI.md](docs/TUI.md)**: TUI guide with keyboard shortcuts, workflow screens, metadata viewer
- **[docs/CLI.md](docs/CLI.md)**: Complete CLI reference with all flags, examples, Docker usage
- **[docs/CONFIGURATION.md](docs/CONFIGURATION.md)**: Config files, environment variables, precedence rules
- **[docs/METADATA.md](docs/METADATA.md)**: Metadata sources (JSON, EPUB, MP3, M4B), field mapping, hybrid mode
- **[docs/LAYOUTS.md](docs/LAYOUTS.md)**: All 6 layout options with examples and recommendations
- **[docs/RENAME_FEATURE.md](docs/RENAME_FEATURE.md)**: File renaming with templates
- **[docs/METADATA_COMMAND.md](docs/METADATA_COMMAND.md)**: Interactive metadata viewer

### Development Documentation
- **[CLAUDE.md](CLAUDE.md)**: This file - guidance for AI assistants and developers
- **[CONTRIBUTING.md](CONTRIBUTING.md)**: Contribution guidelines
- **[LICENSE](LICENSE)**: MIT License

### Screenshot Placeholders (`docs/screenshots/`)
- GUI screenshots to be added by maintainers:
  - `gui-main-screen.png` - Book list with metadata preview
  - `gui-field-mapping.png` - Field mapping configuration dialog
  - `gui-preview-changes.png` - Preview screen with conflict detection
  - `gui-template-builder.png` - Filename template builder

## Dependencies

Key external dependencies:
- **Cobra** (`github.com/spf13/cobra`): CLI framework
- **Viper** (`github.com/spf13/viper`): Configuration management
- **Bubbletea** (`github.com/charmbracelet/bubbletea`): TUI framework
- **Lipgloss** (`github.com/charmbracelet/lipgloss`): TUI styling
- **dhowden/tag** (`github.com/dhowden/tag`): Audio file metadata extraction
- **pirmd/epub** (`github.com/pirmd/epub`): EPUB file parsing
- **fatih/color** (`github.com/fatih/color`): Terminal color output
