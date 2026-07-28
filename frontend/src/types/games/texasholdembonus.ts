// Type declarations for texasholdembonus. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, MaskedCard } from '../common';

/** Texas Hold'em Bonus Poker API response. */
export interface TexasHoldemBonusResponse extends BaseGameResponse {
  /** Player's two hole cards. */
  playerHand: Card[];
  /** Dealer's hole cards: masked as `MaskedCard` until the showdown. */
  dealerHand: (Card | MaskedCard)[];
  /** Community cards (flop / turn / river). Length grows from 0 → 5 over phases. */
  community: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  bonusBet: number;
  flopBet: number;
  turnBet: number;
  riverBet: number;
  totalPlayBet: number;
  result: number;
  antePayout: number;
  playPayout: number;
  bonusPayout: number;
  totalPayout: number;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Ultimate Texas Hold'em (アルティメット・テキサスホールデム) ---
