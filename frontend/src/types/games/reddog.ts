// Type declarations for reddog. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Red Dog API response. */
export interface RedDogResponse extends BaseGameResponse {
  /** Initial 2 cards. */
  initialCards: Card[];
  /** Third card revealed at end (or after raise/stay). */
  thirdCard?: Card;
  phase: number;
  chips: number;
  ante: number;
  raise: number;
  /** Spread = |rank2 - rank1| - 1, 0 when consecutive or pair. */
  spread: number;
  result: number;
  totalPayout: number;
}

// --- Casino War (カジノウォー) ---
