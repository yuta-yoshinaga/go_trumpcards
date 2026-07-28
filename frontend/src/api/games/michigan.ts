// API client for michigan. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MichiganResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options accepted by {@link michiganApi}.exec on `reset`. */
export interface MichiganConfigInput {
  /** Number of players at the table (3–8, default 4). */
  playerCount?: number;
  /** Total chips each player distributes across the four boodles per round (default 8). */
  ante?: number;
  /** Chips each player starts the match with (default 200). */
  startingChips?: number;
  /** Rounds played before the richest player wins the match (default 10). */
  targetRounds?: number;
}

/**
 * API client for the Michigan (Newmarket) /michigan/exec endpoint.
 *
 * Michigan is a "stops" chip-betting game.
 *   - `reset` starts a fresh game, optionally applying a {@link MichiganConfigInput}.
 *   - `bet` → `(boodleBets)` distributes the human's chips across the four
 *     boodles (order A♥, K♣, Q♦, J♠); the array must sum to `betBudget`.
 *   - `play` → `(cardIndex)` plays the hand card at that index (must be in
 *     `playableIndices`).
 *   - `nextround` deals the next round (chips persist server-side).
 *   - `log` and `hint` carry no extra fields.
 */
export const michiganApi = {
  exec: (
    command: 'reset' | 'bet' | 'play' | 'nextround' | 'log' | 'hint',
    boodleBets?: number[],
    cardIndex?: number,
    config?: MichiganConfigInput,
  ) => gameExec<MichiganResponse>('michigan', { command, boodleBets, cardIndex, config }),
};
