// Type declarations for stalactites. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Stalactites. */
export interface StalactitesHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Stalactites game state returned from the API. */
export interface StalactitesResponse extends BaseGameResponse {
  tableau: (Card | null)[][];
  cells: (Card | null)[];
  foundation: Card[][];
  /** How many cards may move as one stack right now (from the domain). */
  /** Foundation base rank for this deal (Stalactites is not Ace-based). */
  baseRank: number;
  maxMovableCards: number;
  /**
   * The same limit when the destination is an empty column -- lower, because
   * that column cannot also serve as a staging slot. 0 when there is none.
   */
  maxMovableCardsToEmptyColumn: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: StalactitesHint;
}

// --- Eight Off (エイトオフ) ---
