// Type declarations for freebet. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Phases, matching the Go domain. */
export const FREE_BET_PHASE = { bet: 0, play: 1, result: 2 } as const;

/** Per-hand outcomes, matching the Go domain. */
export const FREE_BET_RESULT = {
  none: 0,
  win: 1,
  lose: 2,
  push: 3,
  blackjack: 4,
  /** The dealer busted with exactly 22, so the hand pushes. */
  dealer22Push: 5,
} as const;

/** A suggestion for the current decision. */
export interface FreeBetHint {
  /** `hit`, `stand`, `freeDouble` or `freeSplit`. */
  action: string;
  reason: string;
}

/** One player hand. Free splits produce more of these. */
export interface FreeBetHand {
  cards: Card[];
  score: number;
  /** The player's own stake. **This is what a loss costs.** */
  bet: number;
  /**
   * The house's contribution from a free double or free split.
   *
   * **Never add this to `bet`.** It earns a payout on a win and costs the
   * player nothing on a loss, so a single combined figure cannot answer the
   * question this game is about: how much of the stake is actually at risk.
   */
  freeBet: number;
  isSoft: boolean;
  stood: boolean;
  doubled: boolean;
  busted: boolean;
  /** Twenty-one on the first two cards. Pays 3:2. */
  blackjack: boolean;
  /** 0=none, 1=win, 2=lose, 3=push, 4=blackjack, 5=dealer-22 push. */
  result: number;
}

/** Free Bet Blackjack game settings. */
export interface FreeBetConfig {
  initialChips: number;
  defaultAnte: number;
}

/** Response payload for `/freebet/exec`. */
export interface FreeBetResponse extends BaseGameResponse {
  /** 0=Bet, 1=Play, 2=Result. */
  phase: number;
  hands: FreeBetHand[];
  activeHand: number;
  dealerCards: Card[];
  dealerScore: number;
  /**
   * The dealer busted with exactly **22**, so every surviving hand pushes.
   *
   * This is the price of the free doubles and splits; 23 or more is an
   * ordinary bust and pays.
   */
  dealerPushed22: boolean;
  /** **Do not re-derive.** Hard 9-11 on the first two cards only. */
  canFreeDouble: boolean;
  /** **Do not re-derive.** Any pair except ten-valued ones. */
  canFreeSplit: boolean;
  anteBet: number;
  payout: number;
  chips: number;
  roundNumber: number;
  remainingCards: number;
  gameEndFlag: boolean;
  config?: FreeBetConfig;
}
