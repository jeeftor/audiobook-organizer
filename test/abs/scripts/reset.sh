#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
STAGING_ROOT="${ABS_FIXTURE_CACHE_DIR:-$ROOT_DIR/staging-data}"
REPO_ROOT=$(CDPATH='' cd -- "$ROOT_DIR/../.." && pwd)
CLEAR_STAGING=0
EMPTY_RUNTIME=0

section() {
	printf '\n==> %s\n' "$1"
}

detail() {
	printf '    %s\n' "$1"
}

remove_with_sudo_fallback() {
	if "$@"; then
		return 0
	fi

	if command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
		detail "normal cleanup hit protected Docker-owned files; retrying with sudo"
		sudo "$@"
		return $?
	fi

	printf 'Cleanup failed and passwordless sudo is unavailable. Stop ABS and fix ownership under %s.\n' "$ROOT_DIR" >&2
	exit 1
}

remove_children() {
	target_dir=$1

	remove_with_sudo_fallback find "$target_dir" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
}

remove_path() {
	remove_with_sudo_fallback rm -rf "$@"
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
remove_children "$ROOT_DIR/state/plain/config"
remove_children "$ROOT_DIR/state/plain/metadata"
remove_children "$ROOT_DIR/state/metadata-enabled/config"
remove_children "$ROOT_DIR/state/metadata-enabled/metadata"
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
	remove_path "$STAGING_ROOT/audiobooks" "$STAGING_ROOT/books"
	mkdir -p "$STAGING_ROOT/audiobooks" "$STAGING_ROOT/books"
	: > "$STAGING_ROOT/audiobooks/.gitkeep"
	: > "$STAGING_ROOT/books/.gitkeep"
else
	section "Preserving staged fixture cache"
	detail "$STAGING_ROOT"
fi

section "Rebuilding runtime library folders"
remove_path "$ROOT_DIR/runtime/plain" "$ROOT_DIR/runtime/metadata"
mkdir -p \
	"$ROOT_DIR/runtime/plain/audiobooks" \
	"$ROOT_DIR/runtime/plain/books" \
	"$ROOT_DIR/runtime/metadata/audiobooks" \
	"$ROOT_DIR/runtime/metadata/books" \
	"$ROOT_DIR/runtime/import-input/audiobooks" \
	"$ROOT_DIR/runtime/import-input/books" \
	"$ROOT_DIR/runtime/flat-input/audiobooks" \
	"$ROOT_DIR/runtime/flat-input/books"
: > "$ROOT_DIR/runtime/plain/audiobooks/.gitkeep"
: > "$ROOT_DIR/runtime/plain/books/.gitkeep"
: > "$ROOT_DIR/runtime/metadata/audiobooks/.gitkeep"
: > "$ROOT_DIR/runtime/metadata/books/.gitkeep"
: > "$ROOT_DIR/runtime/import-input/audiobooks/.gitkeep"
: > "$ROOT_DIR/runtime/import-input/books/.gitkeep"
: > "$ROOT_DIR/runtime/flat-input/audiobooks/.gitkeep"
: > "$ROOT_DIR/runtime/flat-input/books/.gitkeep"

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

copy_fixture \
	"$REPO_ROOT/testdata/m4b/strange_audiobook_5_Mystery_Series_Mystery_of_the_Lost_City_Jane_Doe.m4b" \
	"$ROOT_DIR/runtime/import-input/audiobooks/dropbox/jane-doe-mess/source.m4b"
detail "embedded import fixture: dropbox/jane-doe-mess/source.m4b"
copy_fixture \
	"$REPO_ROOT/testdata/m4b/strange_audiobook_20_Saga_of_Endless_Horizons_The_Epic_Tale_That_Spans_Generations_Alexander_von_Longname.m4b" \
	"$ROOT_DIR/runtime/import-input/audiobooks/dropbox/longname-mess/source.m4b"
detail "embedded import fixture: dropbox/longname-mess/source.m4b"
copy_fixture \
	"$REPO_ROOT/testdata/epub/title-author.epub" \
	"$ROOT_DIR/runtime/import-input/books/dropbox/cool-stuff/source.epub"
detail "embedded ebook import fixture: dropbox/cool-stuff/source.epub"
copy_fixture \
	"$REPO_ROOT/testdata/epub/title-author-series1.epub" \
	"$ROOT_DIR/runtime/import-input/books/dropbox/testing-knowledge/source.epub"
detail "embedded ebook import fixture: dropbox/testing-knowledge/source.epub"

copy_fixture \
	"$REPO_ROOT/testdata/mp3flat/charlesdexterward_01_lovecraft_64kb.mp3" \
	"$ROOT_DIR/runtime/flat-input/audiobooks/inbox/charlesdexterward_01_lovecraft_64kb.mp3"
detail "flat import fixture: inbox/charlesdexterward_01_lovecraft_64kb.mp3"
copy_fixture \
	"$REPO_ROOT/testdata/mp3flat/falstaffswedding1766version_1_kenrick_64kb.mp3" \
	"$ROOT_DIR/runtime/flat-input/audiobooks/inbox/falstaffswedding1766version_1_kenrick_64kb.mp3"
detail "flat import fixture: inbox/falstaffswedding1766version_1_kenrick_64kb.mp3"
copy_fixture \
	"$REPO_ROOT/testdata/mp3flat/perouse_01_scott_64kb.mp3" \
	"$ROOT_DIR/runtime/flat-input/audiobooks/inbox/perouse_01_scott_64kb.mp3"
detail "flat import fixture: inbox/perouse_01_scott_64kb.mp3"

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
