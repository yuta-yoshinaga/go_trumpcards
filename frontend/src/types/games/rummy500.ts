// Type declarations for rummy500. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Rummy 500 player data with hand, laid melds and scores. */
export interface Rummy500PlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  laidMelds: Rummy500Meld[];
}

/** A meld (set or run) laid by a player in Rummy 500. */
export interface Rummy500Meld {
  cards: Card[];
}

/** Rummy 500 game configuration. */
export interface Rummy500Config {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Rummy 500 game state returned from the API. */
export interface Rummy500Response extends BaseGameResponse {
  players: Rummy500PlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardPile: Card[];
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundEnderIdx: number;
  config: Rummy500Config;
}
