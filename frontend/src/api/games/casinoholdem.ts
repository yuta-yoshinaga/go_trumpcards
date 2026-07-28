// API client for casinoholdem. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CasinoHoldemResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Casino Hold'em /casinoholdem/exec endpoint. */
export const casinoholdemApi = {
  exec: (command: 'reset' | 'bet' | 'call' | 'fold' | 'log', amount?: number, bonusBet?: number) =>
    gameExec<CasinoHoldemResponse>('casinoholdem', { command, amount, bonusBet }),
};
