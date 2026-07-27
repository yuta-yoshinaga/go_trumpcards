// API client for cuarenta. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CuarentaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Cuarenta game settings. */
export interface CuarentaConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Commands accepted by the Cuarenta /cuarenta/exec endpoint. */
export type CuarentaCommand = 'reset' | 'play' | 'next' | 'log';

/**
 * API client for the Cuarenta /cuarenta/exec endpoint.
 *
 * Cuarenta is an Ecuadorian 4-player, 2-team capture game played with a 40-card
 * deck (no 8/9/10). On your turn you `play` a hand card (`play` → `{ handIndex }`):
 * it captures all same-rank table cards (with caída / ronda / limpia bonuses) or
 * is laid on the table. `next` starts the next round, `reset` applies the config
 * (CPU difficulty, target score), and `log` fetches the action log.
 *   - `play` → `{ handIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `log` carry no extra fields.
 */
export const cuarentaApi = {
  exec: (command: CuarentaCommand, opts?: { handIndex?: number; config?: CuarentaConfigInput }) =>
    gameExec<CuarentaResponse>('cuarenta', {
      command,
      handIndex: opts?.handIndex,
      config: opts?.config,
    }),
};
