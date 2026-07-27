// API client for tripeaks. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TriPeaksResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the TriPeaks /tripeaks/exec endpoint. */
export const tripeaksApi = {
  exec: (
    command: 'reset' | 'draw' | 'remove' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n',
    row?: number,
    col?: number,
    n?: number,
  ) => gameExec<TriPeaksResponse>('tripeaks', { command, row, col, n }),
};
