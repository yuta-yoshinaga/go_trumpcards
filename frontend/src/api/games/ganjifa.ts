// API client for ganjifa. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GanjifaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Ganjifa game settings. */
export interface GanjifaConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the Ganjifa /ganjifa/exec endpoint. */
export type GanjifaCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Ganjifa /ganjifa/exec endpoint.
 *
 * Ganjifa is a 3-player Mughal-era Indian trick-taker on a circular 96-card
 * deck (8 suits x 12 ranks). There is **no bidding** — trump is auto-declared
 * from the dealer's longest suit — so `play` is the only decision command.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const ganjifaApi = {
  exec: (
    command: GanjifaCommand,
    opts?: {
      cardIndex?: number;
      config?: GanjifaConfigInput;
    },
  ) =>
    gameExec<GanjifaResponse>('ganjifa', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
