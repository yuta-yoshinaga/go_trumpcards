// Type declarations for acesup. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card in an Aces Up column with action availability flags. */
export interface AcesUpCard {
  card: Card;
  top: boolean;
  removable: boolean;
  movable: boolean;
}

/** A suggested hint in Aces Up. */
export interface AcesUpHint {
  type: 'remove' | 'move' | 'draw';
  col: number;
}

/** Full Aces Up game state returned from the API. */
export interface AcesUpResponse extends BaseGameResponse {
  columns: AcesUpCard[][];
  stockCount: number;
  discardCount: number;
  /** The most recently removed card (top of the discard pile); absent when nothing has been discarded. */
  discardTop?: Card | null;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: AcesUpHint;
}

// --- Pig's Tail ---
