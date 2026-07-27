// Type declarations for bigtwo. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Big Two player data. */
export interface BigTwoPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
}

/** A play or pass action in Big Two. */
export interface BigTwoAction {
  playerIdx: number;
  playedCards: Card[] | null;
}

/** Big Two game rule configuration. */
export interface BigTwoConfig {
  cpuDifficulty: number;
}

/** Input type alias for Big Two configuration. */
export type BigTwoConfigInput = BigTwoConfig;

/** Full Big Two game state returned from the API. */
export interface BigTwoResponse extends BaseGameResponse {
  players: BigTwoPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  tablePlayType: number;
  lastPlayPlayerIdx: number;
  gameEndFlag: boolean;
  cpuActions: BigTwoAction[];
  humanAction: BigTwoAction | null;
  config: BigTwoConfig;
}
