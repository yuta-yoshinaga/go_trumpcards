// Type declarations for canasta. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Canasta game configuration. */
export interface CanastaConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A single meld on the table in Canasta. */
export interface CanastaMeldData {
  cards: Card[];
  isNatural: boolean;
  isCanasta: boolean;
  rank: number;
}

/** Server-calculated breakdown of a player's score for the completed round. */
export interface CanastaScoreBreakdown {
  meldCards: number;
  canastaBonus: number;
  red3Bonus: number;
  goOutBonus: number;
  handPenalty: number;
}

/** Canasta player data with melds and red 3s. */
export interface CanastaPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: CanastaMeldData[];
  red3Count: number;
  red3s: Card[];
  roundScore: number;
  cumulativeScore: number;
  scoreBreakdown: CanastaScoreBreakdown;
  hasCanasta: boolean;
  hasInitMeld: boolean;
}

/** Full Canasta game state returned from the API. */
export interface CanastaResponse extends BaseGameResponse {
  /** Server-calculated minimum for the human player’s initial meld. */
  minMeld: number;
  players: CanastaPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  /** Discard pile ordered oldest (bottom) first, current top last. */
  discardPile: Card[];
  drawPileCount: number;
  discardPileCount: number;
  isFrozen: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: CanastaConfig;
}

// --- Samba (サンバ) ---
