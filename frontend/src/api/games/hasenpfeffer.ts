// API client for hasenpfeffer. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { HasenpfefferConfig, HasenpfefferResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Hasenpfeffer /hasenpfeffer/exec endpoint.
 *
 * **Passing is `bid` with `0`, and it is a real value** — the server rejects a
 * missing `bid` rather than defaulting it, because defaulting would pass on
 * behalf of a player who never chose to.
 */
export const hasenpfefferApi = {
  exec: (
    command: 'reset' | 'bid' | 'discard' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<HasenpfefferConfig>,
    suit?: number,
    bid?: number,
  ) => gameExec<HasenpfefferResponse>('hasenpfeffer', { command, cardIndex, config, suit, bid }),
};
