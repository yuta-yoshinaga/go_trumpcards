// API client for sedma. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SedmaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Sedma game settings. */
export interface SedmaConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Sedma /sedma/exec endpoint. */
export type SedmaCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Sedma /sedma/exec endpoint.
 *
 * Sedma is a Czech/Slovak 32-card no-trump capture trick-taker, 4 players in 2
 * teams. There is no trump suit and no follow obligation — any card is legal.
 * A card captures the trick if its rank equals the lead card's rank or it is a
 * 7 (wild); the last player to capture wins the trick.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const sedmaApi = {
  exec: (
    command: SedmaCommand,
    opts?: {
      cardIndex?: number;
      config?: SedmaConfigInput;
    },
  ) =>
    gameExec<SedmaResponse>('sedma', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
