// Type declarations for badugi. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, BettingHumanProfileData, Card } from '../common';

/** Badugi seat snapshot returned by the /badugi/exec API. */
export interface BadugiPlayerData {
  id: number;
  isHuman: boolean;
  cards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  /** BadugiHand.Size (1..4) after showdown, 0 otherwise. */
  handSize: number;
  handName: string;
  /** Cards exchanged in the most recent draw. */
  drawCount: number;
  /** Cumulative draws across all three draw rounds. */
  totalDraws: number;
  playStyleName: string;
  /** Best-subset selection revealed at showdown. */
  bestCards?: Card[];
}

/** CPU betting action in Badugi. */
export interface BadugiCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
  drawIndex: number;
  roundLabel: string;
}

/** CPU draw result in Badugi. */
export interface BadugiCpuExchange {
  playerIdx: number;
  drawIndex: number;
  exchangeCount: number;
}

/** Badugi showdown result for a single player. */
export interface BadugiResult {
  playerIdx: number;
  handSize: number;
  handName: string;
  wonAmount: number;
}

/** Badugi side pot with eligible player seats. */
export interface BadugiSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Meta-AI statistics for Badugi CPU adaptation. */
export interface BadugiMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  bluffRate: number;
  foldRate: number;
  hesitationMean: number;
}

/** Badugi phase discriminator: 0 Init, 1 Deal, 2 Bet, 3 Draw, 4 Showdown, 5 End. */
export type BadugiPhaseId = 0 | 1 | 2 | 3 | 4 | 5;

/** Full Badugi game state returned by the API. */
export interface BadugiResponse extends BaseGameResponse {
  players: BadugiPlayerData[];
  pot: number;
  sidePots: BadugiSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: BadugiPhaseId;
  drawIndex: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  ante: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: BadugiResult[];
  cpuActions: BadugiCpuAction[];
  cpuExchanges: BadugiCpuExchange[];
  metaAI?: BadugiMetaAI;
  profile?: BettingHumanProfileData;
}
