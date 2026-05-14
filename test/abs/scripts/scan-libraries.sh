#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
ENV_FILE=${ABS_ENV_FILE:-"$ROOT_DIR/.env.testing"}

section() {
	printf '\n==> %s\n' "$1"
}

detail() {
	printf '    %s\n' "$1"
}

if [ -f "$ENV_FILE" ]; then
	set -a
	# shellcheck disable=SC1090
	. "$ENV_FILE"
	set +a
fi

token_status() {
	token=$1
	if [ -n "$token" ]; then
		printf 'set (%s chars)' "$(printf '%s' "$token" | wc -c | tr -d ' ')"
	else
		printf 'missing'
	fi
}

api_get() {
	name=$1
	url=$2
	token=$3
	path=$4
	output_file=$5

	status=$(
		curl -s \
			-o "$output_file" \
			-w '%{http_code}' \
			-H "Authorization: Bearer $token" \
			"$url$path" || true
	)

	case "$status" in
		200)
			return 0
			;;
		000)
			printf '\nCould not connect to %s at %s%s.\n' "$name" "$url" "$path" >&2
			printf 'Run make abs-dev-wait or make abs-dev-reset-scan to start the ABS services first.\n' >&2
			return 1
			;;
		401)
			printf '\nAuthentication failed for %s at %s%s.\n' "$name" "$url" "$path" >&2
			printf 'Set the correct token in test/abs/.env.testing or override ABS_ENV_FILE.\n' >&2
			printf 'If both DBs have the same API key row but one token still fails, the JWT may be signed by a different ABS instance secret; create/copy the token from that instance.\n' >&2
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

api_post() {
	name=$1
	url=$2
	token=$3
	path=$4
	output_file=$5

	status=$(
		curl -s \
			-X POST \
			-o "$output_file" \
			-w '%{http_code}' \
			-H "Authorization: Bearer $token" \
			"$url$path" || true
	)

	case "$status" in
		200)
			return 0
			;;
		000)
			printf '\nCould not connect to %s at %s%s.\n' "$name" "$url" "$path" >&2
			printf 'Run make abs-dev-wait or make abs-dev-reset-scan to start the ABS services first.\n' >&2
			return 1
			;;
		401)
			printf '\nAuthentication failed for %s at %s%s.\n' "$name" "$url" "$path" >&2
			printf 'Set the correct token in test/abs/.env.testing or override ABS_ENV_FILE.\n' >&2
			printf 'Current token status: plain=%s metadata=%s\n' \
				"$(token_status "${ABS_PLAIN_TOKEN:-}")" \
				"$(token_status "${ABS_METADATA_TOKEN:-}")" >&2
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

api_login() {
	name=$1
	url=$2
	output_file=$3

	body=$(jq -n \
		--arg username "$ABS_ROOT_USERNAME" \
		--arg password "$ABS_ROOT_PASSWORD" \
		'{username: $username, password: $password}')

	status=$(
		curl -s \
			-X POST \
			-o "$output_file" \
			-w '%{http_code}' \
			-H "Content-Type: application/json" \
			-d "$body" \
			"$url/login" || true
	)

	case "$status" in
		200)
			jq -r '.user.token // empty' "$output_file"
			return 0
			;;
		000)
			printf '\nCould not connect to %s at %s/login.\n' "$name" "$url" >&2
			printf 'Run make abs-dev-wait or make abs-dev-reset-scan to start the ABS services first.\n' >&2
			return 1
			;;
		401)
			printf '\nLogin failed for %s at %s/login.\n' "$name" "$url" >&2
			printf 'Set ABS_%s_TOKEN or ABS_ROOT_USERNAME/ABS_ROOT_PASSWORD in %s.\n' \
				"$(printf '%s' "$name" | tr '[:lower:]-' '[:upper:]_')" "$ENV_FILE" >&2
			return 1
			;;
		*)
			printf '\nABS login failed for %s: HTTP %s %s/login\n' "$name" "$status" "$url" >&2
			if [ -s "$output_file" ]; then
				printf 'Response: ' >&2
				cat "$output_file" >&2
				printf '\n' >&2
			fi
			return 1
			;;
	esac
}

sqlite_token() {
	sqlite_path=$1
	username=$2
	escaped_username=$(printf "%s" "$username" | sed "s/'/''/g")
	sqlite3 "$sqlite_path" \
		"select token from users where username = '$escaped_username' and isActive = 1 limit 1;" 2>/dev/null || true
}

base64url() {
	openssl base64 -A | tr '+/' '-_' | tr -d '='
}

api_key_token() {
	sqlite_path=$1
	api_key_row=$(sqlite3 -separator '	' "$sqlite_path" \
		"select id, name, strftime('%s', createdAt) from apiKeys where isActive = 1 order by createdAt limit 1;" 2>/dev/null || true)
	if [ -z "$api_key_row" ]; then
		return 0
	fi

	key_id=$(printf '%s' "$api_key_row" | cut -f1)
	key_name=$(printf '%s' "$api_key_row" | cut -f2)
	key_iat=$(printf '%s' "$api_key_row" | cut -f3)
	token_secret=$(sqlite3 "$sqlite_path" \
		"select json_extract(value, '$.tokenSecret') from settings where key = 'server-settings';" 2>/dev/null || true)
	if [ -z "$key_id" ] || [ -z "$key_name" ] || [ -z "$key_iat" ] || [ -z "$token_secret" ]; then
		return 0
	fi

	header=$(printf '{"alg":"HS256","typ":"JWT"}' | base64url)
	payload_json=$(jq -cn \
		--arg keyId "$key_id" \
		--arg name "$key_name" \
		--argjson iat "$key_iat" \
		'{keyId: $keyId, name: $name, type: "api", iat: $iat}')
	payload=$(printf '%s' "$payload_json" | base64url)
	signing_input="$header.$payload"
	signature=$(printf '%s' "$signing_input" | openssl dgst -sha256 -hmac "$token_secret" -binary | base64url)

	printf '%s.%s\n' "$signing_input" "$signature"
}

scan_instance() {
	name=$1
	url=$2
	token=$3
	sqlite_path=$4
	token_source=$5
	timeout=${ABS_SCAN_TIMEOUT:-120}
	login_timeout=${ABS_LOGIN_TIMEOUT:-90}

	if [ ! -f "$sqlite_path" ]; then
		printf 'Missing SQLite database for %s: %s\n' "$name" "$sqlite_path" >&2
		exit 1
	fi

	section "Scanning $name ABS instance"
	detail "url: $url"
	detail "sqlite: $sqlite_path"
	detail "configured token: $(token_status "$token") [$token_source]"

	tmp_response=$(mktemp)
	trap 'rm -f "$tmp_response"' EXIT

	if [ -z "$token" ]; then
		token=$(api_key_token "$sqlite_path")
		if [ -n "$token" ]; then
			token_source="sqlite-api-key"
		else
			token=$(sqlite_token "$sqlite_path" "$ABS_ROOT_USERNAME")
			if [ -n "$token" ]; then
				token_source="sqlite-user:$ABS_ROOT_USERNAME"
			else
				elapsed=0
				while [ "$elapsed" -lt "$login_timeout" ]; do
					if token=$(api_login "$name" "$url" "$tmp_response"); then
						break
					fi
					detail "login not ready for $name after ${elapsed}s; retrying"
					sleep 2
					elapsed=$((elapsed + 2))
				done
				token_source="login:$ABS_ROOT_USERNAME"
			fi
		fi
		if [ -z "$token" ]; then
			printf 'Could not obtain login token for %s within %ss.\n' "$name" "$login_timeout" >&2
			exit 1
		fi
	fi

	detail "active token: $(token_status "$token") [$token_source]"
	if ! api_get "$name" "$url" "$token" "/api/libraries" "$tmp_response"; then
		if [ "$token_source" != "sqlite-api-key" ]; then
			detail "token auth failed for $name; retrying with generated API key"
			token=$(api_key_token "$sqlite_path")
			token_source="sqlite-api-key"
		fi

		if [ -n "$token" ]; then
			detail "active token: $(token_status "$token") [$token_source]"
			if api_get "$name" "$url" "$token" "/api/libraries" "$tmp_response"; then
				detail "authentication OK"
			else
				token=
			fi
		fi

		if [ -z "$token" ]; then
			printf 'Could not authenticate to %s with configured token or generated API key.\n' "$name" >&2
			exit 1
		fi
	else
		detail "authentication OK"
	fi

	sqlite3 -separator '	' "$sqlite_path" \
		"select l.id, l.name, f.path from libraries l join libraryFolders f on f.libraryId = l.id order by l.name, f.path;" |
	while IFS='	' read -r library_id library_name library_path; do
		detail "trigger scan: $library_name ($library_id) $library_path"
		api_post "$name" "$url" "$token" "/api/libraries/$library_id/scan?force=1" "$tmp_response"

		case "$library_path" in
			/audiobooks)
				expected=${ABS_EXPECT_AUDIOBOOKS:-2}
				;;
			/books)
				expected=${ABS_EXPECT_BOOKS:-3}
				;;
			*)
				expected=1
				;;
		esac

		detail "waiting for $library_name scan results"
		elapsed=0
		while [ "$elapsed" -lt "$timeout" ]; do
			api_get "$name" "$url" "$token" "/api/libraries/$library_id/items?limit=1" "$tmp_response"
			total=$(jq -r '.total // 0' "$tmp_response")

			if [ "$total" -ge "$expected" ]; then
				detail "ready: $library_name has $total/$expected items"
				break
			fi

			detail "poll: $library_name has $total/$expected items after ${elapsed}s"
			sleep 2
			elapsed=$((elapsed + 2))
		done

		if [ "$elapsed" -ge "$timeout" ]; then
			printf 'Timed out waiting for %s %s to reach %s items\n' "$name" "$library_name" "$expected" >&2
			exit 1
		fi
	done

	rm -f "$tmp_response"
	trap - EXIT
}

if ! command -v jq >/dev/null 2>&1; then
	printf 'jq is required for ABS scan polling.\n' >&2
	exit 1
fi
if ! command -v openssl >/dev/null 2>&1; then
	printf 'openssl is required for ABS API key token generation.\n' >&2
	exit 1
fi

ABS_PLAIN_URL=${ABS_PLAIN_URL:-http://localhost:13378}
ABS_METADATA_URL=${ABS_METADATA_URL:-http://localhost:13379}
ABS_PLAIN_SQLITE=${ABS_PLAIN_SQLITE:-test/abs/state/plain/config/absdatabase.sqlite}
ABS_METADATA_SQLITE=${ABS_METADATA_SQLITE:-test/abs/state/metadata-enabled/config/absdatabase.sqlite}
ABS_ROOT_USERNAME=${ABS_ROOT_USERNAME:-root}
ABS_ROOT_PASSWORD=${ABS_ROOT_PASSWORD:-password}
ABS_TOKEN=${ABS_TOKEN:-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJrZXlJZCI6ImQ5N2NjODJiLTVkYWUtNDRhOC1iMzM5LWM4N2EzZWNhOTY0YiIsIm5hbWUiOiJSb290QVBJS2V5IiwidHlwZSI6ImFwaSIsImlhdCI6MTc3ODc3NTU2NH0.zHa8RFgO4JKDzZIdfbNukd1waQtzmdxLc_ihvv6zcuQ}
if [ "${ABS_PLAIN_TOKEN:-}" ]; then
	plain_token_source="ABS_PLAIN_TOKEN"
else
	ABS_PLAIN_TOKEN=$ABS_TOKEN
	plain_token_source="ABS_TOKEN"
fi
if [ "${ABS_METADATA_TOKEN:-}" ]; then
	metadata_token_source="ABS_METADATA_TOKEN"
else
	ABS_METADATA_TOKEN=$ABS_TOKEN
	metadata_token_source="ABS_TOKEN"
fi

section "ABS scan configuration"
detail "plain url: $ABS_PLAIN_URL"
detail "metadata url: $ABS_METADATA_URL"
detail "env file: $ENV_FILE"
detail "plain token: $(token_status "${ABS_PLAIN_TOKEN:-}") [$plain_token_source]"
detail "metadata token: $(token_status "${ABS_METADATA_TOKEN:-}") [$metadata_token_source]"
detail "expected audiobooks: ${ABS_EXPECT_AUDIOBOOKS:-2}"
detail "expected books: ${ABS_EXPECT_BOOKS:-3}"

scan_instance "plain" "$ABS_PLAIN_URL" "${ABS_PLAIN_TOKEN:-}" "$ABS_PLAIN_SQLITE" "$plain_token_source"
scan_instance "metadata-enabled" "$ABS_METADATA_URL" "${ABS_METADATA_TOKEN:-}" "$ABS_METADATA_SQLITE" "$metadata_token_source"

section "ABS scans complete"
