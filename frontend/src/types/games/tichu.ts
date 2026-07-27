// Type declarations for tichu. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tichu player action record. */
export interface TichuAction {
  playerIdx: number;
  playedCards: Card[] | null;
  declType: number;
  isPass: boolean;
}

/** Tichu player data. */
export interface TichuPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  team: number;
  rank: number;
  declType: number;
  cardCount: number;
  cards: Card[];
}

/** Tichu config. */
export interface TichuConfig {
  cpuDifficulty: number;
}

/** Tichu API response. */
export interface TichuResponse extends BaseGameResponse {
  players: TichuPlayerData[];
  phase: string;
  currentTurn: number;
  tableCards: Card[];
  tableCombo: string;
  lastPlayIdx: number;
  startLeader: number;
  finishOrder: number[];
  scores: number[];
  isOneTwo: boolean;
  bombCount: number;
  gameEndFlag: boolean;
  config: TichuConfig;
  cpuActions: TichuAction[];
  humanAction: TichuAction | null;
}
