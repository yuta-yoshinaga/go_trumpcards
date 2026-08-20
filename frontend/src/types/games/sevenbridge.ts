// Type declarations for sevenbridge. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A meld (set or run) shared across players in Seven Bridge. */
export interface SevenBridgeMeld {
  cards: Card[];
}

/** Seven Bridge player data with hand, melds and scores. */
export interface SevenBridgePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: SevenBridgeMeld[];
  roundScore: number;
  cumulativeScore: number;
}

/** Seven Bridge game configuration. */
export interface SevenBridgeConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Seven Bridge game state returned from the API. */
export interface SevenBridgeResponse extends BaseGameResponse {
  players: SevenBridgePlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  /** Whether this turn took the discard with a pon/chi claim rather than drawing (#5547). */
  claimedThisTurn?: boolean;
  config: SevenBridgeConfig;
}

// --- Rummy 500 ---
