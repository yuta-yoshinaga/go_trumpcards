// Type declarations for chinchon. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Chinchón player data with round and cumulative scores and elimination flag. */
export interface ChinchonPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  eliminated: boolean;
}

/** A meld (set or run) laid down by the knocker in Chinchón. */
export interface ChinchonMeld {
  cards: Card[];
}

/** Chinchón game configuration. */
export interface ChinchonConfig {
  cpuDifficulty: number;
  playerCount: number;
  knockThreshold: number;
  eliminationLimit: number;
}

/** Full Chinchón game state returned from the API. */
export interface ChinchonResponse extends BaseGameResponse {
  players: ChinchonPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  knockerIdx: number;
  knockerMelds: ChinchonMeld[];
  config: ChinchonConfig;
}
