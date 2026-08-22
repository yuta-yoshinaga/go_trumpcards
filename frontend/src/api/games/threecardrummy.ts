// API client for threecardrummy. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ThreeCardRummyResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Three Card Rummy /threecardrummy/exec endpoint. */
export const threecardrummyApi = {
  exec: (
    command: 'reset' | 'bet' | 'rebet' | 'play' | 'fold' | 'log' | 'hint',
    amount?: number,
    lowBonusBet?: number,
  ) => gameExec<ThreeCardRummyResponse>('threecardrummy', { command, amount, lowBonusBet }),
};
