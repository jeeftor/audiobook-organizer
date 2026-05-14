package organizer

import (
	"strings"
)

// AuthorFormatter handles author name formatting
type AuthorFormatter struct {
	format AuthorFormat
}

// AuthorFormat specifies how to format author names
type AuthorFormat int

const (
	AuthorFormatFirstLast AuthorFormat = iota // "Brandon Sanderson"
	AuthorFormatLastFirst                     // "Sanderson, Brandon"
	AuthorFormatPreserve                      // Keep original format
)

// NewAuthorFormatter creates formatter with specified format
func NewAuthorFormatter(format AuthorFormat) *AuthorFormatter {
	return &AuthorFormatter{
		format: format,
	}
}

// FormatAuthor converts author name to desired format
func (af *AuthorFormatter) FormatAuthor(authorName string) string {
	switch af.format {
	case AuthorFormatFirstLast:
		return ConvertToFirstLast(authorName)
	case AuthorFormatLastFirst:
		return ConvertToLastFirst(authorName)
	case AuthorFormatPreserve:
		return authorName
	default:
		return authorName
	}
}

// DetectFormat determines if name is "Last, First" or "First Last"
func DetectFormat(authorName string) AuthorFormat {
	// If contains comma, assume "Last, First" format
	if strings.Contains(authorName, ",") {
		return AuthorFormatLastFirst
	}
	return AuthorFormatFirstLast
}

// ConvertToFirstLast converts "Last, First" → "First Last"
func ConvertToFirstLast(authorName string) string {
	if !strings.Contains(authorName, ",") {
		return authorName // Already First Last
	}

	parts := strings.Split(authorName, ",")
	if len(parts) != 2 {
		return authorName // Can't convert, return as-is
	}

	// "Last, First" → "First Last"
	return strings.TrimSpace(parts[1]) + " " + strings.TrimSpace(parts[0])
}

// ConvertToLastFirst converts "First Last" → "Last, First"
func ConvertToLastFirst(authorName string) string {
	if strings.Contains(authorName, ",") {
		return authorName // Already Last, First
	}

	parts := strings.Fields(authorName)
	if len(parts) < 2 {
		return authorName // Single name, can't convert
	}

	// Assume last word is last name, rest is first name
	// "Brandon von Sanderson" → "Sanderson, Brandon von"
	lastName := parts[len(parts)-1]
	firstName := strings.Join(parts[:len(parts)-1], " ")

	return lastName + ", " + firstName
}
