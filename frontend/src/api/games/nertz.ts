// API client for nertz. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { NertzConfig as NertzConfigType, NertzMoveZone, NertzResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Nertz / Pounce game settings. */
export interface NertzConfigInput {
  playerCount?: number;
  drawCount?: number;
  targetScore?: number;
  cpuDifficulty?: number;
  cpuTickMoves?: number;
}

/** Source/target zone identifier for a Nertz move. */
export type { NertzMoveZone };

/** API client for the Nertz / Pounce /nertz/exec endpoint. */
export const nertzApi = {
  exec: (
    command: 'reset' | 'nr' | 'tick' | 'd' | 'm' | 'u' | 'h' | 'log',
    args?: {
      playerIdx?: number;
      from?: NertzMoveZone;
      to?: NertzMoveZone;
      config?: NertzConfigInput;
    },
  ) =>
    gameExec<NertzResponse>('nertz', {
      command,
      ...(args || {}),
    }),
};

// NertzConfigType import is used only for type re-export.
export type { NertzConfigType };
