// Package organizer provides a public API for the audiobook organizer functionality.
// This wraps the internal organizer package to allow external modules (like the GUI) to use it.
package organizer

import (
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// Re-export types from internal package
type (
	OrganizerConfig = organizer.OrganizerConfig
	Organizer       = organizer.Organizer
	Metadata        = organizer.Metadata
	FieldMapping    = organizer.FieldMapping
	LogEntry        = organizer.LogEntry
	Summary         = organizer.Summary
	MoveSummary     = organizer.MoveSummary
)

// Re-export functions
var (
	NewOrganizer        = organizer.NewOrganizer
	DefaultFieldMapping = organizer.DefaultFieldMapping
	AudioFieldMapping   = organizer.AudioFieldMapping
	EpubFieldMapping    = organizer.EpubFieldMapping
	NewMetadata         = organizer.NewMetadata
)

// CreateSanitizerFunc creates a path sanitizer function based on configuration.
// This is used by LayoutCalculator to clean path components.
func CreateSanitizerFunc(config *OrganizerConfig) func(string) string {
	// Create a temporary organizer just to use its SanitizePath method
	// This is a lightweight operation since we only need the sanitizer
	org, err := NewOrganizer(config)
	if err != nil {
		// If organizer creation fails, return a no-op sanitizer
		return func(s string) string { return s }
	}
	return org.SanitizePath
}
