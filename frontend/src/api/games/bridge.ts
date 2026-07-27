// API client for bridge. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BridgeResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Bridge game settings. */
export interface BridgeConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Bridge /bridge/exec endpoint. */
export const bridgeApi = {
  exec: (
    command: 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndex?: number,
    bidType?: number,
    bidLevel?: number,
    bidSuit?: number,
    config?: BridgeConfigInput,
  ) =>
    gameExec<BridgeResponse>('bridge', {
      command,
      cardIndex,
      bidType,
      bidLevel,
      bidSuit,
      config,
    }),
};
