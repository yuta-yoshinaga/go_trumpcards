// Type declarations for blackjackswitch. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single hand within a Blackjack Switch round. */
export interface BlackJackSwitchHand {
  /** Cards in this hand. Null entries represent face-down cards (e.g. dealer hole). */
  cards: (Card | null)[];
  score: number;
  bet: number;
  stood: boolean;
  doubled: boolean;
  busted: boolean;
  /** True when the hand is a 2-card 21 (natural). Pays 1:1 in Blackjack Switch. */
  isBJ: boolean;
  /** Domain GameResult: 1=Win, 0=Draw, -1=Lose. */
  result: number;
  /** Per-hand payout (bet returned + winnings). */
  payout: number;
}

/** Blackjack Switch game state response. */
export interface BlackJackSwitchResponse extends BaseGameResponse {
  /** Two player hands; the player may switch the second card between them. */
  hands: BlackJackSwitchHand[];
  /** Dealer's cards. The hole card is null until the round ends. */
  dealerCards: (Card | null)[];
  /** Visible dealer score (up-card only mid-round; full score at end). */
  dealerScore: number;
  /** 1=BET, 2=SWITCH, 3=ACTION, 4=END. */
  phase: number;
  /** Index of the player hand currently acting (during ACTION phase). */
  currentHandIdx: number;
  chips: number;
  /** True when the player exercised the switch this round. */
  switched: boolean;
  /** True when the dealer ended on 22 (push rule, not natural BJ). */
  dealerPushed22: boolean;
  /** Aggregate of hand results: 1=net win, 0=draw, -1=net loss. */
  overallResult: number;
  /** Sum of per-hand payouts. */
  totalPayout: number;
}
