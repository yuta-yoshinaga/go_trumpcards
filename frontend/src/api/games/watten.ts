// API client for watten. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { WattenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Watten (ヴァッテン) game configuration input shape. */
export interface WattenConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
  maxRaises?: number;
}

/**
 * API client for the Watten /watten/exec endpoint.
 *
 * Watten is a Bavarian/Austrian 4-player, 2-team trick-taker with a raise/bluff
 * stake mechanic. `declare` carries the Schlag `rank` and critical `suit`, `play`
 * carries a `cardIndex`, `raise` takes no args, and `respond` carries `hold`
 * (true = hold/accept, false = fold/concede).
 */
export const wattenApi = {
  exec: (
    command: 'reset' | 'declare' | 'play' | 'raise' | 'respond' | 'nextround' | 'hint',
    rank?: number,
    suit?: number,
    cardIndex?: number,
    hold?: boolean,
    config?: WattenConfigInput,
  ) =>
    gameExec<WattenResponse>('watten', {
      command,
      rank,
      suit,
      cardIndex,
      hold,
      config,
    }),
};
