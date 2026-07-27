// API client for indianrummy. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { IndianRummyResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Indian Rummy game settings. */
export interface IndianRummyConfigInput {
  playerCount?: number;
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** API client for the Indian Rummy /indianrummy/exec endpoint. */
export const indianRummyApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'declare' | 'nextround' | 'log',
    cardIndex?: number,
    config?: IndianRummyConfigInput,
  ) =>
    gameExec<IndianRummyResponse>('indianrummy', {
      command,
      cardIndex,
      config,
    }),
};
