package organizer

import (
	"fmt"
	"regexp"
	"strings"
)

// Template represents a parsed filename template
type Template struct {
	raw    string
	tokens []templateToken
}

// templateToken represents a parsed component (literal text or placeholder)
type templateToken struct {
	isPlaceholder bool
	value         string
	fallback      string // Optional fallback for missing fields
}

// TemplateRenderer handles rendering templates with metadata
type TemplateRenderer struct {
	template        *Template
	authorFormatter *AuthorFormatter
}

// TemplateField describes an available template field
type TemplateField struct {
	Name        string
	Description string
	Example     string
}

// Template parsing regex - matches {field_name} or {field_name|fallback}
var templateRegex = regexp.MustCompile(`\{([^}]+)\}`)

// ParseTemplate parses a template string into tokens
func ParseTemplate(templateStr string) (*Template, error) {
	tokens := []templateToken{}
	lastIdx := 0

	matches := templateRegex.FindAllStringSubmatchIndex(templateStr, -1)
	for _, match := range matches {
		// Add literal text before placeholder
		if match[0] > lastIdx {
			tokens = append(tokens, templateToken{
				isPlaceholder: false,
				value:         templateStr[lastIdx:match[0]],
			})
		}

		// Parse placeholder (support fallback: {field|fallback})
		fieldSpec := templateStr[match[2]:match[3]]
		parts := strings.Split(fieldSpec, "|")

		token := templateToken{
			isPlaceholder: true,
			value:         strings.TrimSpace(parts[0]),
		}
		if len(parts) > 1 {
			token.fallback = strings.TrimSpace(parts[1])
		}
		tokens = append(tokens, token)

		lastIdx = match[1]
	}

	// Add remaining literal text
	if lastIdx < len(templateStr) {
		tokens = append(tokens, templateToken{
			isPlaceholder: false,
			value:         templateStr[lastIdx:],
		})
	}

	return &Template{raw: templateStr, tokens: tokens}, nil
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(template *Template, authorFormatter *AuthorFormatter) *TemplateRenderer {
	return &TemplateRenderer{
		template:        template,
		authorFormatter: authorFormatter,
	}
}

// Render applies metadata to template and returns filename
func (tr *TemplateRenderer) Render(metadata Metadata) (string, error) {
	var result strings.Builder

	for _, token := range tr.template.tokens {
		if !token.isPlaceholder {
			// Literal text
			result.WriteString(token.value)
			continue
		}

		// Resolve placeholder
		value := tr.resolveField(token.value, metadata)
		if value == "" && token.fallback != "" {
			value = token.fallback
		}

		result.WriteString(value)
	}

	return result.String(), nil
}

// resolveField resolves a template field name to its value from metadata
func (tr *TemplateRenderer) resolveField(fieldName string, metadata Metadata) string {
	switch strings.ToLower(fieldName) {
	case "author":
		if len(metadata.Authors) > 0 {
			return tr.authorFormatter.FormatAuthor(metadata.Authors[0])
		}
		return ""

	case "authors":
		if len(metadata.Authors) == 0 {
			return ""
		}
		formatted := make([]string, len(metadata.Authors))
		for i, author := range metadata.Authors {
			formatted[i] = tr.authorFormatter.FormatAuthor(author)
		}
		return strings.Join(formatted, ", ")

	case "title":
		return metadata.Title

	case "series":
		// Return series name without number
		series := metadata.GetValidSeries()
		if series == "" {
			return ""
		}
		return CleanSeriesName(series)

	case "series_full":
		return metadata.GetValidSeries()

	case "series_number":
		return GetSeriesNumberFromMetadata(metadata)

	case "album":
		return metadata.Album

	case "track":
		if metadata.TrackNumber > 0 {
			return fmt.Sprintf("%02d", metadata.TrackNumber)
		}
		return ""

	case "year":
		if year, ok := metadata.RawData["year"].(int); ok {
			return fmt.Sprintf("%d", year)
		}
		if year, ok := metadata.RawData["year"].(float64); ok {
			return fmt.Sprintf("%d", int(year))
		}
		return ""

	case "narrator":
		if narrator, ok := metadata.RawData["narrator"].(string); ok {
			return narrator
		}
		return ""

	default:
		// Try RawData for custom fields
		if val, ok := metadata.RawData[fieldName]; ok {
			return fmt.Sprintf("%v", val)
		}
		return ""
	}
}

// GetAvailableFields returns list of supported template fields
func GetAvailableFields() []TemplateField {
	return []TemplateField{
		{
			Name:        "author",
			Description: "First author (with format control)",
			Example:     "Brandon Sanderson",
		},
		{
			Name:        "authors",
			Description: "All authors (comma-separated)",
			Example:     "Stephen King, Peter Straub",
		},
		{
			Name:        "title",
			Description: "Book title",
			Example:     "The Final Empire",
		},
		{
			Name:        "series",
			Description: "Series name (without number)",
			Example:     "Mistborn",
		},
		{
			Name:        "series_full",
			Description: "Series name with number",
			Example:     "Mistborn #1",
		},
		{
			Name:        "series_number",
			Description: "Just the series number",
			Example:     "1",
		},
		{
			Name:        "album",
			Description: "Album field (audio files)",
			Example:     "The Stormlight Archive",
		},
		{
			Name:        "track",
			Description: "Track number (zero-padded)",
			Example:     "01",
		},
		{
			Name:        "year",
			Description: "Publication year",
			Example:     "2006",
		},
		{
			Name:        "narrator",
			Description: "Narrator (if available)",
			Example:     "Michael Kramer",
		},
	}
}

// ValidateTemplate checks if template is syntactically valid
func ValidateTemplate(templateStr string) error {
	// For now, we're lenient with template validation
	// Just check that we can parse it
	_, err := ParseTemplate(templateStr)
	return err
}
