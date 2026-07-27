// API client for golf. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GolfResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Golf Solitaire /golf/exec endpoint. */
export const golfApi = {
  exec: (
    command: 'reset' | 'draw' | 'remove' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n',
    col?: number,
    n?: number,
  ) => gameExec<GolfResponse>('golf', { command, col, n }),
};
