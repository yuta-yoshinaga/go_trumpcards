// Type declarations for holdem. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, BettingHumanProfileData, BettingMetaAI, Card } from '../common';

/** Texas Hold'em player data with stats and hand info. */
export interface HoldemPlayerData {
  id: number;
  isHuman: boolean;
  cards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  handRank: number;
  handName: string;
  bestHand: Card[];
  /** Best 5-card low hand (Omaha Hi-Lo only; populated at showdown when qualified). */
  lowBestHand?: Card[];
  /** True if the player has a qualifying low hand (Omaha Hi-Lo only). */
  lowQualifies?: boolean;
  playStyleName: string;
  totalHands: number;
  vpip: number;
  pfr: number;
  threeBet: number;
  af: string;
}

/** CPU betting action in Texas Hold'em. */
export interface HoldemCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
}

/** Hold'em round result for a single player.
 *
 * Hi-Lo (Omaha 8 or Better) split-pot games populate the optional Low* and
 * HiWonAmount/LowWonAmount fields; for non-Hi-Lo games they are absent. */
export interface HoldemResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  kickers: string;
  bestHand: Card[];
  wonAmount: number;
  mucked: boolean;
  lowBestHand?: Card[];
  lowKickers?: string;
  lowQualifies?: boolean;
  hiWonAmount?: number;
  lowWonAmount?: number;
}

/** Side pot in Hold'em with eligible players. */
export interface HoldemSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Full Texas Hold'em game state returned from the API. */
export interface HoldemResponse extends BaseGameResponse {
  players: HoldemPlayerData[];
  communityCards: Card[];
  pot: number;
  sidePots: HoldemSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: HoldemResult[];
  cpuActions: HoldemCpuAction[];
  handCount: number;
  smallBlind: number;
  bigBlind: number;
  tournamentMode: boolean;
  blindLevelHands: number;
  blindMultiplier: number;
  tableSize: number;
  rebuyPhaseType: number;
  rebuyChips: number;
  rebuyMaxCount: number;
  rebuyCounts: number[];
  addonChips: number;
  rebuyAvailable: boolean;
  addonAvailable: boolean;
  rebuyEnabled: boolean;
  addonEnabled: boolean;
  rebuyPeriodHands: number;
  addonAfterHand: number;
  addonUsed: boolean[];
  muckAvailable: boolean;
  /** True when the variant is Omaha Hi-Lo (8 or Better) — split-pot logic active. */
  isHiLo?: boolean;
  equity?: HoldemEquity;
  potOdds?: number;
  metaAI?: BettingMetaAI;
  profile?: BettingHumanProfileData;
}

/** Equity calculation result for Hold'em hand. */
export interface HoldemEquity {
  winProbability: number;
  handOdds: HoldemHandOdds[];
}

/** Probability of achieving a specific hand rank in Hold'em. */
export interface HoldemHandOdds {
  handRank: number;
  handName: string;
  probability: number;
}

// --- Pineapple Poker ---
