// API client for tarocchini. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TarocchiniResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Tarocchini game settings. */
export interface TarocchiniConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the Tarocchini /tarocchini/exec endpoint. */
export type TarocchiniCommand = 'reset' | 'scarto' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Tarocchini /tarocchini/exec endpoint.
 *
 * Tarocchini is a 4-player Bolognese trick-taker in fixed 2v2 teams on a
 * 62-card tarot deck. There is **no bidding**: the dealer buries 2 surplus
 * cards and play begins.
 *   - `scarto` → `{ cardIndices: number[] }` (exactly 2, no trumps or Matto)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const tarocchiniApi = {
  exec: (
    command: TarocchiniCommand,
    opts?: {
      cardIndex?: number;
      cardIndices?: number[];
      config?: TarocchiniConfigInput;
    },
  ) =>
    gameExec<TarocchiniResponse>('tarocchini', {
      command,
      cardIndex: opts?.cardIndex,
      cardIndices: opts?.cardIndices,
      config: opts?.config,
    }),
};
