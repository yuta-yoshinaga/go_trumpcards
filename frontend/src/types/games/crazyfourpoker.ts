// Type declarations for crazyfourpoker. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Phases, matching the Go domain. */
export const CRAZY_FOUR_POKER_PHASE = { bet: 0, decide: 1, result: 2 } as const;

/** Round outcomes, matching the Go domain. */
export const CRAZY_FOUR_POKER_RESULT = {
  none: 0,
  fold: 1,
  win: 2,
  lose: 3,
  push: 4,
  dealerNotQualified: 5,
} as const;

/** Ante step. Payouts include 1.5:1, so the stake is kept a multiple of this. */
export const CRAZY_FOUR_POKER_ANTE_UNIT = 10;

/** A suggestion for the play decision. `multiplier: 0` means fold. */
export interface CrazyFourPokerHint {
  multiplier: number;
  /** `acesOrBetter`, `marginal` or `fold`. */
  reason: string;
}

/** Crazy 4 Poker game settings. */
export interface CrazyFourPokerConfig {
  initialChips: number;
  defaultAnte: number;
}

/** Response payload for `/crazyfourpoker/exec`. */
export interface CrazyFourPokerResponse extends BaseGameResponse {
  /** 0=Bet, 1=Decide, 2=Result. */
  phase: number;
  /** Five cards; four of them play. */
  playerHand: Card[];
  /** **Empty until the showdown** — the server does not send it while deciding. */
  dealerHand: Card[];
  playerBest: Card[];
  /** Empty until the showdown. */
  dealerBest: Card[];
  /** 1=High Card .. 8=Four of a Kind. */
  playerHandRank: number;
  /** Zero until the showdown. */
  dealerHandRank: number;
  /** Whether the player holds a pair of aces or better. */
  hasAcesOrBetter: boolean;
  /**
   * The highest play multiplier available right now — 1, or 3 with aces.
   *
   * **Do not re-derive this from the hand.** The rule lives in the domain; a
   * second copy on the page drifts and silently loosens the 3x rule.
   */
  maxMultiplier: number;
  /**
   * Whether the player's own four is king high or better.
   *
   * Computed by the server so the page never re-derives "king high", which is
   * the same rule the dealer qualifies on.
   */
  playerQualifies: boolean;
  dealerQualifies: boolean;
  anteBet: number;
  /** Always equal to the ante. */
  superBet: number;
  queensUpBet: number;
  playBet: number;
  playMultiplier: number;
  /** 0=none, 1=fold, 2=win, 3=lose, 4=push, 5=dealer did not qualify. */
  result: number;
  /** Total returned this round, including returned stakes. */
  payout: number;
  chips: number;
  minTotalWager: number;
  roundNumber: number;
  remainingCards: number;
  gameEndFlag: boolean;
  config?: CrazyFourPokerConfig;
}
