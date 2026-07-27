// Type declarations for shithead. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A Shithead player's per-game state. */
export interface ShitheadPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  handCount: number;
  handCards: Card[];
  faceUpCards: Card[];
  faceDownCount: number;
}

/** A single Shithead action (play or pickup). */
export interface ShitheadAction {
  playerIdx: number;
  source: string;
  playedCards: Card[];
  pickup: boolean;
  burned: boolean;
  skipped: boolean;
}

/** Shithead local rule configuration. */
export interface ShitheadConfig {
  magicTwo: boolean;
  magicSeven: boolean;
  magicEight: boolean;
  magicTen: boolean;
  fourOfAKindBurn: boolean;
  cpuDifficulty: number;
}

/** Full Shithead game state returned from the API. */
export interface ShitheadResponse extends BaseGameResponse {
  players: ShitheadPlayerData[];
  currentTurn: number;
  currentSource: string;
  discardPile: Card[];
  stockSize: number;
  skipNext: boolean;
  sevenActive: boolean;
  gameEndFlag: boolean;
  config: ShitheadConfig;
  cpuActions: ShitheadAction[];
  humanAction?: ShitheadAction;
}

// --- Spider Solitaire (スパイダーソリティア) ---
