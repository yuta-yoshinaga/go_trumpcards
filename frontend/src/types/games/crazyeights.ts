// Type declarations for crazyeights. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Crazy Eights player data with scores. */
export interface CrazyEightsPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
}

/** Crazy Eights game configuration. */
export interface CrazyEightsConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Crazy Eights game state returned from the API. */
export interface CrazyEightsResponse extends BaseGameResponse {
  players: CrazyEightsPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  chosenSuit: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: CrazyEightsConfig;
}

// --- Prší (プルシー) ---
