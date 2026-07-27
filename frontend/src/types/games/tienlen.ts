// Type declarations for tienlen. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tien Len player data. */
export interface TienLenPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
}

/** A play or pass action in Tien Len. */
export interface TienLenAction {
  playerIdx: number;
  playedCards: Card[] | null;
}

/** Tien Len game rule configuration. */
export interface TienLenConfig {
  cpuDifficulty: number;
}

/** Input type alias for Tien Len configuration. */
export type TienLenConfigInput = TienLenConfig;

/** Full Tien Len game state returned from the API. */
export interface TienLenResponse extends BaseGameResponse {
  players: TienLenPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  tablePlayType: number;
  lastPlayPlayerIdx: number;
  gameEndFlag: boolean;
  cpuActions: TienLenAction[];
  humanAction: TienLenAction | null;
  config: TienLenConfig;
}
