#!/bin/sh
set -eu

TIMEOUT=${ABS_WAIT_TIMEOUT:-120}

section() {
	printf '\n==> %s\n' "$1"
}

wait_for_url() {
	url=$1
	elapsed=0

	printf '    waiting for %s (timeout %ss)\n' "$url" "$TIMEOUT"
	while [ "$elapsed" -lt "$TIMEOUT" ]; do
		if curl -fsS "$url/ping" >/dev/null 2>&1; then
			printf '    ready: %s\n' "$url"
			return 0
		fi

		sleep 2
		elapsed=$((elapsed + 2))
	done

	printf 'Timed out waiting for Audiobookshelf at %s\n' "$url" >&2
	return 1
}

if [ "${ABS_TEST_URL:-}" ]; then
	section "Waiting for Audiobookshelf"
	wait_for_url "$ABS_TEST_URL"
else
	section "Waiting for ABS services"
	wait_for_url "http://localhost:${ABS_PLAIN_PORT:-13378}"
	wait_for_url "http://localhost:${ABS_METADATA_PORT:-13379}"
fi
