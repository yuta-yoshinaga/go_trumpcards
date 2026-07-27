// Type declarations for cassino. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Cassino player data. */
export interface CassinoPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  capturedCount: number;
  sweepCount: number;
  totalScore: number;
}

/** A build on the Cassino table. */
export interface CassinoBuild {
  ownerIdx: number;
  value: number;
  groups: Card[][];
  isMulti: boolean;
}

/** A take, build, or trail action in Cassino. */
export interface CassinoAction {
  playerIdx: number;
  type: 'take' | 'build' | 'trail';
  playedCard: Card | null;
  capturedCards: Card[];
  buildValue: number;
  isSweep: boolean;
}

/** Cassino score detail (per round). */
export interface CassinoScoreDetail {
  cards: Record<number, number>;
  spades: Record<number, number>;
  aces: Record<number, number>;
  hasBigCasino: number;
  hasLittleCasino: number;
  sweeps: Record<number, number>;
  gained: Record<number, number>;
}

/** Cassino game rule configuration. */
export interface CassinoConfig {
  targetScore: number;
  multiBuildEnabled: boolean;
  sweepBonusEnabled: boolean;
  cpuDifficulty: number;
}

/** Full Cassino game state returned from the API. */
export interface CassinoResponse extends BaseGameResponse {
  players: CassinoPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  builds: CassinoBuild[];
  lastCaptureIdx: number;
  gameEndFlag: boolean;
  phase: string;
  config: CassinoConfig;
  cpuActions: CassinoAction[];
  humanAction: CassinoAction | null;
  remainingDeck: number;
  packsDealt: number;
  roundWinners: number[];
  lastRoundDetail: CassinoScoreDetail | null;
}
