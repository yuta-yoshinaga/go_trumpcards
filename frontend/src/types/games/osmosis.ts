// Type declarations for osmosis. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Osmosis. */
export interface OsmosisHint {
  fromZone: string;
  fromCol: number;
  toCol: number;
}

/** Full Osmosis game state returned from the API. */
export interface OsmosisResponse extends BaseGameResponse {
  reserve: Card[][];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  baseRank: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  /** Whether no card anywhere can still reach a foundation (#4808). */
  isStalemate: boolean;
  hint?: OsmosisHint;
}

// --- Bristol (ブリストル) ---
