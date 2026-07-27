// API client for loo. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { LooResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Loo (Lanterloo) game settings. */
export interface LooConfigInput {
  cpuDifficulty?: number;
  ante?: number;
}

/** Commands accepted by the Loo /loo/exec endpoint. */
export type LooCommand = 'reset' | 'decide' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Loo (Lanterloo) /loo/exec endpoint.
 *
 * Loo is a 4-player 52-card pot-based gambling trick-taker. Trump is set from the
 * turn-up card (no bidding, no trump naming). Each player decides play or pass.
 *   - `decide` → `{ play: boolean }` (true=play, false=pass)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const looApi = {
  exec: (
    command: LooCommand,
    opts?: {
      cardIndex?: number;
      play?: boolean;
      config?: LooConfigInput;
    },
  ) =>
    gameExec<LooResponse>('loo', {
      command,
      cardIndex: opts?.cardIndex,
      play: opts?.play,
      config: opts?.config,
    }),
};
