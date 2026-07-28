// API client for manille. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ManilleResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Manille game settings. */
export interface ManilleConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Manille /manille/exec endpoint. */
export type ManilleCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Manille /manille/exec endpoint.
 *
 * Manille is a French/Belgian 4-player (2 vs 2) trump trick-taker. The only
 * play action is playing a card (must follow suit / overtrump unless the
 * partner already holds the trick); there are no declarations and no Roem.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const manilleApi = {
  exec: (
    command: ManilleCommand,
    opts?: {
      cardIndex?: number;
      config?: ManilleConfigInput;
    },
  ) =>
    gameExec<ManilleResponse>('manille', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
