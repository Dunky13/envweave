# CLI help at every depth, and docs aligned with the shipped surface

Branch `t3code/f403c26d`. Two deliverables: `--help` that works for every
command and sub-subcommand of the multicall binary, and the docs site, landing
page and README brought back in line with what the code dispatches today.

## What changed

### Help mechanism

- `internal/cli/help.go` (new): a slicer over the frozen client usage text.
  `parseUsage` reads the text into sections, entries and notes using its
  existing layout (unindented `heading:` lines, two-space `hikyo ...`
  synopses, deeper-indented continuation lines, two-space prose). `Help(w,
  path)` renders every entry whose leading literal words match `path`, grouped
  under its section with the section's notes, and trims unmatched trailing
  words so `values set KEY --help` shows `values set`. Alternations
  (`project list|show|create`) answer for each alternative. `HelpRequested`
  stops at `--` so `hikyo run -- child --help` is the child's flag.
- `internal/cli/verbs.go` `Run`: `help <words...>` is rewritten to
  `<words...> --help`; `--help`/`-h`/`-help` anywhere before `--` answers on
  stdout with exit 0 before any flag parsing, resolution, network or
  terminal. The usage text moved into the `usageText` const `Usage` prints.
- `cmd/hikyo/main.go` `runHelp`: same contract for the multicall modes. Host
  groups print their own frozen usage (`app.AdminUsage`, `BackupUsage`,
  `RestoreUsage`, new `EscrowUsage`). `server`, `migrate`, `upgrade`,
  `config-rollout` keep the flag package's generated usage (complete and
  always current; note it goes to stderr for `server` and `migrate`) and now
  exit 0 on `flag.ErrHelp`. `help client` prints the full client reference.
  The update check is skipped when help is requested. Multicall usage text
  rewritten: no em-dash, every dispatched mode listed (`escrow`,
  `config-rollout`, `--version`, `upgrade operator rotate`, admin
  `privacy|config|grant`, backup `upgrade-export|upgrade-drill`, restore
  `drill`), client verbs appended from `cli.Verbs` as wrapped words.

### Help content (findings from a mechanical handler-vs-help audit)

- Client help: added `rotate-scanning-key` (dispatched, undocumented) and the
  flags that existed in code but not in help: login `--name --trust-file
  --device`, establish-credential/recovery/context `--trust-file`,
  instance-config `--idempotency-key --confirm-restored-credentials`, key
  create/declare `--forbidden-in --group ...`, the `--acknowledge` family,
  values publish `--preview-token --confirm-protected`, import
  `--kube-context`, pin `--expires-at --override-schema`, approval policy
  member flags, adapter `--create-environment --move --cancel-move`,
  dynamic-provider `--tls-mode`, instance-config provider create/update/
  disable/refresh-metadata flags, scim binding NameID flags.
- `BackupUsage`: `upgrade-export`, `upgrade-drill` and their flags, the
  `--dev` note. `RestoreUsage`: `--dev` note. `AdminUsage`: privacy actions
  other than export need `--confirm`; `correct` takes optional username and
  display name. Usage-error strings updated to name every subverb.

### Tests

- `internal/cli/golden_test.go`: `TestEveryVerbHasHelp` (every `cli.Verbs`
  entry renders help; unknown verb does not), `TestHelpSlicesTheFrozenText`,
  `TestHelpRequestedStopsAtSeparator`, eight new exit-code matrix rows.
  Goldens `help.txt` and `exit-codes.txt` regenerated with `-update` and
  reviewed.
- `cmd/hikyo/main_test.go`: `TestUsageCoversEveryMode` (every dispatched mode
  has a synopsis and a help section, no em-dash, every client verb listed),
  `TestRunHelpRewritesTheHelpWord`.

### Docs

- `cli-reference.mdx`: new "Getting help" section; server/host table now
  lists `upgrade`, `escrow`, `operator`, `config-rollout`, `updater`,
  `--version`; domain table lists `dynamic-provider`, `lease`, `revision`,
  `pin`, `approval`, `definitions`, `import`, `remote`, `rotate-*`; adapter row
  names GitHub Actions; import wizard described as shipped; removed the
  "commands not yet implemented" section (`render`, `sync`, `adopt` were
  never stubs); managed configuration paragraph describes runtime and node
  settings, not mail only.
- `browser-operations.mdx`: dynamic-secret provider and lease management is
  no longer a tracked browser exception (#595 landed); host-local class lists
  `sudo hikyo upgrade`, `escrow verify`, `config-rollout`; managed
  configuration paragraph as above.
- `getting-started.mdx`: `cd Hikyo` (clone directory case); the `--dev`
  callout no longer claims an environment variable for the admin command.
- `index.mdx`: managed configuration row reworded, configuration rollouts
  row added, em-dashes removed.
- Em-dashes removed from prose in `values-workflows`, `configuration`,
  `dynamic-secrets`, `self-hosting` (repo rule, AGENTS.md). The comparison
  table glyphs on the landing page are not prose and were left.
- `installation.mdx`: signed nightlies exist; install-path table gains the
  `sudo hikyo upgrade` row; callouts say "stable" where they said "official".
- `self-hosting.mdx` and `configuration.mdx`: TLS is imported once into
  managed node configuration; there is no 10-second file poll, `SIGHUP`
  refuses to reload, and the reload-failure counter is `0` by construction
  (`internal/app/tls.go`, `app.go` `ReloadTLS`). Adapter egress policy is
  provider-generic, not Forgejo-only.
- `high-availability.mdx`: "Rolling upgrades" replaced by the full-stop,
  single-migrator, exact-target procedure the chart and upgrades page state.
- `architecture.mdx`: roles table lists upgrade, escrow, operator and
  config-rollout; `restore drill` named as the one root-key exception.
- `upgrades.mdx`: dropped the two-hop "intermediate binary" note (legacy
  bridges now land directly in the current nightly).
- `roadmap.mdx`: snapshot refreshed to today's `main`; the "open pull
  requests" table is now "merged since the previous snapshot"; the upgrade
  foundations paragraph lists what landed.
- `build-from-source.mdx`: `cd Hikyo`.
- Audit claims rejected after verification: `/implementation-status/` does
  build (prepare-content.mjs writes it); `upgrade-nightly.sh` is published
  (same script copies it to `public/`).
- Landing page `index.astro`: page title no longer uses an em-dash.
- `README.md`: `sudo hikyo upgrade` in the operate block; `--help` pointer.

## Verification

- `go test ./cmd/hikyo ./internal/cli ./internal/app` pass.
- `pnpm --dir docs/site run check` and `run build` pass on Node 26.7.0;
  `scripts/ci/check-doc-status.mjs --check` verifies 33 ledger entries.
- Manual probes on the built binary: `hikyo --help`, `hikyo help`, `hikyo help
  client`, `hikyo account factor --help`, `hikyo instance-config --help` (four
  sections), `hikyo values set KEY --stdin --help`, `hikyo help adapter
  target`, `hikyo backup export --help`, `hikyo escrow --help`, `hikyo server
  --help` (exit 0), `hikyo teleport --help` (exit 2), `hikyo run -- ... --help`
  (not help).

## Deliberate limits

- `hikyo server --help`, `hikyo migrate --help` and `hikyo config-rollout
  --help` print through the flag package to stderr (exit 0). Routing them to
  stdout means threading a writer through `config.Load`, which every test
  constructs; not worth it for three commands. The CLI reference says so.
- `hikyo upgrade operator ... --help` answers from the multicall usage text
  because that path loads server configuration before its flag set.
- Client-verb help never opens the controlling terminal: main passes a nil
  session when help is requested.
- Codex review: native `codex exec` hit the account usage limit (reset
  Sep 13); the cross-model review has not run. Fallback model is Marc's call.
- Client help lines are synopses, not exhaustive flag tables. The additions
  above cover every flag the audit found missing; the per-verb flag parser
  still refuses unknown flags with a usage error naming the accepted set.
