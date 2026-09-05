// API client for binokel. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BinokelResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Binokel game settings. */
export interface BinokelConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Binokel /binokel/exec endpoint. */
export const binokelApi = {
  exec: (
    command: 'reset' | 'bid' | 'pass' | 'discard' | 'trump' | 'meld' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndex?: number,
    config?: BinokelConfigInput,
    bidAmount?: number,
    suit?: number,
    discardIndices?: number[],
  ) =>
    gameExec<BinokelResponse>('binokel', {
      command,
      cardIndex,
      config,
      bidAmount,
      suit,
      discardIndices,
    }),
};
