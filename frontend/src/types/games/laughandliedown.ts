// Type declarations for laughandliedown. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Laugh and Lie Down seat. */
export interface LaughAndLieDownPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. Always sent, including while {@link LaughAndLieDownPlayer.hidden} is true. */
  cardCount: number;
  /** Empty while {@link LaughAndLieDownPlayer.hidden} is true. */
  cards: Card[];
  /**
   * Cards captured. Public for every seat: the difference from eight IS the
   * settlement.
   */
  wonCount: number;
  /** Could not capture, so their whole hand went to the table. */
  laidDown: boolean;
  /** Net chips. Zero until the game ends. */
  score: number;
  /** Whether this seat's HAND is withheld. */
  hidden: boolean;
}

/** Suggested move for the human seat. */
export interface LaughAndLieDownHintPayload {
  cardIndex?: number;
  /** Table cards the suggestion would take: 1 or 3. */
  takeCount: number;
  /** Reason identifier, e.g. `laughandliedown.hint.take_three`. */
  reason: string;
}

/** Full Laugh and Lie Down game state returned from the API. */
export interface LaughAndLieDownResponse extends BaseGameResponse {
  players: LaughAndLieDownPlayer[];
  /**
   * The face-up table. Starts at twelve and grows with the hands of players who
   * lay down. There is no face-down stock.
   */
  layout: Card[];
  /** 0 = Play, 1 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  /** Hand indices that match some table card by rank. */
  validIndices: number[];
  /**
   * Subset of {@link LaughAndLieDownResponse.validIndices} whose rank has three
   * or more cards on the table, so the three-card take is offered. Sent so the
   * page never recounts the table.
   */
  threeTakeIndices: number[];
  /** Stakes one more than the others, and sweeps the leftovers. */
  dealerIdx: number;
  /** Last seat still holding cards; -1 while undecided. */
  lastInIdx: number;
  /** What the player at `lastInIdx` receives. Sent so the figure is not written down twice. */
  lastInBonus: number;
  /**
   * Total staked. Equals the last-in award plus the over/short total, which is
   * what verifies the settlement table.
   */
  pot: number;
  gameEndFlag: boolean;
  hint?: LaughAndLieDownHintPayload;
}
