// Type declarations for yukon. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';
import type { KlondikeTableauCard } from './klondike';

/** A suggested move hint in Yukon. */
export interface YukonHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** API response shape for a Yukon game. */
export interface YukonResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: YukonHint;
}

// --- Russian Solitaire (ロシアンソリティア) ---
