// Type declarations for scorpion. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse } from '../common';
import type { KlondikeTableauCard } from './klondike';

/** A suggested move hint in Scorpion. */
export interface ScorpionHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** API response shape for a Scorpion game. */
export interface ScorpionResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  stockCount: number;
  completedSuits: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: ScorpionHint;
}

// --- Wasp (ワスプ) ---
