package models

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderMetadataForFile renders metadata for a specific file with column-aware truncation
func (m *RenameFieldMappingModel) renderMetadataForFile(candidateIdx, displayNum, columnWidth int) string {
	if candidateIdx >= len(m.candidates) {
		return "No data"
	}

	candidate := m.candidates[candidateIdx]
	metadata := candidate.Metadata

	var content strings.Builder

	// Color styles
	titleLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	authorLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	seriesLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	defaultLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAFF"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))

	content.WriteString(titleStyle.Render(fmt.Sprintf("Metadata Preview (#%d):", displayNum)) + "\n\n")

	// File info - truncate based on column width (more generous)
	filename := filepath.Base(candidate.CurrentPath)
	maxFilenameLen := columnWidth - 8 // Account for "File: " label and minimal padding
	if maxFilenameLen < 20 {
		maxFilenameLen = 20 // Minimum reasonable length
	}
	if len(filename) > maxFilenameLen {
		filename = filename[:maxFilenameLen-3] + "..."
	}
	content.WriteString(defaultLabelStyle.Render("File: ") + valueStyle.Render(filename) + "\n")
	content.WriteString(defaultLabelStyle.Render("Source Type: ") + valueStyle.Render(metadata.SourceType) + "\n\n")

	// Raw metadata fields (sorted, limited to fit in column)
	rawLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	content.WriteString(titleStyle.Render("Raw Metadata Fields:") + "\n")

	fieldMapping := m.config.FieldMapping

	// Collect and sort keys, filtering out unwanted fields
	var keys []string
	excludedFields := map[string]bool{
		"chapters":    true, // Large array of chapter data
		"description": true, // Long HTML description
		"tags":        true, // Usually not needed for renaming
	}

	for key, val := range metadata.RawData {
		if val == nil || val == "" {
			continue
		}
		// Skip excluded fields
		if excludedFields[key] {
			continue
		}
		keys = append(keys, key)
	}

	// Add series if exists
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

	// Show all fields
	for _, key := range keys {
		var val interface{}
		if key == "series" && len(metadata.Series) > 0 {
			if rawVal, ok := metadata.RawData[key]; ok && rawVal != nil && rawVal != "" {
				val = rawVal
			} else {
				val = metadata.Series[0]
			}
		} else {
			val = metadata.RawData[key]
		}

		// Determine field indicator
		fieldIndicator := ""
		if key == fieldMapping.TitleField {
			fieldIndicator = " " + titleLabelStyle.Render("<- TITLE")
		} else if key == fieldMapping.SeriesField {
			fieldIndicator = " " + seriesLabelStyle.Render("<- SERIES")
		} else if key == fieldMapping.TrackField {
			fieldIndicator = " " + defaultLabelStyle.Render("<- TRACK")
		} else {
			for _, af := range fieldMapping.AuthorFields {
				if key == af {
					fieldIndicator = " " + authorLabelStyle.Render("<- AUTHOR")
					break
				}
			}
		}

		// Truncate value if too long (more generous)
		valStr := fmt.Sprintf("%v", val)
		maxValLen := columnWidth - len(key) - 12 // Account for key, indicators, minimal padding
		if maxValLen < 20 {
			maxValLen = 20 // Minimum reasonable length
		}
		if len(valStr) > maxValLen {
			valStr = valStr[:maxValLen-3] + "..."
		}

		content.WriteString(fmt.Sprintf("  %s: %s%s\n", rawLabelStyle.Render(key), valStr, fieldIndicator))
	}

	return content.String()
}
