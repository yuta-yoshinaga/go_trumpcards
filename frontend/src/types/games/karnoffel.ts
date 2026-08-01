// Type declarations for karnoffel. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Karnöffel seat. */
export interface KarnoffelPlayer {
  id: number;
  isHuman: boolean;
  /** 0 for seats 0 and 2, 1 for seats 1 and 3 — partners sit opposite. */
  team: number;
  cardCount: number;
  /** Your own hand only, until the hand is over. */
  cards: Card[];
  /**
   * The single card dealt face up in front of this seat. **The lowest of the
   * four decides the chosen suit**, so every seat's is public.
   */
  upCard: Card | null;
  tricksWon: number;
  isDealer: boolean;
  isCurrentTurn: boolean;
}

/** How one hand finished. */
export interface KarnoffelHandResult {
  /** The team that took the hand; -1 if neither reached three tricks. */
  winnerTeam: number;
  tricks: [number, number];
  chosenSuit: number;
}

/** Full Karnöffel game state returned from the API. */
export interface KarnoffelResponse extends BaseGameResponse {
  players: KarnoffelPlayer[];
  /** 0 = Play, 1 = HandEnd, 2 = GameEnd. */
  phase: number;
  handNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  /** 1 = Spade, 2 = Clover, 3 = Heart, 4 = Diamond. */
  chosenSuit: number;
  trick: Card[];
  /**
   * Hand indices the human may legally play. **There is no obligation to
   * follow suit**, but the devil cannot lead the first trick, so the server
   * decides this.
   */
  validPlays: number[];
  trickLeaderIdx: number;
  trickNumber: number;
  teamTricks: [number, number];
  /** Hands each team has taken. */
  handsWon: [number, number];
  lastResult: KarnoffelHandResult | null;
  /** Tricks that take a hand (three). */
  tricksToWin: number;
  /** Cards per hand — **five**, not twelve. */
  handSize: number;
  /** Hands needed to win the game. */
  targetHands: number;
  gameEndFlag: boolean;
  /** Winning team; -1 while the game is live. */
  winnerTeam: number;
  config: KarnoffelConfigOutput;
}

/** Settings echoed back with the game state. */
export interface KarnoffelConfigOutput {
  cpuDifficulty: number;
  targetHands: number;
}
