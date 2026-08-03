// Type declarations for missmilligan. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Miss Milligan. */
export interface MissMilliganTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Miss Milligan. */
export interface MissMilliganHint {
  /** `'tableau'`, `'waived'`, or `'stock'` (meaning: deal a row). */
  fromZone: string;
  /** Source column, or -1 when the source is the waived set or the stock. */
  fromCol: number;
  /** Head of the run to carry, or -1 when the source is not the tableau. */
  cardIndex: number;
  /** `'foundation'` or `'tableau'`. */
  toZone: string;
  /** Destination index, or -1 for a deal, which targets no single column. */
  toIdx: number;
}

/** Full Miss Milligan game state returned from the API. */
export interface MissMilliganResponse extends BaseGameResponse {
  tableau: MissMilliganTableauCard[][];
  stockCount: number;
  /** Eight foundations, two per suit, opened by Aces. */
  foundation: Card[][];
  /** Cards held aside by waiving. Nothing may be dealt while this is non-empty. */
  waived: Card[];
  /** Whether waiving is available: the stock is gone and nothing is held. */
  canWaive: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: MissMilliganHint;
}

/** Source or target zone for a Miss Milligan card move. */
export interface MissMilliganMoveZone {
  zone: string;
  col?: number;
  /** Head of the run to carry; omit to move only the top card. */
  cardIndex?: number;
}
