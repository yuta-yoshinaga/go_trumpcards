// Type declarations for hearts. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Hearts player data with scores and trick count. */
export interface HeartsPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  /** Captured penalty cards so far: every heart plus the Q♠ (J♦ excluded). */
  penaltyCards: Card[];
  /**
   * True when the omnibus rule is on and this player captured the J♦ (-10).
   * The J♦ is a bonus, not a penalty, so it is deliberately absent from
   * `penaltyCards` -- without this flag nothing on screen showed who took it
   * (#5491). Always false while the rule is off.
   */
  tookOmnibusJD: boolean;
}

/** A card played in a Hearts trick. */
export interface HeartsTrickCard {
  playerIdx: number;
  card: Card;
}

/** Hearts game configuration. */
export interface HeartsConfig {
  cpuDifficulty: number;
  pointLimit: number;
  omnibusJD: boolean;
}

/** A suggested hint for Hearts. */
export interface HeartsHint {
  cardIndices: number[];
  reason: string;
}

/** Full Hearts game state returned from the API. */
export interface HeartsResponse extends BaseGameResponse {
  players: HeartsPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: HeartsTrickCard[];
  heartsBroken: boolean;
  passDirection: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  config: HeartsConfig;
  hint?: HeartsHint;
}

// --- Gong Zhu (拱猪) ---
