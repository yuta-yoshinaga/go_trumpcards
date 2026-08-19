// Type declarations for israeliwhist. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Israeli Whist trick. */
export interface IsraeliWhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at an Israeli Whist table. Four players, every one for themselves. */
export interface IsraeliWhistPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Tricks called in the auction, or `-1` when this seat never bid. */
  auctionBid: number;
  /** Suit offered in the auction, or `0`. */
  auctionSuit: number;
  /** Dropped out of the auction. Final — a passed seat cannot bid again. */
  passed: boolean;
  /** Target called in the second round, or `-1` before this seat has called. */
  bid: number;
  trickCount: number;
  /** Change from the round just scored. */
  roundScore: number;
  totalScore: number;
}

/**
 * A suggestion. During the auction and the calling round it carries no
 * `cardIndex` and puts the recommended number in `value` (and the suit in
 * `suit` when it recommends a bid); during play it names a card.
 */
export interface IsraeliWhistHint {
  cardIndex?: number;
  /**
   * `israeliwhistAuctionBid` / `israeliwhistAuctionPass` in the auction;
   * `israeliwhistBid` / `israeliwhistMeetQuota` / `israeliwhistAvoidRestricted`
   * when calling; `israeliwhistWinTrick` or `israeliwhistDuck` during play.
   */
  reason: string;
  /** Tricks to bid or call. `0` during play. */
  value: number;
  /** Suit to bid in the auction; `0` otherwise. */
  suit: number;
}

/** Round-count setting. */
export interface IsraeliWhistConfig {
  /** Rounds played before the game ends (1..20, default 4). */
  rounds: number;
}

/** Full Israeli Whist game state returned from the API. */
export interface IsraeliWhistResponse extends BaseGameResponse {
  players: IsraeliWhistPlayer[];
  /** `0` = Auction, `1` = Bid, `2` = Play, `3` = RoundEnd, `4` = GameEnd. */
  phase: number;
  roundNumber: number;
  /**
   * Whether the round that just ended scored double.
   *
   * All four hitting their bid, or all four missing, doubles everyone's change —
   * the swing that used to live only in the action log (#5752).
   */
  doubled?: boolean;
  /** True when the doubling came from every seat hitting (false: every seat missing). */
  doubledAllExact?: boolean;
  trickNumber: number;
  /** `0` until the auction closes. */
  trumpSuit: number;
  /** Seat that won the auction, or `-1` before it closes. */
  declarerIdx: number;
  /** Highest auction call so far; becomes the declarer's quota. */
  highBid: number;
  /** Suit of the highest auction call. */
  highSuit: number;
  /** Lowest call you may make — the quota when you are the declarer, else `0`. */
  minimumBid: number;
  /**
   * The one call the last bidder may not make, because it would bring the
   * total to 13; `-1` when nobody is under that restriction.
   */
  restrictedBid: number;
  currentPlayerIdx: number;
  auctionPlayerIdx: number;
  bidPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: IsraeliWhistTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerIdx: number;
  hint?: IsraeliWhistHint;
  config: IsraeliWhistConfig;
}
