// Type declarations for caribbeandraw. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, MaskedCard } from '../common';

/** Caribbean Draw Poker API response. */
export interface CaribbeanDrawResponse extends BaseGameResponse {
  playerHand: Card[];
  /** Dealer hand: until the end phase only the first card is revealed and the
   * remaining slots are `MaskedCard`. After the end phase all 5 are real `Card`s. */
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
  /**
   * Fee paid for exchanging cards this round; 0 when the player stood pat.
   *
   * It is a cost, not a payout, so it never appears inside `totalPayout` —
   * without it on screen the chip count drops by an amount the result panel
   * cannot explain.
   */
  drawCost: number;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Casino Hold'em (カジノホールデム) ---
