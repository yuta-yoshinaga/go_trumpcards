// Type declarations for fourteenout. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single card in a Fourteen Out column. */
export interface FourteenOutBoardCell {
  /** The card at this position. */
  card: Card | null;
}

/** Suggested hint for Fourteen Out. */
export interface FourteenOutHint {
  /** Always "remove" — this game has no stock and therefore no deal action. */
  action: 'remove';
  /** First column of the recommended pair. */
  fromCol: number;
  /** Second column of the recommended pair. */
  toCol: number;
}

/** Fourteen Out API response. */
export interface FourteenOutResponse extends BaseGameResponse {
  /**
   * The 12 columns, left to right. Only the LAST card of each is in play;
   * a cleared column is an empty array. Column lengths differ — the first four
   * are dealt five cards and the rest four — so this is not a fixed grid.
   */
  columns: FourteenOutBoardCell[][];
  /** 0 = playing, 1 = game clear, 2 = game over. */
  phase: number;
  /** Cards removed so far (must hit 52 to win). */
  removedCount: number;
  /** How many pairs summing to 14 are currently available. */
  removablePairs: number;
  /** Whether the last action can be undone. */
  canUndo: boolean;
  /** True when no two exposed cards sum to 14. With no stock, this is a loss. */
  isStalemate: boolean;
  /** Server-generated hint, present only on `/fourteenout/exec` with `command: "hint"`. */
  hint?: FourteenOutHint;
}

// --- Let It Ride (レット・イット・ライド) ---
