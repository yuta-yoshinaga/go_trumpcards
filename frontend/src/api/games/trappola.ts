// API client for trappola. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TrappolaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Trappola game settings. */
export interface TrappolaConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** API client for the Trappola /trappola/exec endpoint. */
export const trappolaApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndices?: number[],
    cardIndex?: number,
    config?: TrappolaConfigInput,
  ) =>
    gameExec<TrappolaResponse>('trappola', {
      command,
      cardIndices,
      cardIndex,
      config,
    }),
};
