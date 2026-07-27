// Type declarations for highcardflush. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** High Card Flush API response. */
export interface HighCardFlushResponse extends BaseGameResponse {
  playerHand: Card[];
  dealerHand: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  flushBonusBet: number;
  straightFlushBet: number;
  raiseBet: number;
  result: number;
  antePayout: number;
  raisePayout: number;
  flushBonusPayout: number;
  straightFlushPayout: number;
  totalPayout: number;
  dealerQualified: boolean;
  playerFlushLen: number;
  dealerFlushLen: number;
  playerStraightFlushLen: number;
  maxRaiseMultiplier: number;
}

// --- Gaps / Montana (ギャップス) ---
