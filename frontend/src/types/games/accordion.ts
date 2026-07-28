// Type declarations for accordion. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single pile in Accordion. Only the top card is revealed; size tracks stacked depth. */
export interface AccordionPile {
  cards: Card[];
  size: number;
}

/** A suggested move hint in Accordion. */
export interface AccordionHint {
  fromIdx: number;
  toIdx: number;
}

/** API response shape for an Accordion game. */
export interface AccordionResponse extends BaseGameResponse {
  piles: AccordionPile[];
  pileCount: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: AccordionHint;
}

// --- Trash (トラッシュ) ---
