package organizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RenamerConfig contains all configuration for renaming operations
type RenamerConfig struct {
	BaseDir             string       // Directory to scan for files
	Template            string       // Filename template string
	DryRun              bool         // Preview mode, don't execute
	Verbose             bool         // Detailed output
	AuthorFormat        AuthorFormat // How to format author names
	Recursive           bool         // Recursively process subdirectories
	FieldMapping        FieldMapping // Field mapping for metadata
	ReplaceSpace        string       // Character to replace spaces
	StrictMode          bool         // Error on missing template fields
	PreservePath        bool         // Only rename filename, keep directory
	PromptEnabled       bool         // Prompt before renaming each file
	UseEmbeddedMetadata bool         // Force embedded metadata, ignore metadata.json
}

// Validate checks if the configuration is valid and returns helpful error messages
func (c *RenamerConfig) Validate() error {
	// Check base directory
	if c.BaseDir == "" {
		return fmt.Errorf(
			"base directory is required\n\nPlease specify a directory to scan:\n  --dir=/path/to/audiobooks",
		)
	}

	// Check if directory exists
	info, err := os.Stat(c.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(
				"base directory does not exist: %s\n\nPlease check the path and try again",
				c.BaseDir,
			)
		}
		if os.IsPermission(err) {
			return fmt.Errorf(
				"permission denied accessing: %s\n\nTry running with appropriate permissions:\n  sudo audiobook-organizer rename --dir=%s",
				c.BaseDir,
				c.BaseDir,
			)
		}
		return fmt.Errorf("error accessing base directory %s: %w", c.BaseDir, err)
	}

	// Verify it's a directory
	if !info.IsDir() {
		return fmt.Errorf(
			"%s is not a directory\n\nPlease specify a directory, not a file",
			c.BaseDir,
		)
	}

	// Validate template if provided
	if c.Template != "" {
		if err := ValidateTemplate(c.Template); err != nil {
			return fmt.Errorf(
				"invalid template: %w\n\nTemplate must use valid field placeholders like {author}, {title}, {series}\nSee available fields with: audiobook-organizer rename --help-template",
				err,
			)
		}
	}

	// Validate author format
	switch c.AuthorFormat {
	case AuthorFormatFirstLast, AuthorFormatLastFirst, AuthorFormatPreserve:
		// Valid formats
	default:
		return fmt.Errorf(
			"invalid author format: %d\n\nValid options are:\n  first-last  (e.g., Brandon Sanderson)\n  last-first  (e.g., Sanderson, Brandon)\n  preserve    (keep original format)",
			c.AuthorFormat,
		)
	}

	// Validate replace_space character (should be single char or empty)
	if len(c.ReplaceSpace) > 1 {
		return fmt.Errorf(
			"replace_space must be a single character, got: %q\n\nExamples:\n  --replace_space=_\n  --replace_space=.\n  --replace_space=-",
			c.ReplaceSpace,
		)
	}

	return nil
}

// Renamer performs file renaming operations
type Renamer struct {
	config           RenamerConfig
	templateRenderer *TemplateRenderer
	logEntries       []RenameLogEntry
	summary          RenameSummary
}

// RenameLogEntry tracks a rename operation for undo
type RenameLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	OldPath   string    `json:"old_path"`
	NewPath   string    `json:"new_path"`
	Metadata  Metadata  `json:"metadata,omitempty"`
}

// RenameCandidate represents a file that can be renamed
type RenameCandidate struct {
	CurrentPath  string
	ProposedPath string
	Metadata     Metadata
	IsNoOp       bool   // File already has target name
	IsConflict   bool   // Duplicate target name
	Error        string // If preview generation failed
}

// RenameSummary tracks rename operation results
type RenameSummary struct {
	FilesScanned   int
	FilesRenamed   int
	FilesSkipped   int
	ConflictsFound int
	Errors         []string
}

// ConflictResolver handles filename conflicts
type ConflictResolver struct {
	seen map[string]int // filename → occurrence count
}

// NewConflictResolver creates a conflict resolver
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{
		seen: make(map[string]int),
	}
}

// CheckConflict checks if filename conflicts, returns resolved name and conflict flag
func (cr *ConflictResolver) CheckConflict(filename string) (string, bool) {
	count := cr.seen[filename]
	cr.seen[filename]++

	if count == 0 {
		return filename, false // No conflict
	}

	// Generate unique filename: "file.m4b" → "file (2).m4b"
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	newFilename := fmt.Sprintf("%s (%d)%s", base, count+1, ext)

	return newFilename, true // Conflict detected and resolved
}

// NewRenamer creates a new renaming engine
func NewRenamer(config *RenamerConfig) (*Renamer, error) {
	// Validate configuration first
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Parse template
	template, err := ParseTemplate(config.Template)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	// Create template renderer
	authorFormatter := NewAuthorFormatter(config.AuthorFormat)
	renderer := NewTemplateRenderer(template, authorFormatter)

	return &Renamer{
		config:           *config,
		templateRenderer: renderer,
		logEntries:       []RenameLogEntry{},
	}, nil
}

// Execute performs the rename operation
func (r *Renamer) Execute() error {
	// 1. Scan files
	candidates, err := r.ScanFiles()
	if err != nil {
		return err
	}

	// 2. Filter out no-ops
	toRename := filterRenameableCandidates(candidates)

	// 3. Detect conflicts
	conflicts := detectConflicts(toRename)
	r.summary.ConflictsFound = len(conflicts)

	// 4. Execute renames
	for _, candidate := range toRename {
		// Skip if prompt is enabled and user declines
		if r.config.PromptEnabled {
			if !r.promptForRename(candidate) {
				r.summary.FilesSkipped++
				continue
			}
		}

		if err := r.RenameFile(candidate.CurrentPath, candidate.ProposedPath); err != nil {
			r.summary.Errors = append(r.summary.Errors, err.Error())
			continue
		}
		r.summary.FilesRenamed++
	}

	// 5. Save log
	if !r.config.DryRun && len(r.logEntries) > 0 {
		if err := r.SaveLog(); err != nil {
			return err
		}
	}

	return nil
}

// ScanFiles finds renameable files in directory
func (r *Renamer) ScanFiles() ([]RenameCandidate, error) {
	var candidates []RenameCandidate

	err := filepath.Walk(r.config.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			if !r.config.Recursive && path != r.config.BaseDir {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if supported file type
		ext := strings.ToLower(filepath.Ext(path))
		if !IsSupportedFile(ext) {
			return nil
		}

		r.summary.FilesScanned++

		// Determine metadata provider based on configuration
		// ALWAYS use NewMetadataProvider which auto-detects and does hybrid extraction
		var provider MetadataProvider
		provider = NewMetadataProvider(path, r.config.UseEmbeddedMetadata)

		// Note: NewMetadataProvider will automatically:
		// - Detect file type (audio, epub, json)
		// - For audio files, check for metadata.json in parent dir (unless UseEmbeddedMetadata is true)
		// - Do hybrid extraction (JSON + embedded) if metadata.json exists and UseEmbeddedMetadata is false
		// - Use only embedded metadata if no metadata.json or UseEmbeddedMetadata is true

		// Extract metadata
		metadata, err := provider.GetMetadata()
		if err != nil {
			candidates = append(candidates, RenameCandidate{
				CurrentPath: path,
				Error:       fmt.Sprintf("Failed to extract metadata: %v", err),
			})
			return nil
		}

		// Apply field mapping if configured
		if !r.config.FieldMapping.IsEmpty() {
			metadata.ApplyFieldMapping(r.config.FieldMapping)
		}

		// Generate new path
		newPath, err := r.GenerateNewPath(path, metadata)
		if err != nil {
			candidates = append(candidates, RenameCandidate{
				CurrentPath: path,
				Metadata:    metadata,
				Error:       fmt.Sprintf("Failed to render template: %v", err),
			})
			return nil
		}

		// Check if already correct name
		isNoOp := path == newPath

		candidates = append(candidates, RenameCandidate{
			CurrentPath:  path,
			ProposedPath: newPath,
			Metadata:     metadata,
			IsNoOp:       isNoOp,
		})

		return nil
	})

	return candidates, err
}

// GenerateNewPath generates the new path for a file based on metadata
func (r *Renamer) GenerateNewPath(currentPath string, metadata Metadata) (string, error) {
	// Render template to get new filename (without extension)
	newFilename, err := r.templateRenderer.Render(metadata)
	if err != nil {
		return "", err
	}

	// Sanitize filename
	newFilename = r.sanitizeFilename(newFilename)

	// Preserve extension
	ext := filepath.Ext(currentPath)
	if !strings.HasSuffix(newFilename, ext) {
		newFilename += ext
	}

	// Construct new path
	var newPath string
	if r.config.PreservePath {
		// Only rename filename, keep directory
		dir := filepath.Dir(currentPath)
		newPath = filepath.Join(dir, newFilename)
	} else {
		// Move to base directory with new filename
		newPath = filepath.Join(r.config.BaseDir, newFilename)
	}

	return newPath, nil
}

// sanitizeFilename sanitizes a filename
func (r *Renamer) sanitizeFilename(filename string) string {
	// Replace spaces if configured
	if r.config.ReplaceSpace != "" {
		filename = strings.ReplaceAll(filename, " ", r.config.ReplaceSpace)
	}

	// Use SanitizePath logic (create temporary organizer for this)
	// For now, just do basic sanitization
	invalidChars := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	return filename
}

// RenameFile renames a single file
func (r *Renamer) RenameFile(oldPath, newPath string) error {
	if r.config.Verbose {
		PrintBlue("Renaming: %s → %s", filepath.Base(oldPath), filepath.Base(newPath))
	}

	if r.config.DryRun {
		return nil
	}

	// Perform rename
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename %s: %w", oldPath, err)
	}

	// Log operation
	r.logEntries = append(r.logEntries, RenameLogEntry{
		Timestamp: time.Now(),
		OldPath:   oldPath,
		NewPath:   newPath,
	})

	return nil
}

// SaveLog saves rename operations to log file
func (r *Renamer) SaveLog() error {
	logPath := filepath.Join(r.config.BaseDir, ".abook-rename.log")
	data, err := json.MarshalIndent(r.logEntries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(logPath, data, 0o644)
}

// UndoRenames reverses rename operations from log
func (r *Renamer) UndoRenames() error {
	logPath := filepath.Join(r.config.BaseDir, ".abook-rename.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("no rename log found at %s", logPath)
	}

	var entries []RenameLogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("error parsing log: %v", err)
	}

	// Process in reverse order
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if r.config.Verbose {
			PrintYellow(
				"Undoing: %s → %s",
				filepath.Base(entry.NewPath),
				filepath.Base(entry.OldPath),
			)
		}

		if !r.config.DryRun {
			if err := os.Rename(entry.NewPath, entry.OldPath); err != nil {
				return fmt.Errorf("failed to undo rename: %w", err)
			}
		}
	}

	// Remove log file after successful undo
	if !r.config.DryRun {
		return os.Remove(logPath)
	}

	return nil
}

// GetSummary returns the rename summary
func (r *Renamer) GetSummary() RenameSummary {
	return r.summary
}

// promptForRename prompts user for confirmation before renaming
func (r *Renamer) promptForRename(candidate RenameCandidate) bool {
	fmt.Printf("\nRename file?\n")
	fmt.Printf("  From: %s\n", filepath.Base(candidate.CurrentPath))
	fmt.Printf("  To:   %s\n", filepath.Base(candidate.ProposedPath))
	fmt.Printf("Proceed? [y/N]: ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y" || response == "yes"
}

// Helper functions

// filterRenameableCandidates filters out no-ops and errors
func filterRenameableCandidates(candidates []RenameCandidate) []RenameCandidate {
	var renameable []RenameCandidate
	for _, candidate := range candidates {
		if candidate.Error != "" || candidate.IsNoOp {
			continue
		}
		renameable = append(renameable, candidate)
	}
	return renameable
}

// detectConflicts detects and resolves filename conflicts
func detectConflicts(candidates []RenameCandidate) []RenameCandidate {
	resolver := NewConflictResolver()
	var conflicts []RenameCandidate

	for i := range candidates {
		filename := filepath.Base(candidates[i].ProposedPath)
		resolvedName, isConflict := resolver.CheckConflict(filename)

		if isConflict {
			candidates[i].IsConflict = true
			candidates[i].ProposedPath = filepath.Join(
				filepath.Dir(candidates[i].ProposedPath),
				resolvedName,
			)
			conflicts = append(conflicts, candidates[i])
		}
	}

	return conflicts
}
