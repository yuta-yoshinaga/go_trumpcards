// Type declarations for diplomat. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Diplomat. */
export interface DiplomatHint {
  /** `'tableau'`, `'waste'`, or `'stock'`. */
  fromZone: string;
  /** Source pile, or -1 when the source is the waste or the stock. */
  fromIdx: number;
  /** `'foundation'`, `'tableau'`, or `'waste'` (for a draw). */
  toZone: string;
  /** Destination index, or -1 for a draw, which targets no single pile. */
  toIdx: number;
}

/** Full Diplomat game state returned from the API. */
export interface DiplomatResponse extends BaseGameResponse {
  /** Eight piles, dealt one card each. Only the top card is available. */
  tableau: Card[][];
  /** Eight foundations, two per suit, opened by Aces and built up to Kings. */
  foundation: Card[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: DiplomatHint;
}

/** Source or target zone for a Diplomat card move. */
export interface DiplomatMoveZone {
  zone: string;
  col?: number;
}
