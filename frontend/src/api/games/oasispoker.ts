// API client for oasispoker. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OasisPokerResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Oasis Poker /oasispoker/exec endpoint. */
export const oasispokerApi = {
  exec: (
    command: 'reset' | 'bet' | 'exchange' | 'stand' | 'play' | 'fold' | 'log' | 'hint',
    amount?: number,
    jackpotBet?: number,
    indices?: number[],
  ) => gameExec<OasisPokerResponse>('oasispoker', { command, amount, jackpotBet, indices }),
};
