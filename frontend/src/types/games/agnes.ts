// Type declarations for agnes. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single card on an Agnes Sorel tableau column (carries face-up state). */
export interface AgnesTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Agnes Sorel. */
export interface AgnesHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Agnes Sorel game state returned from the API. */
export interface AgnesResponse extends BaseGameResponse {
  tableau: AgnesTableauCard[][];
  stockCount: number;
  foundation: Card[][];
  baseRank: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: AgnesHint;
}

// --- Osmosis (オズモシス / 浸透) ---
