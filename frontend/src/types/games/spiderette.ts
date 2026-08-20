// Type declarations for spiderette. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Hint returned by the Spiderette /hint endpoint. */
export interface SpideretteHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** Tableau card with face-up state in Spiderette. */
export interface SpideretteTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** Full Spiderette Solitaire game state returned from the API. */
export interface SpideretteResponse extends BaseGameResponse {
  tableau: SpideretteTableauCard[][];
  stockCount: number;
  completedSuits: number;
  score: number;
  /**
   * How the score moves: where it starts, what a move costs, what a completed
   * suit pays. Sent so the explanation cannot quote figures the game stopped
   * using.
   */
  scoring: { start: number; movePenalty: number; suitBonus: number };
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: SpideretteHint;
}

// --- Indian Poker (インディアンポーカー) ---
