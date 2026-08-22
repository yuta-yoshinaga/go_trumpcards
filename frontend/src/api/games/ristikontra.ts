// API client for ristikontra. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RistikontraResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Pişti game settings. */
export interface RistikontraConfigInput {
  playerCnt?: number;
  cpuDifficulty?: number;
}

/** Commands accepted by the Pişti /ristikontra/exec endpoint. */
export type RistikontraCommand = 'reset' | 'play' | 'next' | 'log';

/**
 * API client for the Pişti /ristikontra/exec endpoint.
 *
 * Pişti is a Turkish 2–4 player capture (fishing) game. On your turn you `play`
 * a hand card onto the central pile (`play` → `{ handIndex }`); matching the
 * pile's top rank, or playing a Jack, captures the whole pile, and capturing a
 * lone card scores a Pişti bonus. `next` starts the next game after one ends;
 * `reset` applies the config (player count, CPU difficulty); `log` fetches the
 * action log.
 *   - `play` → `{ handIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `log` carry no extra fields.
 */
export const ristikontraApi = {
  exec: (command: RistikontraCommand, opts?: { handIndex?: number; config?: RistikontraConfigInput }) =>
    gameExec<RistikontraResponse>('ristikontra', {
      command,
      handIndex: opts?.handIndex,
      config: opts?.config,
    }),
};
