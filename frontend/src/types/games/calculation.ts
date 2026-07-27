// Type declarations for calculation. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Calculation. */
export interface CalculationHint {
  fromZone: string;
  wasteIdx: number;
  foundationIdx: number;
}

/** Full Calculation game state returned from the API. */
export interface CalculationResponse extends BaseGameResponse {
  foundations: Card[][];
  wastes: Card[][];
  stockCount: number;
  stockTop?: Card;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: CalculationHint;
}

/** Source or target zone for a Calculation card move. */
export interface CalculationMoveZone {
  zone: 'stock' | 'waste' | 'foundation';
  idx?: number;
}
