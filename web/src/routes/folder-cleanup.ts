/**
 * Folder cleanup proposes a folder for every key still sitting at the
 * catalogue root, from the key's own name. Imports (`.env`, the wizard's file
 * connectors) never set a folder, so a freshly imported project is one flat
 * list where the names already carry structure: `HIKYO_ARGON2_MEMORY_KIB`,
 * `HIKYO_ARGON2_TIME`, `HIKYO_BACKUP_RPO`. This module reads that structure
 * back and proposes `Argon2` and `Backup`; the operator edits the dry run
 * before anything is written.
 *
 * Pure: no React, no transport, so the heuristic is tested on its own.
 */

export type FolderProposal = {
  /** The key's immutable id, what the metadata PATCH is addressed to. */
  readonly id: string;
  readonly name: string;
  /** Empty string means "stay at the root". */
  readonly folder: string;
};

/**
 * proposeFolders returns one proposal per root-level key, in input order.
 *
 * Rules, in order:
 * 1. Only keys with an empty folder path are candidates; keys already in a
 *    folder are never moved.
 * 2. When EVERY candidate shares the same first `_` segment and has at least
 *    three segments, that first segment is a project namespace (`HIKYO_`) and
 *    is stripped before looking for a folder.
 * 3. The folder is the next segment, title-cased (`ARGON2` -> `Argon2`), and
 *    only when at least `minMembers` candidates share it. A key whose name is
 *    exhausted by the folder segment (`HIKYO_VERSION`) has nothing left to
 *    name it inside the folder and stays at the root, as does a key whose
 *    prefix is unique: a one-key folder is noise, not structure.
 */
export function proposeFolders(
  keys: readonly { readonly id: string; readonly name: string; readonly folder_path: string }[],
  minMembers = 2,
): readonly FolderProposal[] {
  const candidates = keys.filter((key) => key.folder_path === '');
  if (candidates.length === 0) return [];

  const segmented = candidates.map((key) => key.name.split('_').filter((part) => part !== ''));
  const first = segmented[0]?.[0];
  const namespaced =
    first !== undefined &&
    segmented.every((parts) => parts.length >= 3 && parts[0] === first);
  const drop = namespaced ? 1 : 0;

  // The folder segment must leave at least one segment behind to name the key.
  const folderOf = new Map<string, string>();
  const members = new Map<string, number>();
  candidates.forEach((key, index) => {
    const parts = segmented[index] ?? [];
    const segment = parts[drop];
    if (segment === undefined || parts.length - drop < 2) return;
    const folder = titleCase(segment);
    folderOf.set(key.id, folder);
    members.set(folder, (members.get(folder) ?? 0) + 1);
  });

  return candidates.map((key) => {
    const folder = folderOf.get(key.id);
    return {
      id: key.id,
      name: key.name,
      folder: folder !== undefined && (members.get(folder) ?? 0) >= minMembers ? folder : '',
    };
  });
}

/** `ARGON2` -> `Argon2`, `mail` -> `Mail`. Digits and other symbols are kept as is. */
export function titleCase(segment: string): string {
  const lower = segment.toLowerCase();
  return lower.charAt(0).toUpperCase() + lower.slice(1);
}
