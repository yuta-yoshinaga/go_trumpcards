// Type declarations for followthequeen. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, BettingHumanProfileData, BettingMetaAI, Card } from '../common';

/** Player data in Follow the Queen. */
export interface FollowTheQueenPlayerData {
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

/** CPU betting action in Follow the Queen. */
export interface FollowTheQueenCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
}

/** Follow the Queen round result for a single player. */
export interface FollowTheQueenResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  kickers: string;
  bestHand: Card[];
  /** Chips won in this hand. */
  wonAmount: number;
  mucked: boolean;
}

/** Side pot in Follow the Queen with eligible players. */
export interface FollowTheQueenSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Full Follow the Queen game state returned from the API. */
export interface FollowTheQueenResponse extends BaseGameResponse {
  players: FollowTheQueenPlayerData[];
  communityCard: Card | null;
  /**
   * The rank that is currently wild for everyone, or 0 while no face-up Queen
   * has set one. Sent by the server rather than derived here: the page cannot
   * see the face-down cards the deal walked past.
   */
  wildRank: number;
  /**
   * The human's best hand rank right now, or -1 when they have fewer than five
   * cards or have folded. Sent by the server because the wild rule lives in the
   * domain: the frontend's generic evaluator does not know that Queens are
   * always wild, so computing it here reads a hand two ranks weaker than it is.
   */
  humanHandRank: number;
  pot: number;
  sidePots: FollowTheQueenSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: FollowTheQueenResult[];
  cpuActions: FollowTheQueenCpuAction[];
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
