// API client for basra. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BasraResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Basra (Bastra) game settings (CPU difficulty only). */
export interface BasraConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the Basra /basra/exec endpoint. */
export type BasraCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Basra (Bastra) /basra/exec endpoint.
 *
 * Basra is a 4-player 52-card fishing/capture game.
 *   - `play` → `{ cardIndex, tableIndices? }` (tableIndices = table cards to
 *     capture; omit to trail, a Jack always sweeps)
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const basraApi = {
  exec: (
    command: BasraCommand,
    opts?: {
      cardIndex?: number;
      tableIndices?: number[];
      config?: BasraConfigInput;
    },
  ) =>
    gameExec<BasraResponse>('basra', {
      command,
      cardIndex: opts?.cardIndex,
      tableIndices: opts?.tableIndices,
      config: opts?.config,
    }),
};
