// Type declarations for shamrocks. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Shamrocks. */
export interface ShamrocksHint {
  fromFan: number;
  toFan: number;
  toFoundation: boolean;
}

/** Full Shamrocks game state returned from the API. */
export interface ShamrocksResponse extends BaseGameResponse {
  /** Fans of cards (top card is last); the count varies after a redeal. */
  fans: Card[][];
  /** The 4 foundations (A→K by suit). */
  foundation: Card[][];
  /** Remaining gather-and-reshuffle redeals (0–3). */
  redealsLeft: number;
  /** Current phase (0=Playing, 1=GameClear, 2=GameOver). */
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: ShamrocksHint;
}

// --- Simple Simon (シンプル・サイモン) ---
