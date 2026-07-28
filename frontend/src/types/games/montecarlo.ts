// Type declarations for montecarlo. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Single cell of the 5x5 Monte Carlo board. */
export interface MonteCarloBoardCell {
  /** Card on this cell, or `null` when empty (gap awaiting compression). */
  card: Card | null;
}

/** Suggested hint for Monte Carlo Solitaire. */
export interface MonteCarloHint {
  /** "remove" suggests a pair to take off; "deal" suggests pressing the Deal button. */
  action: 'remove' | 'deal';
  /** First cell of the pair (for `action === 'remove'`). */
  fromR?: number;
  fromC?: number;
  /** Second cell of the pair (for `action === 'remove'`). */
  toR?: number;
  toC?: number;
}

/** Monte Carlo Solitaire API response. */
export interface MonteCarloResponse extends BaseGameResponse {
  /** 5x5 board. Empty cells (post-removal, pre-deal) have `card === null`. */
  board: MonteCarloBoardCell[][];
  /** 0 = playing, 1 = game clear, 2 = game over. */
  phase: number;
  /** Cards remaining in the stock (52 - drawn so far). */
  stockCount: number;
  /** Cards removed from the board so far (must hit 52 to win). */
  removedCount: number;
  /** Number of times the player has pressed Deal. */
  dealCount: number;
  /** Whether the last action can be undone. */
  canUndo: boolean;
  /** True when no remove pairs exist and the stock cannot help. */
  isStalemate: boolean;
  /** Server-generated hint, present only on `/montecarlo/exec` with `command: "hint"`. */
  hint?: MonteCarloHint;
}

// --- Let It Ride (レット・イット・ライド) ---
