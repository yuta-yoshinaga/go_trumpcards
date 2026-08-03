// Type declarations for sirtommy. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in SirTommy. */
export interface SirTommyHint {
  fromZone: string;
  wasteIdx: number;
  foundationIdx: number;
}

/** Full SirTommy game state returned from the API. */
export interface SirTommyResponse extends BaseGameResponse {
  foundations: Card[][];
  wastes: Card[][];
  stockCount: number;
  stockTop?: Card;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: SirTommyHint;
}

/** Source or target zone for a SirTommy card move. */
export interface SirTommyMoveZone {
  zone: 'stock' | 'waste' | 'foundation';
  idx?: number;
}
