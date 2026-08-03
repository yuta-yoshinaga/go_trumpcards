// API client for aluette. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { AluetteResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Aluette game settings. */
export interface AluetteConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Aluette /aluette/exec endpoint. */
export type AluetteCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Aluette /aluette/exec endpoint.
 *
 * Aluette is a 4-player Breton trick-taker in fixed 2v2 teams on a 48-card
 * Spanish-suited deck. There is **no trump suit, no bidding and no follow
 * obligation** — six named cards (the *luettes*) simply outrank everything,
 * and suit is otherwise irrelevant.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const aluetteApi = {
  exec: (
    command: AluetteCommand,
    opts?: {
      cardIndex?: number;
      config?: AluetteConfigInput;
    },
  ) =>
    gameExec<AluetteResponse>('aluette', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
