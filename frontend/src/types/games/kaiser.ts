// Type declarations for kaiser. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One entry in a hand's bidding. */
export interface KaiserBid {
  player: number;
  /** Points bid — **not** a number of tricks. A pass is recorded as 0. */
  value: number;
  /** 0 = with trump, 1 = no trump, 2 = low no trump. */
  contract: number;
}

/** One Kaiser seat. */
export interface KaiserPlayer {
  id: number;
  isHuman: boolean;
  /** 0 for seats 0 and 2, 1 for seats 1 and 3 — partners sit opposite. */
  team: number;
  cardCount: number;
  /** Your own hand only; the opponents' stay empty until the settlement. */
  cards: Card[];
  isDealer: boolean;
  isDeclarer: boolean;
  isCurrentTurn: boolean;
}

/** Full Kaiser game state returned from the API. */
export interface KaiserResponse extends BaseGameResponse {
  players: KaiserPlayer[];
  /** 0 = Bid, 1 = Discard, 2 = Play, 3 = HandEnd, 4 = GameEnd. */
  phase: number;
  handNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  bids: KaiserBid[];
  /** The standing bid, or null while nobody has bid. */
  highBid: KaiserBid | null;
  /** Seat that won the auction; -1 while undecided. */
  declarerIdx: number;
  /** 1 = Spade, 2 = Clover, 3 = Heart, 4 = Diamond; 0 for no trump or before it is named. */
  trumpSuit: number;
  /** 0 = with trump, 1 = no trump, 2 = low no trump (ranking reverses, seven high). */
  contract: number;
  /**
   * Cards still face down in the kitty — two at the deal, zero once the
   * declarer takes them. A 32-card pack could not produce one at all.
   */
  kittySize: number;
  trick: Card[];
  trickLeaderIdx: number;
  trickNumber: number;
  /**
   * Hand indices the human may legally play. Sent by the server because
   * following suit is compulsory (trumping is not).
   */
  validPlays: number[];
  /**
   * Points each team took this hand. The two always sum to 10 — eight tricks
   * plus 5 for the ♥5 minus 3 for the ♠3.
   */
  teamHandPoints: [number, number];
  teamScores: [number, number];
  /** Seat that took the +5 ♥5; -1 while it is unplayed. */
  heartFiveBy: number;
  /** Seat that took the −3 ♠3; -1 while it is unplayed. */
  spadeThreeBy: number;
  /**
   * Whether the declaring side reached its bid. When false the side loses the
   * **bid amount** rather than scoring what it took.
   */
  bidMade: boolean;
  /** 52, raised to 62 once any no-trump bid succeeds. */
  targetScore: number;
  /** 7 — the floor is raised from 6 because the kitty is such an advantage. */
  minBid: number;
  maxBid: number;
  gameEndFlag: boolean;
  /** Winning team; -1 while the game is live. */
  winnerTeam: number;
}
