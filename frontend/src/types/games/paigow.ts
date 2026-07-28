// Type declarations for paigow. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Pai Gow Poker API response. */
export interface PaiGowResponse extends BaseGameResponse {
  playerCards: Card[];
  dealerCards: Card[];
  playerHighHand: Card[];
  playerLowHand: Card[];
  dealerHighHand: Card[];
  dealerLowHand: Card[];
  phase: number;
  chips: number;
  bet: number;
  result: number;
  highHandResult: number;
  lowHandResult: number;
  payout: number;
  commission: number;
  playerHighRank: number;
  playerLowRank: number;
  dealerHighRank: number;
  dealerLowRank: number;
}
