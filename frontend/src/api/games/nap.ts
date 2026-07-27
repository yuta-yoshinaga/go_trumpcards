// API client for nap. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { NapResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Nap (Napoleon) game settings. */
export interface NapConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Nap /nap/exec endpoint. */
export type NapCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Nap (Napoleon) /nap/exec endpoint.
 *
 * Nap is a British 4-player 5-card gambling trick-taker with a bidding phase.
 * Each player bids once (Pass/Two/Three/Four/Nap = how many of the 5 tricks they
 * will take); the highest bidder becomes the declarer who picks trump and leads.
 *   - `bid` → `{ bid: number }` (0=Pass 2=Two 3=Three 4=Four 5=Nap)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const napApi = {
  exec: (
    command: NapCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: NapConfigInput;
    },
  ) =>
    gameExec<NapResponse>('nap', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
