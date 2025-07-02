#!/bin/bash

# Original EPUB file
ORIGINAL="title-author.epub"

# Check if original file exists
if [ ! -f "$ORIGINAL" ]; then
  echo "Error: $ORIGINAL does not exist."
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

# Array of metadata variations
declare -a METADATA_VARIATIONS=(
  # Characters invalid in filenames (sanitized for filename only)
  "--title 'The Book: With Colons?' --series 'Series/With/Slashes' --authors 'Author*With|Invalid'"
  "--title 'Book & Symbols % $ # @ !' --series 'Series™ ©2025' --authors 'Author+Plus-Minus±§'"
  # Non-ASCII characters
  "--title 'Café au lait' --series 'Résumé of Ångström' --authors 'José Martínez'"
  # Very long names
  "--title 'This is an extremely long title that goes on and on and might cause issues with path length limits in some operating systems' --series 'The Never-ending Series of Books That Keep Getting Published Year After Year After Year' --authors 'Hubert Blaine Wolfeschlegelsteinhausenbergerdorff Sr.'"
  # Control characters (some may be stripped by ebook-meta)
  "--title $'Book\nWith\tControl\rCharacters' --series $'Series\0With\bNull' --authors $'Author\u001FWith\u0007Bell'"
  # Spaces and tabs
  "--title ' Book With Many Spaces ' --series 'Series With Tabs' --authors 'Author With Trailing Spaces '"
  # Quotes and backslashes
  "--title 'Book \"Quoted\" Title' --series 'Series\\With\\Backslashes' --authors 'Author \'Single\' and \"Double\" Quotes'"
  # Empty fields
  "--title '' --series '' --authors ''"
  # Mixed variations
  "--title 'Book: Café & Symbols!' --series 'Ångström\'s Series' --authors 'José * Martínez' --index 2.5 --rating 3 --isbn '1234567890'"
  "--title '' --series 'Series™ With No Title' --authors 'Author With Spaces ' --publisher 'Strange Books Inc.' --language 'fr'"
  "--title 'Long Title With Colons: Part 1' --series '' --authors 'Hubert Blaine & José Martínez' --comments 'This is a weird book!'"
  # Additional problematic cases
  "--title 'Book.With.Dots' --series 'Series.With.Dots' --authors 'Author.With.Dots'"
  "--title ' Book With Leading Spaces' --series ' Series With Leading Spaces' --authors ' Author With Leading Spaces'"
  "--title 'Book With Trailing Spaces ' --series 'Series With Trailing Spaces ' --authors 'Author With Trailing Spaces '"
  "--title 'Book With Multiple  Spaces' --series 'Series  With  Multiple   Spaces' --authors 'Author With   Multiple    Spaces'"
  "--title 'Book With Emoji 📚🔍' --series 'Series With Emoji 📚📖' --authors 'Author With Emoji 👨‍💻'"
  "--title 'Book With HTML <b>Tags</b>' --series 'Series With <i>HTML</i>' --authors 'Author With <script>alert(\"XSS\")</script>'"
)

# Counter for naming files
COUNTER=1

# Loop through metadata variations
for META in "${METADATA_VARIATIONS[@]}"; do
  # Generate a unique filename
  TITLE=$(echo "$META" | grep -o -- "--title '[^']*'" | sed "s/--title '//;s/'//" | head -c 50)
  if [ -z "$TITLE" ]; then
    TITLE="book_$COUNTER"
  fi
  SANITIZED_TITLE=$(sanitize_filename "$TITLE")
  OUTPUT="strange_book_${COUNTER}_${SANITIZED_TITLE}.epub"

  # Copy the original file
  cp "$ORIGINAL" "$OUTPUT"

  # Apply metadata
  echo "Creating $OUTPUT with metadata: $META"
  $EBOOK_META "$OUTPUT" $META

  # Increment counter
  ((COUNTER++))
done

echo "Done! Created $((COUNTER-1)) EPUB files with strange metadata."
