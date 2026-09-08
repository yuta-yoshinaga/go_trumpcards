// API client for tongits. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TongitsResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Tongits game settings. */
export interface TongitsConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Tongits /tongits/exec endpoint. */
export const tongitsApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'knock' | 'nextround' | 'log',
    cardIndex?: number,
    config?: TongitsConfigInput,
  ) =>
    gameExec<TongitsResponse>('tongits', {
      command,
      cardIndex,
      config,
    }),
};
