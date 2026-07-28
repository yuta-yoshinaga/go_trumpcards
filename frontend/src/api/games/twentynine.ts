// API client for twentynine. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TwentyNineResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Twenty-Nine (29) game settings. */
export interface TwentyNineConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Twenty-Nine (29) /twentynine/exec endpoint. */
export type TwentyNineCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Twenty-Nine (29) /twentynine/exec endpoint.
 *
 * Twenty-Nine is an Indian/Bangladeshi 4-player, 2-team trick-taker with a
 * bidding phase and a hidden trump. Players bid Pass/16/20/24/28; the highest
 * bidder's team picks a hidden trump suit (revealed only mid-play) and plays
 * eight tricks.
 *   - `bid` → `{ bid: number }` (0=Pass 16 20 24 28)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const twentyNineApi = {
  exec: (
    command: TwentyNineCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: TwentyNineConfigInput;
    },
  ) =>
    gameExec<TwentyNineResponse>('twentynine', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
