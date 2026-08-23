// API client for madrasso. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MadrassoResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Madrasso game settings. */
export interface MadrassoConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** API client for the Madrasso /madrasso/exec endpoint. */
export const madrassoApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndices?: number[],
    cardIndex?: number,
    config?: MadrassoConfigInput,
  ) =>
    gameExec<MadrassoResponse>('madrasso', {
      command,
      cardIndices,
      cardIndex,
      config,
    }),
};
