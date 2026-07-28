// Type declarations for memory. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Memory player data with pair count and captured-pair representative cards. */
export interface MemoryPlayerData {
  id: number;
  isHuman: boolean;
  pairCount: number;
  /** One representative card per captured pair, in ascending rank order. */
  pairs: Card[];
}

/** A card on the Memory game board. */
export interface MemoryBoardCard {
  card: Card | null;
  faceUp: boolean;
  taken: boolean;
}

/** Memory game configuration. */
export interface MemoryConfig {
  cpuDifficulty: number;
  /**
   * Pairs dealt to the board, 6-26. 26 is the full deck. Narrow screens start
   * lower because 52 cards cannot fit a 375x667 viewport while every card keeps a
   * 44x44 tap target — see ADR-0035.
   */
  pairCount: number;
}

/** Full Memory game state returned from the API. */
export interface MemoryResponse extends BaseGameResponse {
  players: MemoryPlayerData[];
  board: MemoryBoardCard[];
  phase: number;
  currentPlayerIdx: number;
  firstFlipPos: number;
  secondFlipPos: number;
  lastMatchResult: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  turnNumber: number;
  config: MemoryConfig;
}

// --- Klondike (ソリティア) ---
