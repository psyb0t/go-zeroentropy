#!/bin/bash
# Downconverts the JSON OpenAPI 3.1 document for oapi-codegen's 3.0 loader.
# Nullable anyOf branches become nullable schemas (refs use allOf so nullable
# is not ignored); const becomes a singleton enum. The homogeneous page_span
# prefixItems tuple becomes integer items with its existing length bounds.
# Two download responses have empty JSON schemas. Map these to json.RawMessage
# to retain their JSON bytes without decoding arbitrary untyped objects; the
# script fails if either schema is no longer empty, so the patch gets reviewed.
# All paths and operations are retained. The original download stays untouched.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR
readonly SPEC_IN="${SPEC_DIR}/openapi.json"
readonly SPEC_OUT="${SPEC_DIR}/openapi.fixed.json"

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

# shellcheck disable=SC2016 # $branch and $path are jq variables.
readonly PATCH_FILTER='
walk(
  if type == "object" then
    (if has("const") then .enum = [.const] | del(.const) else . end)
    | (if (.anyOf? | type == "array") and
          any(.anyOf[]; .type? == "null") then
        .nullable = true
        | .anyOf |= map(select(.type? != "null"))
        | if (.anyOf | length) == 1 then
            .anyOf[0] as $branch | del(.anyOf)
            | if ($branch | has("$ref")) then .allOf = [$branch]
              else . + $branch end
          else . end
       else . end)
    | (if has("prefixItems") then
        if (.prefixItems | unique | length) == 1 then
          .items = .prefixItems[0] | del(.prefixItems)
        else error("heterogeneous prefixItems needs an explicit patch") end
       else . end)
  else . end
)
| .openapi = "3.0.3"
| reduce ["/documents/get-document", "/documents/get-page-image"][] as $path
    (.;
      if .paths[$path].get.responses["200"].content["application/json"].schema == {}
      then .paths[$path].get.responses["200"].content["application/json"].schema["x-go-type"] = "json.RawMessage"
      else error("\($path) no longer has an empty JSON success schema; review the json.RawMessage patch") end)
'

main() {
	local tmp
	tmp="$(mktemp "${SPEC_DIR}/.openapi.fixed.XXXXXX")"
	# shellcheck disable=SC2064 # expand tmp before the local leaves scope
	trap "rm -f -- '${tmp}'" EXIT

	jq "$PATCH_FILTER" "$SPEC_IN" >"$tmp"
	chmod 0644 "$tmp"
	mv -- "$tmp" "$SPEC_OUT"
	log INFO "downconverted spec out=${SPEC_OUT##*/}"
}

main "$@"
