// Type declarations for curdsandwhey. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Simple Simon. */
export interface CurdsAndWheyHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** Full Simple Simon game state returned from the API. */
export interface CurdsAndWheyResponse extends BaseGameResponse {
  /** The 10 tableau columns (top card is last). */
  columns: Card[][];
  /** Number of complete K-A suits removed (0-4). */
  completedSuits: number;
  /** Current phase (0=Playing, 1=GameClear, 2=GameOver). */
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: CurdsAndWheyHint;
}

// --- Double Klondike (ダブル・クロンダイク) ---
