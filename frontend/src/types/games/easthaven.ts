// Type declarations for easthaven. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';
import type { KlondikeTableauCard } from './klondike';

/** A suggested move hint in Easthaven. */
export interface EasthavenHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** API response shape for an Easthaven game. */
export interface EasthavenResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  foundation: Card[][];
  stockCount: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: EasthavenHint;
}

// --- Accordion (アコーディオン) ---
