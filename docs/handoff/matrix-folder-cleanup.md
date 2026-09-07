# Matrix folder cleanup

Dogfooding found that imports (`.env`, the wizard's file connectors) leave every
key at the catalogue root. Key names already carry structure
(`HIKYO_ARGON2_MEMORY_KIB`, `HIKYO_BACKUP_RPO`), so the matrix gains a
"Cleanup" button that proposes folders from those names as an editable dry run,
then moves the ticked keys.

## What was built

- `web/src/routes/folder-cleanup.ts`: pure heuristic, `proposeFolders(keys)`.
  Only root-level keys are candidates. When every candidate shares its first
  `_` segment and has three or more segments, that segment is a project
  namespace and is stripped. The folder is the next segment, title-cased, and
  only proposed when at least two candidates share it; a key with no segment
  left after the folder segment, or with a unique prefix, is proposed at the
  root. Tests in `folder-cleanup.test.ts`.
- `web/src/api/catalogue.ts`: `useMoveKeysToFolders(ref)`. Sequential loop:
  create each missing folder row once (409 tolerated), then one
  `updateKeyMetadata` PATCH per key by id. Stops at the first 429 and reports
  the rest as not attempted; any other refusal is recorded on that key and the
  loop continues. Tests in `catalogue-move.test.tsx`.
- `web/src/routes/FolderCleanupDialog.tsx`: the dry run. Checkbox per row,
  folder text input with a datalist of proposed and existing folders, blank
  means root. After a run, moved keys leave the list; refused keys stay with
  their refusal. Tests in `FolderCleanupDialog.test.tsx`.
- `web/src/routes/Matrix.tsx`: "Cleanup" button beside "Folders & linked
  keys", hidden when declarations are locked (git-managed or system project)
  or when no key is at the root.
- `web/e2e/flows/matrix.spec.ts`: one flow, "proposes folders for root-level
  keys and moves the ticked ones".
- Sensitivity inventory refreshed (`sensitiveInventory.json`,
  `private-client-inventory.json`, `private-client-lifetime.html`): the new
  hook carries key ids, names and folder paths only.

## Decisions

- Folders, not key groups. Key groups are co-publish coupling per
  `docs/adr/schema-model.md`; the ask is display grouping, which is the
  folder path, non-semantic metadata under `definitions-edit`.
- One-key folders are not proposed. The operator can still type one.
- Folder rows are created before the PATCH so the "Folders & linked keys"
  dialog lists what the matrix shows. Keys carry a path, not a folder id, so
  this is cosmetic consistency, not a constraint.

## Known ceiling

Each metadata PATCH is one schema revision, charged against the project's
60-per-hour revision budget (`internal/service/keys.go`, section 151). A
cleanup of more than roughly 60 root keys stops at the first 429 with the rest
reported as not attempted; running Cleanup again an hour later resumes. The
upgrade path is a server-side bulk metadata move that produces one revision for
N keys. Not built: it needs a new OpenAPI operation, generated client, service
and store code, and the parity and contract tests that go with them.

## Not built

- Import-time folder suggestion. The wizard's review step already shows a
  per-key folder for connector sources; a "suggest" toggle there would reuse
  `proposeFolders`. Post-import Cleanup covers the reported case.
