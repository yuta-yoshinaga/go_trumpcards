// Type declarations for napoleonssquare. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Napoleon's Square. */
export interface NapoleonsSquareTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Napoleon's Square. */
export interface NapoleonsSquareHint {
  /** `"waste"`, `"tableau"`, or `"stock"` (meaning: turn a card). */
  fromZone: string;
  /** Source tableau column, or -1 when the source is the waste or stock. */
  fromCol: number;
  /** Head of the run to carry, or -1 when the source is not the tableau. */
  cardIndex: number;
  /** `"foundation"`, `"tableau"`, or `"waste"`. */
  toZone: string;
  /** Destination foundation or column index, or -1 for the waste. */
  toCol: number;
}

/** Full Napoleon's Square game state returned from the API. */
export interface NapoleonsSquareResponse extends BaseGameResponse {
  /** Twelve columns arranged as a square around the foundations. */
  tableau: NapoleonsSquareTableauCard[][];
  stockCount: number;
  waste: Card[];
  /** Eight foundations, two per suit, each seeded with its Ace at deal time. */
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: NapoleonsSquareHint;
}

/** Source or target zone for a Napoleon's Square card move. */
export interface NapoleonsSquareMoveZone {
  zone: string;
  col?: number;
  /** Head of the same-suit run to carry; omit to move only the top card. */
  cardIndex?: number;
}
