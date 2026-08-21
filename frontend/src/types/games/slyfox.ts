// Type declarations for slyfox. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Sly Fox. */
export interface SlyFoxHint {
  /** `'tableau'` for a reserve slot, or `'stock'` to deal the next card. */
  fromZone: string;
  /** Reserve slot index, or -1 when the source is the stock. */
  fromIdx: number;
  /** `'foundation'`, or `'tableau'` when the advice is where to deal. */
  toZone: string;
  /** Destination index. */
  toIdx: number;
}

/** Full Sly Fox game state returned from the API. */
export interface SlyFoxResponse extends BaseGameResponse {
  /** Twenty reserve slots; only the top card of each is available. */
  tableau: Card[][];
  foundation: Card[][];
  /**
   * Per-foundation build direction: `true` builds up from the Ace, `false`
   * builds down from the King. Sent explicitly rather than derived from the
   * index so a reordering cannot silently mislabel the piles.
   */
  foundationAscending: boolean[];
  /** Cards left to deal. There is no waste — a dealt card is placed at once. */
  stockCount: number;
  /**
   * Cards dealt onto the reserve this round. A card sent straight to a
   * foundation is not counted, so a lucky draw does not spend the round.
   */
  dealtThisCycle: number;
  /** How many cards make one round (20). */
  dealCycle: number;
  /**
   * Whether the reserve still cannot feed the foundations. The UI reads this
   * to disable the reserve instead of sending a move the server will refuse.
   */
  reserveLocked: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: SlyFoxHint;
}

/** Source or target zone for a Sly Fox card move or deal. */
export interface SlyFoxMoveZone {
  zone: 'tableau' | 'foundation';
  idx?: number;
}
