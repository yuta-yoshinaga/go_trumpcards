// Type declarations for ninetynine. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Ninety-Nine player data with scores. */
export interface NinetyNinePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  buriedCount: number;
}

/** A card played in a Ninety-Nine trick. */
export interface NinetyNineTrickCard {
  playerIdx: number;
  card: Card;
}

/** Ninety-Nine game configuration. */
export interface NinetyNineConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Ninety-Nine. */
export interface NinetyNineHint {
  cardIndex?: number;
  buryIndices?: number[];
  reason: string;
}

/** Full Ninety-Nine game state returned from the API. */
export interface NinetyNineResponse extends BaseGameResponse {
  players: NinetyNinePlayerData[];
  phase: number;
  dealNumber: number;
  targetScore: number;
  handSize: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  currentTrick: NinetyNineTrickCard[];
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  hint?: NinetyNineHint;
  config: NinetyNineConfig;
}

// --- Three Card Poker (スリーカードポーカー) ---
