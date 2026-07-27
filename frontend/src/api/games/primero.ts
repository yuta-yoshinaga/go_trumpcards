// API client for primero. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PrimeroResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options accepted by {@link primeroApi}.exec on `reset`. */
export interface PrimeroConfigInput {
  /** Number of players at the table (2–6, default 4). */
  playerCount?: number;
  /** Chips each player antes into the pot per round (default 10). */
  ante?: number;
  /** Chips each player starts the match with (default 200). */
  startingChips?: number;
  /** Rounds played before the richest player wins the match (default 10). */
  targetRounds?: number;
}

/**
 * API client for the Primero /primero/exec endpoint.
 *
 * Primero is a Renaissance (16th-century) 4-card poker-vying pot game.
 *   - `reset` starts a fresh game, optionally applying a {@link PrimeroConfigInput}.
 *   - `bet` → `(action)` submits the human's betting action (`"call"` /
 *     `"raise"` / `"fold"`). Raise uses a fixed increment (no amount field).
 *   - `nextround` deals the next round (chips persist server-side).
 *   - `log` and `hint` carry no extra fields.
 */
export const primeroApi = {
  exec: (
    command: 'reset' | 'bet' | 'nextround' | 'log' | 'hint',
    action?: 'call' | 'raise' | 'fold',
    config?: PrimeroConfigInput,
  ) => gameExec<PrimeroResponse>('primero', { command, action, config }),
};
