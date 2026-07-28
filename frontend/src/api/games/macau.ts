// API client for macau. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MacauResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Macau game settings. */
export interface MacauConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Macau /macau/exec endpoint. */
export const macauApi = {
  exec: (
    command: 'reset' | 'play' | 'draw' | 'suit' | 'declare' | 'skipdeclare' | 'nextround',
    cardIndex?: number,
    suit?: number,
    config?: MacauConfigInput,
  ) =>
    gameExec<MacauResponse>('macau', {
      command,
      cardIndex,
      suit,
      config,
    }),
};
