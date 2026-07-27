// API client for threethirteen. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ThreeThirteenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Three Thirteen game settings. */
export interface ThreeThirteenConfigInput {
  cpuDifficulty?: number;
  playerCount?: number;
}

/** API client for the Three Thirteen /threethirteen/exec endpoint. */
export const threethirteenApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'knock' | 'nextround' | 'log',
    cardIndex?: number,
    config?: ThreeThirteenConfigInput,
  ) =>
    gameExec<ThreeThirteenResponse>('threethirteen', {
      command,
      cardIndex,
      config,
    }),
};
