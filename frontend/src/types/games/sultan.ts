// Type declarations for sultan. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Sultan of Turkey. */
export interface SultanHint {
  fromZone: string;
  fromIdx: number;
  toFoundation: number;
}

/** Full Sultan of Turkey game state returned from the API. */
export interface SultanResponse extends BaseGameResponse {
  foundation: Card[][];
  divan: (Card | null)[];
  stockCount: number;
  waste: Card[];
  redealCount: number;
  canRedeal: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: SultanHint;
}

/** Source zone for a Sultan of Turkey card move (divan slot or waste). */
export interface SultanMoveZone {
  zone: string;
  divanIdx?: number;
}

// --- Crescent (クレセント・ソリティア) ---
