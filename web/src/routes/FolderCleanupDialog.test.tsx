// @vitest-environment happy-dom
import { act } from 'react';
import { createRoot } from 'react-dom/client';
import { afterEach, describe, expect, it, vi } from 'vitest';

import type { FolderMove, FolderMoveOutcome } from '../api/catalogue.ts';
import { FolderCleanupDialog } from './FolderCleanupDialog.tsx';

Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });

const proposals = [
  { id: 'id_HIKYO_ARGON2_TIME', name: 'HIKYO_ARGON2_TIME', folder: 'Argon2' },
  { id: 'id_HIKYO_ARGON2_MEMORY_KIB', name: 'HIKYO_ARGON2_MEMORY_KIB', folder: 'Argon2' },
  { id: 'id_HIKYO_EXTERNAL_ORIGIN', name: 'HIKYO_EXTERNAL_ORIGIN', folder: '' },
];

afterEach(() => {
  document.body.innerHTML = '';
});

function set(element: HTMLInputElement, value: string): void {
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set;
  if (setter === undefined) throw new Error('no value setter');
  setter.call(element, value);
  element.dispatchEvent(new Event('input', { bubbles: true }));
}

async function render(onApply: (moves: readonly FolderMove[]) => Promise<readonly FolderMoveOutcome[]>) {
  const onClose = vi.fn();
  const container = document.createElement('div');
  document.body.appendChild(container);
  const root = createRoot(container);
  await act(async () => {
    root.render(
      <FolderCleanupDialog
        proposals={proposals}
        existingFolders={['Legacy']}
        busy={false}
        onApply={onApply}
        onClose={onClose}
      />,
    );
  });
  const button = (label: RegExp) => {
    const found = [...container.querySelectorAll('button')].find((node) => label.test(node.textContent ?? ''));
    if (found === undefined) throw new Error(`no button ${String(label)}`);
    return found;
  };
  const input = (label: string) => {
    const found = container.querySelector<HTMLInputElement>(`input[aria-label="${label}"]`);
    if (found === null) throw new Error(`no input ${label}`);
    return found;
  };
  return { container, onClose, button, input, unmount: () => act(async () => root.unmount()) };
}

describe('FolderCleanupDialog', () => {
  it('ticks proposed keys only, and moves exactly the ticked ones with their edited folders', async () => {
    const onApply = vi.fn<(moves: readonly FolderMove[]) => Promise<readonly FolderMoveOutcome[]>>(
      async (moves) => moves.map((move) => ({ id: move.id, error: null })),
    );
    const view = await render(onApply);
    expect(view.input('Move HIKYO_ARGON2_TIME').checked).toBe(true);
    expect(view.input('Move HIKYO_EXTERNAL_ORIGIN').checked).toBe(false);
    expect(view.button(/^Move/).textContent).toBe('Move 2 key(s)');
    // Datalist offers existing and proposed folders once each.
    const options = [...view.container.querySelectorAll('datalist option')].map((node) => node.getAttribute('value'));
    expect(options).toEqual(['Argon2', 'Legacy']);

    // Untick one, retype another, type a folder for the root key (which ticks it).
    await act(async () => {
      view.input('Move HIKYO_ARGON2_TIME').click();
    });
    await act(async () => {
      set(view.input('Folder for HIKYO_ARGON2_MEMORY_KIB'), 'Hashing');
      set(view.input('Folder for HIKYO_EXTERNAL_ORIGIN'), 'Web');
    });
    expect(view.input('Move HIKYO_EXTERNAL_ORIGIN').checked).toBe(true);
    expect(view.button(/^Move/).textContent).toBe('Move 2 key(s)');

    await act(async () => {
      view.button(/^Move/).click();
    });
    expect(onApply).toHaveBeenCalledWith([
      { id: 'id_HIKYO_ARGON2_MEMORY_KIB', name: 'HIKYO_ARGON2_MEMORY_KIB', folder: 'Hashing' },
      { id: 'id_HIKYO_EXTERNAL_ORIGIN', name: 'HIKYO_EXTERNAL_ORIGIN', folder: 'Web' },
    ]);
    // Every move succeeded: the dialog closes itself.
    expect(view.onClose).toHaveBeenCalledTimes(1);
    await view.unmount();
  });

  it('keeps refused keys with their refusal and drops moved ones', async () => {
    const onApply = vi.fn<(moves: readonly FolderMove[]) => Promise<readonly FolderMoveOutcome[]>>(
      async (moves) =>
        moves.map((move) => ({
          id: move.id,
          error: move.name === 'HIKYO_ARGON2_TIME' ? 'Not moved: budget used up.' : null,
        })),
    );
    const view = await render(onApply);
    await act(async () => {
      view.button(/^Move/).click();
    });
    expect(view.onClose).not.toHaveBeenCalled();
    const keys = [...view.container.querySelectorAll('.catalogue-manage__row .mono')].map((node) => node.textContent);
    expect(keys).toEqual(['HIKYO_ARGON2_TIME', 'HIKYO_EXTERNAL_ORIGIN']);
    expect(view.container.querySelector('[role="alert"]')?.textContent).toContain('Not moved: budget used up.');
    expect(view.container.textContent).toContain('Moved 1 so far.');
    // The refused key is still ticked, so a retry is one click.
    expect(view.button(/^Move/).textContent).toBe('Move 1 key(s)');
    await view.unmount();
  });
});
