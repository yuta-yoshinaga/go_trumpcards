// API client for rikken. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RikkenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Rikken /rikken/exec endpoint.
 *
 * **`contract: 0` is a pass, not an omission.** The server distinguishes the
 * two, so the value must always be sent on a bid.
 */
export const rikkenApi = {
  exec: (
    command: 'reset' | 'bid' | 'call' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    contract?: number,
    suit?: number,
  ) => gameExec<RikkenResponse>('rikken', { command, cardIndex, contract, suit }),
};
