// Type declarations for teendopaanch. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current 3-2-5 trick. */
export interface TeenDoPaanchTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a 3-2-5 table. */
export interface TeenDoPaanchPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /**
   * Tricks this seat owes this round: `3`, `2` or `5`. **Assigned, not bid** —
   * there is no bidding phase, and the three targets rotate between rounds.
   */
  target: number;
  trickCount: number;
  /** Rounds in which this seat made its target. **The game is decided on this.** */
  met: number;
}

/**
 * A suggestion. While choosing trump it carries no `cardIndex` and puts the
 * recommended suit in `suit`; during play it names a card.
 */
export interface TeenDoPaanchHint {
  cardIndex?: number;
  /**
   * `teendopaanchSelectTrump` before play; `teendopaanchWinTrick` or
   * `teendopaanchDuck` (you already have your target) during play.
   */
  reason: string;
  /** Suit to make trump. `0` during play. */
  suit: number;
}

/** Round-count setting. */
export interface TeenDoPaanchConfig {
  /** Rounds to play (3..30, default 3 — one turn each at 3, 2 and 5). */
  rounds: number;
}

/** Full 3-2-5 game state returned from the API. */
export interface TeenDoPaanchResponse extends BaseGameResponse {
  players: TeenDoPaanchPlayer[];
  /** `0` = Trump, `1` = Play, `2` = RoundEnd, `3` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  /** `0` until the 5-target seat has chosen. */
  trumpSuit: number;
  /** Seat that owes 5 tricks. **That seat, and only it, names trump.** */
  fivePlayerIdx: number;
  /**
   * How many cards changed hands for last round's shortfall. Surplus tricks
   * buy an opponent's best cards; nothing on the board shows this happened.
   */
  lastExchange: number;
  /**
   * Who handed cards to whom in the exchange that opened this round.
   *
   * The total alone cannot say whose best card was taken, or what was pulled out
   * of your own hand (#5757).
   */
  lastExchangePairs?: { giver: number; taker: number; count: number }[];
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  currentTrick: TeenDoPaanchTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerIdx: number;
  hint?: TeenDoPaanchHint;
  config: TeenDoPaanchConfig;
}
