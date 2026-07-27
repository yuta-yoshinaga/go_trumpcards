// Type declarations for sevens. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Sevens player data with pass count and card info. */
export interface SevensPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  passesUsed: number;
  maxPasses: number;
  cards: Card[];
  lastPlayedJoker: boolean;
}

/** A play or pass action in Sevens. */
export interface SevensAction {
  playerIdx: number;
  playedCard: Card | null; // null = pass
  targetSuit: number;
  targetValue: number;
  forcedPass: boolean;
}

/** Sevens game rule configuration. */
export interface SevensConfig {
  tunnelEnabled: boolean;
  tunnelSkipWidth: number;
  jokerCount: number;
  cpuStrategy: number;
  maxPasses: number;
  noJokerFinish: boolean;
  jokerReclaimEnabled: boolean;
  endStopEnabled: boolean;
  jokerConsecutiveBanned: boolean;
}

/** Full Sevens game state returned from the API. */
export interface SevensResponse extends BaseGameResponse {
  players: SevensPlayerData[];
  currentTurn: number;
  tableMinVals: number[]; // index 0 unused; 1=SPADE, 2=CLOVER, 3=HEART, 4=DIAMOND
  tableMaxVals: number[];
  tablePlaced: number[]; // bitmask per suit; bit i = value i placed
  config: SevensConfig;
  gameEndFlag: boolean;
  cpuActions: SevensAction[];
  humanAction: SevensAction | null;
}
