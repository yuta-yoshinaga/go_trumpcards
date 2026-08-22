// API client for sevenTwentySeven. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SevenTwentySevenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options accepted by {@link sevenTwentySevenApi}.exec on `reset`. */
export interface SevenTwentySevenConfigInput {
  /** Number of players at the table (2–7, default 4). */
  playerCount?: number;
  /** Chips each player antes into the pot per round (1–1000, default 10). */
  ante?: number;
  /** Chips each player starts the match with (10–100000, default 200). */
  startingChips?: number;
  /** Rounds played before the richest player wins the match (1–100, default 10). */
  targetRounds?: number;
}

/**
 * API client for the SevenTwentySeven /seventwentyseven/exec endpoint.
 *
 *   - `reset` starts a fresh game, optionally applying a {@link SevenTwentySevenConfigInput}.
 *   - `card` takes one more card; `stand` ends the human's drawing for the round.
 *     **Neither carries a parameter** — there is nothing to size or choose.
 *   - `nextround` deals the next round (chips persist server-side).
 *   - `log` and `hint` carry no extra fields.
 */
export const sevenTwentySevenApi = {
  exec: (command: 'reset' | 'card' | 'stand' | 'nextround' | 'log' | 'hint', config?: SevenTwentySevenConfigInput) =>
    gameExec<SevenTwentySevenResponse>('seventwentyseven', { command, config }),
};
