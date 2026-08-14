// Type declarations for shelem. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Shelem trick. */
export interface ShelemTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Shelem table. */
export interface ShelemPlayer {
  id: number;
  isHuman: boolean;
  /** `0` or `1`. Seats 0+2 are one partnership, 1+3 the other. */
  team: number;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Points bid, or `-1` when this seat never bid. */
  bid: number;
  /** Dropped out of the bidding. Final — a passed seat cannot bid again. */
  passed: boolean;
  /** Claimed every trick instead of naming a number. */
  declaredShelem: boolean;
  trickCount: number;
}

/**
 * A suggestion. While bidding or discarding it carries no `cardIndex` and puts
 * the recommended points in `value` (or the trump suit in `suit`); during play
 * it names a card.
 */
export interface ShelemHint {
  cardIndex?: number;
  /**
   * `shelemBid` / `shelemPass` while bidding; `shelemDiscard` while sorting the
   * widow; `shelemWinTrick` or `shelemFeedPartner` during play.
   */
  reason: string;
  /** Points to bid. `0` otherwise. */
  value: number;
  /** Suit to make trump when discarding. `0` otherwise. */
  suit: number;
}

/** Target-score setting. */
export interface ShelemConfig {
  /** Points needed to win (100..2000, default 500). */
  target: number;
}

/** Full Shelem game state returned from the API. */
export interface ShelemResponse extends BaseGameResponse {
  players: ShelemPlayer[];
  /** `0` = Bid, `1` = Discard, `2` = Play, `3` = RoundEnd, `4` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  /** `0` until the declarer has named it. */
  trumpSuit: number;
  /** Seat that won the bidding, or `-1` before it settles. */
  declarerIdx: number;
  /** Points contracted for. */
  contract: number;
  /** The contract is Shelem — every trick — rather than a points total. */
  shelemBid: boolean;
  /** Lowest bid that would beat the standing one; already clamped to the maximum. */
  minBid: number;
  /** Cards still face down; `0` once the declarer has taken them. */
  widowSize: number;
  /** How many cards the declarer must discard (4). */
  discardCount: number;
  /** Running total per team, index 0 and 1. */
  scores: number[];
  /** Card points taken this round per team, index 0 and 1. */
  roundPoints: number[];
  /** Tricks taken this round per team, index 0 and 1. */
  teamTricks: number[];
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: ShelemTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerTeam: number;
  hint?: ShelemHint;
  config: ShelemConfig;
}
