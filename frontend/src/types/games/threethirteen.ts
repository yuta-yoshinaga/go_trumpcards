// Type declarations for threethirteen. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Three Thirteen player data with deadwood, round, and cumulative scores. */
export interface ThreeThirteenPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  deadwood: number;
  roundScore: number;
  cumulativeScore: number;
}

/** Three Thirteen game configuration. */
export interface ThreeThirteenConfig {
  cpuDifficulty: number;
  playerCount: number;
}

/** Full Three Thirteen game state returned from the API. */
export interface ThreeThirteenResponse extends BaseGameResponse {
  players: ThreeThirteenPlayerData[];
  phase: number;
  round: number;
  wildRank: number;
  dealCount: number;
  currentPlayerIdx: number;
  knockerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: ThreeThirteenConfig;
}

// --- Memory (神経衰弱) ---
