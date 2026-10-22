// Type declarations for curdsandwhey. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Curds and Whey. */
export interface CurdsAndWheyHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** Full Curds and Whey game state returned from the API. */
export interface CurdsAndWheyResponse extends BaseGameResponse {
  /** Auto-move destinations keyed by `tableau-<col>-<cardIndex>` (run starts only); `tableau` is in priority order, so `tableau[0]` is the preferred destination, and `foundation` is always empty. */
  legalTargets: Record<string, { tableau: number[]; foundation: number[] }>;
  /** The 13 tableau columns (top card is last). */
  columns: Card[][];
  /** Number of complete K-A suits removed (0-4). */
  completedSuits: number;
  /** Current phase (0=Playing, 1=GameClear, 2=GameOver). */
  phase: number;
  moveCount: number;
  canUndo: boolean;
  /** Start indices of movable runs, one list per tableau column. */
  runStarts: number[][];
  hint?: CurdsAndWheyHint;
}

// --- Double Klondike (ダブル・クロンダイク) ---
