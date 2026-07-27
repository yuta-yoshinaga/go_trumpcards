// Type declarations for pyramid. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card in the pyramid with removal and exposure status. */
export interface PyramidCard {
  card: Card | null;
  removed: boolean;
  exposed: boolean;
}

/** A suggested pair/king removal hint in Pyramid. */
export interface PyramidHint {
  type: string;
  row1: number;
  col1: number;
  row2: number;
  col2: number;
}

/** Full Pyramid game state returned from the API. */
export interface PyramidResponse extends BaseGameResponse {
  pyramid: PyramidCard[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: PyramidHint;
}

// --- TriPeaks (トリピークス) ---
