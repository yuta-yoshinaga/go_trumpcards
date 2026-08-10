// Type declarations for cribbagesquares. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Single cell of the 4x4 Cribbage Squares board. */
export interface CribbageSquaresBoardCell {
  /** Placed card, or `null` when the cell is empty. */
  card: Card | null;
}

/** One hand's cribbage score breakdown (sync: `domain.CribbageScoreDetail`). */
export interface CribbageSquaresScore {
  fifteens: number;
  pairs: number;
  runs: number;
  flush: number;
  nobs: number;
  total: number;
}

/** Server-side Cribbage Squares hint (sync: `domain.CribbageSquaresHint`). */
export interface CribbageSquaresHint {
  row: number;
  col: number;
  /** Points the placement adds to its row and column. */
  score: number;
  /** Whether the score is positive (it works with the cards already there). */
  synergy: boolean;
}

/** Cribbage Squares API response. */
export interface CribbageSquaresResponse extends BaseGameResponse {
  /** 4x4 board. Empty cells have `card === null`. */
  board: CribbageSquaresBoardCell[][];
  /** Next card to place, or `null` once all 16 cards have been placed. */
  currentCard: Card | null;
  /**
   * The 17th card, turned only once the grid is full.
   *
   * `null` for the whole of the playing phase — every hand is built without
   * knowing its own fifth card, which is the point of the game.
   */
  starter: Card | null;
  /** Number of cards placed so far (0..16). */
  placedCount: number;
  /** 0 = playing, 1 = complete. */
  phase: number;
  /** Whether the last action can be undone. */
  canUndo: boolean;
  /** Score per row (length 4). Zero until the starter is turned. */
  rowScores: number[];
  /** Score per column (length 4). Zero until the starter is turned. */
  colScores: number[];
  /** Per-row breakdown, same order as `rowScores`. */
  rowDetails: CribbageSquaresScore[];
  /** Per-column breakdown, same order as `colScores`. */
  colDetails: CribbageSquaresScore[];
  /** Sum of all row and column scores. */
  totalScore: number;
  /** Total needed to clear the game (61). Sent so the page holds no copy. */
  winScore: number;
  /** Whether `totalScore` reached `winScore`. */
  isWin: boolean;
  hint?: CribbageSquaresHint;
}
