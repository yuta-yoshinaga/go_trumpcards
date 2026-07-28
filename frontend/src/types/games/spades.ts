// Type declarations for spades. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Spades player data with bid, scores, and bags. */
export interface SpadesPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  bags: number;
}

/** A card played in a Spades trick. */
export interface SpadesTrickCard {
  playerIdx: number;
  card: Card;
}

/** Spades game configuration. */
export interface SpadesConfig {
  cpuDifficulty: number;
  pointLimit: number;
  nilBonus: number;
  bagPenaltyThreshold: number;
}

/** A suggested hint for Spades. */
export interface SpadesHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Spades game state returned from the API. */
export interface SpadesResponse extends BaseGameResponse {
  players: SpadesPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  currentTrick: SpadesTrickCard[];
  spadesBroken: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  config: SpadesConfig;
  hint?: SpadesHint;
}

// --- Call Break ---
