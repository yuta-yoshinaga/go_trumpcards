// Type declarations for baseballpoker. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Phases, matching the Go domain. */
export const BASEBALL_PHASE = { betting: 0, buyIn: 1, showdown: 2, gameEnd: 3 } as const;

/** Betting actions, matching the Go domain (shared betting constants). */
export const BASEBALL_ACTION = {
  fold: 0,
  check: 1,
  call: 2,
  bet: 3,
  raise: 4,
} as const;

/** A suggestion for the current decision. */
export interface BaseballHint {
  /** `fold`, `check`, `call`, `bet`, `raise` or `pay`. */
  action: string;
  reason: string;
}

/** One seat at the table. */
export interface BaseballSeat {
  name: string;
  isHuman: boolean;
  chips: number;
  bet: number;
  /**
   * The seat's cards, positional.
   *
   * **A card the viewer may not see is `null`, not omitted.** Face-up cards
   * are present for every seat — they are what stud players read — while a
   * face-down card of another seat arrives as `null` so the order of the
   * revealed cards stays legible.
   */
  cards: (Card | null)[];
  /** Runs parallel to {@link cards}; true where that card is face up. */
  faceUp: boolean[];
  /** Extra cards this seat received from face-up 4s. */
  bonusCards: number;
  folded: boolean;
  allIn: boolean;
  isTurn: boolean;
  /** True while this seat is the one being asked to buy the pot. */
  isBuying: boolean;
  /** Set at showdown only. */
  handRank: number;
  /** Whether the best five used a wild card. Set at showdown only. */
  usedWild: boolean;
  /** The best five cards. Set at showdown only. */
  bestHand: Card[];
  wonAmount: number;
}

/** Baseball Poker game settings. */
export interface BaseballConfig {
  seats: number;
  initialChips: number;
  ante: number;
}

/** Response payload for `/baseballpoker/exec`. */
export interface BaseballPokerResponse extends BaseGameResponse {
  /** 0=Betting, 1=BuyIn, 2=Showdown, 3=GameEnd. */
  phase: number;
  seats: BaseballSeat[];
  /** How many face-up cards have been dealt (1..4). */
  street: number;
  /** How many there will be in total (4). */
  streetTotal: number;
  /**
   * The card values that are wild (3 and 9).
   *
   * **Sent by the server so the page never hardcodes them** — a copy here
   * would drift from the evaluator and mark the wrong cards.
   */
  wildValues: number[];
  /** The face-up value that pays an extra card (4). */
  bonusValue: number;
  /** The face-up value that forces the buy-or-fold choice (3). */
  buyInValue: number;
  pot: number;
  currentBet: number;
  /** What the human seat must put in to call. */
  toCall: number;
  raiseCount: number;
  /** False once the per-round raise cap is reached. **Do not re-derive.** */
  canRaise: boolean;
  turnSeat: number;
  humanSeat: number;
  isHumanTurn: boolean;
  /** The seat being asked to buy the pot, or -1 when nobody is. */
  buyerSeat: number;
  /** What the buyer must pay. Fixed when asked, not when answering. */
  buyCost: number;
  /** True while the human seat is the one being asked. */
  isBuying: boolean;
  handNumber: number;
  remainingCards: number;
  winnerSeat: number;
  gameEndFlag: boolean;
  config?: BaseballConfig;
}
