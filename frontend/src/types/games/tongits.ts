// Type declarations for tongits. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tongits player data with scores. */
export interface TongitsPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: TongitsMeld[];
  roundScore: number;
  cumulativeScore: number;
}

/** A meld (set or run) in Tongits. */
export interface TongitsMeld {
  cards: Card[];
}

/** Tongits game configuration. */
export interface TongitsConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Tongits game state returned from the API. */
export interface TongitsResponse extends BaseGameResponse {
  players: TongitsPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  isTongits: boolean;
  /** Reason the round ended: 0 none, 1 Tongits, 2 challenge, 3 stock out. */
  roundEndReason: number;
  /** Round winner index, or -1 for a draw. */
  roundWinner: number;
  /** Remaining hand points for the human on their discard turn, or -1. */
  remainingPoints: number;
  config: TongitsConfig;
}

// --- Thirty-One (サーティワン / Scat) ---
