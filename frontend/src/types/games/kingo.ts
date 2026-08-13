// Type declarations for kingo. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Phases, matching the Go domain. */
export const KINGO_PHASE = { bet: 0, result: 1, gameEnd: 2 } as const;

/** Hand ranks, matching the Go domain. */
export const KINGO_RANK = { none: 0, pair: 1, arashi: 2 } as const;

/** Cards dealt to each seat. */
export const KINGO_HAND_SIZE = 3;

/** A suggestion for the current decision. */
export interface KingoHint {
  /** `bet`, `deal` or `next`. */
  action: string;
  /** The stake it suggests, when `action` is `bet`. */
  amount: number;
  reason: string;
}

/** One seat at the table. */
export interface KingoSeat {
  name: string;
  isHuman: boolean;
  chips: number;
  /** This round's stake. **Always 0 for the banker — it does not bet.** */
  bet: number;
  /**
   * The three cards.
   *
   * **Empty for every seat during the bet phase.** Nothing is dealt until the
   * bets are in, so no seat has a private hand — this game has no hidden
   * information to withhold, only information that does not exist yet.
   */
  cards: Card[];
  /** 0=none, 1=pair, 2=arashi. Set once the round resolves. */
  rank: number;
  /** The number this seat collected. Set once the round resolves. */
  matchedValue: number;
  isBanker: boolean;
  /** This round's net for the seat; negative when it paid. */
  wonAmount: number;
}

/** Kingo game settings. */
export interface KingoConfig {
  seats: number;
  initialChips: number;
  minBet: number;
  rounds: number;
}

/** Response payload for `/kingo/exec`. */
export interface KingoResponse extends BaseGameResponse {
  /** 0=Bet, 1=Result, 2=GameEnd. */
  phase: number;
  seats: KingoSeat[];
  bankerSeat: number;
  roundNumber: number;
  /** Total rounds; at least the seat count so the bank reaches every seat. */
  rounds: number;
  humanSeat: number;
  /** **Decides whether the human owes a bet or a deal.** */
  isHumanBanker: boolean;
  isHumanTurn: boolean;
  /** Always 3. */
  handSize: number;
  /**
   * Multipliers for the two combinations.
   *
   * **Sent by the server so the page never hardcodes them** — a copy here
   * would drift from the odds the payout was derived from.
   */
  payoutArashi: number;
  payoutPair: number;
  remainingCards: number;
  winnerSeat: number;
  gameEndFlag: boolean;
  config?: KingoConfig;
}
