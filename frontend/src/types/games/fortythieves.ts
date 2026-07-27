// Type declarations for fortythieves. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Forty Thieves with face-up/face-down state. */
export interface FortyThievesTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Forty Thieves. */
export interface FortyThievesHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Forty Thieves game state returned from the API. */
export interface FortyThievesResponse extends BaseGameResponse {
  tableau: FortyThievesTableauCard[][];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: FortyThievesHint;
}

/** Source or target zone for a Forty Thieves card move. */
export interface FortyThievesMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Forty and Eight (フォーティ・アンド・エイト) ---
