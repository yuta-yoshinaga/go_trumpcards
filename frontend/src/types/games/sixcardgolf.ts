// Type declarations for sixcardgolf. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Six Card Golf grid slot. */
export interface SixCardGolfSlot {
  card: Card | null;
  faceUp: boolean;
}

/** Six Card Golf player data. */
export interface SixCardGolfPlayerData {
  id: number;
  isHuman: boolean;
  grid: SixCardGolfSlot[];
  roundScore: number;
  cumulativeScore: number;
  allFaceUp: boolean;
}

/** Six Card Golf game config. */
export interface SixCardGolfConfig {
  playerCount: number;
  cpuDifficulty: number;
  rounds: number;
}

/** Six Card Golf API response. */
export interface SixCardGolfResponse extends BaseGameResponse {
  players: SixCardGolfPlayerData[];
  phase: number;
  roundNumber: number;
  totalRounds: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  drawnCard: Card | null;
  drawnFromDiscard: boolean;
  canFlip: boolean;
  finalTurnTrigger: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: SixCardGolfConfig;
}
