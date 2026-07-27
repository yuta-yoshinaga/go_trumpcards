// API client for twotenjack. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TwoTenJackResponse } from '../../types/card';
import { gameExec } from '../gameExec';
import { spadesApi } from './spades';

/** Configuration options for Two Ten Jack game settings. */
export interface TwoTenJackConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Two Ten Jack /twotenjack/exec endpoint.
 *
 * Argument order mirrors {@link spadesApi}: command, trumpSuit, cardIndex, config.
 * This keeps compatibility with {@link hooks/useTrickGameBase.useTrickGameBase | useTrickGameBase} which invokes play as
 * `(command, undefined, cardIndex)`.
 */
export const twoTenJackApi = {
  exec: (
    command: 'reset' | 'declare' | 'play' | 'next' | 'nextround' | 'hint',
    trumpSuit?: number,
    cardIndex?: number,
    config?: TwoTenJackConfigInput,
  ) =>
    gameExec<TwoTenJackResponse>('twotenjack', {
      command,
      trumpSuit,
      cardIndex,
      config,
    }),
};
