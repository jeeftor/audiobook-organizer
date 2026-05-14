#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
ENV_FILE=${ABS_ENV_FILE:-"$ROOT_DIR/.env.local"}

section() {
	printf '\n==> %s\n' "$1" >&2
}

detail() {
	printf '    %s\n' "$1" >&2
}

require_tool() {
	if ! command -v "$1" >/dev/null 2>&1; then
		printf '%s is required for ABS API configuration.\n' "$1" >&2
		exit 1
	fi
}

token_status() {
	token=$1
	if [ -n "$token" ]; then
		printf 'set (%s chars)' "$(printf '%s' "$token" | wc -c | tr -d ' ')"
	else
		printf 'missing'
	fi
}

request_json() {
	method=$1
	name=$2
	url=$3
	path=$4
	token=$5
	body=$6
	output_file=$7

	if [ -n "$token" ]; then
		status=$(
			curl -s \
				-X "$method" \
				-o "$output_file" \
				-w '%{http_code}' \
				-H "Authorization: Bearer $token" \
				-H "Content-Type: application/json" \
				-d "$body" \
				"$url$path" || true
		)
	else
		status=$(
			curl -s \
				-X "$method" \
				-o "$output_file" \
				-w '%{http_code}' \
				-H "Content-Type: application/json" \
				-d "$body" \
				"$url$path" || true
		)
	fi

	case "$status" in
		200)
			return 0
			;;
		000)
			printf '\nCould not connect to %s at %s%s.\n' "$name" "$url" "$path" >&2
			return 1
			;;
		*)
			printf '\nABS API request failed for %s: HTTP %s %s%s\n' "$name" "$status" "$url" "$path" >&2
			if [ -s "$output_file" ]; then
				printf 'Response: ' >&2
				cat "$output_file" >&2
				printf '\n' >&2
			fi
			return 1
			;;
	esac
}

get_json() {
	name=$1
	url=$2
	path=$3
	token=$4
	output_file=$5

	if [ -n "$token" ]; then
		status=$(
			curl -s \
				-o "$output_file" \
				-w '%{http_code}' \
				-H "Authorization: Bearer $token" \
				"$url$path" || true
		)
	else
		status=$(
			curl -s \
				-o "$output_file" \
				-w '%{http_code}' \
				"$url$path" || true
		)
	fi

	case "$status" in
		200)
			return 0
			;;
		000)
			printf '\nCould not connect to %s at %s%s.\n' "$name" "$url" "$path" >&2
			return 1
			;;
		*)
			printf '\nABS API request failed for %s: HTTP %s %s%s\n' "$name" "$status" "$url" "$path" >&2
			if [ -s "$output_file" ]; then
				printf 'Response: ' >&2
				cat "$output_file" >&2
				printf '\n' >&2
			fi
			return 1
			;;
	esac
}

ensure_initialized() {
	name=$1
	url=$2
	username=$3
	password=$4
	output_file=$5

	get_json "$name" "$url" "/status" "" "$output_file"
	if [ "$(jq -r '.isInit // false' "$output_file")" = "true" ]; then
		detail "$name root user already exists"
		return
	fi

	body=$(jq -n \
		--arg username "$username" \
		--arg password "$password" \
		'{newRoot: {username: $username, password: $password}}')
	request_json POST "$name" "$url" "/init" "" "$body" "$output_file"
	detail "$name root user initialized"
}

login() {
	name=$1
	url=$2
	username=$3
	password=$4
	output_file=$5

	body=$(jq -n \
		--arg username "$username" \
		--arg password "$password" \
		'{username: $username, password: $password}')
	request_json POST "$name" "$url" "/login" "" "$body" "$output_file"
	jq -r '.user.token // empty' "$output_file"
}

apply_settings() {
	name=$1
	url=$2
	token=$3
	store_metadata=$4
	output_file=$5

	body=$(jq -n \
		--argjson storeMetadataWithItem "$store_metadata" \
		'{storeMetadataWithItem: $storeMetadataWithItem, scannerDisableWatcher: true}')
	request_json PATCH "$name" "$url" "/api/settings" "$token" "$body" "$output_file"
	detail "$name storeMetadataWithItem=$store_metadata"
}

ensure_library() {
	name=$1
	url=$2
	token=$3
	library_name=$4
	folder_path=$5
	icon=$6
	provider=$7
	output_file=$8

	get_json "$name" "$url" "/api/libraries" "$token" "$output_file"
	if jq -e \
		--arg library_name "$library_name" \
		--arg folder_path "$folder_path" \
		'.libraries[]? | select(.name == $library_name and any(.folders[]?; .fullPath == $folder_path))' \
		"$output_file" >/dev/null; then
		detail "$name library exists: $library_name -> $folder_path"
		return
	fi

	body=$(jq -n \
		--arg name "$library_name" \
		--arg fullPath "$folder_path" \
		--arg icon "$icon" \
		--arg provider "$provider" \
		'{
			name: $name,
			folders: [{fullPath: $fullPath}],
			icon: $icon,
			mediaType: "book",
			provider: $provider,
			settings: {
				coverAspectRatio: 1,
				disableWatcher: true,
				skipMatchingMediaWithAsin: false,
				skipMatchingMediaWithIsbn: false,
				autoScanCronExpression: null
			}
		}')
	request_json POST "$name" "$url" "/api/libraries" "$token" "$body" "$output_file"
	detail "$name library created: $library_name -> $folder_path"
}

configure_instance() {
	name=$1
	url=$2
	store_metadata=$3
	output_file=$4

	section "Configuring $name ABS instance"
	detail "url: $url"

	ensure_initialized "$name" "$url" "$ABS_ROOT_USERNAME" "$ABS_ROOT_PASSWORD" "$output_file"
	token=$(login "$name" "$url" "$ABS_ROOT_USERNAME" "$ABS_ROOT_PASSWORD" "$output_file")
	if [ -z "$token" ]; then
		printf 'Login succeeded for %s but no token was returned.\n' "$name" >&2
		exit 1
	fi
	detail "login token: $(token_status "$token")"

	apply_settings "$name" "$url" "$token" "$store_metadata" "$output_file"
	ensure_library "$name" "$url" "$token" "Audiobooks" "/audiobooks" "audiobookshelf" "audible" "$output_file"
	ensure_library "$name" "$url" "$token" "Ebooks" "/books" "book" "google" "$output_file"

	printf '%s' "$token"
}

require_tool curl
require_tool jq

if [ -f "$ENV_FILE" ]; then
	set -a
	# shellcheck disable=SC1090
	. "$ENV_FILE"
	set +a
fi

ABS_PLAIN_URL=${ABS_PLAIN_URL:-http://localhost:13378}
ABS_METADATA_URL=${ABS_METADATA_URL:-http://localhost:13379}
ABS_ROOT_USERNAME=${ABS_ROOT_USERNAME:-root}
ABS_ROOT_PASSWORD=${ABS_ROOT_PASSWORD:-password}
ABS_PLAIN_SQLITE=${ABS_PLAIN_SQLITE:-test/abs/state/plain/config/absdatabase.sqlite}
ABS_METADATA_SQLITE=${ABS_METADATA_SQLITE:-test/abs/state/metadata-enabled/config/absdatabase.sqlite}

tmp_response=$(mktemp)
trap 'rm -f "$tmp_response"' EXIT

section "ABS API configuration"
detail "env file: $ENV_FILE"
detail "root username: $ABS_ROOT_USERNAME"
detail "plain url: $ABS_PLAIN_URL"
detail "metadata url: $ABS_METADATA_URL"

plain_token=$(configure_instance "plain" "$ABS_PLAIN_URL" false "$tmp_response")
metadata_token=$(configure_instance "metadata-enabled" "$ABS_METADATA_URL" true "$tmp_response")

mkdir -p "$(dirname -- "$ENV_FILE")"
umask 077
cat > "$ENV_FILE" <<EOF
ABS_PLAIN_URL=$ABS_PLAIN_URL
ABS_METADATA_URL=$ABS_METADATA_URL
ABS_PLAIN_TOKEN=$plain_token
ABS_METADATA_TOKEN=$metadata_token
ABS_PLAIN_SQLITE=$ABS_PLAIN_SQLITE
ABS_METADATA_SQLITE=$ABS_METADATA_SQLITE
ABS_ROOT_USERNAME=$ABS_ROOT_USERNAME
ABS_ROOT_PASSWORD=$ABS_ROOT_PASSWORD
EOF

section "ABS API configuration complete"
detail "wrote tokens and paths to $ENV_FILE"
