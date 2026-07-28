// Type declarations for seahaventowers. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Seahaven Towers. */
export interface SeahavenTowersHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Seahaven Towers game state returned from the API. */
export interface SeahavenTowersResponse extends BaseGameResponse {
  tableau: (Card | null)[][];
  reservedCells: (Card | null)[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: SeahavenTowersHint;
}
