// Type declarations for zheng. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Zheng Shangyou player data. */
export interface ZhengPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
}

/** A play or pass action in Zheng Shangyou (null playedCards = pass). */
export interface ZhengAction {
  playerIdx: number;
  playedCards: Card[] | null;
}

/** Zheng Shangyou game rule configuration. */
export interface ZhengConfig {
  cpuDifficulty: number;
}

/** Input type alias for Zheng Shangyou configuration. */
export type ZhengConfigInput = ZhengConfig;

/** Full Zheng Shangyou game state returned from the API. */
export interface ZhengResponse extends BaseGameResponse {
  players: ZhengPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  tablePlayType: number;
  lastPlayPlayerIdx: number;
  gameEndFlag: boolean;
  cpuActions: ZhengAction[];
  humanAction: ZhengAction | null;
  config: ZhengConfig;
}
