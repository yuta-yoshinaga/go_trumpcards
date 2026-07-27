// Type declarations for klondike. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card in a Klondike tableau column with face-up status. */
export interface KlondikeTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Klondike. */
export interface KlondikeHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Klondike game state returned from the API. */
export interface KlondikeResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
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
  hint?: KlondikeHint;
}

// --- Canfield (キャンフィールド) ---
