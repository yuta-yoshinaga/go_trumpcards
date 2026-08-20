// API client for burraco. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BurracoResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Burraco game settings. */
export interface BurracoConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Burraco /burraco/exec endpoint. */
export const burracoApi = {
  exec: (
    command:
      | 'reset'
      | 'drawstock'
      | 'drawdiscard'
      | 'meld'
      | 'skipmeld'
      | 'discard'
      | 'goout'
      | 'nextround'
      | 'log'
      | 'hint',
    cardIndex?: number,
    config?: BurracoConfigInput,
    naturalPairIndices?: number[],
    meldGroups?: number[][],
  ) =>
    gameExec<BurracoResponse>('burraco', {
      command,
      cardIndex,
      config,
      naturalPairIndices,
      meldGroups,
    }),
};
