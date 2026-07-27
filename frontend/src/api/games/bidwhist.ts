// API client for bidwhist. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BidWhistResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Bid Whist game settings. */
export interface BidWhistConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Optional parameters for a Bid Whist action. */
export interface BidWhistParams {
  bidTricks?: number;
  bidDirection?: number;
  trumpSuit?: number;
  discardIndices?: number[];
  cardIndex?: number;
  config?: BidWhistConfigInput;
}

/** API client for the Bid Whist game. Calls POST /bidwhist/exec. */
export const bidWhistApi = {
  exec: (
    command:
      | 'reset'
      | 'b'
      | 'bid'
      | 'pa'
      | 'pass'
      | 't'
      | 'trump'
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
    params: BidWhistParams = {},
  ) => gameExec<BidWhistResponse>('bidwhist', { command, ...params }),
};
