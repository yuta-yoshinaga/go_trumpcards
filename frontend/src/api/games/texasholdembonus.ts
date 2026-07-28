// API client for texasholdembonus. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TexasHoldemBonusResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Texas Hold'em Bonus Poker /texasholdembonus/exec endpoint. */
export const texasholdembonusApi = {
  exec: (command: 'reset' | 'bet' | 'play' | 'fold' | 'check' | 'raise' | 'log', amount?: number, bonusBet?: number) =>
    gameExec<TexasHoldemBonusResponse>('texasholdembonus', { command, amount, bonusBet }),
};
