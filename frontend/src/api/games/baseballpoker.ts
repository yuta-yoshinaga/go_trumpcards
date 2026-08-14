// API client for baseballpoker. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BaseballPokerResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Baseball Poker /baseballpoker/exec endpoint.
 *
 * `amount` is required for `bet` and `raise` and ignored by the other actions.
 * The buy-the-pot answer is sent as the named `pay` / `buyfold` commands rather
 * than a numeric field, so a missing value can never read as "pay".
 */
export const baseballpokerApi = {
  exec: (
    command: 'reset' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'pay' | 'buyfold' | 'next' | 'hint' | 'log',
    params?: { amount?: number },
  ) => gameExec<BaseballPokerResponse>('baseballpoker', { command, ...params }),
};
