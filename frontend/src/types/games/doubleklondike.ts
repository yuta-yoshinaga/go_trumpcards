// Type declarations for doubleklondike. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A tableau card in Double Klondike; face-down cards hide their value. */
export interface DoubleKlondikeTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Double Klondike. */
export interface DoubleKlondikeHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Double Klondike game state returned from the API. */
export interface DoubleKlondikeResponse extends BaseGameResponse {
  /** The 9 tableau columns (top card is last). */
  tableau: DoubleKlondikeTableauCard[][];
  /** Cards left in the stock. */
  stockCount: number;
  /** The waste pile (top card is last). */
  waste: Card[];
  /** The 8 foundations (A-K by suit, two per suit). */
  foundation: Card[][];
  /** Current phase (0=Playing, 1=GameClear, 2=GameOver). */
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  hint?: DoubleKlondikeHint;
}

// --- Black Hole (ブラックホール) ---
