// Type declarations for sixbidsolo. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One entry in a hand's auction. */
export interface SixBidSoloBid {
  player: number;
  /**
   * 0 = pass, 1 = Solo, 2 = Heart Solo, 3 = Misère, 4 = Guarantee Solo,
   * 5 = Spread Misère, 6 = Call Solo — the six bids in ascending order.
   */
  kind: number;
}

/** One Six-Bid Solo seat. */
export interface SixBidSoloPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /**
   * Your own hand only, until the settlement. **A spread misère also exposes
   * the declarer's hand** — that is what the bid is buying.
   */
  cards: Card[];
  /** Card points taken this hand. */
  points: number;
  tricksWon: number;
  /** Running total across hands. */
  score: number;
  isDealer: boolean;
  isDeclarer: boolean;
  isCurrentTurn: boolean;
}

/** Breakdown of one hand's settlement. */
export interface SixBidSoloHandResult {
  kind: number;
  declarer: number;
  /** Card points the declarer took, **including the widow**. */
  declarerPoints: number;
  /** What the widow added. **Zero at either misère** — it is excluded there. */
  widowPoints: number;
  target: number;
  made: boolean;
  /** What each opponent pays or receives. */
  value: number;
  deltas: [number, number, number];
}

/** Full Six-Bid Solo game state returned from the API. */
export interface SixBidSoloResponse extends BaseGameResponse {
  players: SixBidSoloPlayer[];
  /** 0 = Bid, 1 = Declare, 2 = Play, 3 = HandEnd, 4 = GameEnd. */
  phase: number;
  handNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  bids: SixBidSoloBid[];
  /** The standing bid, or null while nobody has bid. */
  highBid: SixBidSoloBid | null;
  /** Seat that won the auction; -1 while undecided. */
  declarerIdx: number;
  /** 1 = Spade, 2 = Clover, 3 = Heart, 4 = Diamond; 0 at either misère. */
  trumpSuit: number;
  declared: boolean;
  /** The card a call solo named; its holder had to exchange it. */
  calledCard: Card | null;
  /** Whether a spread misère has laid the declarer's hand down. */
  spreadOpen: boolean;
  /** Populated only once the hand settles — it stays face down during play. */
  widow: Card[];
  /** How many cards are face down (three), sent even while they are concealed. */
  widowSize: number;
  trick: Card[];
  /** Hand indices the human may legally play (following suit is compulsory). */
  validPlays: number[];
  trickLeaderIdx: number;
  trickNumber: number;
  lastResult: SixBidSoloHandResult | null;
  /**
   * Target card points per bid, indexed by kind. **A plain bid needs 61, not
   * 60**, and guarantee solo depends on the trump suit (74 at hearts, 80
   * elsewhere), so the server sends the resolved table.
   */
  bidTargets: number[];
  /** Card points in play (120). */
  totalPoints: number;
  /** The base a plain bid must EXCEED (60). */
  baseTarget: number;
  /** Cards per hand — **eleven, with three more in the widow**. */
  handSize: number;
  /** Hands in a game. */
  targetHands: number;
  gameEndFlag: boolean;
  /** Winning seat; -1 while the game is live. */
  winnerIdx: number;
  config: SixBidSoloConfigOutput;
}

/** Settings echoed back with the game state. */
export interface SixBidSoloConfigOutput {
  cpuDifficulty: number;
  targetHands: number;
}
