// API client for shithead. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ShitheadConfig as ShitheadConfigType, ShitheadResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Shithead game settings. */
export interface ShitheadConfigInput {
  magicTwo?: boolean;
  magicSeven?: boolean;
  magicEight?: boolean;
  magicTen?: boolean;
  fourOfAKindBurn?: boolean;
  cpuDifficulty?: number;
}

/** API client for the Shithead /shithead/exec endpoint. */
export const shitheadApi = {
  exec: (
    command: 'reset' | 'play' | 'log',
    args?: {
      indices?: number[];
      config?: ShitheadConfigInput;
    },
  ) =>
    gameExec<ShitheadResponse>('shithead', {
      command,
      ...(args || {}),
    }),
};

// ShitheadConfigType import is used only for type re-export.
export type { ShitheadConfigType };
