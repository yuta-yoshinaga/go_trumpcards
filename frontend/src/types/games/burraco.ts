// Type declarations for burraco. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Burraco game configuration. */
export interface BurracoConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A single meld on the table in Burraco. */
export interface BurracoMeldData {
  cards: Card[];
  isNatural: boolean;
  isBurraco: boolean;
  rank: number;
}

/** Burraco player data with melds, red 3s, and pozzetto status. */
export interface BurracoPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: BurracoMeldData[];
  red3Count: number;
  red3s: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasBurraco: boolean;
  hasInitMeld: boolean;
  tookPozzetto: boolean;
}

/** Full Burraco game state returned from the API. */
export interface BurracoResponse extends BaseGameResponse {
  players: BurracoPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  /** The full discard pile, oldest (bottom) first. In Burraco the whole pile is
   * taken at once, so its contents are public information for all players. */
  discardPile: Card[];
  drawPileCount: number;
  discardPileCount: number;
  pozzettoCount: number;
  isFrozen: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: BurracoConfig;
}
