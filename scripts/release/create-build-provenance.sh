#!/bin/sh
set -eu
[ "$#" -eq 2 ] || { printf 'usage: %s DIRECTORY CANDIDATE\n' "$0" >&2; exit 2; }
directory=$1 candidate=$2
script_dir=$(CDPATH='' cd -- "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$script_dir/../lib/release.sh"
validate_release_candidate_record "$candidate"
: "${GITHUB_REPOSITORY:?}" "${GITHUB_REF:?}" "${GITHUB_SHA:?}" "${GITHUB_RUN_ID:?}" "${GITHUB_RUN_ATTEMPT:?}"
version=$(jq -r '.version' "$candidate")
commit=$(jq -r '.commit' "$candidate")
[ "$GITHUB_REF" = "refs/tags/v$version" ] && [ "$GITHUB_SHA" = "$commit" ] || exit 1
printf '%s\n' "$GITHUB_REPOSITORY" | grep -Eq '^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$' || exit 2
printf '%s\n' "$GITHUB_RUN_ID" | grep -Eq '^[1-9][0-9]*$' || exit 2
printf '%s\n' "$GITHUB_RUN_ATTEMPT" | grep -Eq '^[1-9][0-9]*$' || exit 2
[ ! -e "$directory/build-provenance.json" ] || { printf 'provenance: output already exists\n' >&2; exit 1; }
scratch=$(mktemp -d "${TMPDIR:-/tmp}/hikyo-build-provenance.XXXXXX")
trap 'rm -rf "$scratch"' EXIT HUP INT TERM
: >"$scratch/subjects"
for path in "$directory"/*; do
  [ -e "$path" ] || [ -L "$path" ] || continue
  [ ! -L "$path" ] || { printf 'provenance: symlinks are forbidden\n' >&2; exit 1; }
  # GoReleaser leaves platform build directories beside the flat assets.
  [ ! -d "$path" ] || continue
  [ -f "$path" ] || { printf 'provenance: only regular assets allowed\n' >&2; exit 1; }
  name=$(basename "$path")
  safe_release_name "$name" || exit 1
  case "$name" in stable-policy.sigstore.json) ;; release-manifest.json | *.sigstore.json) continue ;; esac
  jq -nc --arg name "$name" --arg sha "$(sha256_file "$path")" '{name:$name,digest:{sha256:$sha}}' >>"$scratch/subjects"
done
jq -s --arg repository "https://github.com/$GITHUB_REPOSITORY" --arg ref "$GITHUB_REF" --arg commit "$commit" \
  --arg builder "https://github.com/$GITHUB_REPOSITORY/.github/workflows/release.yml@$GITHUB_REF" \
  --arg invocation "https://github.com/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID/attempts/$GITHUB_RUN_ATTEMPT" '
  {"_type":"https://in-toto.io/Statement/v1",subject:sort_by(.name),predicateType:"https://slsa.dev/provenance/v1",
   predicate:{buildDefinition:{buildType:"https://hikyo.dev/build/github-actions/v1",
     externalParameters:{repository:$repository,ref:$ref},
     resolvedDependencies:[{uri:("git+"+$repository+"@"+$ref),digest:{gitCommit:$commit}}]},
     runDetails:{builder:{id:$builder},metadata:{invocationId:$invocation}}}}
' "$scratch/subjects" >"$directory/build-provenance.json"
