// Type declarations for bideuchre. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One entry in a hand's auction. */
export interface BidEuchreBid {
  player: number;
  /** Tricks bid, 3-6. 0 is a pass. */
  value: number;
}

/** One Bid Euchre seat. */
export interface BidEuchrePlayer {
  id: number;
  isHuman: boolean;
  /** 0 for seats 0 and 2, 1 for seats 1 and 3 — partners sit opposite. */
  team: number;
  cardCount: number;
  /**
   * Your own hand only until the settlement. **There is no kitty** — the whole
   * 24-card pack is dealt out, so nothing else is ever face down.
   */
  cards: Card[];
  tricksWon: number;
  isDealer: boolean;
  isDeclarer: boolean;
  isCurrentTurn: boolean;
}

/** Breakdown of one hand's settlement. */
export interface BidEuchreHandResult {
  /**
   * What each team scored. A set costs the declaring side **its bid, not the
   * tricks it took**; the defenders score their own tricks either way.
   */
  points: [number, number];
  tricks: [number, number];
  made: boolean;
  bid: number;
}

/** Full Bid Euchre game state returned from the API. */
export interface BidEuchreResponse extends BaseGameResponse {
  players: BidEuchrePlayer[];
  /** 0 = Bid, 1 = ChooseTrump, 2 = Play, 3 = HandEnd, 4 = GameEnd. */
  phase: number;
  handNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  /** The dealer alone may take the contract by **equalling** the standing bid. */
  dealerIdx: number;
  bids: BidEuchreBid[];
  /** The standing bid, or null while nobody has bid. */
  highBid: BidEuchreBid | null;
  /** Seat that won the auction; -1 while undecided. */
  declarerIdx: number;
  /**
   * The declaration itself: 0 = Spade, 1 = Club, 2 = Diamond, 3 = Heart,
   * 4 = NoTrump high, 5 = **NoTrump LOW** (the ranking reverses and the nine is
   * highest). `trumpSuit` alone cannot tell the two no-trump forms apart.
   */
  trump: number;
  /** 1 = Spade, 2 = Clover, 3 = Heart, 4 = Diamond; 0 at either no trump. */
  trumpSuit: number;
  trumpChosen: boolean;
  trick: Card[];
  /**
   * Hand indices the human may legally play. Following suit is compulsory and
   * the **left bower counts as a trump**, so the server decides this.
   */
  validPlays: number[];
  trickLeaderIdx: number;
  trickNumber: number;
  teamTricks: [number, number];
  /** Running totals by team. */
  scores: [number, number];
  lastResult: BidEuchreHandResult | null;
  /** Points needed to win (32). */
  gameTarget: number;
  /** The floor, three tricks. */
  minBid: number;
  maxBid: number;
  /** Six — the whole 24-card pack is dealt out, leaving no kitty. */
  handSize: number;
  gameEndFlag: boolean;
  /** Winning team; -1 while the game is live. */
  winnerTeam: number;
  config: BidEuchreConfigOutput;
}

/** Settings echoed back with the game state. */
export interface BidEuchreConfigOutput {
  cpuDifficulty: number;
  /** Whether the declarer may name a no-trump form. */
  allowNoTrump: boolean;
}
