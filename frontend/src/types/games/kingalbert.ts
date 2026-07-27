// Type declarations for kingalbert. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in King Albert. */
export interface KingAlbertTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in King Albert. */
export interface KingAlbertHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full King Albert game state returned from the API. */
export interface KingAlbertResponse extends BaseGameResponse {
  tableau: KingAlbertTableauCard[][];
  reserve: (Card | null)[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: KingAlbertHint;
}

/** Source or target zone for a King Albert card move. */
export interface KingAlbertMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Flower Garden ---
