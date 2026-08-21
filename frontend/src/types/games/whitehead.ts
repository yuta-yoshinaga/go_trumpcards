// Type declarations for whitehead. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card in a Whitehead tableau column with face-up status. */
export interface WhiteheadTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Whitehead. */
export interface WhiteheadHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Whitehead game state returned from the API. */
export interface WhiteheadResponse extends BaseGameResponse {
  tableau: WhiteheadTableauCard[][];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  drawCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  score: number;
  scoringMode: number;
  hint?: WhiteheadHint;
}

// --- Canfield (キャンフィールド) ---
