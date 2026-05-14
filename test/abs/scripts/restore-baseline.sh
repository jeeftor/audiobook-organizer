#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)

section() {
	printf '\n==> %s\n' "$1"
}

restore_config() {
	name=$1
	source_dir=$2
	target_dir=$3

	if [ ! -f "$source_dir/absdatabase.sqlite" ]; then
		printf 'Missing %s baseline config: %s/absdatabase.sqlite\n' "$name" "$source_dir" >&2
		exit 1
	fi

	mkdir -p "$target_dir"
	find "$target_dir" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
	cp -R "$source_dir/." "$target_dir/"
	find "$target_dir" -name .DS_Store -type f -delete
	: > "$target_dir/.gitkeep"
	printf '    restored %s config: %s\n' "$name" "$target_dir"
}

section "Restoring baseline ABS config"
restore_config "plain" \
	"$ROOT_DIR/baseline-config/plain/config" \
	"$ROOT_DIR/state/plain/config"

restore_config "metadata-enabled" \
	"$ROOT_DIR/baseline-config/metadata-enabled/config" \
	"$ROOT_DIR/state/metadata-enabled/config"

mkdir -p "$ROOT_DIR/state/plain/metadata" "$ROOT_DIR/state/metadata-enabled/metadata"
: > "$ROOT_DIR/state/plain/metadata/.gitkeep"
: > "$ROOT_DIR/state/metadata-enabled/metadata/.gitkeep"
sync

section "Baseline restore complete"
