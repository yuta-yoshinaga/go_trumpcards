// API client for solowhist. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SoloWhistResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Solo Whist game settings. */
export interface SoloWhistConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Solo Whist /solowhist/exec endpoint. */
export type SoloWhistCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Solo Whist /solowhist/exec endpoint.
 *
 * Solo Whist is a British 4-player 52-card trick-taker with a bidding phase.
 * Each player bids once (Pass/Solo/Misère/Abundance); the highest bidder is the
 * declarer who plays alone against the other 3 defenders.
 *   - `bid` → `{ bid: number }` (0=Pass 1=Solo 2=Misère 3=Abundance)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const soloWhistApi = {
  exec: (
    command: SoloWhistCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: SoloWhistConfigInput;
    },
  ) =>
    gameExec<SoloWhistResponse>('solowhist', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
