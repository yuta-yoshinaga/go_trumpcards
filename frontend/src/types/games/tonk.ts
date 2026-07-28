// Type declarations for tonk. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tonk player data with scores. */
export interface TonkPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
}

/** A meld (set or run) in Tonk. */
export interface TonkMeld {
  cards: Card[];
}

/** Tonk game configuration. */
export interface TonkConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Tonk game state returned from the API. */
export interface TonkResponse extends BaseGameResponse {
  players: TonkPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  knockerIdx: number;
  knockerMelds: TonkMeld[];
  knockerDeadwood: Card[];
  opponentMelds: TonkMeld[];
  opponentDeadwood: Card[];
  isTonk: boolean;
  isUndercut: boolean;
  config: TonkConfig;
}

// --- Thirty-One (サーティワン / Scat) ---
