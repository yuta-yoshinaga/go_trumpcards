// Type declarations for duchess. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Duchess. */
export interface DuchessTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Duchess. */
export interface DuchessHint {
  /** `'reserve'`, `'waste'`, `'tableau'`, or `'stock'` (meaning: turn a card). */
  fromZone: string;
  /** Reserve fan or tableau column, or -1 when the source is the waste or stock. */
  fromIdx: number;
  /** Head of the run to carry, or -1 when the source is not the tableau. */
  cardIndex: number;
  /** `'foundation'`, `'tableau'`, or `'waste'` (for a draw). */
  toZone: string;
  /** Destination index, or -1 for a draw, which targets no single pile. */
  toIdx: number;
}

/** Full Duchess game state returned from the API. */
export interface DuchessResponse extends BaseGameResponse {
  /** Four reserve fans of three cards; only the top of each is available. */
  reserve: Card[][];
  tableau: DuchessTableauCard[][];
  /** Four foundations, one per suit, all starting at {@link DuchessResponse.baseRank}. */
  foundation: Card[][];
  stockCount: number;
  waste: Card[];
  /** The rank all four foundations start from. 0 means it has not been chosen yet. */
  baseRank: number;
  /** While true, picking the base rank off a reserve fan is the only legal action. */
  awaitingBaseRank: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  /**
   * Whether auto-complete can move at least one card right now.
   *
   * The page used to guess this from foundation pile heights, which is not the
   * domain's condition: Duchess seeds no foundations, so one card can already be
   * enough, and a tall pile with nothing to feed it still fails (#5557).
   */
  canAutoComplete?: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: DuchessHint;
}

/** Source or target zone for a Duchess card move. */
export interface DuchessMoveZone {
  zone: string;
  col?: number;
  /** Head of the run to carry; omit to move only the top card. */
  cardIndex?: number;
}
