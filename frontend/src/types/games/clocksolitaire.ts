// Type declarations for clocksolitaire. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card in a Clock Solitaire pile with face-up status. */
export interface ClockSolitaireCard {
  card: Card | null;
  faceUp: boolean;
}

/** Full Clock Solitaire game state returned from the API. */
export interface ClockSolitaireResponse extends BaseGameResponse {
  piles: ClockSolitaireCard[][];
  faceUpCount: number[];
  phase: number;
  stepCount: number;
  currentCard?: Card;
  canUndo?: boolean;
}
