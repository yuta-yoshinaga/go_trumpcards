// API client for doubleattack. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DoubleAttackResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Extra Bet Blackjack /doubleattack/exec endpoint.
 *
 * **`amount: 0` declines the extra bet** and must be sent explicitly — the
 * server rejects an `attack` with no amount. `bustIt` may be omitted, which
 * means no side bet.
 *
 * The extra bet is sent as-is; the server rejects one above the ante rather
 * than clamping it.
 */
export const doubleattackApi = {
  exec: (
    command: 'reset' | 'bet' | 'attack' | 'hit' | 'stand' | 'double' | 'split' | 'next' | 'hint' | 'log',
    params?: { ante?: number; bustIt?: number; amount?: number },
  ) => gameExec<DoubleAttackResponse>('doubleattack', { command, ...params }),
};
