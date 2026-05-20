#!/usr/bin/env sh
set -eu

latest_stable_tag() {
	git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname |
		grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' |
		head -n 1 || true
}

base_version="${BETA_BASE_VERSION:-$(latest_stable_tag)}"
if [ -z "$base_version" ]; then
	base_version="v0.0.0"
fi

short_sha="${BETA_SHORT_SHA:-$(git rev-parse --short HEAD)}"
branch_name="${BETA_BRANCH_NAME:-}"
if [ -z "$branch_name" ]; then
	if [ -n "${GITHUB_REF:-}" ]; then
		branch_name="${GITHUB_REF#refs/heads/}"
	else
		branch_name="$(git rev-parse --abbrev-ref HEAD)"
	fi
fi
branch_name="$(printf '%s' "$branch_name" | tr '/' '-')"

version="${base_version}-beta.${short_sha}"
release_name="${base_version}-beta (${branch_name})"

{
	printf 'version=%s\n' "$version"
	printf 'release_name=%s\n' "$release_name"
	printf 'branch=%s\n' "$branch_name"
} | tee -a "${GITHUB_OUTPUT:-/dev/null}"
