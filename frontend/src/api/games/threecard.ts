// API client for threecard. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ThreeCardResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Three Card Poker /threecard/exec endpoint. */
export const threecardApi = {
  exec: (command: 'reset' | 'bet' | 'play' | 'fold' | 'log', amount?: number, pairPlusBet?: number) =>
    gameExec<ThreeCardResponse>('threecard', { command, amount, pairPlusBet }),
};
