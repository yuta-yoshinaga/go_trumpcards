// Type declarations for quinze. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Quinze phase constants (sync: internal/domain/Quinze.go). */
export const QuinzePhase = {
  BET: 1,
  PLAYER_TURN: 2,
  BANKER_TURN: 3,
  END: 4,
} as const;

/** One Quinze hand. */
export interface QuinzeHand {
  /**
   * Null entries while {@link QuinzeHand.hidden} is true. The server does
   * not send the cards of a hand the player may not see; only the COUNT
   * survives, because how many cards a seat drew is visible at the table.
   */
  cards: (Card | null)[];
  bet: number;
  /**
   * Total in points, as an integer. Aces are worth 1 point and face cards are
   * worth 10 points. Landing on exactly 15 takes the bank. 0 while hidden.
   */
  totalPoints: number;
  /** The same total rendered for display, e.g. "15". Empty while hidden. */
  totalLabel: string;
  stood: boolean;
  payout: number;
  /** While true the hand's cards and total are withheld by the server. */
  hidden: boolean;
}

/** One Quinze seat. */
export interface QuinzeSeat {
  name: string;
  isCpu: boolean;
  /** Absent before the deal. */
  hand?: QuinzeHand;
}

/** Full Quinze game state returned from the API. */
export interface QuinzeResponse extends BaseGameResponse {
  seats: QuinzeSeat[];
  bankerHand?: QuinzeHand;
  bankerIdx: number;
  /** While true the human banks: no stake, and they decide the draw at the end. */
  isHumanBanker: boolean;
  chips: number;
  activeSeat: number;
  /** Seat that takes the bank next deal, or -1. Only an exact 15 moves it. */
  nextBanker: number;
  lastResult: string;
  phase: number;
  /** Target score of 15 points, sent so the target is not hardcoded on both sides. */
  targetPoints: number;
  canHit: boolean;
  canStand: boolean;
  /** Total points at which the CPU seats and the banker stand. */
  cpuStandPoints: number;
}
