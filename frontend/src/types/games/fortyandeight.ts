// Type declarations for fortyandeight. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Forty and Eight with face-up/face-down state. */
export interface FortyAndEightTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Forty and Eight. */
export interface FortyAndEightHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Forty and Eight game state returned from the API. */
export interface FortyAndEightResponse extends BaseGameResponse {
  tableau: FortyAndEightTableauCard[][];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  redealUsed: boolean;
  canRedeal: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: FortyAndEightHint;
}

/** Source or target zone for a Forty and Eight card move. */
export interface FortyAndEightMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Sultan of Turkey (スルタン) ---
