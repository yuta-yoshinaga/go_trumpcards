// Type declarations for gaps. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested next-move hint in Gaps. */
export interface GapsHint {
  fromRow: number;
  fromCol: number;
  toRow: number;
  toCol: number;
}

/** Full Gaps game state returned from the API. */
export interface GapsResponse extends BaseGameResponse {
  /** 4-row x 13-col grid. `null` cells are gaps. */
  grid: (Card | null)[][];
  redealsUsed: number;
  redealsRemaining: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: GapsHint;
}

// --- Four Card Poker (フォーカードポーカー) ---
