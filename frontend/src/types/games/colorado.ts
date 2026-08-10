// Type declarations for colorado. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Colorado. */
export interface ColoradoHint {
  fromZone: string;
  /** Tableau pile index, or -1 when the source is the waste or the stock. */
  fromIdx: number;
  toZone: string;
  /** Destination index, or -1 when the destination is the waste. */
  toIdx: number;
}

/** Full Colorado game state returned from the API. */
export interface ColoradoResponse extends BaseGameResponse {
  tableau: Card[][];
  foundation: Card[][];
  /**
   * Per-foundation build direction: `true` builds up from the Ace, `false`
   * builds down from the King. Sent explicitly rather than derived from the
   * index so a reordering cannot silently mislabel the piles.
   */
  foundationAscending: boolean[];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: ColoradoHint;
}

/** Source or target zone for a Colorado card move. */
export interface ColoradoMoveZone {
  zone: 'waste' | 'tableau' | 'foundation' | 'stock';
  idx?: number;
}
