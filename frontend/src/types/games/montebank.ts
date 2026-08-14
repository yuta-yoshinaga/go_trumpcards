// Type declarations for montebank. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Phases, matching the Go domain. */
export const MONTE_BANK_PHASE = { bet: 0, result: 1, gameEnd: 2 } as const;

/** Round outcomes, matching the Go domain. */
export const MONTE_BANK_RESULT = { none: 0, win: 1, lose: 2 } as const;

/** Multiplier paid on a suit match. */
export const MONTE_BANK_PAYOUT = 3;

/** A suggestion for which layout card to back. */
export interface MonteBankHint {
  pickIdx: number;
  reason: string;
}

/** One exposed layout card, with the numbers that decide whether backing it is good. */
export interface MonteBankLayoutCard {
  card: Card;
  /**
   * How many of this suit are showing in the layout.
   *
   * **This is the only number that decides whether a bet is good**, and the
   * entire house edge comes out of it: one is exactly break-even at 3:1, two
   * costs 11.1%, three 22.2%, four 33.3%.
   */
  suitCount: number;
  /** How many of this suit are left in the deck. */
  remainingOfSuit: number;
  /** True when `suitCount` is 1. **Server-computed — do not re-derive.** */
  isEven: boolean;
  isPicked: boolean;
}

/** Monte Bank game settings. */
export interface MonteBankConfig {
  initialChips: number;
  defaultBet: number;
}

/** Response payload for `/montebank/exec`. */
export interface MonteBankResponse extends BaseGameResponse {
  /** 0=Bet, 1=Result, 2=GameEnd. */
  phase: number;
  layout: MonteBankLayoutCard[];
  /** The card turned to settle the round. Absent before the bet. */
  gate?: Card;
  /** 0-based layout position backed this round, or -1 before the bet. */
  pick: number;
  bet: number;
  /** 0=none, 1=win, 2=lose. */
  result: number;
  payout: number;
  chips: number;
  roundNumber: number;
  remainingCards: number;
  gameEndFlag: boolean;
  payoutMultiplier: number;
  config?: MonteBankConfig;
}
