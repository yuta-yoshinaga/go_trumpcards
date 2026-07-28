// Type declarations for russianpoker. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, MaskedCard } from '../common';

/** Russian Poker game state from the /russianpoker/exec endpoint. */
export interface RussianPokerResponse extends BaseGameResponse {
  playerHand: Card[];
  dealerHand: (Card | MaskedCard)[];
  phase: number;
  chips: number;
  anteBet: number;
  exchangeCount: number;
  exchangeFee: number;
  bought6th: boolean;
  buy6thFee: number;
  forceExchanged: boolean;
  forceExchangeFee: number;
  playBet: number;
  result: number;
  antePayout: number;
  playPayout: number;
  totalPayout: number;
  dealerQualified: boolean;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Beleaguered Castle (包囲された城) ---
