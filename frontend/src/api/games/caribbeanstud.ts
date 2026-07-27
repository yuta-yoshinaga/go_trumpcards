// API client for caribbeanstud. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CaribbeanStudResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Caribbean Stud Poker /caribbeanstud/exec endpoint. */
export const caribbeanstudApi = {
  exec: (command: 'reset' | 'bet' | 'play' | 'fold' | 'log', amount?: number, jackpotBet?: number) =>
    gameExec<CaribbeanStudResponse>('caribbeanstud', { command, amount, jackpotBet }),
};
