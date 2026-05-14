#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
STAGING_ROOT="${ABS_FIXTURE_CACHE_DIR:-$ROOT_DIR/staging-data}"
CLEAR_STAGING=0
EMPTY_RUNTIME=0

section() {
	printf '\n==> %s\n' "$1"
}

detail() {
	printf '    %s\n' "$1"
}

usage() {
	cat <<EOF
Usage: $0 [--clear-staging] [--empty-runtime]

Reset the local Audiobookshelf test instance.

Options:
  --clear-staging  Also delete staged/downloaded public-domain audiobook and EPUB files.
  --empty-runtime  Leave mounted audiobook and book folders empty after reset.
EOF
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

while [ "$#" -gt 0 ]; do
	case "$1" in
		--clear-staging)
			CLEAR_STAGING=1
			;;
		--empty-runtime)
			EMPTY_RUNTIME=1
			;;
		-h|--help)
			usage
			exit 0
			;;
		*)
			usage >&2
			exit 2
			;;
	esac
	shift
done

section "Stopping ABS containers"
if ! docker compose -f "$COMPOSE_FILE" down --remove-orphans; then
	printf 'Docker compose down failed; refusing to reset mounted ABS state while containers may still be running.\n' >&2
	exit 1
fi

section "Resetting ABS state directories"
mkdir -p \
	"$ROOT_DIR/state/plain/config" \
	"$ROOT_DIR/state/plain/metadata" \
	"$ROOT_DIR/state/metadata-enabled/config" \
	"$ROOT_DIR/state/metadata-enabled/metadata"
find "$ROOT_DIR/state/plain/config" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
find "$ROOT_DIR/state/plain/metadata" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
find "$ROOT_DIR/state/metadata-enabled/config" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
find "$ROOT_DIR/state/metadata-enabled/metadata" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
: > "$ROOT_DIR/state/plain/config/.gitkeep"
: > "$ROOT_DIR/state/plain/metadata/.gitkeep"
: > "$ROOT_DIR/state/metadata-enabled/config/.gitkeep"
: > "$ROOT_DIR/state/metadata-enabled/metadata/.gitkeep"
detail "plain config: $ROOT_DIR/state/plain/config"
detail "plain metadata cache: $ROOT_DIR/state/plain/metadata"
detail "metadata-enabled config: $ROOT_DIR/state/metadata-enabled/config"
detail "metadata-enabled metadata cache: $ROOT_DIR/state/metadata-enabled/metadata"

if [ "$CLEAR_STAGING" -eq 1 ]; then
	section "Clearing staged fixture cache"
	rm -rf "$STAGING_ROOT/audiobooks" "$STAGING_ROOT/books"
	mkdir -p "$STAGING_ROOT/audiobooks" "$STAGING_ROOT/books"
	: > "$STAGING_ROOT/audiobooks/.gitkeep"
	: > "$STAGING_ROOT/books/.gitkeep"
else
	section "Preserving staged fixture cache"
	detail "$STAGING_ROOT"
fi

section "Rebuilding runtime library folders"
rm -rf "$ROOT_DIR/runtime/plain" "$ROOT_DIR/runtime/metadata"
mkdir -p \
	"$ROOT_DIR/runtime/plain/audiobooks" \
	"$ROOT_DIR/runtime/plain/books" \
	"$ROOT_DIR/runtime/metadata/audiobooks" \
	"$ROOT_DIR/runtime/metadata/books"
: > "$ROOT_DIR/runtime/plain/audiobooks/.gitkeep"
: > "$ROOT_DIR/runtime/plain/books/.gitkeep"
: > "$ROOT_DIR/runtime/metadata/audiobooks/.gitkeep"
: > "$ROOT_DIR/runtime/metadata/books/.gitkeep"

if [ "$EMPTY_RUNTIME" -eq 1 ]; then
	detail "Runtime libraries left empty for ABS setup."
	section "Reset complete"
	detail "$ROOT_DIR"
	exit 0
fi

alice_audio="$STAGING_ROOT/audiobooks/Lewis Carroll/Alice's Adventures in Wonderland (Abridged)/Alice's Adventures in Wonderland (Abridged).m4b"
carol_audio="$STAGING_ROOT/audiobooks/Charles Dickens/A Christmas Carol/A Christmas Carol.m4b"
alice_book="$STAGING_ROOT/books/Lewis Carroll/Alice's Adventures in Wonderland/Alice's Adventures in Wonderland.epub"
frankenstein_book="$STAGING_ROOT/books/Mary Shelley/Frankenstein/Frankenstein.epub"
pride_book="$STAGING_ROOT/books/Jane Austen/Pride and Prejudice/Pride and Prejudice.epub"

copy_fixture "$alice_audio" "$ROOT_DIR/runtime/plain/audiobooks/unsorted-audio/drop-001/not-alice.m4b"
detail "plain audiobook fixture: unsorted-audio/drop-001/not-alice.m4b"
copy_fixture "$carol_audio" "$ROOT_DIR/runtime/plain/audiobooks/loose/holiday_story_final.m4b"
detail "plain audiobook fixture: loose/holiday_story_final.m4b"
copy_fixture "$alice_book" "$ROOT_DIR/runtime/plain/books/imported/ebook-001.epub"
detail "plain book fixture: imported/ebook-001.epub"
copy_fixture "$frankenstein_book" "$ROOT_DIR/runtime/plain/books/random/shelley-book.epub"
detail "plain book fixture: random/shelley-book.epub"
copy_fixture "$pride_book" "$ROOT_DIR/runtime/plain/books/to-sort/austen.epub"
detail "plain book fixture: to-sort/austen.epub"

copy_fixture "$alice_audio" "$ROOT_DIR/runtime/metadata/audiobooks/unsorted-audio/drop-001/not-alice.m4b"
copy_fixture "$carol_audio" "$ROOT_DIR/runtime/metadata/audiobooks/loose/holiday_story_final.m4b"
copy_fixture "$alice_book" "$ROOT_DIR/runtime/metadata/books/imported/ebook-001.epub"
copy_fixture "$frankenstein_book" "$ROOT_DIR/runtime/metadata/books/random/shelley-book.epub"
copy_fixture "$pride_book" "$ROOT_DIR/runtime/metadata/books/to-sort/austen.epub"

write_metadata \
	"$ROOT_DIR/runtime/metadata/audiobooks/unsorted-audio/drop-001" \
	"Alice's Adventures in Wonderland (Abridged)" \
	"Lewis Carroll" \
	"1865" \
	"Public-domain LibriVox abridged recording for ABS integration tests."

write_metadata \
	"$ROOT_DIR/runtime/metadata/audiobooks/loose" \
	"A Christmas Carol" \
	"Charles Dickens" \
	"1843" \
	"Public-domain LibriVox recording for ABS integration tests."

write_metadata \
	"$ROOT_DIR/runtime/metadata/books/imported" \
	"Alice's Adventures in Wonderland" \
	"Lewis Carroll" \
	"1865" \
	"Public-domain Project Gutenberg EPUB for ABS integration tests."

write_metadata \
	"$ROOT_DIR/runtime/metadata/books/random" \
	"Frankenstein; or, The Modern Prometheus" \
	"Mary Wollstonecraft Shelley" \
	"1818" \
	"Public-domain Project Gutenberg EPUB for ABS integration tests."

write_metadata \
	"$ROOT_DIR/runtime/metadata/books/to-sort" \
	"Pride and Prejudice" \
	"Jane Austen" \
	"1813" \
	"Public-domain Project Gutenberg EPUB for ABS integration tests."

detail "metadata runtime mirrors plain fixtures and includes metadata.json sidecars."
section "Reset complete"
detail "$ROOT_DIR"
