# Handoff: system-scope refusals and the first administrator's missing grants

PR: https://github.com/Hikyo-Org/Hikyo/pull/728

Origin: six screenshots from a nightly 35 instance (2026-09-11), every one a
permission-shaped refusal for the first administrator inside the `Hikyo`
organisation and its `Hikyo <instance id>` project.

## Diagnosis (reproduced on a fresh HEAD build, MFA session)

Not a nightly 35 regression: no authz or service change since nightly 30.
Three separate causes.

1. **System scope refuses machine consumers, adapters and SCIM by design.**
   The `Hikyo` organisation and project are the self-configuration hierarchy
   (#686). `internal/authz/self_config.go` applies a closed operation
   allowlist there; service-account, dynamic-provider, lease, adapter and
   SCIM-binding operations are not on it, so they answer 404 whatever the
   caller holds (permission-model ADR, 2026-09-06 amendment). The same calls
   answer 200 in an ordinary organisation. The web app did not know this: it
   offered the pages in the sidebar and rendered denial copy.
2. **Two grants no template seeds.** `audit-read` is outside `admin`
   (audit-model ADR); `instance-directory` is outside `operator`
   (`internal/domain/permission.go`). Granting them through the ordinary
   grant API resolves Audit, Project audit and Remotes. Each grant ends the
   session.
3. **Copy that lied.** Machine access said "needs manage-identities" on any
   query error, to a holder of that capability.

## What this change does

- `web/src/api/selfConfig.ts`: `useSystemScope(enabled)` reads the binding
  from the shared self-config query (one-minute staleness, no polling).
  Enabled only for the session's `instance_operator` hint
  (`useInstanceOperator` in `AuthProvider.tsx`): the status read needs
  `instance-config` and records its denial, and the protected hierarchy is
  hidden from everyone else. Outside a provider (page-level tests) the hint
  reads false; a provider still waiting on whoami reads null, which the gate
  treats as pending.
- `web/src/routes/sidebar-model.ts`: omits `machine-access`, `adapters` and
  `scim` in the system organisation/project (`SYSTEM_SCOPE_REFUSES`).
- `web/src/routes/SystemScope.tsx`: `gateSystemScope` wraps Machine access,
  Deployment adapters and SCIM provisioning so a deep link renders the reason
  instead of the page. Renders nothing until the binding read settles.
- `web/src/routes/Projects.tsx`: no "New project" form in the system
  organisation (the profile refuses extra projects).
- Audit and Remotes refusal copy names the exact grant, its scope and where to
  make it. Machine access copy no longer asserts a cause it cannot know.
- Docs: getting-started § 4 documents the two grants and the system scope;
  browser-operations notes the refused pages.

## Deliberately not changed

- Templates and the allowlist: locked ADR decisions. Seeding `audit-read` or
  `instance-directory` at bootstrap is a separate decision.
- Change approvals stays in the system project: policy reads and votes are
  on the allowlist; policy writes are not, and will refuse there.

## Verification

- `pnpm --dir web exec tsc --noEmit`, `pnpm --dir web exec vitest run` green.
- `pnpm --dir docs/site run check` green.
- Manual: fresh `server --dev`, bootstrap admin, TOTP enrolled, the six
  pages walked before and after the two grants.
