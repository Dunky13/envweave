#!/bin/sh
set -eu
: "${GH_BIN:=gh}"
[ "$#" -eq 3 ] || { printf 'usage: %s REPOSITORY TAG COMMIT\n' "$0" >&2; exit 2; }
repository=$1 tag=$2 commit=$3
script_dir=$(CDPATH='' cd -- "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$script_dir/../lib/release.sh"
is_full_sha "$commit" && is_semver "${tag#v}" && [ "$tag" != "${tag#v}" ] || exit 2
printf '%s\n' "$repository" | grep -Eq '^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$' || exit 2
reference=$("$GH_BIN" api "repos/$repository/git/ref/tags/$tag")
object=$(printf '%s\n' "$reference" | jq -er '.object | select(.type == "tag") | .sha')
is_full_sha "$object" || exit 1
signed_tag=$("$GH_BIN" api "repos/$repository/git/tags/$object")
printf '%s\n' "$signed_tag" | jq -e --arg tag "$tag" --arg commit "$commit" '
  .tag == $tag and .object.type == "commit" and .object.sha == $commit and
  .verification.verified == true and .verification.reason == "valid"
' >/dev/null || { printf 'release: GitHub-verified annotated tag on exact commit required\n' >&2; exit 1; }
"$script_dir/require-green-main.sh" "$repository" "$commit"
printf 'release: signed tag and exact main CI verified\n'
