// Type declarations for crazyquilt. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Crazy Quilt. */
export interface CrazyQuiltHint {
  /** `'quilt'`, `'waste'`, or `'stock'`. */
  fromZone: string;
  /** Quilt cell index, or -1 when the source is the waste or the stock. */
  fromIdx: number;
  /** `'foundation'` or `'waste'`. */
  toZone: string;
  /** Destination index, or -1 when the destination is the waste. */
  toIdx: number;
}

/** Full Crazy Quilt game state returned from the API. */
export interface CrazyQuiltResponse extends BaseGameResponse {
  /**
   * The 64 quilt cells in row-major order (`row * 8 + col`). A cell whose card
   * has been taken is `null`.
   */
  quilt: (Card | null)[];
  /**
   * Per-cell availability, same order as `quilt`.
   *
   * A card is available only when one of its **short** sides is exposed, which
   * depends on whether the cell is laid vertically or horizontally — an open
   * long side frees nothing. The server computes it so the rule lives in one
   * place; do not re-derive it from the layout here.
   */
  available: boolean[];
  /** Eight foundations, seeded with an Ace or a King of each suit before the deal. */
  foundation: Card[][];
  /** Per-foundation direction: `true` builds up from the Ace, `false` down from the King. */
  foundationAscending: boolean[];
  stockCount: number;
  /** Redeals remaining. The waste is turned over without shuffling. */
  redealsLeft: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: CrazyQuiltHint;
}

/** Source or target zone for a Crazy Quilt card move. */
export interface CrazyQuiltMoveZone {
  zone: 'quilt' | 'waste' | 'foundation';
  /** Quilt cell index (0..63). */
  col?: number;
}
