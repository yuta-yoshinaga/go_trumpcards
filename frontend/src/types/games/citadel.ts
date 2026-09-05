// Type declarations for citadel. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Beleaguered Castle. */
export interface CitadelTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Beleaguered Castle. */
export interface CitadelHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Beleaguered Castle game state returned from the API. */
export interface CitadelResponse extends BaseGameResponse {
  tableau: CitadelTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: CitadelHint;
}

/** Source or target zone for a Beleaguered Castle card move. */
export interface CitadelMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Streets and Alleys ---
