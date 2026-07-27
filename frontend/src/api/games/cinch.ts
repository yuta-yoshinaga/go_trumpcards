// API client for cinch. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CinchResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Cinch (Double Pedro) game settings. */
export interface CinchConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** Commands accepted by the Cinch /cinch/exec endpoint. */
export type CinchCommand = 'reset' | 'bid' | 'trump' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Cinch (Double Pedro / High Five) /cinch/exec endpoint.
 *
 * Cinch is a 4-player 52-card All-Fours/Pitch-family bidding trick-taker. A Bid
 * phase (0=pass, 1-14) decides the bidder, who then names trump and leads.
 *   - `bid` → `{ bid: number }` (0=pass, 1-14)
 *   - `trump` → `{ trumpSuit: number }` (1=♠ 2=♣ 3=♥ 4=♦)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const cinchApi = {
  exec: (
    command: CinchCommand,
    opts?: {
      cardIndex?: number;
      bid?: number;
      trumpSuit?: number;
      config?: CinchConfigInput;
    },
  ) =>
    gameExec<CinchResponse>('cinch', {
      command,
      cardIndex: opts?.cardIndex,
      bid: opts?.bid,
      trumpSuit: opts?.trumpSuit,
      config: opts?.config,
    }),
};
