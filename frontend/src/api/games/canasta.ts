// API client for canasta. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CanastaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Canasta game settings. */
export interface CanastaConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Canasta /canasta/exec endpoint. */
export const canastaApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'skipmeld' | 'discard' | 'goout' | 'nextround' | 'log',
    cardIndex?: number,
    config?: CanastaConfigInput,
    naturalPairIndices?: number[],
    meldGroups?: number[][],
  ) =>
    gameExec<CanastaResponse>('canasta', {
      command,
      cardIndex,
      config,
      naturalPairIndices,
      meldGroups,
    }),
};
