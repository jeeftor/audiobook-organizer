#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
TIMEOUT=${ABS_CONFIG_WAIT_TIMEOUT:-30}

section() {
	printf '\n==> %s\n' "$1"
}

detail() {
	printf '    %s\n' "$1"
}

check_service_config() {
	service=$1

	docker compose -f "$COMPOSE_FILE" run --rm --no-deps --entrypoint sh "$service" -c \
		'test -s /config/absdatabase.sqlite && test -d /config/migrations' >/dev/null 2>&1
}

wait_for_service_config() {
	service=$1
	deadline=$(( $(date +%s) + TIMEOUT ))

	detail "checking Docker bind mount for $service"
	while :; do
		if check_service_config "$service"; then
			detail "ready: $service sees /config/absdatabase.sqlite"
			return 0
		fi

		if [ "$(date +%s)" -ge "$deadline" ]; then
			printf 'Timed out waiting for %s to see restored /config/absdatabase.sqlite\n' "$service" >&2
			exit 1
		fi

		sleep 1
	done
}

section "Verifying restored config through Docker"
wait_for_service_config abs-plain
wait_for_service_config abs-metadata
