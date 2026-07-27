// Type declarations for blackhole. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Black Hole. */
export interface BlackHoleHint {
  fan: number;
}

/** Full Black Hole game state returned from the API. */
export interface BlackHoleResponse extends BaseGameResponse {
  /** The 17 fans (top card is last). */
  fans: Card[][];
  /** The central black hole pile (top card is last). */
  blackHole: Card[];
  /** Current phase (0=Playing, 1=GameClear, 2=GameOver). */
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  hint?: BlackHoleHint;
}

// --- FreeCell (フリーセル) ---
