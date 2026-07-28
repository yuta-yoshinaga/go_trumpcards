// Type declarations for canfield. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single card on a Canfield tableau column. */
export interface CanfieldTableauCard {
  card: Card;
}

/** A suggested move hint in Canfield. */
export interface CanfieldHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Canfield game state returned from the API. */
export interface CanfieldResponse extends BaseGameResponse {
  tableau: CanfieldTableauCard[][];
  reserve: Card[];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  baseRank: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: CanfieldHint;
}

// --- Agnes Sorel (アグネス・ソレル) ---
