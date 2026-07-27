// Type declarations for casinoholdem. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, MaskedCard } from '../common';

/** Casino Hold'em API response. */
export interface CasinoHoldemResponse extends BaseGameResponse {
  /** Player's two hole cards. */
  playerHand: Card[];
  /** Dealer's hole cards: masked as `MaskedCard` until the showdown (only after a call). */
  dealerHand: (Card | MaskedCard)[];
  /** Community cards (flop / turn / river). Length grows from 3 (flop) → 5 (showdown). */
  community: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  /** AA Bonus side bet wager. */
  bonusBet: number;
  /** Call bet placed at flop (2× ante). 0 if folded. */
  callBet: number;
  result: number;
  /** Whether the dealer qualified (Pair of Fours or better). */
  dealerQualify: boolean;
  antePayout: number;
  callPayout: number;
  bonusPayout: number;
  totalPayout: number;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Texas Hold'em Bonus Poker (テキサスホールデムボーナスポーカー) ---
