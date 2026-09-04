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
  /**
   * Hi-Lo only. Current best 8-or-better low (5 cards) during play for human player.
   * Evaluated by the backend server; not evaluated in TS to avoid duplicating rules.
   */
  currentLowHand?: Card[];
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
  /** High and low winnings combined, so the chips on screen always add up. */
  wonAmount: number;
  mucked: boolean;
  /** Hi-Lo only. Whether this seat made a qualifying eight-or-better low. */
  lowQualifies?: boolean;
  /** Hi-Lo only. The five cards forming the low. */
  lowBestHand?: Card[];
  /**
   * Hi-Lo only. Chips won as the low half. The high half is
   * {@link SevenCardStudResult.wonAmount} minus this.
   */
  wonLow?: number;
  /** Chicago only. The highest spade this seat held **face-down**, if any. */
  spadeCard?: Card | null;
  /**
   * Chicago only. Chips won as the spade half. The high half is
   * {@link SevenCardStudResult.wonAmount} minus this.
   */
  wonSpade?: number;
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
  /**
   * Whether this session is the Hi-Lo (8 or Better) split. Sent so the page
   * renders the low breakdown without inferring the variant from the route.
   */
  isHiLo?: boolean;
  /**
   * Whether this session is the Chicago split, where half the pot goes to the
   * highest spade in the hole. Sent for the same reason as
   * {@link SevenCardStudResponse.isHiLo} — so the page does not infer the
   * variant from the route.
   */
  isChicago?: boolean;
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
