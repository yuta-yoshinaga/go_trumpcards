// Type declarations for guandan. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Guandan seat. */
export interface GuandanPlayer {
  id: number;
  isHuman: boolean;
  /** 0 for seats 0/2, 1 for seats 1/3 — **partners sit opposite**. */
  team: number;
  cardCount: number;
  /** Your own hand only, until the hand is settled. */
  cards: Card[];
  /**
   * 1-4 once this seat has gone out, 0 while it still holds cards. The order
   * decides **both** the next hand's tribute and how far the winners climb.
   */
  finishedRank: number;
  isCurrentTurn: boolean;
}

/** The combination sitting on the table. */
export interface GuandanCombo {
  /**
   * 0 = none, 1 = single, 2 = pair, 3 = triple, 4 = full house, 5 = straight,
   * 6 = plate, 7 = tube, 8 = bomb, 9 = straight flush, 10 = joker bomb.
   */
  kind: number;
  rank: number;
  size: number;
}

/** One tribute payment for this hand. */
export interface GuandanTribute {
  from: number;
  to: number;
  /** The card paid — the payer's highest, **wilds excluded**. */
  card: Card | null;
  /** The card handed back, or null while the return is still owed. */
  returned: Card | null;
}

/** How the previous hand settled. */
export interface GuandanHandResult {
  /** Seats in finishing order. */
  order: number[];
  winnerTeam: number;
  /** Levels climbed. **One of 1, 2 or 4** — never 3. */
  advance: number;
  /** True when the winners took first *and* second, the only way to climb 4. */
  firstSecond: boolean;
}

/** Full Guandan game state returned from the API. */
export interface GuandanResponse extends BaseGameResponse {
  players: GuandanPlayer[];
  /** 0 = Tribute, 1 = Play, 2 = HandEnd, 3 = GameEnd. */
  phase: number;
  handNumber: number;
  currentPlayerIdx: number;
  /**
   * This hand's level rank. **Cards of this rank beat aces** and lose only to
   * the jokers, and the two hearts among them are wild.
   */
  level: number;
  /** Each team's current level, 2 through 14 (ace). */
  teamLevels: [number, number];
  /** The team whose level is being played. */
  declarerTeam: number;
  lastCombo: GuandanCombo | null;
  lastPlayerIdx: number;
  /** Seats in the order they went out this hand. */
  finished: number[];
  tributes: GuandanTribute[];
  /** True when a payer held both red jokers, which cancels tribute outright. */
  tributeCancelled: boolean;
  lastResult: GuandanHandResult | null;
  minLevel: number;
  /** The last level, an ace — winning here ends the game. */
  maxLevel: number;
  /** Levels gained by taking first and second: **4**. */
  advanceFirstSecond: number;
  /** Levels gained by taking first and third: 2. */
  advanceFirstThird: number;
  /** Levels gained by taking first and fourth: 1. */
  advanceFirstFourth: number;
  gameEndFlag: boolean;
  /** Winning team; -1 while the game is live. */
  winnerTeam: number;
  config: GuandanConfigOutput;
}

/** Settings echoed back with the game state. */
export interface GuandanConfigOutput {
  cpuDifficulty: number;
}
