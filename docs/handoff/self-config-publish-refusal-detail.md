# Handoff: self-config publish refusals name the key

Status: fix committed on this branch; cross-model review pending (native Codex
usage limit reached 2026-09-07, fallback model is Marc's call).

## Symptom

Publishing drafts in the instance's own configuration project (the
`Hikyo ins_...` project) answered `error 400` with no detail. Reported with
`HIKYO_MCP_ALLOWED_ORIGINS=192.168.0.0/24,95.99.20...` and
`HIKYO_MCP_ENABLED=true` staged in Production.

## Root cause

Two layers dropped the reason:

1. `config.applyManagedOwnerValues` runs the startup parser over the managed
   owner values and replaced its key-named failure
   (`HIKYO_MCP_ALLOWED_ORIGINS: "192.168.0.0/24" is not an exact HTTP(S)
   origin`) with the fixed string `managed owner configuration is invalid for
   this node`.
2. `service.validateSelfConfigCells` wrapped that as a bare `domain.ErrInvalid`
   with no `SafeDetail`, so `bad_request` went out without `error.detail` and
   the matrix fell back to its generic sentence.

The staged value itself is wrong: the key takes exact browser origins
(`https://app.example.com`), not IPs or CIDRs. MCP client IP restriction is
not a setting today.

## Fix

- `internal/config/managed_owner.go`: `managedOwnerRefusal` returns the
  parser message verbatim when it opens with a non-secret managed owner key,
  the key alone for a secret one (`HIKYO_DIRECTORY_PROXY` echoes its raw
  value in parser errors), and the fixed wording for everything else
  (parser-only inputs such as `HIKYO_DB`, cross-variable rules).
- `internal/service/self_config_values.go`: `invalidDetail(...)` so the
  message rides the wire; every other `runtimeconfig.Prepare` refusal was
  already key-named and value-free (`mail.invalid`, node overrides,
  bootstrap sources).
- No web change: `matrixMutationError` already renders
  `Publish refused: <detail> Fix the named key in the matrix row editor`.

Tests: `TestSelfConfigInvalidOwnerValuePublishNamesTheKey` (CIDR refused with
key and value in the detail, corrected origin publishes),
`TestSelfConfigInvalidSecretOwnerValuePublishNamesKeyOnly`,
`TestManagedOwnerRefusalNamesKeyAndHidesSecretValues`.

## Disclosure argument

A self-config publish is post-authorization (`OpValuePublish` on the
instance-config scope) and the values are the caller's own drafts. Managed
owner keys have no prefix overlap (checked), so prefix matching cannot pick
the wrong descriptor's secret flag.

## Not in this branch: upgrade from the web UI

Investigated, not built. `docs/adr/signed-upgrade-compatibility.md` (locked
2026-09-05) already decides the shape: WebUI application is stable-only,
instance-admin-only with fresh proof, and only creates a bounded request for
external orchestration; a separately installed root-owned host helper
(systemd, Compose) or a pull-based Flux controller applies it. The server never
gets a Docker socket or systemd/Helm authority. The API skeleton exists and is
deliberately dead (`requestInstanceUpdate` answers 409
`remote-apply-disabled`; `web/src/routes/Remotes.tsx` apply button gated on
`apply_supported`, forced false). Enabling it is a Marc decision on which
adapter to build first.
