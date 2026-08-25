// API client for continentalrummy. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ContinentalRummyResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Continental Rummy /continentalrummy/exec endpoint.
 *
 * **Drawing from the stock and lifting the discard are separate commands**, and
 * **going out still names the card you throw** — going out costs one card like
 * any other turn, so it is never defaulted.
 */
export const continentalrummyApi = {
  exec: (
    command: 'reset' | 'stock' | 'take' | 'discard' | 'goout' | 'next' | 'hint' | 'log',
    params?: { handIndex?: number; config?: { cpuDifficulty?: number; totalRounds?: number } },
  ) => gameExec<ContinentalRummyResponse>('continentalrummy', { command, ...params }),
};
