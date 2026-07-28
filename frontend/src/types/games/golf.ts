// Type declarations for golf. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card in the Golf tableau with removal and exposure status. */
export interface GolfCard {
  card: Card | null;
  removed: boolean;
  exposed: boolean;
}

/** A suggested hint in Golf Solitaire. */
export interface GolfHint {
  type: string;
  col: number;
}

/** Full Golf Solitaire game state returned from the API. */
export interface GolfResponse extends BaseGameResponse {
  layout: GolfCard[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: GolfHint;
}

// --- Aces Up (四つ葉のクローバー) ---
