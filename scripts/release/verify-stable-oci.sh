#!/bin/sh
set -eu
[ "$#" -eq 2 ] || { printf 'usage: %s REPOSITORY MANIFEST\n' "$0" >&2; exit 2; }
repository=$1 manifest=$2
: "${COSIGN_BIN:=cosign}"
script_dir=$(CDPATH='' cd -- "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$script_dir/../lib/release.sh"
version=$(jq -er '.version' "$manifest")
commit=$(jq -er '.source_commit' "$manifest")
is_semver "$version" && is_full_sha "$commit" || exit 2
printf '%s\n' "$repository" | grep -Eq '^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$' || exit 2
for kind in image chart-digest; do
  reference=$(jq -er --arg kind "$kind" '
    [.artifacts[] | select(.kind == $kind)] | select(length == 1) | .[0] |
    (.image // .chart)+"@"+.digest' "$manifest")
  "$COSIGN_BIN" verify --certificate-identity "https://github.com/$repository/.github/workflows/release.yml@refs/tags/v$version" \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com \
    --certificate-github-workflow-sha "$commit" \
    --certificate-github-workflow-repository "$repository" \
    --certificate-github-workflow-ref "refs/tags/v$version" "$reference" >/dev/null
done
