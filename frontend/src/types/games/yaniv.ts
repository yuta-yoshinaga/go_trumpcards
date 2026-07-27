// Type declarations for yaniv. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Yaniv player data with cumulative penalty score and revealed hand total. */
export interface YanivPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  score: number;
  handTotal: number;
  isEliminated: boolean;
}

/** Yaniv game configuration. */
export interface YanivConfig {
  cpuDifficulty: number;
  scoreLimit: number;
}

/** Full Yaniv game state returned from the API. */
export interface YanivResponse extends BaseGameResponse {
  players: YanivPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  pickupCards: Card[];
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  callerIdx: number;
  asafWinnerIdx: number;
  isAsaf: boolean;
  roundScores: number[];
  config: YanivConfig;
}

// --- Seven Bridge (セブンブリッジ) ---
