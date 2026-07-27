// API client for skat. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SkatConfig as SkatConfigType, SkatResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Skat game settings. */
export interface SkatConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** API client for the Skat /skat/exec endpoint. */
export const skatApi = {
  exec: (
    command: 'reset' | 'bid' | 'pickskat' | 'discard' | 'game' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    args?: {
      accept?: boolean;
      pickup?: boolean;
      discardA?: number;
      discardB?: number;
      gameType?: number;
      trumpSuit?: number;
      cardIndex?: number;
      config?: SkatConfigInput;
    },
  ) =>
    gameExec<SkatResponse>('skat', {
      command,
      ...(args || {}),
    }),
};

// SkatConfigType import is used only for type re-export; ensure it's referenced.
export type { SkatConfigType };
