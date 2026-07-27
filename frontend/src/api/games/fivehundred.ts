// API client for fivehundred. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FiveHundredResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for 500 (Five Hundred) game settings. */
export interface FiveHundredConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Optional parameters for a 500 (Five Hundred) action. */
export interface FiveHundredParams {
  bidKind?: number;
  bidTricks?: number;
  bidSuit?: number;
  discardIndices?: number[];
  cardIndex?: number;
  jokerSuit?: number;
  config?: FiveHundredConfigInput;
}

/** API client for the 500 (Five Hundred) game. Calls POST /fivehundred/exec. */
export const fiveHundredApi = {
  exec: (
    command:
      | 'reset'
      | 'b'
      | 'bid'
      | 'pa'
      | 'pass'
      | 'e'
      | 'exchange'
      | 'p'
      | 'play'
      | 'n'
      | 'next'
      | 'nr'
      | 'nextround'
      | 'hint'
      | 'log',
    params: FiveHundredParams = {},
  ) => gameExec<FiveHundredResponse>('fivehundred', { command, ...params }),
};
