// API client for minchiate. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MinchiateResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Minchiate game settings. */
export interface MinchiateConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the Minchiate /minchiate/exec endpoint. */
export type MinchiateCommand = 'reset' | 'scarto' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Minchiate /minchiate/exec endpoint.
 *
 * Minchiate is a 4-player Florentine trick-taker in fixed 2v2 teams on a
 * 97-card tarot deck with **40 trumps**. There is **no bidding**: the dealer
 * buries 13 surplus cards and play begins.
 *   - `scarto` → `{ cardIndices: number[] }` (exactly 13, no trumps or Matto)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const minchiateApi = {
  exec: (
    command: MinchiateCommand,
    opts?: {
      cardIndex?: number;
      cardIndices?: number[];
      config?: MinchiateConfigInput;
    },
  ) =>
    gameExec<MinchiateResponse>('minchiate', {
      command,
      cardIndex: opts?.cardIndex,
      cardIndices: opts?.cardIndices,
      config: opts?.config,
    }),
};
