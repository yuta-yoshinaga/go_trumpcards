// Type declarations for fiftyone. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Single Fifty-one player state from the API. */
export interface FiftyOnePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  score: number;
}

/** Fifty-one game configuration. */
export interface FiftyOneConfig {
  cpuDifficulty: number;
}

/** Full Fifty-one game state returned from the API. */
export interface FiftyOneResponse extends BaseGameResponse {
  players: FiftyOnePlayerData[];
  tableCards: Card[];
  phase: number;
  currentTurn: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  turnNumber: number;
  stopCallerIdx: number;
  lastAction: string;
  lastHandIdx: number;
  lastTableIdx: number;
  config: FiftyOneConfig;
}

// --- Yukon (ユーコン) ---
