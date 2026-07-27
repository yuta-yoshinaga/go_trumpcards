// API client for tonk. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TonkResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Tonk game settings. */
export interface TonkConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Tonk /tonk/exec endpoint. */
export const tonkApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'knock' | 'nextround' | 'log',
    cardIndex?: number,
    config?: TonkConfigInput,
  ) =>
    gameExec<TonkResponse>('tonk', {
      command,
      cardIndex,
      config,
    }),
};
