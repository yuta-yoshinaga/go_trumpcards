// Type declarations for prsi. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Prší player data. */
export interface PrsiPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
}

/** Prší game configuration (no point limit — first to empty hand wins). */
export interface PrsiConfig {
  cpuDifficulty: number;
}

/** Full Prší game state returned from the API. */
export interface PrsiResponse extends BaseGameResponse {
  players: PrsiPlayerData[];
  phase: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  penaltyDrawCount: number;
  pendingSkips: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: PrsiConfig;
}

// --- Macau (マカオ) ---
