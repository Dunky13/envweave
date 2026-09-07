import { describe, expect, it } from 'vitest';

import { proposeFolders, titleCase } from './folder-cleanup.ts';

const root = (...names: string[]) => names.map((name) => ({ id: `id_${name}`, name, folder_path: '' }));

describe('proposeFolders', () => {
  it('strips a shared namespace and folders by the next segment', () => {
    expect(
      proposeFolders(
        root(
          'HIKYO_ARGON2_MEMORY_KIB',
          'HIKYO_ARGON2_TIME',
          'HIKYO_BACKUP_RPO',
          'HIKYO_BACKUP_RETAIN_DAYS',
        ),
      ),
    ).toEqual([
      { id: 'id_HIKYO_ARGON2_MEMORY_KIB', name: 'HIKYO_ARGON2_MEMORY_KIB', folder: 'Argon2' },
      { id: 'id_HIKYO_ARGON2_TIME', name: 'HIKYO_ARGON2_TIME', folder: 'Argon2' },
      { id: 'id_HIKYO_BACKUP_RPO', name: 'HIKYO_BACKUP_RPO', folder: 'Backup' },
      { id: 'id_HIKYO_BACKUP_RETAIN_DAYS', name: 'HIKYO_BACKUP_RETAIN_DAYS', folder: 'Backup' },
    ]);
  });

  it('leaves a one-key prefix at the root', () => {
    expect(proposeFolders(root('HIKYO_MAIL_FROM', 'HIKYO_MAIL_ADDR', 'HIKYO_EXTERNAL_ORIGIN'))).toEqual([
      { id: 'id_HIKYO_MAIL_FROM', name: 'HIKYO_MAIL_FROM', folder: 'Mail' },
      { id: 'id_HIKYO_MAIL_ADDR', name: 'HIKYO_MAIL_ADDR', folder: 'Mail' },
      { id: 'id_HIKYO_EXTERNAL_ORIGIN', name: 'HIKYO_EXTERNAL_ORIGIN', folder: '' },
    ]);
  });

  it('uses the first segment as the folder when no namespace is shared', () => {
    expect(proposeFolders(root('DB_HOST', 'DB_PORT', 'SMTP_HOST', 'SMTP_PASS', 'PORT'))).toEqual([
      { id: 'id_DB_HOST', name: 'DB_HOST', folder: 'Db' },
      { id: 'id_DB_PORT', name: 'DB_PORT', folder: 'Db' },
      { id: 'id_SMTP_HOST', name: 'SMTP_HOST', folder: 'Smtp' },
      { id: 'id_SMTP_PASS', name: 'SMTP_PASS', folder: 'Smtp' },
      { id: 'id_PORT', name: 'PORT', folder: '' },
    ]);
  });

  it('does not treat a two-segment key as namespaced', () => {
    // HIKYO_VERSION has no segment left after HIKYO, so HIKYO is not a
    // namespace here: it is the folder candidate, shared by all three.
    expect(proposeFolders(root('HIKYO_VERSION', 'HIKYO_ARGON2_TIME', 'HIKYO_ARGON2_MEMORY'))).toEqual([
      { id: 'id_HIKYO_VERSION', name: 'HIKYO_VERSION', folder: 'Hikyo' },
      { id: 'id_HIKYO_ARGON2_TIME', name: 'HIKYO_ARGON2_TIME', folder: 'Hikyo' },
      { id: 'id_HIKYO_ARGON2_MEMORY', name: 'HIKYO_ARGON2_MEMORY', folder: 'Hikyo' },
    ]);
  });

  it('never proposes for a key already in a folder and ignores it for the namespace rule', () => {
    expect(
      proposeFolders([
        { id: 'id_OTHER_THING', name: 'OTHER_THING', folder_path: 'Legacy' },
        ...root('HIKYO_MAIL_FROM', 'HIKYO_MAIL_ADDR'),
      ]),
    ).toEqual([
      { id: 'id_HIKYO_MAIL_FROM', name: 'HIKYO_MAIL_FROM', folder: 'Mail' },
      { id: 'id_HIKYO_MAIL_ADDR', name: 'HIKYO_MAIL_ADDR', folder: 'Mail' },
    ]);
  });

  it('returns nothing for an empty catalogue', () => {
    expect(proposeFolders([])).toEqual([]);
  });
});

describe('titleCase', () => {
  it('lowercases then capitalises the first character', () => {
    expect(titleCase('ARGON2')).toBe('Argon2');
    expect(titleCase('mail')).toBe('Mail');
    expect(titleCase('2FA')).toBe('2fa');
  });
});
