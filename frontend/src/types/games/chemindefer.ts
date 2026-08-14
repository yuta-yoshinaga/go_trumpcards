// Type declarations for chemindefer. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Number of seats at a Chemin de Fer table. The bank travels between them. */
export const CHEMIN_DE_FER_SEATS = 6;

/** The seat the human occupies. */
export const CHEMIN_DE_FER_HUMAN_SEAT = 0;

/** Phases, matching the Go domain. */
export const CHEMIN_DE_FER_PHASE = {
  stake: 0,
  bet: 1,
  punterDraw: 2,
  bankerDraw: 3,
  roundEnd: 4,
} as const;

/** Round outcomes, matching the Go domain. */
export const CHEMIN_DE_FER_RESULT = {
  none: 0,
  banker: 1,
  punter: 2,
  tie: 3,
} as const;

/**
 * One seat at a Chemin de Fer table.
 *
 * **`isBanker` is not a property of the seat.** The bank passes to the next
 * seat whenever the punters win, so this flips between rounds.
 */
export interface ChemindeFerPlayer {
  id: number;
  name: string;
  isHuman: boolean;
  chips: number;
  /** Amount staked this round. Cleared once the coup settles. */
  bet: number;
  isBanker: boolean;
  /** The highest bettor, who decides the punter side's third card. */
  isRepresentative: boolean;
}

/** A suggestion for the side currently deciding on a third card. */
export interface ChemindeFerHint {
  draw: boolean;
  /** `punterFive`, `bankerDraw` or `bankerStand`. */
  reason: string;
}

/** Chemin de Fer game settings. */
export interface ChemindeFerConfig {
  rounds: number;
  initialChips: number;
}

/** Response payload for `/chemindefer/exec`. */
export interface ChemindeFerResponse extends BaseGameResponse {
  players: ChemindeFerPlayer[];
  /** 0=Stake, 1=Bet, 2=PunterDraw, 3=BankerDraw, 4=RoundEnd. */
  phase: number;
  bankerIdx: number;
  /** Seat due to bet, or -1 once betting has closed. */
  betTurn: number;
  stake: number;
  remainingStake: number;
  totalBet: number;
  /** Range the banker may stake. */
  stakeMin: number;
  stakeMax: number;
  /** Range the seat on turn may bet. Both are 0 once betting has closed. */
  betMin: number;
  betMax: number;
  /** -1 until the deal. */
  representativeIdx: number;
  /**
   * Whether the punter side has a real decision right now.
   *
   * **True only at a total of 5.** 0-4 must draw and 6-7 must stand, and the
   * server applies those itself rather than asking — so the page must not
   * offer buttons when this is false.
   */
  punterMayChoose: boolean;
  bankerHand: Card[];
  punterHand: Card[];
  bankerTotal: number;
  punterTotal: number;
  punterDrew: boolean;
  /** 0=undecided, 1=banker, 2=punters, 3=tie. */
  result: number;
  roundNumber: number;
  remainingCards: number;
  isHumanTurn: boolean;
  gameEndFlag: boolean;
  config?: ChemindeFerConfig;
}
