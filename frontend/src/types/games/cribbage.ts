// Type declarations for cribbage. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Cribbage player data with scores. */
export interface CribbagePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
}

/** Cribbage score detail breakdown. */
export interface CribbageScoreDetail {
  fifteens: number;
  pairs: number;
  runs: number;
  flush: number;
  nobs: number;
  total: number;
}

/** Cribbage game configuration. */
export interface CribbageConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Cribbage game state returned from the API. */
export interface CribbageResponse extends BaseGameResponse {
  players: CribbagePlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  crib: Card[];
  starter: Card | null;
  pegCount: number;
  pegPlayedCards: Card[];
  showPhaseStep: number;
  handScoreDetails: (CribbageScoreDetail | null)[];
  gameEndFlag: boolean;
  winnerIdx: number;
  config: CribbageConfig;
}

// --- Oh Hell (オー・ヘル) ---
