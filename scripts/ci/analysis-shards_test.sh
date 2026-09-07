#!/bin/sh
set -eu

script_dir=$(CDPATH='' cd -- "$(dirname "$0")" && pwd)
repo_root=$(CDPATH='' cd -- "$script_dir/../.." && pwd)
planner=$repo_root/scripts/ci/analysis-shards
fixture_dir=$(mktemp -d "${TMPDIR:-/tmp}/hikyo-analysis-shards.XXXXXX")
trap 'rm -rf "$fixture_dir"' EXIT HUP INT TERM

mkdir -p \
	"$fixture_dir/extra" \
	"$fixture_dir/internal/app" \
	"$fixture_dir/internal/crypto" \
	"$fixture_dir/internal/isolation" \
	"$fixture_dir/internal/service" \
	"$fixture_dir/internal/store"

printf '%s\n' 'module example.com/shards' 'go 1.27.0' >"$fixture_dir/go.mod"

for package in extra internal/app internal/crypto internal/isolation internal/service internal/store; do
	package_name=${package##*/}
	printf 'package %s\n' "$package_name" >"$fixture_dir/$package/$package_name.go"
done

cat >"$fixture_dir/extra/extra_test.go" <<'EOF'
package extra

import "testing"

func FuzzAuto(f *testing.F) { f.Fuzz(func(*testing.T, []byte) {}) }
EOF

cat >"$fixture_dir/internal/app/app_test.go" <<'EOF'
package app

import "testing"

func TestBoot(t *testing.T)           {}
func TestRestore(t *testing.T)        {}
func TestUpgrade(t *testing.T)        {}
func TestMaintenance(t *testing.T)    {}
func TestScheduler(t *testing.T)      {}
func TestDiagnostics(t *testing.T)    {}
func TestOwnerRuntime(t *testing.T)   {}
func TestAutomaticDrill(t *testing.T) {}
func helperNotATest(t *testing.T)     {}
EOF

cat >"$fixture_dir/internal/crypto/crypto_test.go" <<'EOF'
package crypto

import "testing"

func FuzzOpen(f *testing.F)  { f.Fuzz(func(*testing.T, []byte) {}) }
func FuzzParse(f *testing.F) { f.Fuzz(func(*testing.T, []byte) {}) }
EOF

cat >"$fixture_dir/internal/isolation/isolation_test.go" <<'EOF'
package isolation

import "testing"

func FuzzIsolation(f *testing.F) { f.Fuzz(func(*testing.T, []byte) {}) }
func TestAlpha(t *testing.T)      {}
func TestBravo(t *testing.T)      {}
func TestCharlie(t *testing.T)    {}
func TestDelta(t *testing.T)      {}
func TestEcho(t *testing.T)       {}
func TestFoxtrot(t *testing.T)    {}
func TestGolf(t *testing.T)       {}
func TestHotel(t *testing.T)      {}
func TestIndia(t *testing.T)      {}

// Go does not discover a lowercase rune after the Test prefix.
func Testlower(t *testing.T) {}
EOF

cat >"$fixture_dir/internal/service/service_test.go" <<'EOF'
package service

import "testing"

func FuzzService(f *testing.F) { f.Fuzz(func(*testing.T, []byte) {}) }
EOF

race_actual=$fixture_dir/race-actual
fuzz_actual=$fixture_dir/fuzz-actual
isolation_actual=$fixture_dir/isolation-actual
: >"$race_actual"
: >"$fuzz_actual"
: >"$isolation_actual"

shard=0
while [ "$shard" -lt 3 ]; do
	"$planner" race --root "$fixture_dir" --shard "$shard" --shards 3 |
		awk -v shard="$shard" '{ print shard "\t" $0 }' >>"$race_actual"
	"$planner" fuzz --root "$fixture_dir" --shard "$shard" --shards 3 |
		awk -v shard="$shard" '{ print shard "\t" $0 }' >>"$fuzz_actual"
	"$planner" isolation --root "$fixture_dir" --shard "$shard" --shards 3 |
		awk -v shard="$shard" '{ print shard "\t" $0 }' >>"$isolation_actual"
	shard=$((shard + 1))
done

# Whole packages are assigned once. internal/app is spread over every shard by
# test name; each of its tests must be assigned exactly once and nothing else
# may carry a filter.
tab=$(printf '\t')
if [ -n "$(grep -v "internal/app$tab" "$race_actual" | cut -f2 | sort | uniq -d)" ]; then
	printf 'analysis shard fixture failed: race package assigned more than once\n' >&2
	exit 1
fi
if grep -v "internal/app$tab" "$race_actual" | grep -q "$tab.*$tab"; then
	printf 'analysis shard fixture failed: a package other than app carries a test filter\n' >&2
	exit 1
fi
app_tests=$(grep "internal/app$tab" "$race_actual" | cut -f3 | sed 's/^\^(//; s/)\$$//' | tr '|' '\n' | sort)
if [ -n "$(printf '%s\n' "$app_tests" | uniq -d)" ]; then
	printf 'analysis shard fixture failed: app test assigned more than once\n' >&2
	exit 1
fi
expected_app_tests=$(printf '%s\n' TestAutomaticDrill TestBoot TestDiagnostics TestMaintenance TestOwnerRuntime TestRestore TestScheduler TestUpgrade)
if [ "$app_tests" != "$expected_app_tests" ]; then
	printf 'analysis shard fixture failed: app tests not covered exactly once: %s\n' "$app_tests" >&2
	exit 1
fi
app_shards=$(grep -c "internal/app$tab" "$race_actual" | tr -d ' ')
if [ "$app_shards" -lt 2 ]; then
	printf 'analysis shard fixture failed: app tests were not spread across shards\n' >&2
	exit 1
fi
if [ -n "$(cut -f2- "$fuzz_actual" | sort | uniq -d)" ]; then
	printf 'analysis shard fixture failed: fuzz target assigned more than once\n' >&2
	exit 1
fi
if [ -n "$(cut -f2 "$isolation_actual" | sort | uniq -d)" ]; then
	printf 'analysis shard fixture failed: isolation test assigned more than once\n' >&2
	exit 1
fi

cut -f2 "$race_actual" | sort -u >"$fixture_dir/race-packages"
cat >"$fixture_dir/race-expected" <<'EOF'
example.com/shards/extra
example.com/shards/internal/app
example.com/shards/internal/crypto
example.com/shards/internal/service
example.com/shards/internal/store
EOF
cmp "$fixture_dir/race-expected" "$fixture_dir/race-packages"

# The cumulative service/store race timeout came from co-locating the largest
# authenticated datastore fixtures. Keep them on independent runners; app is
# split by test name so its share on any runner stays bounded.
heavy_shards=$(awk -F '\t' '$2 ~ /^example.com\/shards\/internal\/(service|store)$/ { print $1 }' \
	"$race_actual" | sort -u | wc -l | tr -d ' ')
[ "$heavy_shards" -eq 2 ] || {
	printf 'analysis shard fixture failed: heavy race packages share a runner\n' >&2
	exit 1
}

cut -f2- "$fuzz_actual" | sort >"$fixture_dir/fuzz-targets"
cat >"$fixture_dir/fuzz-expected" <<'EOF'
example.com/shards/extra	FuzzAuto
example.com/shards/internal/crypto	FuzzOpen
example.com/shards/internal/crypto	FuzzParse
example.com/shards/internal/isolation	FuzzIsolation
example.com/shards/internal/service	FuzzService
EOF
cmp "$fixture_dir/fuzz-expected" "$fixture_dir/fuzz-targets"

cut -f2 "$isolation_actual" | sort >"$fixture_dir/isolation-tests"
cat >"$fixture_dir/isolation-expected" <<'EOF'
TestAlpha
TestBravo
TestCharlie
TestDelta
TestEcho
TestFoxtrot
TestGolf
TestHotel
TestIndia
EOF
cmp "$fixture_dir/isolation-expected" "$fixture_dir/isolation-tests"

crypto_shards=$(awk -F '\t' '$2 == "example.com/shards/internal/crypto" { print $1 }' \
	"$fuzz_actual" | sort -u | wc -l | tr -d ' ')
[ "$crypto_shards" -eq 1 ] || {
	printf 'analysis shard fixture failed: one package was split across fuzz shards\n' >&2
	exit 1
}

if "$planner" race --root "$fixture_dir" --shard 3 --shards 3 >/dev/null 2>&1; then
	printf 'analysis shard fixture failed: out-of-range shard accepted\n' >&2
	exit 1
fi

printf 'analysis shard fixture: complete, disjoint race packages, fuzz targets, and isolation tests passed\n'
