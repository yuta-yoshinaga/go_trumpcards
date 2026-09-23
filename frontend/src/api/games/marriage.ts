// API client for marriage. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MarriageResponse } from '../../types/games/marriage';
import { gameExec } from '../gameExec';

/** Configuration options for Marriage game settings. */
export interface MarriageConfigInput {
  playerCount?: number;
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** API client for the Marriage /marriage/exec endpoint. */
export const marriageApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'declare' | 'nextround' | 'log',
    cardIndex?: number,
    config?: MarriageConfigInput,
  ) =>
    gameExec<MarriageResponse>('marriage', {
      command,
      cardIndex,
      config,
    }),
};
