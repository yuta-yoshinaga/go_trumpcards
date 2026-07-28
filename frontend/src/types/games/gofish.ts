// Type declarations for gofish. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Go Fish player data with hand, book count, and completed books. */
export interface GoFishPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bookCount: number;
  books: GoFishBook[];
}

/** A completed 4-of-a-kind book in Go Fish. */
export interface GoFishBook {
  rank: number;
  cards: Card[];
}

/** CPU action record in Go Fish. */
export interface GoFishCpuAction {
  askPlayerIdx: number;
  askTargetIdx: number;
  askRank: number;
  success: boolean;
  cardsReceived: number;
  drawnCard: Card | null;
  bookFormed: boolean;
  bookRank: number;
}

/** Information about the last ask action in Go Fish. */
export interface GoFishLastAsk {
  playerIdx: number;
  targetIdx: number;
  rank: number;
  success: boolean;
  cardsReceived: Card[];
  drawnCard: Card | null;
  bookFormed: boolean;
  bookRank: number;
}

/** Go Fish game configuration. */
export interface GoFishConfig {
  cpuDifficulty: number;
}

/** Full Go Fish game state returned from the API. */
export interface GoFishResponse extends BaseGameResponse {
  players: GoFishPlayerData[];
  phase: number;
  currentTurn: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  turnNumber: number;
  deckRemaining: number;
  lastAsk: GoFishLastAsk | null;
  cpuActions: GoFishCpuAction[];
  humanAction: GoFishCpuAction | null;
  config: GoFishConfig;
}

// --- Canasta (カナスタ) ---
