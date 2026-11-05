#!/bin/sh

set -eu

image_name="audiobook-organizer-web-session-test"
container_name="audiobook-organizer-web-session-test"
port="1338"

cleanup() {
	container_id=$(docker ps -aq --filter "name=^/${container_name}$")
	if [ -n "$container_id" ]; then
		docker rm -f "$container_id" >/dev/null
	fi
}

trap cleanup EXIT

docker build -t "$image_name" .
docker run -d --name "$container_name" -p "${port}:${port}" "$image_name" \
	web --host=0.0.0.0 --port="$port" --no-open >/dev/null

token=""
attempt=0
while [ "$attempt" -lt 30 ]; do
	token=$(docker logs "$container_name" 2>&1 | sed -n 's/.*[?]token=\([a-f0-9]*\).*/\1/p' | head -n 1)
	if [ -n "$token" ]; then
		break
	fi
	attempt=$((attempt + 1))
	sleep 1
done

if [ -z "$token" ]; then
	docker logs "$container_name"
	echo "web session token was not printed by the container" >&2
	exit 1
fi

url="http://127.0.0.1:${port}/api/config/initial"
header="X-Audiobook-Organizer-Token: ${token}"
curl --fail --silent --show-error -H "$header" "$url" >/dev/null

sleep 301

curl --fail --silent --show-error -H "$header" "$url" >/dev/null
restart_count=$(docker inspect --format '{{.RestartCount}}' "$container_name")
if [ "$restart_count" != "0" ]; then
	echo "web container restarted ${restart_count} time(s) during the session check" >&2
	exit 1
fi
