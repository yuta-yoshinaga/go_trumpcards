// API client for prsi. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PrsiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Prší game settings. */
export interface PrsiConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Prší /prsi/exec endpoint. */
export const prsiApi = {
  exec: (command: 'reset' | 'play' | 'draw' | 'log', cardIndex?: number, config?: PrsiConfigInput) =>
    gameExec<PrsiResponse>('prsi', {
      command,
      cardIndex,
      config,
    }),
};
