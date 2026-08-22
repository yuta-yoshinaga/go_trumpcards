// API client for caribbeandraw. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CaribbeanDrawResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Caribbean Draw Poker /caribbeandraw/exec endpoint.
 *
 * `indices` are **0-based** positions in the player's hand and are only read by
 * the `draw` command; omitting them (or passing an empty array) stands pat.
 */
export const caribbeandrawApi = {
  exec: (
    command: 'reset' | 'bet' | 'draw' | 'play' | 'fold' | 'log' | 'hint',
    amount?: number,
    jackpotBet?: number,
    indices?: number[],
  ) => gameExec<CaribbeanDrawResponse>('caribbeandraw', { command, amount, jackpotBet, indices }),
};
