// API client for fortyfives. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FortyFivesResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Auction Forty-Fives game settings. */
export interface FortyFivesConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Auction Forty-Fives /fortyfives/exec endpoint. */
export type FortyFivesCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Auction Forty-Fives /fortyfives/exec endpoint.
 *
 * Auction Forty-Fives is an Irish/Canadian 4-player, 2-team trick-taker with a
 * bidding phase. Players bid Pass/15/20/25 (Jink); the highest bidder's team
 * declares trump and plays five tricks.
 *   - `bid` → `{ bid: number }` (0=Pass 15 20 25)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const fortyFivesApi = {
  exec: (
    command: FortyFivesCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: FortyFivesConfigInput;
    },
  ) =>
    gameExec<FortyFivesResponse>('fortyfives', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
