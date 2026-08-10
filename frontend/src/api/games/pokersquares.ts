// API client for pokersquares. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PokerSquaresResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Poker Squares /pokersquares/exec endpoint. */
export const pokersquaresApi = {
  exec: (command: 'reset' | 'place' | 'undo' | 'giveup' | 'hint' | 'log', row?: number, col?: number) =>
    gameExec<PokerSquaresResponse>('pokersquares', { command, row, col }),
};
