// Type declarations for willothewisp. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Hint returned by the WillOTheWisp /hint endpoint. */
export interface WillOTheWispHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** Phase constants returned by the Will o' the Wisp API. */
export const WillOTheWispPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Tableau card with face-up state in WillOTheWisp. */
export interface WillOTheWispTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** Full WillOTheWisp Solitaire game state returned from the API. */
export interface WillOTheWispResponse extends BaseGameResponse {
  tableau: WillOTheWispTableauCard[][];
  stockCount: number;
  completedSuits: number;
  score: number;
  /**
   * How the score moves: where it starts, what a move costs, what a completed
   * suit pays. Sent so the explanation cannot quote figures the game stopped
   * using.
   */
  scoring: { start: number; movePenalty: number; suitBonus: number };
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: WillOTheWispHint;
}

// --- Indian Poker (インディアンポーカー) ---
