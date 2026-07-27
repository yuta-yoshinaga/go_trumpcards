// Type declarations for crescent. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Crescent (always face-up). */
export interface CrescentTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Crescent. */
export interface CrescentHint {
  fromCol: number;
  toZone: string;
  toCol: number;
  redeal: boolean;
}

/** Full Crescent game state returned from the API. */
export interface CrescentResponse extends BaseGameResponse {
  tableau: CrescentTableauCard[][];
  foundation: Card[][];
  redealsRemaining: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: CrescentHint;
}

/** Source or target zone for a Crescent card move. */
export interface CrescentMoveZone {
  zone: 'tableau' | 'foundation';
  col?: number;
}

// --- Baker's Dozen (ベーカーズ・ダズン) ---
