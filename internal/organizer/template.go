package organizer

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Template represents a parsed filename template
type Template struct {
	raw    string
	tokens []templateToken
}

type templateTokenKind int

const (
	tokenLiteral templateTokenKind = iota
	tokenSimple
	tokenComposite
)

// templateToken represents a parsed component (literal text or placeholder)
type templateToken struct {
	kind      templateTokenKind
	value     string
	fallback  string
	format    string
	composite []templatePart
}

type templatePart struct {
	isField bool
	literal string
	field   string
	format  string
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
var (
	templateRegex       = regexp.MustCompile(`\{([^}]+)\}`)
	dollarTemplateRegex = regexp.MustCompile(`\$\{([^}]+)\}`)
	simpleFieldRegex    = regexp.MustCompile(`^([a-zA-Z_][\w-]*)(?::(\d+))?$`)
)

var knownTemplateFields = []string{
	"series_number",
	"series_count",
	"series_full",
	"narrators",
	"narrator",
	"authors",
	"author",
	"series",
	"title",
	"album",
	"track",
	"year",
}

func init() {
	sort.Slice(knownTemplateFields, func(i, j int) bool {
		return len(knownTemplateFields[i]) > len(knownTemplateFields[j])
	})
}

// ParseTemplate parses a template string into tokens
func ParseTemplate(templateStr string) (*Template, error) {
	templateStr = dollarTemplateRegex.ReplaceAllString(templateStr, `{$1}`)
	if strings.Contains(templateStr, "{}") {
		return nil, fmt.Errorf("empty template placeholder")
	}

	tokens := []templateToken{}
	lastIdx := 0

	matches := templateRegex.FindAllStringSubmatchIndex(templateStr, -1)
	for _, match := range matches {
		if match[0] > lastIdx {
			tokens = append(tokens, templateToken{
				kind:  tokenLiteral,
				value: templateStr[lastIdx:match[0]],
			})
		}

		rawSpec := templateStr[match[2]:match[3]]
		if strings.TrimSpace(rawSpec) == "" {
			return nil, fmt.Errorf("empty template placeholder")
		}

		token, err := parseBraceToken(rawSpec)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)

		lastIdx = match[1]
	}

	if lastIdx < len(templateStr) {
		tokens = append(tokens, templateToken{
			kind:  tokenLiteral,
			value: templateStr[lastIdx:],
		})
	}

	return &Template{raw: templateStr, tokens: tokens}, nil
}

func parseBraceToken(rawSpec string) (templateToken, error) {
	fallback := ""
	spec := rawSpec
	if parts := strings.SplitN(rawSpec, "|", 2); len(parts) == 2 {
		spec = parts[0]
		fallback = strings.TrimSpace(parts[1])
	}

	trimmedSpec := strings.TrimSpace(spec)
	if isSimpleFieldSpec(trimmedSpec) {
		name, format := splitFieldNameAndFormat(trimmedSpec)
		return templateToken{
			kind:     tokenSimple,
			value:    name,
			format:   format,
			fallback: fallback,
		}, nil
	}

	parts, err := parseCompositeContent(spec)
	if err != nil {
		return templateToken{}, err
	}
	if len(parts) == 0 {
		return templateToken{}, fmt.Errorf("invalid template placeholder %q", rawSpec)
	}

	return templateToken{
		kind:      tokenComposite,
		composite: parts,
		fallback:  fallback,
	}, nil
}

func isSimpleFieldSpec(spec string) bool {
	if strings.Contains(spec, " ") {
		return false
	}
	return simpleFieldRegex.MatchString(spec)
}

func splitFieldNameAndFormat(spec string) (string, string) {
	matches := simpleFieldRegex.FindStringSubmatch(spec)
	if len(matches) < 2 {
		return spec, ""
	}
	format := ""
	if len(matches) > 2 {
		format = matches[2]
	}
	return matches[1], format
}

func parseCompositeContent(content string) ([]templatePart, error) {
	parts := []templatePart{}
	pos := 0

	for pos < len(content) {
		fieldName, format, start, end, found := findFieldReferenceAt(content, pos)
		if !found {
			parts = append(parts, templatePart{
				isField: false,
				literal: content[pos:],
			})
			break
		}

		if start > pos {
			parts = append(parts, templatePart{
				isField: false,
				literal: content[pos:start],
			})
		}

		parts = append(parts, templatePart{
			isField: true,
			field:   fieldName,
			format:  format,
		})
		pos = end
	}

	return parts, nil
}

func findFieldReferenceAt(content string, start int) (fieldName, format string, fieldStart, fieldEnd int, found bool) {
	earliest := -1

	for _, candidate := range knownTemplateFields {
		idx := strings.Index(content[start:], candidate)
		if idx == -1 {
			continue
		}
		idx += start
		if earliest != -1 && idx > earliest {
			continue
		}

		end := idx + len(candidate)
		nextFormat := ""
		if end < len(content) && content[end] == ':' {
			formatEnd := end + 1
			for formatEnd < len(content) && content[formatEnd] >= '0' && content[formatEnd] <= '9' {
				formatEnd++
			}
			if formatEnd > end+1 {
				nextFormat = content[end+1 : formatEnd]
				end = formatEnd
			}
		}

		if earliest == -1 || idx < earliest {
			earliest = idx
			fieldName = candidate
			format = nextFormat
			fieldStart = idx
			fieldEnd = end
			found = true
		}
	}

	return fieldName, format, fieldStart, fieldEnd, found
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
		switch token.kind {
		case tokenLiteral:
			result.WriteString(token.value)
		case tokenSimple:
			value := tr.resolveFieldFormatted(token.value, token.format, metadata)
			if value == "" && token.fallback != "" {
				value = token.fallback
			}
			result.WriteString(value)
		case tokenComposite:
			value, err := tr.renderCompositeToken(token, metadata)
			if err != nil {
				return "", err
			}
			result.WriteString(value)
		}
	}

	return result.String(), nil
}

func (tr *TemplateRenderer) renderCompositeToken(token templateToken, metadata Metadata) (string, error) {
	for _, part := range token.composite {
		if !part.isField {
			continue
		}
		if tr.resolveFieldFormatted(part.field, part.format, metadata) == "" {
			return "", nil
		}
	}

	var result strings.Builder
	for _, part := range token.composite {
		if !part.isField {
			result.WriteString(part.literal)
			continue
		}
		result.WriteString(tr.resolveFieldFormatted(part.field, part.format, metadata))
	}

	return result.String(), nil
}

func (tr *TemplateRenderer) resolveFieldFormatted(fieldName, format string, metadata Metadata) string {
	value := tr.resolveField(fieldName, metadata)
	return applyNumericFormat(value, format)
}

// resolveField resolves a template field name to its value from metadata
func (tr *TemplateRenderer) resolveField(fieldName string, metadata Metadata) string {
	normalizedFieldName := normalizeTemplateFieldName(fieldName)
	switch normalizedFieldName {
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
		series := metadata.GetValidSeries()
		if series == "" {
			return ""
		}
		return CleanSeriesName(series)

	case "series_full":
		return metadata.GetValidSeries()

	case "series_number", "series_count":
		return GetSeriesNumberFromMetadata(metadata)

	case "album":
		return metadata.Album

	case "track":
		if metadata.TrackNumber > 0 {
			return fmt.Sprintf("%02d", metadata.TrackNumber)
		}
		return ""

	case "year":
		if year, ok := rawTemplateValue(metadata, fieldName, normalizedFieldName).(int); ok {
			return fmt.Sprintf("%d", year)
		}
		if year, ok := rawTemplateValue(metadata, fieldName, normalizedFieldName).(float64); ok {
			return fmt.Sprintf("%d", int(year))
		}
		return ""

	case "narrator":
		return stringifyTemplateValue(rawTemplateValue(metadata, "narrator", "narrators"))

	case "narrators":
		return stringifyTemplateValue(rawTemplateValue(metadata, fieldName, normalizedFieldName))

	default:
		return stringifyTemplateValue(rawTemplateValue(metadata, fieldName, normalizedFieldName))
	}
}

func applyNumericFormat(value, format string) string {
	if value == "" || format == "" {
		return value
	}

	width, err := strconv.Atoi(format)
	if err != nil || width <= 0 {
		return value
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return fmt.Sprintf("%0*d", width, intValue)
	}

	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		if floatValue == float64(int(floatValue)) {
			return fmt.Sprintf("%0*d", width, int(floatValue))
		}
	}

	return value
}

func normalizeTemplateFieldName(fieldName string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(fieldName)), "-", "_")
}

func rawTemplateValue(metadata Metadata, fieldNames ...string) interface{} {
	if metadata.RawData == nil {
		return nil
	}
	for _, fieldName := range fieldNames {
		if val, ok := metadata.RawData[fieldName]; ok {
			return val
		}
		normalizedFieldName := normalizeTemplateFieldName(fieldName)
		if val, ok := metadata.RawData[normalizedFieldName]; ok {
			return val
		}
	}
	return nil
}

func stringifyTemplateValue(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []string:
		return strings.Join(typed, ", ")
	case []interface{}:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := stringifyTemplateValue(item); text != "" {
				values = append(values, text)
			}
		}
		return strings.Join(values, ", ")
	default:
		return fmt.Sprintf("%v", typed)
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
	_, err := ParseTemplate(templateStr)
	return err
}
