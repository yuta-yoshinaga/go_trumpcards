// API client for colourwhist. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ColourWhistResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Colour Whist /colourwhist/exec endpoint.
 *
 * **`contract: 0` is a pass, not an omission.** Troel (4) is never sent — it is
 * forced at deal time and the server rejects it as a bid.
 */
export const colourwhistApi = {
  exec: (
    command: 'reset' | 'bid' | 'call' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    contract?: number,
    suit?: number,
  ) => gameExec<ColourWhistResponse>('colourwhist', { command, cardIndex, contract, suit }),
};
