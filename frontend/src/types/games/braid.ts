// Type declarations for braid. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Braid. */
export interface BraidHint {
  /** `'direction'`, `'braid'`, `'field'`, `'helper'`, `'waste'`, or `'stock'`. */
  fromZone: string;
  /** Slot index for `'field'` / `'helper'`, otherwise -1. */
  fromIdx: number;
  /** `'foundation'`, `'helper'`, or `'waste'` (for a draw). */
  toZone: string;
  /** Destination index, or -1 when the move targets no single slot. */
  toIdx: number;
}

/** Full Braid game state returned from the API. */
export interface BraidResponse extends BaseGameResponse {
  /** The 20-card braid. Only its last entry is available, and only to a foundation. */
  braid: Card[];
  /**
   * Four slots that refill from the braid's tail — the only thing that
   * consumes the braid. An empty slot is `null` rather than omitted, so the
   * index keeps matching the one a hint refers to.
   */
  fields: (Card | null)[];
  /** Eight slots that can only be filled from the waste. Empty slots are `null`. */
  helpers: (Card | null)[];
  /** Eight foundations built IN SUIT from {@link BraidResponse.baseRank}. */
  foundation: Card[][];
  stockCount: number;
  waste: Card[];
  /** The rank all eight foundations start from, fixed when the game is dealt. */
  baseRank: number;
  /** 0 = not chosen yet, 1 = ascending, 2 = descending. */
  direction: number;
  /** While true, no card may reach a foundation until a direction is chosen. */
  awaitingDirection: boolean;
  /** Redeals still available (starts at 2). */
  redealsLeft: number;
  canRedeal: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: BraidHint;
}

/** Source or target zone for a Braid card move. */
export interface BraidMoveZone {
  zone: string;
  col?: number;
}
