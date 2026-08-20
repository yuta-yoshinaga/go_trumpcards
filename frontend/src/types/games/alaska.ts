// Type declarations for alaska. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';
import type { KlondikeTableauCard } from './klondike';

/** A suggested move hint in Russian Solitaire. */
export interface AlaskaHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** API response shape for a Russian Solitaire game. */
export interface AlaskaResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: AlaskaHint;
}

// --- Cruel (クルーエル) ---
