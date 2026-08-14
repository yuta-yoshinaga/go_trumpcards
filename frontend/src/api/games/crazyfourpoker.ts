// API client for crazyfourpoker. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CrazyFourPokerResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Crazy 4 Poker /crazyfourpoker/exec endpoint.
 *
 * **The ante carries an equal Super Bonus** — there is no separate parameter
 * for it. `queensUp` may be omitted, which means "no side bet".
 *
 * The play multiplier is sent as-is; the server rejects a multiplier the hand
 * does not earn rather than clamping it.
 */
export const crazyfourpokerApi = {
  exec: (
    command: 'reset' | 'bet' | 'play' | 'fold' | 'next' | 'hint' | 'log',
    params?: { ante?: number; queensUp?: number; multiplier?: number; chips?: number },
  ) => gameExec<CrazyFourPokerResponse>('crazyfourpoker', { command, ...params }),
};
