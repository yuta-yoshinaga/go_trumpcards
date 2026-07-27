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
}

// --- Monte Carlo Solitaire (モンテカルロ・ソリティア) ---
