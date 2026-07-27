// API client for anaconda. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { AnacondaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options accepted by {@link anacondaApi}.exec on `reset`. */
export interface AnacondaConfigInput {
  /** Number of players at the table (3–7, default 4). */
  playerCount?: number;
  /** Chips each player antes into the pot per round (default 10). */
  ante?: number;
  /** Chips each player starts the match with (default 200). */
  startingChips?: number;
  /** Rounds played before the richest player wins the match (default 10). */
  targetRounds?: number;
}

/** Bet action accepted by {@link anacondaApi}.exec on the `bet` command. */
export type AnacondaBetAction = 'call' | 'raise' | 'fold';

/**
 * API client for the Anaconda (Pass the Trash) /anaconda/exec endpoint.
 *
 * Anaconda is a poker pot game.
 *   - `reset` starts a fresh game, optionally applying an {@link AnacondaConfigInput}.
 *   - `pass` → `(cardIndices)` passes the selected cards left (3→2→1).
 *   - `keep` → `(cardIndices)` keeps exactly 5 cards (discarding the other 2).
 *   - `bet` → `(action)` calls (also checks), raises, or folds during Roll.
 *   - `nextround` deals the next round (chips persist server-side).
 *   - `log` and `hint` carry no extra fields.
 */
export const anacondaApi = {
  exec: (
    command: 'reset' | 'pass' | 'keep' | 'bet' | 'nextround' | 'log' | 'hint',
    cardIndices?: number[],
    action?: AnacondaBetAction,
    config?: AnacondaConfigInput,
  ) => gameExec<AnacondaResponse>('anaconda', { command, cardIndices, action, config }),
};
