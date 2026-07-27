// Type declarations for wasp. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse } from '../common';
import type { KlondikeTableauCard } from './klondike';

/** A suggested move hint in Wasp. */
export interface WaspHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** API response shape for a Wasp game. */
export interface WaspResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  stockCount: number;
  completedSuits: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: WaspHint;
}

// --- Easthaven (イーストヘイブン) ---
