// API client for preference. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PreferenceResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Préférence game settings. */
export interface PreferenceConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Préférence /preference/exec endpoint. */
export type PreferenceCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Préférence /preference/exec endpoint.
 *
 * Préférence is a Russian/Austrian 3-player 32-card trick-taker with a bidding
 * phase. Each player bids once (Pass/Six/Misère/Seven/Eight); the highest bidder
 * is the declarer who plays alone against the other 2 defenders.
 *   - `bid` → `{ bid: number }` (0=Pass 1=Six 2=Misère 3=Seven 4=Eight)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const preferenceApi = {
  exec: (
    command: PreferenceCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: PreferenceConfigInput;
    },
  ) =>
    gameExec<PreferenceResponse>('preference', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
