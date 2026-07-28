// Type declarations for machiavelli. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Machiavelli player data with scores and deadwood. */
export interface MachiavelliPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  deadwood: number;
}

/** A meld on the shared Machiavelli table (kind: 0=set, 1=run). */
export interface MachiavelliMeld {
  cards: Card[];
  kind: number;
}

/** Machiavelli game configuration. */
export interface MachiavelliConfig {
  playerCount: number;
  cpuDifficulty: number;
  targetRounds: number;
}

/** Full Machiavelli game state returned from the API. */
export interface MachiavelliResponse extends BaseGameResponse {
  players: MachiavelliPlayer[];
  table: MachiavelliMeld[];
  phase: number;
  roundNumber: number;
  targetRounds: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  config: MachiavelliConfig;
}

// --- Panguingue / Pan (パングインゲ) ---
