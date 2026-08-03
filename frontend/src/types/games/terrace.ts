// Type declarations for terrace. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Terrace. */
export interface TerraceHint {
  /** `'reserve'` (the terrace), `'waste'`, `'tableau'`, or `'stock'`. */
  fromZone: string;
  /** Source pile, or -1 when the source is not the tableau. */
  fromIdx: number;
  /** `'foundation'`, `'tableau'`, or `'waste'` (for a draw). */
  toZone: string;
  /** Destination index, or -1 for a draw, which targets no single pile. */
  toIdx: number;
}

/** Full Terrace game state returned from the API. */
export interface TerraceResponse extends BaseGameResponse {
  /** The 11-card terrace. Its top card can only reach a foundation, and it never refills. */
  reserve: Card[];
  /** Nine piles, dealt one card each. */
  tableau: Card[][];
  /** Eight foundations built up in ALTERNATING COLOUR from {@link TerraceResponse.baseRank}. */
  foundation: Card[][];
  stockCount: number;
  waste: Card[];
  /** The rank the foundations start from. 0 means it has not been fixed yet. */
  baseRank: number;
  /** While true, the first card sent to a foundation fixes the base rank. */
  awaitingBaseRank: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: TerraceHint;
}

/** Source or target zone for a Terrace card move. */
export interface TerraceMoveZone {
  zone: string;
  col?: number;
}
