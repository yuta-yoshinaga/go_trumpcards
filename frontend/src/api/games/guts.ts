// API client for guts. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GutsResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options accepted by {@link gutsApi}.exec on `reset`. */
export interface GutsConfigInput {
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
 * API client for the Guts /guts/exec endpoint.
 *
 * Guts is a fast multi-player pot-vying gambling game.
 *   - `reset` starts a fresh game, optionally applying a {@link GutsConfigInput}.
 *   - `declare` → `(declaration)` submits the human's call (0=out/fold,
 *     1=in/stay) and resolves the round.
 *   - `nextround` deals the next round (chips persist server-side).
 *   - `log` and `hint` carry no extra fields.
 */
export const gutsApi = {
  exec: (command: 'reset' | 'declare' | 'nextround' | 'log' | 'hint', declaration?: number, config?: GutsConfigInput) =>
    gameExec<GutsResponse>('guts', { command, declaration, config }),
};
