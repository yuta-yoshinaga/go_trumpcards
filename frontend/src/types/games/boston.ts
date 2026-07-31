// Type declarations for boston. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One entry in a hand's auction. */
export interface BostonBid {
  player: number;
  /** Ladder step. 0 is a pass. **Not a number of tricks.** */
  level: number;
  name: string;
  suit: number;
}

/**
 * One rung of the bidding ladder, as sent by the server.
 *
 * The misère bids sit **between** the trick bids, so a client that rebuilds
 * this list from trick counts alone will get the order wrong.
 */
export interface BostonBidOption {
  level: number;
  name: string;
  /** 1 = trick count, 2 = misère, 3 = piccolissimo. */
  kind: number;
  /** Target trick count — 0 for a misère, 1 for Piccolissimo. */
  tricks: number;
  needsTrump: boolean;
  /** Whether the declarer's hand is shown after the first trick. */
  exposed: boolean;
  canCallPartner: boolean;
  payout: number;
}

/** One Boston seat. */
export interface BostonPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /**
   * Your own hand; also the declarer's under an "on the Table" contract once
   * the first trick is done, and everyone's at the settlement.
   */
  cards: Card[];
  tricksWon: number;
  chips: number;
  isDealer: boolean;
  isDeclarer: boolean;
  isPartner: boolean;
  /** True for the declarer and, when one was called, the partner. */
  isDeclarerSide: boolean;
  isCurrentTurn: boolean;
}

/** Full Boston game state returned from the API. */
export interface BostonResponse extends BaseGameResponse {
  players: BostonPlayer[];
  /** 0 = Bid, 1 = CallPartner, 2 = Play, 3 = HandEnd, 4 = GameEnd. */
  phase: number;
  handNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  bids: BostonBid[];
  /** The standing bid, or null while nobody has bid. */
  highBid: BostonBid | null;
  /** The whole ladder in rank order. */
  bidOptions: BostonBidOption[];
  /** Seat that won the auction; -1 while undecided. */
  declarerIdx: number;
  /** Seat called as partner; -1 when the declarer plays alone. */
  partnerIdx: number;
  /** 1 = Spade, 2 = Clover, 3 = Heart, 4 = Diamond; 0 for a no-trump contract. */
  trumpSuit: number;
  exposed: boolean;
  trick: Card[];
  /** Hand indices the human may legally play (following suit is compulsory). */
  validPlays: number[];
  trickLeaderIdx: number;
  trickNumber: number;
  /**
   * Tricks the declaring side has taken — the declarer's plus the partner's,
   * since a called partner's tricks count toward the contract.
   */
  declarerTricks: number;
  bidMade: boolean;
  handSize: number;
  targetHands: number;
  gameEndFlag: boolean;
  winnerIdx: number;
}
