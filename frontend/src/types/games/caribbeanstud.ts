// Type declarations for caribbeanstud. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, MaskedCard } from '../common';

/** Caribbean Stud Poker API response. */
export interface CaribbeanStudResponse extends BaseGameResponse {
  playerHand: Card[];
  /** Dealer hand: during the action phase only the first card is revealed and
   * the remaining slots are `MaskedCard`. After the end phase all 5 are real `Card`s. */
  dealerHand: (Card | MaskedCard)[];
  phase: number;
  chips: number;
  anteBet: number;
  jackpotBet: number;
  playBet: number;
  result: number;
  antePayout: number;
  playPayout: number;
  jackpotPayout: number;
  totalPayout: number;
  dealerQualified: boolean;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Casino Hold'em (カジノホールデム) ---
