#!/bin/bash
# Downloads the ZeroEntropy OpenAPI spec into api/. The spec is not committed;
# `go generate` fetches it as its first step. Stainless hosts the spec under a
# URL that carries a content hash and changes on every update, so the current
# URL is read from the `.stats.yml` of the ZeroEntropy Python SDK repo.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

readonly STATS_URL="https://raw.githubusercontent.com/ZeroEntropy-AI/zeroentropy-python/main/.stats.yml"
readonly DOWNLOAD_TIMEOUT_SECONDS=60
readonly SPEC_NAME="openapi.json"
SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR

on_error() {
	local exit_code="$1"
	local line="$2"

	printf '{"level":"ERROR","file":"%s","line":%d,"msg":"command failed","exit_code":%d}\n' \
		"${BASH_SOURCE[1]}" "$line" "$exit_code" >&2
}

log() {
	local level="$1"
	shift

	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$level" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		--arg msg "$*" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

main() {
	mkdir -p "$SPEC_DIR"
	local tmp
	tmp="$(mktemp "${SPEC_DIR}/.${SPEC_NAME}.XXXXXX")"
	# shellcheck disable=SC2064 # expand tmp now, the local is gone when EXIT fires
	trap "rm -f -- '${tmp}'" EXIT

	local stats spec_url
	log INFO "downloading stats=${STATS_URL}"
	stats="$(wget -q -T "$DOWNLOAD_TIMEOUT_SECONDS" -O - "$STATS_URL")"
	spec_url="$(printf '%s\n' "$stats" |
		sed -n 's/^openapi_spec_url:[[:space:]]*//p' |
		tr -d "\"'\r")"
	if [ -z "$spec_url" ]; then
		log ERROR "no openapi_spec_url in stats=${STATS_URL}"
		exit 1
	fi

	case "$spec_url" in
	https://storage.googleapis.com/stainless-sdk-openapi-specs/zeroentropy%2F*) ;;
	*)
		log ERROR "unexpected spec URL in SDK metadata"
		exit 1
		;;
	esac

	log INFO "downloading spec=${SPEC_NAME} url=${spec_url}"
	wget -q \
		-T "$DOWNLOAD_TIMEOUT_SECONDS" \
		-O "$tmp" \
		"$spec_url"

	# Stainless serves JSON at its .yml URL. Validate before replacing the spec.
	if ! jq -e '(.openapi | startswith("3.")) and (.paths | type == "object")' "$tmp" >/dev/null; then
		log ERROR "downloaded file is not an OpenAPI document spec=${SPEC_NAME}"
		exit 1
	fi

	chmod 0644 "$tmp"
	mv -- "$tmp" "${SPEC_DIR}/${SPEC_NAME}"
	log INFO "saved spec=${SPEC_NAME} openapi=$(jq -r .openapi "${SPEC_DIR}/${SPEC_NAME}")"
}

main "$@"
