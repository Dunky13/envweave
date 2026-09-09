# Shared-admission CI timing failure

Source: [CI run 34328354454, race shard 0](https://github.com/Hikyo-Org/Hikyo/actions/runs/34328354454/job/102391372130).

`TestSingletonHARestoreBootRetainsCoordinationThroughSourceRepair/initial-ha`
failed with `repaired owner did not retain shared admission`. The job did not
report a Go data race. The test expected request 61 to be refused after 60
requests, but shared admission uses fixed wall-clock minute windows. Requests
spanning a minute boundary legitimately leave allowance in the new window.

The test now checks the repaired owner's durable discovery-counter total across
all windows for its unique source IP. This proves the real runtime retained
shared admission without assuming the requests finish within one minute.
Admission package tests retain coverage of cross-node rate enforcement and
window reset behavior. Production code and rate-limit policy are unchanged.

Investigation evidence:

- A controlled clock rollover reproduced the original assertion failure three
  times; the same-window control passed.
- The original PostgreSQL topology test passed locally under the race detector.
- The replacement assertion passed an actual PostgreSQL restore boot with a
  deliberately forced minute rollover midway through the requests.
- Disconnecting the repaired limiter's shared backend made the replacement
  assertion fail with `shared admission hits = 0, want 60`.
- All temporary clock probes and backend mutations were removed.

Focused verification requires an isolated PostgreSQL instance:

```sh
HIKYO_TEST_POSTGRES_DSN='postgres://USER@HOST:PORT/postgres?sslmode=disable' \
  go test -race ./internal/app \
  -run '^TestSingletonHARestoreBootRetainsCoordinationThroughSourceRepair$' \
  -count=3 -timeout=5m
go test -race ./internal/admission
```
