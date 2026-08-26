// Type declarations for bigben. Split-file layout introduced by
// issue #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Big Ben. */
export interface BigBenTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** One clock face: its cards plus the rank it has to end on. */
export interface BigBenFoundation {
  cards: Card[];
  /**
   * The rank this face must reach. Index 0 is **nine o'clock** and wants a 9;
   * the hours run 9, 10, 11, 12, 1, 2, … 8 clockwise from there.
   */
  targetRank: number;
  /** Whether the face has reached `targetRank` and accepts nothing more. */
  complete: boolean;
}

/** A suggested move hint in Big Ben. */
export interface BigBenHint {
  /** `'tableau'`, or `'stock'` when the advice is to deal. */
  fromZone: string;
  /** Source column, or -1 when the advice is to deal. */
  fromCol: number;
  /** `'foundation'`, `'tableau'`, or `'stock'` (deal — no destination). */
  toZone: string;
  /** Destination clock face or column index, or -1 for a deal. */
  toIdx: number;
}

/** Full Big Ben game state returned from the API. */
export interface BigBenResponse extends BaseGameResponse {
  /** Eight columns building down in suit. An empty one is never refilled. */
  tableau: BigBenTableauCard[][];
  /** Twelve clock faces, index 0 at nine o'clock through index 11 at eight. */
  foundation: BigBenFoundation[];
  /** Cards left in the stock. Dealing tops every column back up to three. */
  stockCount: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: BigBenHint;
}

/** Source or target zone for a Big Ben card move. */
export interface BigBenMoveZone {
  zone: string;
  col?: number;
}
