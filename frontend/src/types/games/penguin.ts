// Type declarations for penguin. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Penguin. */
export interface PenguinHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Penguin game state returned from the API. */
export interface PenguinResponse extends BaseGameResponse {
  tableau: (Card | null)[][];
  freeCells: (Card | null)[];
  foundation: Card[][];
  baseRank: number;
  /** How many cards may move as one stack right now (from the domain). */
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
  hint?: PenguinHint;
}

// --- Seahaven Towers (シーヘイブンタワーズ) ---
