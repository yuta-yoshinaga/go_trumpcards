// Type declarations for continentalrummy. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Seats at the table: one human and three CPUs. */

/** The seat the human occupies. */

/** Cards dealt to each seat, three at a time. */

/**
 * Phases, matching the Go domain.
 *
 * **These are strings, not indices.** A turn is always draw then discard, so
 * `DISCARD` is also where going out happens.
 */
export const CONTINENTAL_RUMMY_PHASE = {
  DRAW: 'draw',
  DISCARD: 'discard',
  ROUND_END: 'roundEnd',
  GAME_END: 'gameEnd',
} as const;

/** One seat. **Only the human's `cards` are populated**; others carry a count. */
export interface ContinentalRummyPlayer {
  id: number;
  isHuman: boolean;
  cards: Card[];
  cardCount: number;
  /** Sequences laid down on going out. Empty until someone goes out. */
  melds: Card[][];
  score: number;
  isDealer: boolean;
}

/** One line of the winner's bonus tally. */
export interface ContinentalRummyBonus {
  /** `win`, `joker`, `noJoker`, `firstTurn`, `dealt` or `oneSuit`. */
  key: string;
  points: number;
}

/**
 * One round's settlement.
 *
 * **The winner collects; the losers' cards are never counted.** `perOpponent`
 * is the tally, and `total` is that times the number of opponents.
 */
export interface ContinentalRummyResult {
  /** The seat that went out, or -1 when the round washed out. */
  winnerIdx: number;
  bonuses: ContinentalRummyBonus[];
  perOpponent: number;
  total: number;
}

/** Score gained by each seat in one completed round. */
export interface ContinentalRummyRoundScore {
  roundNumber: number;
  scores: number[];
}

/** Continental Rummy game settings. */
export interface ContinentalRummyConfig {
  cpuDifficulty: number;
  totalRounds: number;
}

/** Response payload for `/continentalrummy/exec`. */
export interface ContinentalRummyResponse extends BaseGameResponse {
  players: ContinentalRummyPlayer[];
  /** `draw` | `discard` | `roundEnd` | `gameEnd`. */
  phase: string;
  roundNumber: number;
  totalRounds: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  stockCount: number;
  discardTop?: Card;
  /**
   * The card counts that make a legal go-out, straight from the domain.
   *
   * **Do not hardcode these.** Three fives totals fifteen and is still not a
   * go-out, so a client that reconstructs the list from "partitions of 15"
   * gets it wrong.
   */
  layouts: number[][];
  lastResult?: ContinentalRummyResult;
  /** Completed round gains, retained by the server for the whole game. */
  roundScoreHistory: ContinentalRummyRoundScore[];
  gameEndFlag: boolean;
  winnerIdx: number;
  isHumanTurn: boolean;
  /**
   * The one card you can throw to go out, or -1.
   *
   * **The server solves the fifteen-card partition.** Re-deriving it in the
   * page would put the rule in a second place and let the two disagree.
   */
  goOutIdx: number;
  /** Hand indices grouped into legal runs when a go-out discard exists. */
  goOutGroups: number[][];
  /**
   * Whether the dealt fifteen already go out, before drawing.
   *
   * **This is the heavier win.** Laying them without ever drawing pays 10,
   * against 7 for going out after one draw — so it is offered as its own
   * action rather than folded into the ordinary go-out.
   */
  canGoOutOnDeal: boolean;
  /** The card worth throwing, or -1. */
  hintDiscardIdx: number;
  /** `draw_stock`, `take_discard`, `go_out` or `discard_loose`. */
  hintReason: string;
  config?: ContinentalRummyConfig;
}
