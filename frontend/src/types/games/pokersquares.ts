// Type declarations for pokersquares. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Single cell of the 5x5 Poker Squares board. */
export interface PokerSquaresBoardCell {
  /** Placed card, or `null` when the cell is empty. */
  card: Card | null;
}

/** Poker Squares API response. */
export interface PokerSquaresResponse extends BaseGameResponse {
  /** 5x5 board. Empty cells have `card === null`. */
  board: PokerSquaresBoardCell[][];
  /** Next card to place, or `null` once all 25 cards have been placed. */
  currentCard: Card | null;
  /** Number of cards placed so far (0..25). */
  placedCount: number;
  /** 0 = playing, 1 = complete. */
  phase: number;
  /** Whether the last action can be undone. */
  canUndo: boolean;
  /** Score per row (length 5). */
  rowScores: number[];
  /** Score per column (length 5). */
  colScores: number[];
  /** Sum of all row and column scores. */
  totalScore: number;
  /**
   * Server-side placement hint, weighing row/column synergy.
   *
   * **The CUI had this and the Web did not (#4790).** The page's own
   * `useGameHint` heuristic does not look at synergy, so a Web player got the
   * weaker advice of the two.
   */
  hint?: PokerSquaresHint;
}

/** Server-side Poker Squares hint (sync: `domain.PokerSquaresHint`). */
export interface PokerSquaresHint {
  row: number;
  col: number;
  /** Row/column synergy score the placement would create. */
  score: number;
  /** Whether the score is positive (it works with the cards already there). */
  synergy: boolean;
}

// --- Monte Carlo Solitaire (モンテカルロ・ソリティア) ---
