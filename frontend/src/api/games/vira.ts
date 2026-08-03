// API client for vira. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ViraResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Vira game settings. */
export interface ViraConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the Vira /vira/exec endpoint. */
export type ViraCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Vira /vira/exec endpoint.
 *
 * Vira is a 19th-century Swedish 3-player 52-card trick-taker with a bidding
 * phase. Each player bids once (Pass/Gask/Solo/Misère/Vira); the highest bidder
 * is the declarer who plays alone against the other 2 defenders. Settlement runs
 * through a **pot that carries forward between rounds**, so a run of failed
 * contracts makes the next one worth more.
 *   - `bid` → `{ bid: number }` (0=Pass 1=Gask 2=Solo 3=Misère 4=Vira)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const viraApi = {
  exec: (
    command: ViraCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: ViraConfigInput;
    },
  ) =>
    gameExec<ViraResponse>('vira', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
