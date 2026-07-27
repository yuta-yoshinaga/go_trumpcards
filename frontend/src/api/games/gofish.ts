// API client for gofish. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GoFishResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Go Fish game settings. */
export interface GoFishConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Go Fish /gofish/exec endpoint. */
export const goFishApi = {
  exec: (command: 'reset' | 'ask' | 'log', targetIdx?: number, rank?: number, config?: GoFishConfigInput) =>
    gameExec<GoFishResponse>('gofish', { command, targetIdx, rank, config }),
};
