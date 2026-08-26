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
 * API client for the /ristikontra/exec endpoint.
 *
 * Ristikontra is a Finnish capture (fishing) game, always 4 players in fixed
 * 2-vs-2 partnerships. On your turn you `play` a hand card onto the central
 * pile (`play` → `{ handIndex }`); **matching the pile top's rank** captures
 * the whole pile — a Jack is an ordinary card, and there is no capture bonus.
 * Immediately after a capture, laying the rank that made it steals the bundle
 * (the counter the game is named for). `next` starts the next game after one
 * ends; `reset` applies the config (CPU difficulty — the table is always 4);
 * `log` fetches the action log.
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
