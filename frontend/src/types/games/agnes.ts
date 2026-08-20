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
  /**
   * Whether no legal move remains. Decided by the domain's `Agnes.IsStalemate()`
   * -- the page used to re-derive it from the move rules in TypeScript, so the
   * same rule lived in two places and could drift (#5601).
   */
  isStalemate: boolean;
  hint?: AgnesHint;
}

// --- Osmosis (オズモシス / 浸透) ---
