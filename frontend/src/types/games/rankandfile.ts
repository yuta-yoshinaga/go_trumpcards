// Type declarations for rankandfile. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Rank and File with face-up/face-down state. */
export interface RankAndFileTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Rank and File. */
export interface RankAndFileHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Rank and File game state returned from the API. */
export interface RankAndFileResponse extends BaseGameResponse {
  tableau: RankAndFileTableauCard[][];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: RankAndFileHint;
}

/** Source or target zone for a Rank and File card move. */
export interface RankAndFileMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}
