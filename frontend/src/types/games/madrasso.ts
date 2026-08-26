// Type declarations for madrasso. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A Madrasso player's public/own state. */
export interface MadrassoPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  teamId: number;
}

/** A card played in a Madrasso trick. */
export interface MadrassoTrickCard {
  playerIdx: number;
  card: Card;
}

/** Madrasso game configuration. */
export interface MadrassoConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Madrasso. */
export interface MadrassoHint {
  cardIndices: number[];
  reason: string;
}

/** Full Madrasso game state returned from the API. */
export interface MadrassoResponse extends BaseGameResponse {
  players: MadrassoPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: MadrassoTrickCard[];
  lastTrick: MadrassoTrickCard[];
  lastTrickWinner: number;
  leadPlayerIdx: number;
  teamScores: number[];
  /** 現ラウンドで獲得した**整数点**。クローン元 (トレセッテ) の 1/3 点ではない。 */
  teamRoundPoints: number[];
  /** 配りで決まった切り札スート (1=♠ 2=♣ 3=♥ 4=♦、-1=未確定)。 */
  trumpSuit: number;
  playableIndices: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: MadrassoConfig;
  hint?: MadrassoHint;
}

// --- Sheepshead ---
