// Package organizer provides public utility functions
package organizer

import (
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// Path and file utilities
var (
	// IsSupportedFile checks if a file extension is supported
	IsSupportedFile = organizer.IsSupportedFile

	// IsSupportedFileForFlatMode checks if a file extension is supported in flat mode
	IsSupportedFileForFlatMode = organizer.IsSupportedFileForFlatMode
)

// Album detection utilities
var (
	// HasCommonPrefix checks if two strings share a common prefix
	HasCommonPrefix = organizer.HasCommonPrefix

	// HasTrackNumberPattern checks if strings have track number patterns
	HasTrackNumberPattern = organizer.HasTrackNumberPattern
)

// Series utilities
var (
	// ExtractSeriesNumber extracts series number from a series string
	ExtractSeriesNumber = organizer.ExtractSeriesNumber
)

// Layout calculator
type LayoutCalculator = organizer.LayoutCalculator

var (
	// NewLayoutCalculator creates a new layout calculator
	NewLayoutCalculator = organizer.NewLayoutCalculator
)

// File operations
type FileOps = organizer.FileOps

var (
	// NewFileOps creates a new file operations handler
	NewFileOps = organizer.NewFileOps
)

// Renamer types and functions
type (
	Renamer         = organizer.Renamer
	RenamerConfig   = organizer.RenamerConfig
	RenameCandidate = organizer.RenameCandidate
	RenameSummary   = organizer.RenameSummary
)

var (
	// NewRenamer creates a new renamer
	NewRenamer = organizer.NewRenamer
)

// Color printing utilities
var (
	PrintRed     = organizer.PrintRed
	PrintGreen   = organizer.PrintGreen
	PrintYellow  = organizer.PrintYellow
	PrintBlue    = organizer.PrintBlue
	PrintMagenta = organizer.PrintMagenta
	PrintCyan    = organizer.PrintCyan
	PrintBase    = organizer.PrintBase
)
