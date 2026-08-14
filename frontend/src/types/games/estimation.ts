// Type declarations for estimation. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Estimation trick. */
export interface EstimationTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at an Estimation table. Four players, every one for themselves. */
export interface EstimationPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Tricks called, or `-1` before this seat has called. */
  bid: number;
  /** `0` = normal, `1` = Dash Call (0), `2` = Risk (the highest call). */
  callType: number;
  trickCount: number;
  /** Change from the round just scored. */
  roundScore: number;
  totalScore: number;
}

/**
 * A suggestion. While choosing trump or calling it carries no `cardIndex` and
 * puts the recommended suit or number in `value`; during play it names a card.
 */
export interface EstimationHint {
  cardIndex?: number;
  /**
   * `estimationSelectTrump` / `estimationBid` / `estimationDashCall` /
   * `estimationAvoidRestricted` before play; `estimationWinTrick` or
   * `estimationDuck` (you already have your call) during play.
   */
  reason: string;
  /** Suit to make trump, or the number to call. `0` during play. */
  value: number;
}

/** Round-count setting. */
export interface EstimationConfig {
  /** Rounds played before the game ends (1..18, default 5). */
  rounds: number;
}

/** Full Estimation game state returned from the API. */
export interface EstimationResponse extends BaseGameResponse {
  players: EstimationPlayer[];
  /** `0` = Trump, `1` = Bid, `2` = Play, `3` = RoundEnd, `4` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  /** `0` until the dealer has chosen. */
  trumpSuit: number;
  /**
   * The one call the last bidder may not make, because it would bring the
   * total to 13; `-1` when nobody is under that restriction.
   */
  restrictedBid: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: EstimationTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerIdx: number;
  hint?: EstimationHint;
  config: EstimationConfig;
}
