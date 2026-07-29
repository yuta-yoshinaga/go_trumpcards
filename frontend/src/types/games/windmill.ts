// Type declarations for windmill. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Windmill. */
export interface WindmillHint {
  /** `'sail'`, `'waste'`, `'corner'`, or `'stock'` (meaning: turn a card). */
  fromZone: string;
  /** Sail or corner index, or -1 when the source is the waste or the stock. */
  fromIdx: number;
  /** `'center'`, `'corner'`, or `'waste'` (for a draw). */
  toZone: string;
  /** Destination index, or -1 when the destination is not a corner. */
  toIdx: number;
}

/** Full Windmill game state returned from the API. */
export interface WindmillResponse extends BaseGameResponse {
  /** Eight fixed slots. A slot that can no longer be refilled stays `null`. */
  sails: (Card | null)[];
  /** The centre foundation: A-K four times through, 52 cards when complete. */
  center: Card[];
  /** Four corner foundations, each King down to Ace. */
  corners: Card[][];
  stockCount: number;
  waste: Card[];
  /** While true, pulling another corner card back onto the centre is refused. */
  transferBlocked: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: WindmillHint;
}

/** Source or target zone for a Windmill card move. */
export interface WindmillMoveZone {
  zone: string;
  col?: number;
}
