// Type declarations for americantoad. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in American Toad. */
export interface AmericanToadTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in American Toad. */
export interface AmericanToadHint {
  /** `'reserve'`, `'waste'`, `'tableau'`, or `'stock'` (meaning: turn a card). */
  fromZone: string;
  /** Source column, or -1 when the source is not the tableau. */
  fromIdx: number;
  /** Head of the run to carry, or -1 when the source is not the tableau. */
  cardIndex: number;
  /** `'foundation'`, `'tableau'`, or `'waste'` (for a draw). */
  toZone: string;
  /** Destination index, or -1 for a draw, which targets no single pile. */
  toIdx: number;
}

/** Full American Toad game state returned from the API. */
export interface AmericanToadResponse extends BaseGameResponse {
  /** The 20-card reserve. Only the last element is available. */
  reserve: Card[];
  tableau: AmericanToadTableauCard[][];
  /** Eight foundations, two per suit, all starting at {@link AmericanToadResponse.baseRank}. */
  foundation: Card[][];
  stockCount: number;
  waste: Card[];
  /** The rank all eight foundations start from, fixed by the deal. */
  baseRank: number;
  /** How many times the stock has been turned over. Two passes are allowed. */
  passesUsed: number;
  /** Whether the waste can still be recycled into the stock. */
  canRedeal: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: AmericanToadHint;
}

/** Source or target zone for an American Toad card move. */
export interface AmericanToadMoveZone {
  zone: string;
  col?: number;
  /** Head of the run to carry; omit to move only the top card. */
  cardIndex?: number;
}
