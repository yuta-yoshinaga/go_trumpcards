// API client for bolivia. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BoliviaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Bolivia game settings. */
export interface BoliviaConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Bolivia /bolivia/exec endpoint. */
export const boliviaApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'skipmeld' | 'discard' | 'goout' | 'nextround' | 'log',
    cardIndex?: number,
    config?: BoliviaConfigInput,
    naturalPairIndices?: number[],
    meldGroups?: number[][],
  ) =>
    gameExec<BoliviaResponse>('bolivia', {
      command,
      cardIndex,
      config,
      naturalPairIndices,
      meldGroups,
    }),
};
