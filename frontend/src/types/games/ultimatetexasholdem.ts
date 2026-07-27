// Type declarations for ultimatetexasholdem. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, MaskedCard } from '../common';

/** Ultimate Texas Hold'em API response. */
export interface UltimateTexasHoldemResponse extends BaseGameResponse {
  /** Player's two hole cards. */
  playerHand: Card[];
  /** Dealer's hole cards: masked as `MaskedCard` until the showdown. */
  dealerHand: (Card | MaskedCard)[];
  /** Community cards (flop / turn / river). Length grows from 0 → 5 over phases. */
  community: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  blindBet: number;
  tripsBet: number;
  playBet: number;
  folded: boolean;
  result: number;
  dealerQualified: boolean;
  antePayout: number;
  blindPayout: number;
  playPayout: number;
  tripsPayout: number;
  totalPayout: number;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Mississippi Stud (ミシシッピ・スタッド) ---
