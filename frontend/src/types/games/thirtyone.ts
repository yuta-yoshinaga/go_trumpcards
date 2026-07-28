// Type declarations for thirtyone. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Thirty-One player data with lives and best-suit score. */
export interface ThirtyOnePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  lives: number;
  score: number;
  isEliminated: boolean;
}

/** Thirty-One game configuration. */
export interface ThirtyOneConfig {
  cpuDifficulty: number;
  initialLives: number;
}

/** Full Thirty-One game state returned from the API. */
export interface ThirtyOneResponse extends BaseGameResponse {
  players: ThirtyOnePlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  knockerIdx: number;
  thirtyOneIdx: number;
  roundWinnerIdx: number;
  roundLosers: number[];
  config: ThirtyOneConfig;
}

// --- Yaniv (ヤニブ) ---
