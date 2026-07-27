// Type declarations for spider. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Spider Solitaire. */
export interface SpiderHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** Tableau card with face-up state in Spider. */
export interface SpiderTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** Full Spider Solitaire game state returned from the API. */
export interface SpiderResponse extends BaseGameResponse {
  tableau: SpiderTableauCard[][];
  stockCount: number;
  completedSuits: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  score: number;
  difficulty: number;
  hint?: SpiderHint;
}

// --- Spiderette (スパイダレット) ---
