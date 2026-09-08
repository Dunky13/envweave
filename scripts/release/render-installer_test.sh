#!/bin/sh
set -eu

script_dir=$(CDPATH='' cd -- "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$script_dir/../lib/release.sh"

fixture_dir=$(mktemp -d "${TMPDIR:-/tmp}/hikyo-installer-fixture.XXXXXX")
trap 'rm -rf "$fixture_dir"' EXIT HUP INT TERM
printf '{"root":"fixture"}\n' >"$fixture_dir/root.json"
printf '#!/bin/sh\nexit 0\n' >"$fixture_dir/verify.sh"

"$script_dir/render-installer.sh" 0.1.0 "$fixture_dir/root.json" \
	"$fixture_dir/verify.sh" "$fixture_dir/install.sh" >/dev/null

grep -F "version='0.1.0'" "$fixture_dir/install.sh" >/dev/null
grep -F "root_sha256='$(sha256_file "$fixture_dir/root.json")'" "$fixture_dir/install.sh" >/dev/null
grep -F "verifier_sha256='$(sha256_file "$fixture_dir/verify.sh")'" "$fixture_dir/install.sh" >/dev/null
grep -F "release_lib_sha256='$(sha256_file "$script_dir/../lib/release.sh")'" "$fixture_dir/install.sh" >/dev/null
grep -F "current_trust_url=\"https://raw.githubusercontent.com/\$repository/refs/heads/main\"" \
	"$fixture_dir/install.sh" >/dev/null
grep -F "download \"\$current_trust_url/release/trust/metadata.json\"" \
	"$fixture_dir/install.sh" >/dev/null
if grep -E '@(VERSION|ROOT_SHA256|VERIFIER_SHA256|RELEASE_LIB_SHA256)@' "$fixture_dir/install.sh" >/dev/null; then
	printf 'installer fixture failed: unresolved placeholder\n' >&2
	exit 1
fi

printf 'installer fixture: trust root and verifier pinned; current revocations fetched\n'

mkdir "$fixture_dir/verifiers"
for platform in Linux_x86_64 Linux_arm64 Darwin_x86_64 Darwin_arm64; do
	printf 'verifier fixture %s\n' "$platform" >"$fixture_dir/verifiers/hikyo-release-verifier_$platform"
done
"$script_dir/render-installer.sh" 1.0.0 "$fixture_dir/root.json" \
	"$fixture_dir/verify.sh" "$fixture_dir/stable-install.sh" "$fixture_dir/verifiers" >/dev/null
pins=$(sed -n "s/^stable_verifiers='\(.*\)'$/\1/p" "$fixture_dir/stable-install.sh")
for platform in Linux_x86_64 Linux_arm64 Darwin_x86_64 Darwin_arm64; do
	[ "$(printf '%s' "$pins" | jq -r --arg p "$platform" '.[$p]')" = \
		"$(sha256_file "$fixture_dir/verifiers/hikyo-release-verifier_$platform")" ]
done
sh -n "$fixture_dir/stable-install.sh"
mkdir "$fixture_dir/bin"
cat >"$fixture_dir/bin/curl" <<'EOF'
#!/bin/sh
set -eu
while [ "$#" -gt 0 ]; do
	case "$1" in
		https://*) url=$1; shift ;;
		-o) output=$2; shift 2 ;;
		*) shift ;;
	esac
done
case "$url" in
	*/root.json) cp "$INSTALLER_FIXTURE/root.json" "$output" ;;
	*/verify-bundle.sh) cp "$INSTALLER_FIXTURE/verify.sh" "$output" ;;
	*/scripts/lib/release.sh) cp "$INSTALLER_LIBRARY" "$output" ;;
	*/hikyo-release-verifier_*) printf '#!/bin/sh\ntouch "$INSTALLER_FIXTURE/untrusted-executed"\n' >"$output" ;;
	*) printf '{}\n' >"$output" ;;
esac
EOF
chmod +x "$fixture_dir/bin/curl"
if PATH="$fixture_dir/bin:$PATH" INSTALLER_FIXTURE="$fixture_dir" \
	INSTALLER_LIBRARY="$script_dir/../lib/release.sh" \
	sh "$fixture_dir/stable-install.sh" >"$fixture_dir/install-error" 2>&1; then
	printf 'installer fixture failed: substituted executable accepted\n' >&2; exit 1
fi
grep -F 'pinned stable verifier hash mismatch' "$fixture_dir/install-error" >/dev/null
[ ! -e "$fixture_dir/untrusted-executed" ] || { printf 'installer executed unverified binary\n' >&2; exit 1; }

# Valid bootstrap downloads must not subsequently be overwritten by filenames
# from unauthenticated trust metadata, even when the verifier hash is correct.
printf '{"recovery":{"public_key":"recovery.pub"},"bootstrap_primary":{"public_key":"primary.pub"}}\n' >"$fixture_dir/root.json"
"$script_dir/render-installer.sh" 1.0.0 "$fixture_dir/root.json" \
	"$fixture_dir/verify.sh" "$fixture_dir/stable-install.sh" "$fixture_dir/verifiers" >/dev/null
cat >"$fixture_dir/bin/curl" <<'EOF'
#!/bin/sh
set -eu
while [ "$#" -gt 0 ]; do
	case "$1" in
		https://*) url=$1; shift ;;
		-o) output=$2; shift 2 ;;
		*) shift ;;
	esac
done
printf '%s\n' "$url" >>"$INSTALLER_FIXTURE/downloads"
case "$url" in
	*/root.json) cp "$INSTALLER_FIXTURE/root.json" "$output" ;;
	*/verify-bundle.sh) cp "$INSTALLER_FIXTURE/verify.sh" "$output" ;;
	*/scripts/lib/release.sh) cp "$INSTALLER_LIBRARY" "$output" ;;
	*/hikyo-release-verifier_*) cp "$INSTALLER_FIXTURE/verifiers/${url##*/}" "$output" ;;
	*/metadata.json) cp "$INSTALLER_FIXTURE/metadata.json" "$output" ;;
	*) printf '{}\n' >"$output" ;;
esac
EOF
chmod +x "$fixture_dir/bin/curl"
for attack_name in root.json stable-verifier metadata.json catalog.json metadata.sigstore.json \
	stable-policy.json stable-policy.sigstore.json stable-trusted-root.json recovery.pub; do
	jq -nc --arg name "$attack_name" '{primary_keys:[{public_key:$name}]}' >"$fixture_dir/metadata.json"
	: >"$fixture_dir/downloads"
	if PATH="$fixture_dir/bin:$PATH" INSTALLER_FIXTURE="$fixture_dir" \
		INSTALLER_LIBRARY="$script_dir/../lib/release.sh" \
		sh "$fixture_dir/stable-install.sh" >"$fixture_dir/install-error" 2>&1; then
		printf 'installer fixture failed: metadata overwrite accepted: %s\n' "$attack_name" >&2; exit 1
	fi
	grep -E 'unsafe public-key name|primary key cannot replace recovery key' "$fixture_dir/install-error" >/dev/null
	# Only the initial trusted fetch may have requested the reserved basename.
	requests=$(awk -F/ -v name="$attack_name" '$NF == name {count++} END {print count+0}' "$fixture_dir/downloads")
	[ "$requests" -le 1 ] || { printf 'metadata overwrote reserved path: %s\n' "$attack_name" >&2; exit 1; }
done
jq -nc '{primary_keys:[{public_key:"primary.pub"},{public_key:"primary.pub"}]}' >"$fixture_dir/metadata.json"
if PATH="$fixture_dir/bin:$PATH" INSTALLER_FIXTURE="$fixture_dir" \
	INSTALLER_LIBRARY="$script_dir/../lib/release.sh" \
	sh "$fixture_dir/stable-install.sh" >"$fixture_dir/install-error" 2>&1; then
	printf 'installer fixture failed: duplicate public-key inventory accepted\n' >&2; exit 1
fi
grep -F 'invalid or duplicate public-key inventory' "$fixture_dir/install-error" >/dev/null
rm "$fixture_dir/verifiers/hikyo-release-verifier_Linux_arm64"
if "$script_dir/render-installer.sh" 1.0.0 "$fixture_dir/root.json" \
	"$fixture_dir/verify.sh" "$fixture_dir/incomplete.sh" "$fixture_dir/verifiers" >"$fixture_dir/error" 2>&1; then
	printf 'installer fixture failed: incomplete verifier set accepted\n' >&2
	exit 1
fi
grep -F 'missing stable verifier' "$fixture_dir/error" >/dev/null
printf 'installer fixture: executable substitution blocked before execution; incomplete verifier sets rejected\n'
printf 'installer fixture: metadata cannot overwrite pinned bootstrap paths or duplicate key entries\n'
