// API client for ninetynine. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { NinetyNineResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Ninety-Nine game settings. */
export interface NinetyNineConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** API client for the Ninety-Nine /ninetynine/exec endpoint. Bidding submits 3 bury-card indices. */
export const ninetyNineApi = {
  exec: (
    command: 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    buryIndices?: number[],
    cardIndex?: number,
    config?: NinetyNineConfigInput,
  ) => gameExec<NinetyNineResponse>('ninetynine', { command, buryIndices, cardIndex, config }),
};
