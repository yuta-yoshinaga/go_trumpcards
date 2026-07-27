// Type declarations for indianrummy. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Indian Rummy player data with scores, deadwood, and pure-sequence flag. */
export interface IndianRummyPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  deadwood: number;
  hasPureSequence: boolean;
}

/** Indian Rummy game configuration. */
export interface IndianRummyConfig {
  playerCount: number;
  cpuDifficulty: number;
  targetRounds: number;
}

/** Full Indian Rummy game state returned from the API. */
export interface IndianRummyResponse extends BaseGameResponse {
  players: IndianRummyPlayer[];
  phase: number;
  roundNumber: number;
  targetRounds: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  wildJoker: Card | null;
  wildRank: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  declarerIdx: number;
  declarationValid: boolean;
  config: IndianRummyConfig;
}

// --- Machiavelli (マキャヴェッリ) ---
