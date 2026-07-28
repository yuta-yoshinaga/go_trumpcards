// Type declarations for letitride. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, MaskedCard } from '../common';

/** Let It Ride API response. */
export interface LetItRideResponse extends BaseGameResponse {
  playerHand: Card[];
  /** Community cards: masked as `MaskedCard` until revealed by phase progression. */
  communityCards: (Card | MaskedCard)[];
  phase: number;
  chips: number;
  betAmount: number;
  bet1Active: boolean;
  bet2Active: boolean;
  bet3Active: boolean;
  result: number;
  handRank: number;
  bet1Payout: number;
  bet2Payout: number;
  bet3Payout: number;
  totalPayout: number;
}

// --- Red Dog (レッドドッグ) ---
