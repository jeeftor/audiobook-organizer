package guiapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FilePair records the original filename in the source directory and the
// resulting filename in the target directory (which may differ due to renaming).
// This matches the FilePair type written by the core library's LogEntry.
type FilePair struct {
	From string `json:"from"` // original filename in source directory
	To   string `json:"to"`   // filename in target directory (may be renamed)
}

// MoveLogEntry represents a single file move operation
type MoveLogEntry struct {
	Timestamp  time.Time  `json:"timestamp"`
	SourcePath string     `json:"source_path"`
	TargetPath string     `json:"target_path"`
	Files      []FilePair `json:"files"`
}

// UndoLastOperation reverses the most recent organization operation
func (a *App) UndoLastOperation() (map[string]interface{}, error) {
	a.log("Starting undo operation")

	// Find the log file
	logPath := filepath.Join(a.config.OutputDir, ".abook-org.log")

	// Read the log file
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read operation log: %w", err)
	}

	// Parse the log
	var entries []MoveLogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse operation log: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no operations to undo")
	}

	a.log("Found %d operations in log", len(entries))

	// Get the timestamp of the last operation
	lastTimestamp := entries[len(entries)-1].Timestamp

	// Find all entries from the last operation (same timestamp range - within 1 second)
	var lastOpEntries []MoveLogEntry
	for _, entry := range entries {
		if entry.Timestamp.After(lastTimestamp.Add(-1 * time.Second)) {
			lastOpEntries = append(lastOpEntries, entry)
		}
	}

	a.log("Undoing %d file moves from last operation", len(lastOpEntries))

	// Reverse the moves
	undoneCount := 0
	errors := make([]string, 0)

	for _, entry := range lastOpEntries {
		for _, file := range entry.Files {
			sourcePath := filepath.Join(entry.TargetPath, file.To)   // find by target name
			targetPath := filepath.Join(entry.SourcePath, file.From) // restore to original name

			a.log("Moving back: %s -> %s", sourcePath, targetPath)

			// Ensure target directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				errors = append(
					errors,
					fmt.Sprintf("Failed to create directory for %s: %v", file, err),
				)
				continue
			}

			// Move the file back
			if err := os.Rename(sourcePath, targetPath); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to move %s: %v", file, err))
				continue
			}

			undoneCount++
		}
	}

	// Clean up empty directories
	dirsToRemove := make(map[string]bool)
	for _, entry := range lastOpEntries {
		dirsToRemove[entry.TargetPath] = true
	}

	for dir := range dirsToRemove {
		// Try to remove directory (will only succeed if empty)
		os.Remove(dir)
		// Try to remove parent directories too
		parent := filepath.Dir(dir)
		os.Remove(parent)
	}

	// Remove the undone entries from the log
	remainingEntries := make([]MoveLogEntry, 0)
	for _, entry := range entries {
		if entry.Timestamp.Before(lastTimestamp.Add(-1 * time.Second)) {
			remainingEntries = append(remainingEntries, entry)
		}
	}

	// Write updated log
	updatedData, err := json.MarshalIndent(remainingEntries, "", "  ")
	if err != nil {
		a.log("Warning: Failed to update log file: %v", err)
	} else {
		if err := os.WriteFile(logPath, updatedData, 0o644); err != nil {
			a.log("Warning: Failed to write updated log: %v", err)
		}
	}

	a.log("Undo complete: %d files restored, %d errors", undoneCount, len(errors))

	return map[string]interface{}{
		"success":       len(errors) == 0,
		"filesRestored": undoneCount,
		"errors":        errors,
	}, nil
}

// GetUndoInfo returns information about what can be undone
func (a *App) GetUndoInfo() (map[string]interface{}, error) {
	logPath := filepath.Join(a.config.OutputDir, ".abook-org.log")

	data, err := os.ReadFile(logPath)
	if err != nil {
		return map[string]interface{}{
			"canUndo":        false,
			"operationCount": 0,
		}, nil
	}

	var entries []MoveLogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return map[string]interface{}{
			"canUndo":        false,
			"operationCount": 0,
		}, nil
	}

	// Count operations by timestamp
	lastTimestamp := entries[len(entries)-1].Timestamp
	lastOpCount := 0
	for _, entry := range entries {
		if entry.Timestamp.After(lastTimestamp.Add(-1 * time.Second)) {
			lastOpCount += len(entry.Files)
		}
	}

	return map[string]interface{}{
		"canUndo":        true,
		"operationCount": lastOpCount,
		"lastOperation":  lastTimestamp,
	}, nil
}
