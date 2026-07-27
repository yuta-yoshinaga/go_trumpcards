// API client for piquet. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PiquetConfig as PiquetConfigType, PiquetResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Piquet game settings. */
export interface PiquetConfigInput {
  cpuDifficulty?: number;
  dealsPerPartie?: number;
}

/** API client for the Piquet /piquet/exec endpoint. */
export const piquetApi = {
  exec: (
    command: 'reset' | 'e' | 'y' | 'd' | 'p' | 'nd' | 'h' | 'log',
    cardIndex?: number,
    discardIndices?: number[],
    config?: PiquetConfigType,
  ) =>
    gameExec<PiquetResponse>('piquet', {
      command,
      cardIndex,
      discardIndices,
      config,
    }),
};
