// Type declarations for pontoon. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Pontoon hand. A seat holds several once it splits. */
export interface PontoonHand {
  /**
   * Null entries while {@link PontoonHand.hidden} is true. The server does not
   * send the cards of a hand the player may not see, so reading the response
   * cannot reveal the banker's hole cards; only the COUNT survives, because
   * Twist and Buy change the hand size in view of the table.
   */
  cards: (Card | null)[];
  bet: number;
  /** Server-computed total, with the ace already resolved to 1 or 11. 0 while hidden. */
  total: number;
  /** 0 = bust, 1 = points, 2 = five card trick, 3 = pontoon. 0 while hidden. */
  rank: number;
  /** While true the hand's cards, total and rank are withheld by the server. */
  hidden: boolean;
  /** Once true the hand can no longer buy. */
  twisted: boolean;
  stuck: boolean;
  /** Net chips after settlement; 0 while the round is live. */
  payout: number;
}

/** One Pontoon seat. */
export interface PontoonSeat {
  name: string;
  isCpu: boolean;
  hands: PontoonHand[];
}

/** Full Pontoon game state returned from the API. */
export interface PontoonResponse extends BaseGameResponse {
  seats: PontoonSeat[];
  /** The banker's hand. Absent before the deal. */
  bankerHand?: PontoonHand;
  bankerIdx: number;
  /** While true the human banks: no stake, and they decide the draw at the end. */
  isHumanBanker: boolean;
  chips: number;
  activeSeat: number;
  activeHand: number;
  /** Seat that takes the bank next deal, or -1. */
  nextBanker: number;
  lastResult: string;
  phase: number;
  /**
   * The server decides what is legal. Recomputing these on the client would
   * mean re-implementing the 15 minimum to stick and the no-buy-after-twist
   * rule, and two implementations drift.
   */
  canStick: boolean;
  canTwist: boolean;
  canBuy: boolean;
  canSplit: boolean;
}
