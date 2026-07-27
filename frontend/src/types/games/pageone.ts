// Type declarations for pageone. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Page One player data with scores. */
export interface PageOnePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasDeclared: boolean;
}

/** Page One game configuration. */
export interface PageOneConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Page One game state returned from the API. */
export interface PageOneResponse extends BaseGameResponse {
  players: PageOnePlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: PageOneConfig;
}

// --- Gin Rummy (ジンラミー) ---
