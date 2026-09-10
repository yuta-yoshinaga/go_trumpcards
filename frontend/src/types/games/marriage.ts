// Type declarations for marriage. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Marriage player data with scores, deadwood, and pure-sequence flag. */
export interface MarriagePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  deadwood: number;
  hasPureSequence: boolean;
  maal: number;
}

/** Marriage game configuration. */
export interface MarriageConfig {
  playerCount: number;
  cpuDifficulty: number;
  targetRounds: number;
}

/** Full Marriage game state returned from the API. */
export interface MarriageResponse extends BaseGameResponse {
  players: MarriagePlayer[];
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
  humanDeadwood: number;
  humanHasPureSequence: boolean;
  config: MarriageConfig;
}

/** Marriage phases, synchronized with the backend domain values. */
export const MarriagePhase = {
  DRAW: 0,
  DISCARD: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

// --- Machiavelli (マキャヴェッリ) ---
