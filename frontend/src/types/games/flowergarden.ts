// Type declarations for flowergarden. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Flower Garden. */
export interface FlowerGardenTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Flower Garden. */
export interface FlowerGardenHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Flower Garden game state returned from the API. */
export interface FlowerGardenResponse extends BaseGameResponse {
  tableau: FlowerGardenTableauCard[][];
  reserve: (Card | null)[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: FlowerGardenHint;
}

/** Source or target zone for a Flower Garden card move. */
export interface FlowerGardenMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Tarneeb ---
