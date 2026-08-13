// Type declarations for ironcross. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Phases, matching the Go domain. */
export const IRONCROSS_PHASE = { betting: 0, chooseLine: 1, showdown: 2, gameEnd: 3 } as const;

/** Betting actions, matching the Go domain (shared betting constants). */
export const IRONCROSS_ACTION = {
  fold: 0,
  check: 1,
  call: 2,
  bet: 3,
  raise: 4,
} as const;

/**
 * The two arms of the cross, matching the Go domain.
 *
 * **`none` is a real value, not "unset".** A seat that has not chosen yet is
 * `none`; sending it as a choice is rejected by the server.
 */
export const IRONCROSS_LINE = { none: 0, vertical: 1, horizontal: 2 } as const;

/** Hole cards per seat. */
export const IRONCROSS_HOLE_CARDS = 4;

/** Cards in the cross. */
export const IRONCROSS_CROSS_CARDS = 5;

/** A suggestion for the current decision. */
export interface IronCrossHint {
  /** `fold`, `check`, `call`, `bet`, `raise` or `line`. */
  action: string;
  /** Which arm to take, when `action` is `line`. */
  line: number;
  reason: string;
}

/** One seat at the table. */
export interface IronCrossSeat {
  name: string;
  isHuman: boolean;
  chips: number;
  bet: number;
  /**
   * The four hole cards.
   *
   * **Empty for CPU seats until showdown.** The server withholds them rather
   * than relying on the page to hide them.
   */
  cards: Card[];
  folded: boolean;
  allIn: boolean;
  isTurn: boolean;
  /**
   * The arm this seat plays (0=not chosen, 1=vertical, 2=horizontal).
   *
   * **Stays 0 for CPU seats until showdown** — a visible choice would give
   * away the strength the betting is meant to hide.
   */
  line: number;
  /** Set at showdown only. */
  handRank: number;
  /** The best five of the seven available cards. Set at showdown only. */
  bestHand: Card[];
  wonAmount: number;
}

/** Iron Cross game settings. */
export interface IronCrossConfig {
  seats: number;
  initialChips: number;
  ante: number;
}

/** Response payload for `/ironcross/exec`. */
export interface IronCrossResponse extends BaseGameResponse {
  /** 0=Betting, 1=ChooseLine, 2=Showdown, 3=GameEnd. */
  phase: number;
  seats: IronCrossSeat[];
  /**
   * The cross, always five entries and always positional.
   *
   * **A face-down slot is `null`, not omitted.** Index 0 is the centre, 1 the
   * top, 2 the bottom, 3 the left and 4 the right — compacting the array would
   * make it impossible to tell which arm a card belongs to.
   */
  cross: (Card | null)[];
  /** How many of the five are face up. */
  revealedCount: number;
  /** How many there will be in total (5). */
  crossTotal: number;
  /** The cross positions the vertical line uses (top, centre, bottom). */
  verticalIndexes: number[];
  /** The cross positions the horizontal line uses (left, centre, right). */
  horizontalIndexes: number[];
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
  /** True while the table is picking vertical or horizontal. */
  isChoosing: boolean;
  handNumber: number;
  remainingCards: number;
  winnerSeat: number;
  gameEndFlag: boolean;
  config?: IronCrossConfig;
}
