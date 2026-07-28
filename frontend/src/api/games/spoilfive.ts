// API client for spoilfive. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SpoilFiveResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Spoil Five game settings (CPU difficulty only — target points are fixed server-side). */
export interface SpoilFiveConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the Spoil Five /spoilfive/exec endpoint. */
export type SpoilFiveCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Spoil Five /spoilfive/exec endpoint.
 *
 * Spoil Five (Maw) is an Irish play-only trick-taker for 5 players on a 52-card
 * deck (5 cards each). Trump is a turned-up card; the trump 5, trump J, and ♥A
 * are the fixed top trumps and may be held back (Reneging). The first player to
 * win 3 of the 5 tricks takes the pot; otherwise the round is a Spoil (流局) and
 * the pot carries over. First player to targetPoints wins the match.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const spoilFiveApi = {
  exec: (
    command: SpoilFiveCommand,
    opts?: {
      cardIndex?: number;
      config?: SpoilFiveConfigInput;
    },
  ) =>
    gameExec<SpoilFiveResponse>('spoilfive', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
