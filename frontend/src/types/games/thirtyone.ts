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
  /**
   * Hand totals at which each difficulty's CPUs consider knocking.
   *
   * The difficulty setting *is* these numbers, so they come from the domain
   * rather than being written into a translated string that would go stale
   * the moment a constant moves (#5623).
   */
  knockThresholds: {
    easy: number;
    normal: number;
    hard: number;
  };
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
