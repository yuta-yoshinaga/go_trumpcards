// Type declarations for beleagueredcastle. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Beleaguered Castle. */
export interface BeleagueredCastleTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Beleaguered Castle. */
export interface BeleagueredCastleHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Beleaguered Castle game state returned from the API. */
export interface BeleagueredCastleResponse extends BaseGameResponse {
  tableau: BeleagueredCastleTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: BeleagueredCastleHint;
}

/** Source or target zone for a Beleaguered Castle card move. */
export interface BeleagueredCastleMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Streets and Alleys ---
