#!/bin/bash

# Change to the epub directory where the original file should be
cd "$(dirname "$0")/epub" || {
  echo "Error: Could not change to epub directory"
  exit 1
}

# Original EPUB file
ORIGINAL="title-author.epub"

# Check if original file exists
if [ ! -f "$ORIGINAL" ]; then
  echo "Error: $ORIGINAL does not exist in $(pwd)"
  echo "Available files:"
  ls -la *.epub 2>/dev/null || echo "No EPUB files found"
  exit 1
fi

# Base command for ebook-meta
EBOOK_META="/Applications/calibre.app/Contents/MacOS/ebook-meta"

# Check if ebook-meta is available
if [ ! -x "$EBOOK_META" ]; then
  echo "Error: ebook-meta not found at $EBOOK_META"
  exit 1
fi

# Function to sanitize filename (replace invalid characters for filenames)
sanitize_filename() {
  echo "$1" | tr -s '[:space:]' '_' | tr -d '/\\:*?"<>|'
}

# Function to create EPUB with specific metadata
create_epub() {
  local counter=$1
  local title="$2"
  local series="$3"
  local author="$4"
  local description="$5"

  # Generate filename
  local sanitized_title=$(sanitize_filename "$title")
  local output="strange_book_${counter}_${sanitized_title}.epub"

  # Copy the original file
  cp "$ORIGINAL" "$output"

  # Build command arguments
  local args=()

  if [ -n "$title" ]; then
    args+=(--title "$title")
  fi

  if [ -n "$series" ]; then
    args+=(--series "$series")
  fi

  if [ -n "$author" ]; then
    args+=(--authors "$author")
  fi

  # Apply metadata
  echo "Creating $output"
  echo "  Setting: title='$title', series='$series', author='$author'"
  "$EBOOK_META" "$output" "${args[@]}"

  # Read back and parse the actual metadata
  local metadata_output=$("$EBOOK_META" "$output")

  local actual_title=$(echo "$metadata_output" | grep "^Title" | sed 's/^Title[[:space:]]*:[[:space:]]*//')
  local actual_series=$(echo "$metadata_output" | grep "^Series" | sed 's/^Series[[:space:]]*:[[:space:]]*//')
  local actual_author=$(echo "$metadata_output" | grep "^Author(s)" | sed 's/^Author(s)[[:space:]]*:[[:space:]]*//')

  echo "  Actual: title='$actual_title', series='$actual_series', author='$actual_author'"

  # Generate Go test case format
  echo "		{"
  echo "			filename:       \"$output\","
  echo "			expectedTitle:  \"$actual_title\","
  echo "			expectedSeries: \"$actual_series\","
  echo "			expectedAuthor: \"$actual_author\","
  echo "		},"
  echo "---"
}

# Create test files with various metadata
create_epub 1 "The Book: With Colons" "Series/With/Slashes" "Author*With|Invalid" "colons and special chars"

create_epub 2 "Book & Symbols % \$ # @ !" "Series™ with ® symbols" "Author+Plus-Minus±§" "symbols"

create_epub 3 "Café au lait" "Résumé Series" "José" "accents"

create_epub 4 "This is an extremely long title that goes on and on" "The Long Series" "Hubert" "long title"

# Control characters - use printf to ensure proper handling
CONTROL_TITLE=$(printf "Book\\nWith\\tControl\\rCharacters")
CONTROL_SERIES=$(printf "Series\\0With\\bNull")
CONTROL_AUTHOR=$(printf "Author\\u001FWith\\u0007Bell")
create_epub 5 "$CONTROL_TITLE" "$CONTROL_SERIES" "$CONTROL_AUTHOR" "control chars"

create_epub 6 " Book With Many Spaces " "Series With Spaces" "Author With Spaces" "spaces"

create_epub 7 "Book \"Quoted\" Title" "Series\\With\\Backslashes" "Author With \"Quotes\"" "quotes"

create_epub 8 "" "" "" "empty"

create_epub 9 "Book: Café & Symbols!" "Ångström's Collection" "José Martínez" "mixed"

create_epub 10 "" "Series™ with ® symbols" "Author With Symbols" "no title"

create_epub 11 "Long Title: With Colons (Part 1)" "" "Hubert Blaine" "colons no series"

create_epub 12 "Book.With.Dots" "Series.With.Dots" "Author.With.Dots" "dots"

create_epub 13 " Book With Leading Spaces" " Series With Leading Spaces" " Author With Leading Spaces" "leading spaces"

create_epub 14 "Book With Trailing Spaces " "Series With Trailing Spaces " "Author With Trailing Spaces " "trailing spaces"

create_epub 15 "Book  With  Multiple  Spaces" "Series  With  Multiple  Spaces" "Author  With  Multiple  Spaces" "multiple spaces"

create_epub 16 "Book With Emoji 🔍" "Series With Emoji 🔍" "Author With Emoji 🔍" "emoji"

create_epub 17 "Book With HTML <b>Tags</b>" "Series With <i>HTML</i> Tags" "Author With <span>HTML</span> Tags" "html tags"

# Add explicit multi-author test cases
create_epub 18 "Multi-Author Book" "Collaboration Series" "John Doe & Jane Smith" "two authors"

create_epub 19 "Three Author Book" "Team Series" "Alice Johnson & Bob Wilson & Carol Davis" "three authors"

create_epub 20 "Complex Authors" "Mixed Series" "José Martínez & Dr. Sarah O'Connor & Liu Wei" "complex multi-author names"

echo "Done! Created 20 EPUB files with various metadata."
echo ""
echo "Copy the Go test cases above into your test file!"
echo ""
echo "To manually verify any file, run:"
echo "$EBOOK_META <filename>.epub"
