// Type declarations for deucetoseven. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, BettingHumanProfileData, Card } from '../common';

/** 2-7 Triple Draw seat snapshot returned by the /deucetoseven/exec API. */
export interface DeuceToSevenPlayerData {
  id: number;
  isHuman: boolean;
  cards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  /** Poker category (0 High Card … 9 Royal Flush) after showdown, 0 otherwise. */
  handRank: number;
  handName: string;
  /** Cards exchanged in the most recent draw. */
  drawCount: number;
  /** Cumulative draws across all three draw rounds. */
  totalDraws: number;
  playStyleName: string;
}

/** CPU betting action in 2-7 Triple Draw. */
export interface DeuceToSevenCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
  drawIndex: number;
  roundLabel: string;
}

/** CPU draw result in 2-7 Triple Draw. */
export interface DeuceToSevenCpuExchange {
  playerIdx: number;
  drawIndex: number;
  exchangeCount: number;
}

/** 2-7 Triple Draw showdown result for a single player. */
export interface DeuceToSevenResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  wonAmount: number;
}

/** 2-7 Triple Draw side pot with eligible player seats. */
export interface DeuceToSevenSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Meta-AI statistics for 2-7 Triple Draw CPU adaptation. */
export interface DeuceToSevenMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  bluffRate: number;
  foldRate: number;
  hesitationMean: number;
}

/** 2-7 Triple Draw phase discriminator: 0 Init, 1 Deal, 2 Bet, 3 Draw, 4 Showdown, 5 End. */
export type DeuceToSevenPhaseId = 0 | 1 | 2 | 3 | 4 | 5;

/** Full 2-7 Triple Draw game state returned by the API. */
export interface DeuceToSevenResponse extends BaseGameResponse {
  players: DeuceToSevenPlayerData[];
  pot: number;
  sidePots: DeuceToSevenSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: DeuceToSevenPhaseId;
  drawIndex: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  ante: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: DeuceToSevenResult[];
  cpuActions: DeuceToSevenCpuAction[];
  cpuExchanges: DeuceToSevenCpuExchange[];
  metaAI?: DeuceToSevenMetaAI;
  profile?: BettingHumanProfileData;
}
