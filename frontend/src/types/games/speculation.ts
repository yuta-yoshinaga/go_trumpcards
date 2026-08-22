// Type declarations for Speculation. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * "No seat" sentinel used by every seat-valued field except `turnSeat`.
 *
 * **It is -1, never 0.** Seat 0 is the human, so a component that treats a
 * falsy seat as "nobody" points the "holds the best trump" marker straight at
 * the player before a single card has been turned. `bestSeat`, `offerFrom`,
 * `offerTo` and `winnerSeat` all use it (SpeculationWebController.go).
 */
export const SPECULATION_NO_SEAT = -1;

/** Seat index of the human player. Every other seat is CPU (Speculation.go). */
export const SPECULATION_HUMAN_SEAT = 0;

/**
 * Trump suit sentinel for "not decided yet".
 *
 * **Not 0.** 0 is `CardDesignJoker` in the Go deck and 1..4 are ♠♣♥♦, so an
 * undecided trump drawn as suit 0 would render as a real suit.
 */
export const SPECULATION_NO_SUIT = -1;

/**
 * One seat's public information.
 *
 * **The face-down cards themselves are never sent** — only how many are left.
 * Knowing what a rival still holds is exactly the secret the auction is played
 * over, so the server sends `hiddenCount` and nothing else
 * (SpeculationWebPresenter.go).
 */
export interface SpeculationSeat {
  name: string;
  chips: number;
  /** How many cards this seat has still to turn up. */
  hiddenCount: number;
  /** The best trump this seat currently holds, or absent when it holds none. */
  best?: Card;
}

/** Speculation table settings. */
export interface SpeculationConfig {
  players: number;
  initialChips: number;
  /** Ante every seat pays into the pot at the start of a round. */
  stake: number;
  rounds: number;
}

/** Response payload for `/speculation/exec`. */
export interface SpeculationResponse extends BaseGameResponse {
  /** 0=Flip, 1=Auction, 2=Result, 3=GameEnd. */
  phase: number;
  seats: SpeculationSeat[];
  /** 1=♠ 2=♣ 3=♥ 4=♦, or {@link SPECULATION_NO_SUIT} before the trump is turned. */
  trumpSuit: number;
  /** The card turned to fix the trump suit. It belongs to nobody. */
  trumpCard?: Card;
  pot: number;
  /** Seat whose turn it is to turn up a card. */
  turnSeat: number;
  /** Seat holding the best trump, or {@link SPECULATION_NO_SEAT}. */
  bestSeat: number;
  /** Seat making the offer, or {@link SPECULATION_NO_SEAT} when no auction is open. */
  offerFrom: number;
  /** Seat receiving the offer (the card's owner), or {@link SPECULATION_NO_SEAT}. */
  offerTo: number;
  /** Chips offered. 0 when no auction is open. */
  offerAmount: number;
  /** Rounds completed so far (0-based; the round in play is `roundNo + 1`). */
  roundNo: number;
  /** Winner of the round just settled, or {@link SPECULATION_NO_SEAT} when void. */
  winnerSeat: number;
  gameEndFlag: boolean;
  config?: SpeculationConfig;
}
