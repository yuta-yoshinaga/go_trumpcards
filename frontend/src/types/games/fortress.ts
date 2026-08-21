// Type declarations for fortress. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Fortress. */
export interface FortressTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Fortress. */
export interface FortressHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Fortress game state returned from the API. */
export interface FortressResponse extends BaseGameResponse {
  tableau: FortressTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: FortressHint;
}

/** Source or target zone for a Fortress card move. */
export interface FortressMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Streets and Alleys ---
