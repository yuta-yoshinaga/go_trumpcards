// Type declarations for speed. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Speed player data with hand and draw pile info. */
export interface SpeedPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  drawPileSize: number;
}

/** Speed CPU action record. */
export interface SpeedCpuAction {
  cardIndex: number;
  pileIndex: number;
}

/** Speed hint information. */
export interface SpeedHint {
  cardIndex: number;
  pileIndex: number;
  found: boolean;
}

/** Speed game configuration. */
export interface SpeedConfig {
  cpuDifficulty: number;
  autoFlip: boolean;
}

/** Full Speed game state returned from the API. */
export interface SpeedResponse extends BaseGameResponse {
  players: SpeedPlayerData[];
  centerPiles: Card[];
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  cpuActions?: SpeedCpuAction[];
  hint?: SpeedHint;
  config: SpeedConfig;
}
