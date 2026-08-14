// API client for cribbagesquares. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CribbageSquaresResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Cribbage Squares /cribbagesquares/exec endpoint. */
export const cribbagesquaresApi = {
  exec: (command: 'reset' | 'place' | 'undo' | 'giveup' | 'hint' | 'log', row?: number, col?: number) =>
    gameExec<CribbageSquaresResponse>('cribbagesquares', { command, row, col }),
};
