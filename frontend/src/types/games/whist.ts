// Type declarations for whist. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Whist player data with team, scores, and trick count. */
export interface WhistPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  team: number;
}

/** A card played in a Whist trick. */
export interface WhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** Whist game configuration. */
export interface WhistConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Whist. */
export interface WhistHint {
  cardIndex?: number;
  reason: string;
}

/** Full Whist game state returned from the API. */
export interface WhistResponse extends BaseGameResponse {
  players: WhistPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: WhistTrickCard[];
  trumpSuit: number;
  dealerIdx: number;
  teamScores: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: WhistConfig;
  hint?: WhistHint;
}

// --- Catch the Ten (スコッチ・ホイスト) ---
