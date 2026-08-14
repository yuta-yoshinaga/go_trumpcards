// API client for goofspiel. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GoofspielConfig, GoofspielResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Goofspiel /goofspiel/exec endpoint.
 *
 * **`bid` resolves the whole round.** Bidding is simultaneous, so the server
 * takes the CPU bids and the reveal together — there is no separate step.
 */
export const goofspielApi = {
  exec: (
    command: 'reset' | 'bid' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<GoofspielConfig>,
  ) => gameExec<GoofspielResponse>('goofspiel', { command, cardIndex, config }),
};
