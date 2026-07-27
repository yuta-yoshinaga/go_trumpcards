// Type declarations for sevencardstud. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, BettingHumanProfileData, BettingMetaAI, Card } from '../common';

/** Player data in Seven Card Stud. */
export interface SevenCardStudPlayerData {
  id: number;
  isHuman: boolean;
  holeCards: Card[];
  doorCards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  handRank: number;
  handName: string;
  bestHand: Card[];
  playStyleName: string;
  totalHands: number;
  vpip: number;
  pfr: number;
  threeBet: number;
  af: string;
}

/** CPU betting action in Seven Card Stud. */
export interface SevenCardStudCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
}

/** Seven Card Stud round result for a single player. */
export interface SevenCardStudResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  kickers: string;
  bestHand: Card[];
  wonAmount: number;
  mucked: boolean;
}

/** Side pot in Seven Card Stud with eligible players. */
export interface SevenCardStudSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Full Seven Card Stud game state returned from the API. */
export interface SevenCardStudResponse extends BaseGameResponse {
  players: SevenCardStudPlayerData[];
  communityCard: Card | null;
  pot: number;
  sidePots: SevenCardStudSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: SevenCardStudResult[];
  cpuActions: SevenCardStudCpuAction[];
  handCount: number;
  ante: number;
  bringIn: number;
  smallBet: number;
  bigBet: number;
  tournamentMode: boolean;
  anteLevelHands: number;
  anteMultiplier: number;
  tableSize: number;
  bringInPlayerIdx: number;
  rebuyAvailable: boolean;
  addonAvailable: boolean;
  rebuyCounts: number[];
  addonUsed: boolean[];
  rebuyEnabled: boolean;
  addonEnabled: boolean;
  rebuyMaxCount: number;
  rebuyChips: number;
  addonChips: number;
  rebuyPeriodHands: number;
  addonAfterHand: number;
  rebuyPhaseType: number;
  muckAvailable: boolean;
  metaAI?: BettingMetaAI;
  profile?: BettingHumanProfileData;
}

// --- Five Card Stud ---
