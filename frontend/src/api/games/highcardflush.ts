// API client for highcardflush. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { HighCardFlushResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the High Card Flush /highcardflush/exec endpoint. */
export const highcardflushApi = {
  exec: (
    command: 'reset' | 'bet' | 'raise' | 'fold' | 'log' | 'hint',
    amount?: number,
    flushBonusBet?: number,
    straightFlushBet?: number,
    multiplier?: number,
  ) =>
    gameExec<HighCardFlushResponse>('highcardflush', {
      command,
      amount,
      flushBonusBet,
      straightFlushBet,
      multiplier,
    }),
};
