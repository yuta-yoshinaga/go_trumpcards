// Type declarations for president. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** President player data. */
export interface PresidentPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
}

/** A play or pass action in President. */
export interface PresidentAction {
  playerIdx: number;
  playedCards: Card[] | null; // null = pass
}

/** Card exchange action in President. */
export interface PresidentExchangeAction {
  fromPlayerIdx: number;
  toPlayerIdx: number;
  cards: Card[];
}

/** President game rule configuration. */
export interface PresidentConfig {
  revolutionEnabled: boolean;
  cardExchangeEnabled: boolean;
  passFieldFlushEnabled: boolean;
  cpuDifficulty: number;
}

/** Full President game state returned from the API. */
export interface PresidentResponse extends BaseGameResponse {
  players: PresidentPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  lastPlayPlayerIdx: number;
  gameEndFlag: boolean;
  revolutionActive: boolean;
  config: PresidentConfig;
  exchangeActions: PresidentExchangeAction[];
  cpuActions: PresidentAction[];
  humanAction: PresidentAction | null;
}
