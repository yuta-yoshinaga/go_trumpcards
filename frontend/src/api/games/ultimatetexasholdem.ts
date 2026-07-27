// API client for ultimatetexasholdem. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { UltimateTexasHoldemResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Ultimate Texas Hold'em /ultimatetexasholdem/exec endpoint. */
export const ultimatetexasholdemApi = {
  exec: (
    command: 'reset' | 'bet' | 'play' | 'check' | 'fold' | 'log',
    amount?: number,
    tripsBet?: number,
    multiplier?: number,
  ) => gameExec<UltimateTexasHoldemResponse>('ultimatetexasholdem', { command, amount, tripsBet, multiplier }),
};
