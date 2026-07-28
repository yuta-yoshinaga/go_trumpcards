// API client for scarto. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ScartoResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Scarto (スカルト) game settings. */
export interface ScartoConfigInput {
  cpuDifficulty?: number;
  targetDeals?: number;
}

/** Commands accepted by the Scarto /scarto/exec endpoint. */
export type ScartoCommand = 'reset' | 'scarto' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Scarto (スカルト) /scarto/exec endpoint.
 *
 * Scarto is a 3-player Italian tarocchi trick-taker on the 78-card tarot deck.
 * The human is seat 0. There is no bidding, chien, or partnership.
 *   - `scarto` → `{ cardIndices }` (the 3 low pip cards the dealer buries)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const scartoApi = {
  exec: (
    command: ScartoCommand,
    opts?: {
      cardIndices?: number[];
      cardIndex?: number;
      config?: ScartoConfigInput;
    },
  ) =>
    gameExec<ScartoResponse>('scarto', {
      command,
      cardIndices: opts?.cardIndices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
