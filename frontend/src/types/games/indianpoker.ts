// Type declarations for indianpoker. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { ActionLogEntry, BaseGameResponse, Card } from '../common';

/** Bracket data for Indian Poker profile export. */
export interface IndianPokerProfileBracketData {
  aggressive: number;
  total: number;
}

/** Exported Indian Poker human profile data. */
export interface IndianPokerHumanProfileData {
  aggressiveByBracket: [IndianPokerProfileBracketData, IndianPokerProfileBracketData, IndianPokerProfileBracketData];
  foldToBetCount: number;
  foldToBetTotal: number;
  gamesPlayed: number;
  hesitationCount: number;
  hesitationMean: number;
  hesitationM2: number;
}

/** Meta-AI statistics for Indian Poker CPU adaptation. */
export interface IndianPokerMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  bluffRate: number;
  foldRate: number;
  hesitationMean: number;
}

/** Indian Poker player data with card, chips, and betting status. */
export interface IndianPokerPlayerOutput {
  id: number;
  isHuman: boolean;
  card: Card | null;
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  cardRank: number;
  playStyleName: string;
}

/** Indian Poker round result for a single player. */
export interface IndianPokerResultOutput {
  playerIdx: number;
  card: Card | null;
  cardRank: number;
  wonAmount: number;
}

/** CPU betting action in Indian Poker. */
export interface IndianPokerCpuActionOutput {
  playerIdx: number;
  action: number;
  amount: number;
}

/** Side pot in Indian Poker with eligible players. */
export interface IndianPokerSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Full Indian Poker game state returned from the API. */
export interface IndianPokerResponse extends BaseGameResponse {
  players: IndianPokerPlayerOutput[];
  pot: number;
  sidePots: IndianPokerSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: IndianPokerResultOutput[];
  cpuActions: IndianPokerCpuActionOutput[];
  handCount: number;
  ante: number;
  actionLog?: ActionLogEntry[];
  metaAI?: IndianPokerMetaAI;
  profile?: IndianPokerHumanProfileData;
}

// --- Euchre (ユーカー) ---
