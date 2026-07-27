// API client for russianpoker. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RussianPokerResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Russian Poker /russianpoker/exec endpoint. */
export const russianpokerApi = {
  exec: (
    command: 'reset' | 'bet' | 'exchange' | 'buy6th' | 'select' | 'play' | 'fold' | 'force' | 'decline' | 'log',
    amount?: number,
    indices?: number[],
    discardIndex?: number,
  ) => gameExec<RussianPokerResponse>('russianpoker', { command, amount, indices, discardIndex }),
};
