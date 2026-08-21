// API client for narcotic. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { NarcoticResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Narcotic /narcotic/exec endpoint.
 *
 * **`remove` takes no column** — all four exposed cards go together, so there is
 * nothing to choose. Only `move` names a pile. `redeal` gathers the table once
 * the stock is spent, with no limit.
 */
export const narcoticApi = {
  exec: (
    command: 'reset' | 'draw' | 'remove' | 'move' | 'redeal' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n',
    col?: number,
    n?: number,
  ) => gameExec<NarcoticResponse>('narcotic', { command, col, n }),
};
