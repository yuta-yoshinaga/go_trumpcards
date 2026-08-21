// Type declarations for sthelena. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in St. Helena (always face-up). */
export interface StHelenaTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in St. Helena. */
export interface StHelenaHint {
  fromCol: number;
  toZone: string;
  toCol: number;
  redeal: boolean;
}

/** Full St. Helena game state returned from the API. */
export interface StHelenaResponse extends BaseGameResponse {
  /** Twelve columns circling the foundations. An empty one cannot be refilled. */
  tableau: StHelenaTableauCard[][];
  /** Eight foundations: 0..3 build up from the aces, 4..7 build down from the kings. */
  foundation: Card[][];
  redealsRemaining: number;
  /**
   * Whether the first-deal reachability restriction is still in force: the top
   * four columns reach only the king row, the bottom four only the ace row, and
   * the four side columns either. The first redeal lifts it.
   */
  restrictionsActive: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: StHelenaHint;
}

/** Source or target zone for a St. Helena card move. */
export interface StHelenaMoveZone {
  zone: 'tableau' | 'foundation';
  col?: number;
}
