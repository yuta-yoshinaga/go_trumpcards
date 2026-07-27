// Type declarations for twotenjack. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Two Ten Jack player data (4-player team game: seats 0,2 vs 1,3). */
export interface TwoTenJackPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  capturedPoints: number;
}

/** A card played in a Two Ten Jack trick. */
export interface TwoTenJackTrickCard {
  playerIdx: number;
  card: Card;
}

/** Two Ten Jack game configuration. */
export interface TwoTenJackConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Two Ten Jack. */
export interface TwoTenJackHint {
  cardIndex?: number;
  trumpSuit?: number;
  reason: string;
}

/** Full Two Ten Jack game state returned from the API. */
export interface TwoTenJackResponse extends BaseGameResponse {
  players: TwoTenJackPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  declarerIdx: number;
  trumpSuit: number;
  currentTrick: TwoTenJackTrickCard[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: TwoTenJackConfig;
  hint?: TwoTenJackHint;
}

// --- Crazy Eights (クレイジーエイト) ---
