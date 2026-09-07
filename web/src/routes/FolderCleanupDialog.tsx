import { useId, useState } from 'react';

import type { FolderMove, FolderMoveOutcome } from '../api/catalogue.ts';
import { proposeFolders, type FolderProposal } from './folder-cleanup.ts';
import { useModalDialog } from './useModalDialog.ts';

type Row = FolderProposal & { readonly include: boolean; readonly error: string | null };

/**
 * FolderCleanupDialog is the dry run behind the matrix "Cleanup" button: one
 * row per root-level key with the folder `proposeFolders` derived from its
 * name, editable before anything is written. The operator can untick a key,
 * retype its folder (the datalist offers every proposed and existing folder),
 * or blank it to keep the key at the root. Nothing moves until "Move".
 *
 * When every root key shares a first segment, whether that segment is a
 * namespace to strip or the folder itself is not decidable from names, so
 * the dialog shows the heuristic's pick as a checkbox the operator can flip.
 * Flipping recomputes every row, discarding row edits, and says so.
 *
 * After a run, moved keys leave the list and refused keys stay with their
 * refusal beside them, so a partial run (a revision budget that ran out, a
 * folder the scanner blocked) is retried from where it stopped, not from
 * scratch.
 */
export function FolderCleanupDialog({
  keys,
  existingFolders,
  busy,
  onApply,
  onClose,
}: {
  keys: readonly { readonly id: string; readonly name: string; readonly folder_path: string }[];
  existingFolders: readonly string[];
  busy: boolean;
  onApply: (moves: readonly FolderMove[]) => Promise<readonly FolderMoveOutcome[]>;
  onClose: () => void;
}) {
  const dialog = useModalDialog();
  const titleId = useId();
  const listId = useId();
  const [strip, setStrip] = useState<boolean | undefined>(undefined);
  const plan = proposeFolders(keys, { strip });
  const fresh = (proposals: readonly FolderProposal[]): readonly Row[] =>
    proposals.map((proposal) => ({ ...proposal, include: proposal.folder !== '', error: null }));
  const [rows, setRows] = useState<readonly Row[]>(() => fresh(plan.proposals));
  const stripId = useId();
  const toggleStrip = (next: boolean): void => {
    setStrip(next);
    setRows(fresh(proposeFolders(keys, { strip: next }).proposals));
  };
  const [moved, setMoved] = useState(0);
  const [failure, setFailure] = useState<string | null>(null);

  const update = (id: string, patch: Partial<Row>): void =>
    setRows((current) => current.map((row) => (row.id === id ? { ...row, ...patch } : row)));

  const moves = rows
    .filter((row) => row.include && row.folder.trim() !== '')
    .map((row) => ({ id: row.id, name: row.name, folder: row.folder.trim() }));

  const options = [...new Set([...existingFolders, ...rows.map((row) => row.folder.trim())])]
    .filter((folder) => folder !== '')
    .sort();

  const apply = (): void => {
    setFailure(null);
    void onApply(moves)
      .then((outcomes) => {
        const byId = new Map(outcomes.map((outcome) => [outcome.id, outcome.error]));
        const done = outcomes.filter((outcome) => outcome.error === null).length;
        setMoved((count) => count + done);
        setRows((current) =>
          current
            .filter((row) => byId.get(row.id) !== null)
            .map((row) => ({ ...row, error: byId.get(row.id) ?? null })),
        );
        if (outcomes.length > 0 && outcomes.every((outcome) => outcome.error === null)) {
          onClose();
        }
      })
      .catch(() => setFailure('The cleanup could not run. Try again.'));
  };

  return (
    <dialog className="matrix-editor catalogue-manage" ref={dialog} aria-labelledby={titleId} onClose={onClose}>
      <div className="matrix-editor__head">
        <div>
          <h2 id={titleId}>Cleanup: group keys into folders</h2>
          <p>
            Folders proposed from key names, for keys not in a folder yet. Edit or untick a row,
            then move. Nothing changes until you do.
          </p>
        </div>
        <button type="button" className="btn matrix-editor__close" aria-label="Close cleanup" onClick={onClose}>
          ✕
        </button>
      </div>

      {plan.prefix === null || rows.length === 0 ? null : (
        <p className="catalogue-manage__row-main">
          <input
            id={stripId}
            type="checkbox"
            checked={plan.stripped}
            disabled={busy}
            onChange={(event) => toggleStrip(event.target.checked)}
          />
          <label htmlFor={stripId}>
            {`Every key starts with ${plan.prefix}_. Treat it as a namespace and fold by the next segment`}
          </label>
          <span className="catalogue-manage__meta">Flipping this recomputes every row.</span>
        </p>
      )}

      {rows.length === 0 ? (
        <p className="catalogue-manage__empty">
          {moved > 0 ? `Moved ${String(moved)} key(s). Every key is in a folder now.` : 'Every key is already in a folder.'}
        </p>
      ) : (
        <ul className="catalogue-manage__list">
          {rows.map((row) => (
            <li className="catalogue-manage__row" key={row.id}>
              <div className="catalogue-manage__row-main">
                <input
                  type="checkbox"
                  aria-label={`Move ${row.name}`}
                  checked={row.include}
                  disabled={busy}
                  onChange={(event) => update(row.id, { include: event.target.checked })}
                />
                <span className="mono">{row.name}</span>
                <input
                  type="text"
                  aria-label={`Folder for ${row.name}`}
                  list={listId}
                  placeholder="(root)"
                  value={row.folder}
                  disabled={busy}
                  onChange={(event) => update(row.id, { folder: event.target.value, include: true })}
                />
              </div>
              {row.error === null ? null : (
                <p className="alert" role="alert">
                  <span className="alert__glyph" aria-hidden="true">!</span>
                  <span>{row.error}</span>
                </p>
              )}
            </li>
          ))}
        </ul>
      )}
      <datalist id={listId}>
        {options.map((folder) => (
          <option key={folder} value={folder} />
        ))}
      </datalist>

      {failure === null ? null : (
        <p className="alert" role="alert">
          <span className="alert__glyph" aria-hidden="true">!</span>
          <span>{failure}</span>
        </p>
      )}

      <div className="matrix-editor__actions">
        {moved > 0 && rows.length > 0 ? (
          <span className="catalogue-manage__meta">{`Moved ${String(moved)} so far.`}</span>
        ) : null}
        <button type="button" className="btn" disabled={busy} onClick={onClose}>
          {rows.length === 0 ? 'Close' : 'Cancel'}
        </button>
        {rows.length === 0 ? null : (
          <button type="button" className="btn btn--primary" disabled={busy || moves.length === 0} onClick={apply}>
            {busy ? 'Moving…' : `Move ${String(moves.length)} key(s)`}
          </button>
        )}
      </div>
    </dialog>
  );
}
