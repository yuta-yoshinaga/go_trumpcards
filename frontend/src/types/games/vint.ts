// Type declarations for vint. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One entry in a hand's auction. */
export interface VintBid {
  player: number;
  /** Bid level 1-7, contracting for 6 + level tricks. 0 is a pass. */
  level: number;
  /** 0 = Spade, 1 = Club, 2 = Diamond, 3 = Heart, 4 = NoTrump — spades are LOWEST. */
  denom: number;
  trickValue: number;
}

/** One Vint seat. */
export interface VintPlayer {
  id: number;
  isHuman: boolean;
  /** 0 for seats 0 and 2, 1 for seats 1 and 3 — partners sit opposite. */
  team: number;
  cardCount: number;
  /**
   * Your own hand only until the settlement. **There is no dummy** — unlike
   * bridge, even your partner's hand stays concealed.
   */
  cards: Card[];
  tricksWon: number;
  isDealer: boolean;
  isDeclarer: boolean;
  isCurrentTurn: boolean;
}

/** Breakdown of one hand's settlement. */
export interface VintHandResult {
  /**
   * What each team scored below the line for its own tricks. **Both sides
   * score**, whether or not the contract was made.
   */
  trickPoints: [number, number];
  /**
   * Trump A K Q J 10. Scores only from **three or more** (20x / 30x / 40x the
   * trick value); a no-trump contract has none.
   */
  honourPoints: [number, number];
  /**
   * Scored separately from honours: the side with the majority takes 10x the
   * trick value per ace, and a 2-2 split goes to the side with more tricks.
   */
  acePoints: [number, number];
  /** Undertricks × level × 500, scored above the opponents' line. */
  penalty: [number, number];
  made: boolean;
  declarerTricks: number;
  trickValue: number;
}

/** Full Vint game state returned from the API. */
export interface VintResponse extends BaseGameResponse {
  players: VintPlayer[];
  /** 0 = Bid, 1 = Play, 2 = HandEnd, 3 = GameEnd. */
  phase: number;
  handNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  bids: VintBid[];
  /** The standing bid, or null while nobody has bid. */
  highBid: VintBid | null;
  /** Seat that won the auction; -1 while undecided. */
  declarerIdx: number;
  /** 1 = Spade, 2 = Clover, 3 = Heart, 4 = Diamond; 0 at no trump. */
  trumpSuit: number;
  trick: Card[];
  /** Hand indices the human may legally play (following suit is compulsory). */
  validPlays: number[];
  trickLeaderIdx: number;
  trickNumber: number;
  teamTricks: [number, number];
  /**
   * Below-the-line totals by team. **Both teams score here for the tricks they
   * took**, whether or not the contract was made. Reset when a team reaches the
   * target and takes a game.
   */
  below: [number, number];
  /** Above-the-line totals — honours, aces, penalties and game bonuses. */
  above: [number, number];
  /** Games each team has taken. Two takes the rubber. */
  gamesWon: [number, number];
  lastResult: VintHandResult | null;
  /**
   * Level-1 trick value per denomination, indexed by denom (spades 4, clubs 6,
   * diamonds 8, hearts 10, no trump 12). Each level above adds 10.
   */
  trickValues: number[];
  /** Below-the-line points needed for a game (500). */
  gameTarget: number;
  minLevel: number;
  maxLevel: number;
  gameEndFlag: boolean;
  /** Team that took the rubber; -1 while it is live. */
  winnerTeam: number;
}
