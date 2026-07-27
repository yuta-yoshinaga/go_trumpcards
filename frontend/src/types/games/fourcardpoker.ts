// Type declarations for fourcardpoker. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Four Card Poker API response. */
export interface FourCardPokerResponse extends BaseGameResponse {
  /** Player's 5-card hand. */
  playerHand: Card[];
  /** Dealer hand: during the action phase only the upcard is revealed
   * (length 1); after the end phase all 6 cards are revealed. */
  dealerHand: Card[];
  /** Player's best 4-card subset (populated at end phase). */
  playerBest: Card[];
  /** Dealer's best 4-card subset (populated at end phase). */
  dealerBest: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  acesUpBet: number;
  playBet: number;
  playMultiplier: number;
  result: number;
  antePayout: number;
  playPayout: number;
  anteBonusPayout: number;
  acesUpPayout: number;
  totalPayout: number;
  playerHandRank: number;
  dealerHandRank: number;
}
