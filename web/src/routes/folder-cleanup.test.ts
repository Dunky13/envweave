import { describe, expect, it } from 'vitest';

import { proposeFolders, titleCase } from './folder-cleanup.ts';

const root = (...names: string[]) => names.map((name) => ({ id: `id_${name}`, name, folder_path: '' }));
const folders = (plan: ReturnType<typeof proposeFolders>) =>
  plan.proposals.map((proposal) => [proposal.name, proposal.folder]);

describe('proposeFolders', () => {
  it('strips a shared namespace and folders by the next segment', () => {
    const plan = proposeFolders(
      root('HIKYO_ARGON2_MEMORY_KIB', 'HIKYO_ARGON2_TIME', 'HIKYO_BACKUP_RPO', 'HIKYO_BACKUP_RETAIN_DAYS'),
    );
    expect(plan.prefix).toBe('HIKYO');
    expect(plan.stripped).toBe(true);
    expect(plan.proposals[0]).toEqual({ id: 'id_HIKYO_ARGON2_MEMORY_KIB', name: 'HIKYO_ARGON2_MEMORY_KIB', folder: 'Argon2' });
    expect(folders(plan)).toEqual([
      ['HIKYO_ARGON2_MEMORY_KIB', 'Argon2'],
      ['HIKYO_ARGON2_TIME', 'Argon2'],
      ['HIKYO_BACKUP_RPO', 'Backup'],
      ['HIKYO_BACKUP_RETAIN_DAYS', 'Backup'],
    ]);
  });

  it('leaves a one-key prefix at the root', () => {
    expect(folders(proposeFolders(root('HIKYO_MAIL_FROM', 'HIKYO_MAIL_ADDR', 'HIKYO_AUDIT_DAYS', 'HIKYO_AUDIT_MODE', 'HIKYO_EXTERNAL_ORIGIN')))).toEqual([
      ['HIKYO_MAIL_FROM', 'Mail'],
      ['HIKYO_MAIL_ADDR', 'Mail'],
      ['HIKYO_AUDIT_DAYS', 'Audit'],
      ['HIKYO_AUDIT_MODE', 'Audit'],
      ['HIKYO_EXTERNAL_ORIGIN', ''],
    ]);
  });

  it('uses the first segment as the folder when no prefix is shared', () => {
    const plan = proposeFolders(root('DB_HOST', 'DB_PORT', 'SMTP_HOST', 'SMTP_PASS', 'PORT'));
    expect(plan.prefix).toBeNull();
    expect(plan.stripped).toBe(false);
    expect(folders(plan)).toEqual([
      ['DB_HOST', 'Db'],
      ['DB_PORT', 'Db'],
      ['SMTP_HOST', 'Smtp'],
      ['SMTP_PASS', 'Smtp'],
      ['PORT', ''],
    ]);
  });

  it('does not treat a two-segment key as a shared prefix', () => {
    // HIKYO_VERSION has no segment left after HIKYO, so HIKYO is not a
    // prefix here: it is the folder candidate, shared by all three.
    const plan = proposeFolders(root('HIKYO_VERSION', 'HIKYO_ARGON2_TIME', 'HIKYO_ARGON2_MEMORY'));
    expect(plan.prefix).toBeNull();
    expect(folders(plan)).toEqual([
      ['HIKYO_VERSION', 'Hikyo'],
      ['HIKYO_ARGON2_TIME', 'Hikyo'],
      ['HIKYO_ARGON2_MEMORY', 'Hikyo'],
    ]);
  });

  it('keeps a shared prefix as the folder when stripping yields fewer than two folders', () => {
    const plan = proposeFolders(root('APP_MAIL_FROM', 'APP_MAIL_HOST', 'APP_MAIL_PORT'));
    expect(plan.prefix).toBe('APP');
    expect(plan.stripped).toBe(false);
    expect(folders(plan)).toEqual([
      ['APP_MAIL_FROM', 'App'],
      ['APP_MAIL_HOST', 'App'],
      ['APP_MAIL_PORT', 'App'],
    ]);
  });

  it('lets the caller decide the prefix question either way', () => {
    // Structurally identical to the namespace case: names alone cannot say
    // whether DB is a namespace or the folder, so the caller may override.
    const keys = root('DB_HOST_PRIMARY', 'DB_HOST_REPLICA', 'DB_PORT_PRIMARY', 'DB_PORT_REPLICA');
    expect(proposeFolders(keys).stripped).toBe(true);
    expect(folders(proposeFolders(keys, { strip: false }))).toEqual([
      ['DB_HOST_PRIMARY', 'Db'],
      ['DB_HOST_REPLICA', 'Db'],
      ['DB_PORT_PRIMARY', 'Db'],
      ['DB_PORT_REPLICA', 'Db'],
    ]);
    const forced = proposeFolders(root('APP_MAIL_FROM', 'APP_MAIL_HOST', 'APP_MAIL_PORT'), { strip: true });
    expect(forced.stripped).toBe(true);
    expect(folders(forced)).toEqual([
      ['APP_MAIL_FROM', 'Mail'],
      ['APP_MAIL_HOST', 'Mail'],
      ['APP_MAIL_PORT', 'Mail'],
    ]);
    // Without a shared prefix there is nothing to strip, whatever the caller says.
    expect(proposeFolders(root('DB_HOST', 'SMTP_HOST'), { strip: true }).stripped).toBe(false);
  });

  it('never proposes for a key already in a folder and ignores it for the prefix rule', () => {
    expect(
      folders(
        proposeFolders([
          { id: 'id_OTHER_THING', name: 'OTHER_THING', folder_path: 'Legacy' },
          ...root('HIKYO_MAIL_FROM', 'HIKYO_MAIL_ADDR', 'HIKYO_AUDIT_DAYS', 'HIKYO_AUDIT_MODE'),
        ]),
      ),
    ).toEqual([
      ['HIKYO_MAIL_FROM', 'Mail'],
      ['HIKYO_MAIL_ADDR', 'Mail'],
      ['HIKYO_AUDIT_DAYS', 'Audit'],
      ['HIKYO_AUDIT_MODE', 'Audit'],
    ]);
  });

  it('returns nothing for an empty catalogue', () => {
    expect(proposeFolders([])).toEqual({ prefix: null, stripped: false, proposals: [] });
  });
});

describe('titleCase', () => {
  it('lowercases then capitalises the first character', () => {
    expect(titleCase('ARGON2')).toBe('Argon2');
    expect(titleCase('mail')).toBe('Mail');
    expect(titleCase('2FA')).toBe('2fa');
  });
});
