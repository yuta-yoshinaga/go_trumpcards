// API client for thirtyone. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ThirtyOneResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Thirty-One game settings. */
export interface ThirtyOneConfigInput {
  cpuDifficulty?: number;
  initialLives?: number;
}

/** API client for the Thirty-One /thirtyone/exec endpoint. */
export const thirtyoneApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'knock' | 'nextround' | 'log',
    cardIndex?: number,
    config?: ThirtyOneConfigInput,
  ) =>
    gameExec<ThirtyOneResponse>('thirtyone', {
      command,
      cardIndex,
      config,
    }),
};
