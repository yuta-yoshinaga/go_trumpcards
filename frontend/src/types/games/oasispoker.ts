// Type declarations for oasispoker. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, MaskedCard } from '../common';

/** Oasis Poker API response. */
export interface OasisPokerResponse extends BaseGameResponse {
  playerHand: Card[];
  /** Dealer hand: during bet/exchange/action phases only the first card is revealed and
   * the remaining slots are `MaskedCard`. After the end phase all 5 are real `Card`s. */
  dealerHand: (Card | MaskedCard)[];
  phase: number;
  chips: number;
  anteBet: number;
  jackpotBet: number;
  /** Number of cards exchanged this round (0..5). */
  exchangeCount: number;
  /** Fee collected for exchanging cards (ante × exchangeCount). */
  exchangeFee: number;
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

// --- Russian Poker (ロシアンポーカー) ---
