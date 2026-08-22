// API client for ramsch. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RamschConfig as RamschConfigType, RamschResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Ramsch game settings. */
export interface RamschConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** API client for the Ramsch /ramsch/exec endpoint. */
export const ramschApi = {
  exec: (
    command: 'reset' | 'bid' | 'pickramsch' | 'discard' | 'game' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    args?: {
      accept?: boolean;
      pickup?: boolean;
      discardA?: number;
      discardB?: number;
      gameType?: number;
      trumpSuit?: number;
      cardIndex?: number;
      config?: RamschConfigInput;
    },
  ) =>
    gameExec<RamschResponse>('ramsch', {
      command,
      ...(args || {}),
    }),
};

// RamschConfigType import is used only for type re-export; ensure it's referenced.
export type { RamschConfigType };
