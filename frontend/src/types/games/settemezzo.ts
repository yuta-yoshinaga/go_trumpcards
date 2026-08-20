// Type declarations for settemezzo. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Sette e Mezzo hand. */
export interface SetteEMezzoHand {
  /**
   * Null entries while {@link SetteEMezzoHand.hidden} is true. The server does
   * not send the cards of a hand the player may not see; only the COUNT
   * survives, because how many cards a seat drew is visible at the table.
   */
  cards: (Card | null)[];
  bet: number;
  /**
   * Total in HALF-points, as an integer. Face cards are worth half a point, and
   * landing on exactly 7.5 is what takes the bank, so the comparison has to be
   * exact rather than close. 0 while hidden.
   */
  totalHalves: number;
  /** The same total rendered for display, e.g. "7.5". Empty while hidden. */
  totalLabel: string;
  /** The matta's assigned value in halves; 0 when the hand holds none. */
  mattaHalves: number;
  hasMatta: boolean;
  stood: boolean;
  payout: number;
  /** While true the hand's cards and total are withheld by the server. */
  hidden: boolean;
}

/** One Sette e Mezzo seat. */
export interface SetteEMezzoSeat {
  name: string;
  isCpu: boolean;
  /** Absent before the deal. */
  hand?: SetteEMezzoHand;
}

/** Full Sette e Mezzo game state returned from the API. */
export interface SetteEMezzoResponse extends BaseGameResponse {
  seats: SetteEMezzoSeat[];
  bankerHand?: SetteEMezzoHand;
  bankerIdx: number;
  /** While true the human banks: no stake, and they decide the draw at the end. */
  isHumanBanker: boolean;
  chips: number;
  activeSeat: number;
  /** Seat that takes the bank next deal, or -1. Only an exact 7.5 moves it. */
  nextBanker: number;
  lastResult: string;
  phase: number;
  /** 7.5 in halves (15), sent so the target is not hardcoded on both sides. */
  targetHalves: number;
  canHit: boolean;
  canStand: boolean;
  /** Total (in halves) at which the CPU seats and the banker stand. */
  cpuStandHalves: number;
  canSetMatta: boolean;
}
