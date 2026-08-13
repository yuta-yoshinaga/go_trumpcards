// Type declarations for banluck. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Phases, matching the Go domain. */
export const BAN_LUCK_PHASE = { bet: 0, play: 1, roundEnd: 2, gameEnd: 3 } as const;

/**
 * Hand ranks, matching the Go domain. **Higher is stronger.**
 *
 * Card count decides the rank, not the total: `banBan` (A+A) totals 12 yet
 * outranks everything, and `fiveDragon` beats an ordinary 20 at any total.
 */
export const BAN_LUCK_RANK = {
  bust: 0,
  point: 1,
  fiveDragon: 2,
  banLuck: 3,
  banBan: 4,
} as const;

/** Outcome against the banker. The banker's own seat is always `push`. */
export const BAN_LUCK_OUTCOME = { lose: 0, push: 1, win: 2 } as const;

/** The banker cannot stand below this total. */
export const BAN_LUCK_BANKER_MUST_HIT_UNDER = 15;

/** A suggestion for the current decision. */
export interface BanLuckHint {
  /** `hit` or `stand`. */
  action: string;
  reason: string;
}

/** One seat at the table. */
export interface BanLuckSeat {
  name: string;
  isHuman: boolean;
  chips: number;
  /** The live stake. **Reset to 0 once the round settles** — use `roundBet` on the result screen. */
  bet: number;
  cards: Card[];
  score: number;
  /** 0=bust, 1=point, 2=fiveDragon, 3=banLuck, 4=banBan. */
  rank: number;
  /** Against the banker: 0=lose, 1=push, 2=win. */
  outcome: number;
  /** What this seat staked in the settled round. */
  roundBet: number;
  delta: number;
  busted: boolean;
  stood: boolean;
  isBanker: boolean;
  isTurn: boolean;
}

/** Ban Luck game settings. */
export interface BanLuckConfig {
  seats: number;
  initialChips: number;
  rounds: number;
  defaultBet: number;
}

/** Response payload for `/banluck/exec`. */
export interface BanLuckResponse extends BaseGameResponse {
  /** 0=Bet, 1=Play, 2=RoundEnd, 3=GameEnd. */
  phase: number;
  seats: BanLuckSeat[];
  bankerSeat: number;
  turnSeat: number;
  humanSeat: number;
  isHumanTurn: boolean;
  /**
   * The human seat is the banker and is below 15, so it cannot stand.
   *
   * **Do not re-derive this from the score.** The rule lives in the domain.
   */
  mustHit: boolean;
  roundNumber: number;
  remainingCards: number;
  winnerSeat: number;
  gameEndFlag: boolean;
  config?: BanLuckConfig;
}
