// API client for chinesepoker. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ChinesePokerResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Chinese Poker /chinesepoker/exec endpoint. */
export const chinesepokerApi = {
  exec: (
    command: 'reset' | 'bet' | 'set' | 'log',
    amount?: number,
    frontIndices?: number[],
    middleIndices?: number[],
  ) => gameExec<ChinesePokerResponse>('chinesepoker', { command, amount, frontIndices, middleIndices }),
};
