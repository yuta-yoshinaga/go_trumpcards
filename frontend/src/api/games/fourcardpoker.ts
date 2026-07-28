// API client for fourcardpoker. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FourCardPokerResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Four Card Poker /fourcardpoker/exec endpoint. */
export const fourcardpokerApi = {
  exec: (
    command: 'reset' | 'bet' | 'play' | 'fold' | 'log',
    amount?: number,
    acesUpBet?: number,
    playMultiplier?: number,
  ) =>
    gameExec<FourCardPokerResponse>('fourcardpoker', {
      command,
      amount,
      acesUpBet,
      playMultiplier,
    }),
};
