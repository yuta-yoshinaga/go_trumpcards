// Type declarations for somerset. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Somerset. */
export interface SomersetTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Somerset. */
export interface SomersetHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Somerset game state returned from the API. */
export interface SomersetResponse extends BaseGameResponse {
  tableau: SomersetTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: SomersetHint;
}

/** Source or target zone for a Somerset card move. */
export interface SomersetMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Streets and Alleys ---
