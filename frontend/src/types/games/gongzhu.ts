// Type declarations for gongzhu. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Gong Zhu player data with scores and trick count. */
export interface GongZhuPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Point cards this player has captured so far (hearts, ♠Q pig, ♦J sheep, ♣10 doubler). Public info revealed as tricks are taken. */
  capturedPointCards: Card[];
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Gong Zhu trick. */
export interface GongZhuTrickCard {
  playerIdx: number;
  card: Card;
}

/** Which point cards have been exposed (stakes doubled). */
export interface GongZhuExposure {
  pig: boolean;
  sheep: boolean;
  ace: boolean;
  doubler: boolean;
}

/** Gong Zhu game configuration. */
export interface GongZhuConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Gong Zhu. */
export interface GongZhuHint {
  cardIndices: number[];
  reason: string;
}

/** How one player's round score was built up (from the domain's own scorer). */
export interface GongZhuScoreBreakdown {
  heartCount: number;
  heartsSum: number;
  allHearts: boolean;
  aceExposed: boolean;
  hasPig: boolean;
  pigExposed: boolean;
  hasSheep: boolean;
  sheepExposed: boolean;
  hasDoubler: boolean;
  doublerMultiplier: number;
  doublerStandalone: number;
  subtotal: number;
  total: number;
}

/** Full Gong Zhu game state returned from the API. */
export interface GongZhuResponse extends BaseGameResponse {
  /**
   * Per-player score breakdown, present only at round end.
   *
   * Produced by the same function that assigns the score, so the panel and the
   * number can never disagree (#5630).
   */
  scoreBreakdowns?: GongZhuScoreBreakdown[];
  players: GongZhuPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: GongZhuTrickCard[];
  heartsBroken: boolean;
  exposed: GongZhuExposure;
  exposableIndices: number[];
  /** いま出せる手札の位置（マストフォローの可視化。#4812）。 */
  playableIndices: number[];
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  config: GongZhuConfig;
  hint?: GongZhuHint;
}

// --- Tressette ---
