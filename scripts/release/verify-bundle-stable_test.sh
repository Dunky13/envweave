#!/bin/sh
set -eu

script_dir=$(CDPATH='' cd -- "$(dirname "$0")" && pwd)
fixture=$(mktemp -d "${TMPDIR:-/tmp}/hikyo-stable-dispatch.XXXXXX")
trap 'rm -rf "$fixture"' EXIT HUP INT TERM
mkdir "$fixture/trust" "$fixture/bundle"
printf '{}\n' >"$fixture/trust/root.json"
printf '{"event":{"signed_by":"github-actions-stable"}}\n' >"$fixture/trust/metadata.json"
printf '{}\n' >"$fixture/trust/metadata.sigstore.json"
printf '{"version":"1.0.0","source_commit":"%040d"}\n' 1 >"$fixture/bundle/release-manifest.json"
cat >"$fixture/verifier" <<'EOF'
#!/bin/sh
printf '%s\n' "$@" >"$VERIFIER_ARGS"
exit "${VERIFIER_STATUS:-0}"
EOF
chmod +x "$fixture/verifier"
export VERIFIER_ARGS="$fixture/args"
export COSIGN_BIN=deliberately-unavailable-cosign
unset HIKYO_STABLE_VERIFIER

verify() {
	"$script_dir/verify-bundle.sh" --root "$fixture/trust/root.json" \
		--metadata "$fixture/trust/metadata.json" \
		--metadata-signature "$fixture/trust/metadata.sigstore.json" \
		--state "$fixture/state" --bundle "$fixture/bundle" "$@"
}

if verify --latest >"$fixture/error" 2>&1; then
	printf 'stable dispatch accepted missing verifier\n' >&2; exit 1
fi
grep -F 'set HIKYO_STABLE_VERIFIER' "$fixture/error" >/dev/null
export HIKYO_STABLE_VERIFIER="$fixture/verifier"
verify --latest --published
grep -Fx -- '--latest' "$fixture/args" >/dev/null
grep -Fx -- '--published' "$fixture/args" >/dev/null
grep -Fx -- '1.0.0' "$fixture/args" >/dev/null
verify --historical 1.0.0
grep -Fx -- '--historical' "$fixture/args" >/dev/null
if grep -Fx -- '--latest' "$fixture/args" >/dev/null; then
	printf 'stable dispatch conflated historical and latest verification\n' >&2; exit 1
fi
verify --trust-only
grep -Fx -- '--trust-only' "$fixture/args" >/dev/null
if verify --historical 2.0.0 >"$fixture/error" 2>&1; then
	printf 'stable dispatch accepted mismatched historical version\n' >&2; exit 1
fi
grep -F 'historical version does not match bundle' "$fixture/error" >/dev/null
export VERIFIER_STATUS=23
result=0
verify --latest >"$fixture/error" 2>&1 || result=$?
[ "$result" -eq 23 ] || { printf 'stable verifier failure was masked\n' >&2; exit 1; }
# Presence of a delegation also selects strict verification for historical
# metadata; editing an unauthenticated event field cannot enable fallback.
printf '{}\n' >"$fixture/trust/metadata.json"
printf '{}\n' >"$fixture/trust/stable-policy.json"
result=0
verify --latest >"$fixture/error" 2>&1 || result=$?
[ "$result" -eq 23 ] || { printf 'stable delegation selected legacy fallback\n' >&2; exit 1; }
printf 'stable dispatch: trusted verifier required; exact version and failure propagation passed\n'
