// Type declarations for bakersdozen. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Baker's Dozen. */
export interface BakersDozenTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Baker's Dozen. */
export interface BakersDozenHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Baker's Dozen game state returned from the API. */
export interface BakersDozenResponse extends BaseGameResponse {
  tableau: BakersDozenTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: BakersDozenHint;
}

/** Source or target zone for a Baker's Dozen card move. */
export interface BakersDozenMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}
