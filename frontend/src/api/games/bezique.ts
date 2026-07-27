// API client for bezique. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BeziqueResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Bezique game settings. */
export interface BeziqueConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Commands accepted by the Bezique /bezique/exec endpoint. */
export type BeziqueCommand = 'reset' | 'play' | 'meld' | 'skip' | 'next' | 'hint' | 'log' | 'config';

/**
 * API client for the Bezique /bezique/exec endpoint.
 *
 * Bezique is a 2-player ancestor of Pinochle played with a 64-card deck. In
 * phase 1 (while stock remains) the trick winner may declare ONE meld for
 * points (marriage 20 / royal marriage 40 / Bezique 40 / four aces 100 / kings
 * 80 / queens 60 / jacks 40) or skip; then both draw. When the stock empties,
 * phase 2 is strict must-follow with no melds and a +10 last-trick bonus.
 * Scores accumulate across deals to a target (default 1000).
 *   - `play` → `{ cardIndex: number }`
 *   - `meld` → `{ meldIndex: number }`
 *   - `reset` / `config` → `{ config }`
 *   - `skip` / `next` / `hint` / `log` carry no extra fields.
 */
export const beziqueApi = {
  exec: (
    command: BeziqueCommand,
    opts?: {
      cardIndex?: number;
      meldIndex?: number;
      config?: BeziqueConfigInput;
    },
  ) =>
    gameExec<BeziqueResponse>('bezique', {
      command,
      cardIndex: opts?.cardIndex,
      meldIndex: opts?.meldIndex,
      config: opts?.config,
    }),
};
