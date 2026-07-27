// Type declarations for conquian. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A table meld (set or run) in Conquian. */
export interface ConquianMeld {
  cards: Card[];
}

/** Conquian player data with face-up table melds and rounds won. */
export interface ConquianPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: ConquianMeld[];
  wins: number;
}

/** Conquian game configuration. */
export interface ConquianConfig {
  cpuDifficulty: number;
  targetWins: number;
}

/** Full Conquian game state returned from the API. */
export interface ConquianResponse extends BaseGameResponse {
  players: ConquianPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  tookDiscard: boolean;
  config: ConquianConfig;
}

// --- Chinchón (チンチョン) ---
