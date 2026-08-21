// Type declarations for saliclaw. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Salic Law. */
export interface SalicLawHint {
  /** `'tableau'`, or `'stock'` when the advice is to deal another card. */
  fromZone: string;
  /** Source column, or -1 when the source is the stock. */
  fromIdx: number;
  /** `'foundation'`, `'tableau'`, or `'stock'` (deal another card). */
  toZone: string;
  /** Destination index, or -1 for a deal, which targets no single column. */
  toIdx: number;
}

/** Full Salic Law game state returned from the API. */
export interface SalicLawResponse extends BaseGameResponse {
  /**
   * Eight columns, each based on a king. A column past `openPiles` is empty
   * because its king has not been dealt yet.
   */
  tableau: Card[][];
  /** Eight foundations, built up from ace to jack ignoring suit. */
  foundation: Card[][];
  /** Cards left to deal. There is no waste; a dealt card lands on a column. */
  stockCount: number;
  /** The eight queens taken out of play. Decoration only -- never movable. */
  queens: Card[];
  /** How many columns have had their base king dealt. */
  openPiles: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: SalicLawHint;
}

/** Source or target zone for a Salic Law card move. */
export interface SalicLawMoveZone {
  zone: string;
  col?: number;
}
