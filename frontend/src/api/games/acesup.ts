// API client for acesup. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { AcesUpResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Aces Up /acesup/exec endpoint. */
export const acesupApi = {
  exec: (
    command: 'reset' | 'draw' | 'remove' | 'move' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n',
    col?: number,
    n?: number,
  ) => gameExec<AcesUpResponse>('acesup', { command, col, n }),
};
