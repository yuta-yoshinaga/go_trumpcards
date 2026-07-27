// Type declarations for tripeaks. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card in the TriPeaks tableau with removal and exposure status. */
export interface TriPeaksCard {
  card: Card | null;
  removed: boolean;
  exposed: boolean;
}

/** A suggested hint in TriPeaks. */
export interface TriPeaksHint {
  type: string;
  row: number;
  col: number;
}

/** Full TriPeaks game state returned from the API. */
export interface TriPeaksResponse extends BaseGameResponse {
  layout: TriPeaksCard[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: TriPeaksHint;
}
