package organizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasCommonPrefixExported(t *testing.T) {
	tests := []struct {
		name     string
		str1     string
		str2     string
		expected bool
	}{
		{
			name:     "dash separator",
			str1:     "My Audiobook - Track 01",
			str2:     "My Audiobook - Track 02",
			expected: true,
		},
		{
			name:     "colon separator",
			str1:     "The Great Book: Chapter 1",
			str2:     "The Great Book: Chapter 2",
			expected: true,
		},
		{
			name:     "comma separator",
			str1:     "Series Name, Part 1",
			str2:     "Series Name, Part 2",
			expected: true,
		},
		{
			name:     "no common prefix",
			str1:     "Alpha Book",
			str2:     "Beta Book",
			expected: false,
		},
		{
			name:     "completely different",
			str1:     "Foundation",
			str2:     "Dune",
			expected: false,
		},
		{
			name:     "similar but no separator",
			str1:     "My Book",
			str2:     "My Book",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasCommonPrefix(tt.str1, tt.str2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasTrackNumberPatternExported(t *testing.T) {
	tests := []struct {
		name     string
		str1     string
		str2     string
		expected bool
	}{
		{
			name:     "track keyword",
			str1:     "Book Track 1",
			str2:     "Book Track 2",
			expected: true,
		},
		{
			name:     "part keyword",
			str1:     "Foundation Part 1",
			str2:     "Foundation Part 2",
			expected: true,
		},
		{
			name:     "chapter keyword",
			str1:     "Chapter 1",
			str2:     "Chapter 2",
			expected: true,
		},
		{
			name:     "disc keyword",
			str1:     "Audiobook Disc 1",
			str2:     "Audiobook Disc 2",
			expected: true,
		},
		{
			name:     "episode keyword",
			str1:     "Podcast Episode 1",
			str2:     "Podcast Episode 2",
			expected: true,
		},
		{
			name:     "volume keyword",
			str1:     "Series Volume 1",
			str2:     "Series Volume 2",
			expected: true,
		},
		{
			name:     "case insensitive",
			str1:     "Book TRACK 1",
			str2:     "Book track 2",
			expected: true,
		},
		{
			name:     "no pattern match",
			str1:     "Alpha",
			str2:     "Bravo",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasTrackNumberPattern(tt.str1, tt.str2)
			assert.Equal(t, tt.expected, result)
		})
	}
}
