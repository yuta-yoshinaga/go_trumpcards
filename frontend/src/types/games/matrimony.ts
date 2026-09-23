// Type declarations for matrimony. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Matrimony game phases, matching the Go domain enum. */
export const MatrimonyPhase = { PLAYING: 0, GAME_CLEAR: 1, GAME_OVER: 2 } as const;

/** A suggested move hint in Matrimony. */
export interface MatrimonyHint {
  /** `'tableau'`, `'waste'`, or `'stock'`. */
  fromZone: string;
  /** Tableau slot, or -1 when the source is the waste or the stock. */
  fromIdx: number;
  /** `'foundation'`, `'tableau'`, or `'waste'` (for a draw). */
  toZone: string;
  /** Destination index, or -1 for a draw, which targets no single pile. */
  toIdx: number;
}

/** Full Matrimony game state returned from the API. */
export interface MatrimonyResponse extends BaseGameResponse {
  /**
   * Sixteen slots holding exactly one card each; an empty slot is `null`.
   *
   * This is a flat array, not an array of piles — a slot cannot be stacked.
   */
  tableau: (Card | null)[];
  /** Four foundations: two Q♠ descending and two J♦ ascending. */
  foundation: Card[][];
  stockCount: number;
  redealCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: MatrimonyHint;
}

/** Source or target zone for a Matrimony card move. */
export interface MatrimonyMoveZone {
  zone: 'tableau' | 'waste' | 'stock' | 'foundation';
  /** Tableau slot (0..15). */
  col?: number;
}
