// Type declarations for sjavs. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Sjavs seat. */
export interface SjavsPlayer {
  id: number;
  isHuman: boolean;
  /** 0 or 1; seats across the table are partners. */
  team: number;
  /** Hand size. Always sent, including while {@link SjavsPlayer.hidden} is true. */
  cardCount: number;
  /** Empty while {@link SjavsPlayer.hidden} is true. */
  cards: Card[];
  /**
   * This seat's bid, 0 for a pass or no bid yet. Public — a bid is spoken at
   * the table, and what your partner holds is the read.
   */
  bid: number;
  hidden: boolean;
}

/** One card played into the current trick. */
export interface SjavsTrickCard {
  playerIdx: number;
  card: Card;
}

/** Settlement of one hand. */
export interface SjavsHandResult {
  declarerTeam: number;
  declarerPoints: number;
  /** -1 when the hand tied 60-60 and nobody scored. */
  scoringTeam: number;
  amount: number;
  /** All eight tricks. */
  vol: boolean;
  trumpWasClubs: boolean;
}

/** Suggested move for the human seat. */
export interface SjavsHintPayload {
  cardIndex?: number;
  bidLength?: number;
  /** Reason identifier, e.g. `sjavs.hint.bid`. */
  reason: string;
}

/** Full Sjavs game state returned from the API. */
export interface SjavsResponse extends BaseGameResponse {
  players: SjavsPlayer[];
  /** 0 = Bid, 1 = Play, 2 = HandEnd, 3 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  /** -1 until bidding ends. */
  trumpSuit: number;
  /**
   * Trumps in the declared suit: 13 for a red suit, 12 for a black one,
   * counting the six permanent trumps. Zero while undecided. Sent so the page
   * never recounts — counting only the trump suit always falls short.
   */
  trumpCount: number;
  /** Seat that declared trumps; -1 while undecided. */
  bidderIdx: number;
  bidLength: number;
  /** Shortest biddable trump suit. Below this you must pass. */
  minBid: number;
  /** The human's longest trump suit, which caps their bid. */
  myLongest: number;
  trick: SjavsTrickCard[];
  trickNo: number;
  /**
   * Hand indices that may be played. Carries the following rule, including
   * that trumps form their own suit.
   */
  validIndices: number[];
  /**
   * Indices of the human's hand that are trumps, including the six permanent
   * ones. Empty until a trump is named. Sent because those six cannot be
   * recognised from their suit, and the client must not keep its own list.
   */
  trumpIndices: number[];
  /** Card points this hand, per team. Always sums to 120. */
  teamPoints: number[];
  /** Each team's distance from winning the rubber, counted DOWN from 24. */
  remaining: number[];
  /** Rubbers won, per team. A cross records a rubber, not a hand. */
  crosses: number[];
  /** Extra points a 60-60 hand added to the next game. */
  carryOver: number;
  handResult?: SjavsHandResult;
  gameEndFlag: boolean;
  /** -1 while the rubber is undecided. */
  winnerTeam: number;
  /** The losers never left 24. */
  doubleVictory: boolean;
  hint?: SjavsHintPayload;
}
