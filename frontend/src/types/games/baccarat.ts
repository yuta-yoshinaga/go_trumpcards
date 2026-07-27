// Type declarations for baccarat. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Result of a Baccarat side bet (player pair, banker pair). */
export interface BaccaratSideBetResult {
  betType: number;
  resultType: number;
  resultName: string;
  betAmount: number;
  payout: number;
}

/** Full Baccarat game state returned from the API. */
export interface BaccaratResponse extends BaseGameResponse {
  playerHand: Card[];
  bankerHand: Card[];
  playerHandValue: number;
  bankerHandValue: number;
  phase: number;
  chips: number;
  betAmount: number;
  betType: number;
  result: number;
  payout: number;
  history: number[];
  playerPairBet: number;
  bankerPairBet: number;
  sideBetResults: BaccaratSideBetResult[];
}

// --- Napoleon (ナポレオン) ---
