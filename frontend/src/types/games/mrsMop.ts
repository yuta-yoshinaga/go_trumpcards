// Type declarations for mrsmop. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in MrsMop Solitaire. */
export interface MrsMopHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** Tableau card with face-up state in MrsMop. */
export interface MrsMopTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** Full MrsMop Solitaire game state returned from the API. */
export interface MrsMopResponse extends BaseGameResponse {
  tableau: MrsMopTableauCard[][];
  stockCount: number;
  completedSuits: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  score: number;
  difficulty: number;
  hint?: MrsMopHint;
}

// --- MrsMopette (スパイダレット) ---
