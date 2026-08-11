// Type declarations for royalcotillion. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Royal Cotillion. */
export interface RoyalCotillionHint {
  /** `'tableau'`, `'reserve'`, `'waste'`, or `'stock'`. */
  fromZone: string;
  /** Slot or reserve index, or -1 when the source is the waste or the stock. */
  fromIdx: number;
  /** `'foundation'`, `'tableau'`, or `'waste'` (for a draw). */
  toZone: string;
  /** Destination index, or -1 for a draw, which targets no single pile. */
  toIdx: number;
}

/** Full Royal Cotillion game state returned from the API. */
export interface RoyalCotillionResponse extends BaseGameResponse {
  /**
   * Sixteen slots holding exactly one card each; an empty slot is `null`.
   *
   * This is a flat array, not an array of piles — a slot cannot be stacked.
   */
  tableau: (Card | null)[];
  /**
   * Four reserve piles of three. Only the last card of each is playable, and
   * an emptied pile is never refilled.
   */
  reserve: Card[][];
  /**
   * Eight foundations. Each builds **by twos and wraps**, so it passes through
   * all thirteen ranks: A,3,5,7,9,J,K,2,4,6,8,10,Q (or the deuce-start
   * rotation). Eight piles of thirteen is exactly the 104 cards.
   */
  foundation: Card[][];
  /**
   * Per-foundation series: `true` starts from the Ace, `false` from the deuce.
   * Sent explicitly rather than derived from the index so a reordering cannot
   * silently mislabel the piles.
   */
  foundationOdd: boolean[];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: RoyalCotillionHint;
}

/** Source or target zone for a Royal Cotillion card move. */
export interface RoyalCotillionMoveZone {
  zone: 'tableau' | 'reserve' | 'waste' | 'stock' | 'foundation';
  /** Tableau slot (0..15) or reserve pile (0..3). */
  col?: number;
}
