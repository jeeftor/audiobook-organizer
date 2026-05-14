#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
STAGING_AUDIOBOOK_DIR="$ROOT_DIR/staging-data/audiobooks"
STAGING_BOOK_DIR="$ROOT_DIR/staging-data/books"
PLAIN_AUDIOBOOK_DIR="$ROOT_DIR/runtime/plain/audiobooks"
PLAIN_BOOK_DIR="$ROOT_DIR/runtime/plain/books"
METADATA_AUDIOBOOK_DIR="$ROOT_DIR/runtime/metadata/audiobooks"
METADATA_BOOK_DIR="$ROOT_DIR/runtime/metadata/books"

download() {
	url=$1
	target=$2
	legacy_target=$3

	if [ -f "$target" ]; then
		printf 'exists: %s\n' "$target"
		return
	fi

	mkdir -p "$(dirname -- "$target")"

	if [ -f "$legacy_target" ]; then
		printf 'stage existing: %s\n' "$target"
		cp "$legacy_target" "$target"
		return
	fi

	printf 'download: %s\n' "$target"
	curl -fL --retry 3 --connect-timeout 20 -o "$target.tmp" "$url"
	mv "$target.tmp" "$target"
}

write_metadata() {
	target_dir=$1
	title=$2
	author=$3
	year=$4
	description=$5

	mkdir -p "$target_dir"
	cat > "$target_dir/metadata.json" <<EOF
{
  "title": "$title",
  "authors": ["$author"],
  "publishedYear": "$year",
  "description": "$description"
}
EOF
}

copy_fixture() {
	source_file=$1
	target_file=$2

	if [ ! -f "$source_file" ]; then
		printf 'Missing staged fixture: %s\n' "$source_file" >&2
		exit 1
	fi

	mkdir -p "$(dirname -- "$target_file")"
	cp "$source_file" "$target_file"
}

refresh_library() {
	rm -rf "$PLAIN_AUDIOBOOK_DIR" "$PLAIN_BOOK_DIR" "$METADATA_AUDIOBOOK_DIR" "$METADATA_BOOK_DIR"
	mkdir -p "$PLAIN_AUDIOBOOK_DIR" "$PLAIN_BOOK_DIR" "$METADATA_AUDIOBOOK_DIR" "$METADATA_BOOK_DIR"
	: > "$PLAIN_AUDIOBOOK_DIR/.gitkeep"
	: > "$PLAIN_BOOK_DIR/.gitkeep"
	: > "$METADATA_AUDIOBOOK_DIR/.gitkeep"
	: > "$METADATA_BOOK_DIR/.gitkeep"

	copy_fixture \
		"$alice_audio/Alice's Adventures in Wonderland (Abridged).m4b" \
		"$PLAIN_AUDIOBOOK_DIR/unsorted-audio/drop-001/not-alice.m4b"
	copy_fixture \
		"$carol_audio/A Christmas Carol.m4b" \
		"$PLAIN_AUDIOBOOK_DIR/loose/holiday_story_final.m4b"
	copy_fixture \
		"$alice_book/Alice's Adventures in Wonderland.epub" \
		"$PLAIN_BOOK_DIR/imported/ebook-001.epub"
	copy_fixture \
		"$frankenstein_book/Frankenstein.epub" \
		"$PLAIN_BOOK_DIR/random/shelley-book.epub"
	copy_fixture \
		"$pride_book/Pride and Prejudice.epub" \
		"$PLAIN_BOOK_DIR/to-sort/austen.epub"

	copy_fixture \
		"$alice_audio/Alice's Adventures in Wonderland (Abridged).m4b" \
		"$METADATA_AUDIOBOOK_DIR/unsorted-audio/drop-001/not-alice.m4b"
	copy_fixture \
		"$carol_audio/A Christmas Carol.m4b" \
		"$METADATA_AUDIOBOOK_DIR/loose/holiday_story_final.m4b"
	copy_fixture \
		"$alice_book/Alice's Adventures in Wonderland.epub" \
		"$METADATA_BOOK_DIR/imported/ebook-001.epub"
	copy_fixture \
		"$frankenstein_book/Frankenstein.epub" \
		"$METADATA_BOOK_DIR/random/shelley-book.epub"
	copy_fixture \
		"$pride_book/Pride and Prejudice.epub" \
		"$METADATA_BOOK_DIR/to-sort/austen.epub"
}

alice_audio="$STAGING_AUDIOBOOK_DIR/Lewis Carroll/Alice's Adventures in Wonderland (Abridged)"
carol_audio="$STAGING_AUDIOBOOK_DIR/Charles Dickens/A Christmas Carol"
alice_book="$STAGING_BOOK_DIR/Lewis Carroll/Alice's Adventures in Wonderland"
frankenstein_book="$STAGING_BOOK_DIR/Mary Shelley/Frankenstein"
pride_book="$STAGING_BOOK_DIR/Jane Austen/Pride and Prejudice"

legacy_alice_audio="$ROOT_DIR/library/audiobooks/Lewis Carroll/Alice's Adventures in Wonderland (Abridged)"
legacy_carol_audio="$ROOT_DIR/library/audiobooks/Charles Dickens/A Christmas Carol"
legacy_alice_book="$ROOT_DIR/library/books/Lewis Carroll/Alice's Adventures in Wonderland"
legacy_frankenstein_book="$ROOT_DIR/library/books/Mary Shelley/Frankenstein"
legacy_pride_book="$ROOT_DIR/library/books/Jane Austen/Pride and Prejudice"

download \
	"https://archive.org/download/alices_adventures/AlicesAdventuresInWonderlandV2_librivox.m4b" \
	"$alice_audio/Alice's Adventures in Wonderland (Abridged).m4b" \
	"$legacy_alice_audio/Alice's Adventures in Wonderland (Abridged).m4b"

download \
	"https://archive.org/download/A_Christmas_Carol/ChristmasCarol_librivox.m4b" \
	"$carol_audio/A Christmas Carol.m4b" \
	"$legacy_carol_audio/A Christmas Carol.m4b"

download \
	"https://www.gutenberg.org/ebooks/11.epub.images" \
	"$alice_book/Alice's Adventures in Wonderland.epub" \
	"$legacy_alice_book/Alice's Adventures in Wonderland.epub"

download \
	"https://www.gutenberg.org/ebooks/84.epub.images" \
	"$frankenstein_book/Frankenstein.epub" \
	"$legacy_frankenstein_book/Frankenstein.epub"

download \
	"https://www.gutenberg.org/ebooks/1342.epub.noimages" \
	"$pride_book/Pride and Prejudice.epub" \
	"$legacy_pride_book/Pride and Prejudice.epub"

find "$STAGING_AUDIOBOOK_DIR" "$STAGING_BOOK_DIR" -name metadata.json -type f -delete

refresh_library

write_metadata \
	"$METADATA_AUDIOBOOK_DIR/unsorted-audio/drop-001" \
	"Alice's Adventures in Wonderland (Abridged)" \
	"Lewis Carroll" \
	"1865" \
	"Public-domain LibriVox abridged recording for ABS integration tests."

write_metadata \
	"$METADATA_AUDIOBOOK_DIR/loose" \
	"A Christmas Carol" \
	"Charles Dickens" \
	"1843" \
	"Public-domain LibriVox recording for ABS integration tests."

write_metadata \
	"$METADATA_BOOK_DIR/imported" \
	"Alice's Adventures in Wonderland" \
	"Lewis Carroll" \
	"1865" \
	"Public-domain Project Gutenberg EPUB for ABS integration tests."

write_metadata \
	"$METADATA_BOOK_DIR/random" \
	"Frankenstein; or, The Modern Prometheus" \
	"Mary Wollstonecraft Shelley" \
	"1818" \
	"Public-domain Project Gutenberg EPUB for ABS integration tests."

write_metadata \
	"$METADATA_BOOK_DIR/to-sort" \
	"Pride and Prejudice" \
	"Jane Austen" \
	"1813" \
	"Public-domain Project Gutenberg EPUB for ABS integration tests."

printf '\nSeed data is staged under %s\n' "$ROOT_DIR/staging-data"
printf 'Plain runtime library refreshed under %s\n' "$ROOT_DIR/runtime/plain"
printf 'Metadata runtime library refreshed under %s\n' "$ROOT_DIR/runtime/metadata"
