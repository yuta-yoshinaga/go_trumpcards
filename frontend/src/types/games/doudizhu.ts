// Type declarations for doudizhu. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Dou Dizhu player action record. */
export interface DoudizhuAction {
  playerIdx: number;
  playedCards: Card[] | null;
  bidValue: number;
}

/** Dou Dizhu player data. */
export interface DoudizhuPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  isLandlord: boolean;
  cardCount: number;
  cards: Card[];
}

/** Dou Dizhu config. */
export interface DoudizhuConfig {
  cpuDifficulty: number;
}

/** Dou Dizhu API response. */
export interface DoudizhuResponse extends BaseGameResponse {
  players: DoudizhuPlayerData[];
  phase: string;
  currentTurn: number;
  tableCards: Card[];
  tableCombo: string;
  kittyCards: Card[];
  landlordIdx: number;
  baseBid: number;
  highestBid: number;
  bombCount: number;
  scores: number[];
  gameEndFlag: boolean;
  config: DoudizhuConfig;
  cpuActions: DoudizhuAction[];
  humanAction: DoudizhuAction | null;
}
