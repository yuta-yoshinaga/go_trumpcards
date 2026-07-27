// Type declarations for cruel. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';
import type { KlondikeTableauCard } from './klondike';

/** A suggested move hint in Cruel. */
export interface CruelHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** API response shape for a Cruel game. */
export interface CruelResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: CruelHint;
}

// --- Scorpion (スコーピオン) ---
