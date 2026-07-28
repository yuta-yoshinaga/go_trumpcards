// Type declarations for threecard. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Three Card Poker API response. */
export interface ThreeCardResponse extends BaseGameResponse {
  playerHand: Card[];
  dealerHand: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  pairPlusBet: number;
  playBet: number;
  result: number;
  antePayout: number;
  playPayout: number;
  anteBonusPayout: number;
  pairPlusPayout: number;
  totalPayout: number;
  dealerQualified: boolean;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Caribbean Stud Poker (カリビアンスタッドポーカー) ---
