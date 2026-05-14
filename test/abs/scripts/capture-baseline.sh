#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"

copy_config() {
	name=$1
	source_dir=$2
	target_dir=$3

	if [ ! -f "$source_dir/absdatabase.sqlite" ]; then
		printf 'Skipping %s baseline: %s/absdatabase.sqlite not found\n' "$name" "$source_dir" >&2
		return
	fi

	rm -rf "$target_dir"
	mkdir -p "$target_dir"
	cp -R "$source_dir/." "$target_dir/"
	find "$target_dir" -name .DS_Store -type f -delete
	: > "$target_dir/.gitkeep"
	printf 'Captured %s baseline config: %s\n' "$name" "$target_dir"
}

if ! docker compose -f "$COMPOSE_FILE" down --remove-orphans; then
	printf 'Warning: Docker compose down failed; copying current local config anyway.\n' >&2
fi

copy_config "plain" \
	"$ROOT_DIR/state/plain/config" \
	"$ROOT_DIR/baseline-config/plain/config"

copy_config "metadata-enabled" \
	"$ROOT_DIR/state/metadata-enabled/config" \
	"$ROOT_DIR/baseline-config/metadata-enabled/config"

cat > "$ROOT_DIR/.env.local" <<EOF
ABS_PLAIN_URL=http://localhost:13378
ABS_METADATA_URL=http://localhost:13379
ABS_PLAIN_TOKEN=${ABS_PLAIN_TOKEN:-}
ABS_METADATA_TOKEN=${ABS_METADATA_TOKEN:-}
ABS_PLAIN_SQLITE=test/abs/state/plain/config/absdatabase.sqlite
ABS_METADATA_SQLITE=test/abs/state/metadata-enabled/config/absdatabase.sqlite
EOF

printf 'Wrote ignored local env file: %s\n' "$ROOT_DIR/.env.local"
printf 'Set ABS_PLAIN_TOKEN and ABS_METADATA_TOKEN in that file as needed.\n'
