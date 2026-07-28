// API client for chinchon. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ChinchonResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Chinchón game settings. */
export interface ChinchonConfigInput {
  cpuDifficulty?: number;
  playerCount?: number;
  knockThreshold?: number;
  eliminationLimit?: number;
}

/** API client for the Chinchón /chinchon/exec endpoint. */
export const chinchonApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'knock' | 'layoff' | 'nextround' | 'log',
    cardIndex?: number,
    config?: ChinchonConfigInput,
    cardIndices?: number[],
  ) =>
    gameExec<ChinchonResponse>('chinchon', {
      command,
      cardIndex,
      cardIndices,
      config,
    }),
};
