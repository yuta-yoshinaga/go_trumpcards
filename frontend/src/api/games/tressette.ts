// API client for tressette. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TressetteResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Tressette game settings. */
export interface TressetteConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** API client for the Tressette /tressette/exec endpoint. */
export const tressetteApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndices?: number[],
    cardIndex?: number,
    config?: TressetteConfigInput,
  ) =>
    gameExec<TressetteResponse>('tressette', {
      command,
      cardIndices,
      cardIndex,
      config,
    }),
};
